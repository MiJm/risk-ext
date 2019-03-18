package views

import (
	"encoding/json"
	"fmt"
	"risk-ext/models"
	"strconv"
	"time"

	"risk-ext/config"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type UsersViewForApp struct {
	Views
}

func (this *UsersViewForApp) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"NOLOGIN": A{1}},
		"POST":   MA{"NOLOGIN": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//小程序登录
func (this *UsersViewForApp) Get(ctx iris.Context) (statuCode int, data M) {
	// isApp := ctx.FormValueDefault("app", "")
	// if isApp != "" {
	// 	statuCode, data = this.Login(ctx)
	// 	return
	// }
	data = make(M)
	statuCode = 400
	//openId := ctx.FormValue("openId")
	code := ctx.FormValue("code") //微信code
	var userData = struct {
		Type int8   `json:"type"` //用户类型 0=manager 1=member 2=C端用户
		Data string `json:"data"` //用户内容json
	}{}
	if code != "" {
		reponse, err := new(models.Users).GetAccessToken(code)
		if err != nil {
			data["code"] = 0
			data["msg"] = err.Error()
			data["data"] = nil
			return
		}
		statuCode = 200
		userInfos, err := new(models.Users).GetUsers(reponse.OpenId, true)
		if err != nil {
			config.Redis.HSet("wechatuser", reponse.OpenId, reponse.AccessToken)
			//去执行新客注册操作
			data["code"] = -1
			data["msg"] = "OK"
			data["data"] = reponse.OpenId
			return
		}
		if userInfos.UserStatus == 0 {
			statuCode = 400
			data["code"] = 0
			data["msg"] = "该用户已被禁用"
			data["data"] = nil
			return
		}
		oldToken := userInfos.UserToken
		userToken := bson.NewObjectId().Hex()
		userInfos.UserLogin = uint32(time.Now().Unix())
		userInfos.UserToken = userToken
		if userInfos.UserUnionId == "" {
			if reponse.UnionId == "" {
				rep, err := new(models.Users).GetWxUserInfo(reponse.AccessToken, reponse.OpenId)
				if err == nil {
					userInfos.UserUnionId = rep.UnionId
				}
			} else {
				userInfos.UserUnionId = reponse.UnionId
			}
		}
		err = userInfos.Update()
		if err == nil {
			config.Redis.HDel(coll, oldToken+"_2")
			userStr, _ := json.Marshal(userInfos)
			userData.Type = 2
			userData.Data = string(userStr)
			userInfos.Redis.Save(coll, userToken+"_2", userData)
		}
		data["code"] = 1
		data["msg"] = "OK"
		data["data"] = userInfos
		return
	}
	token := ctx.FormValue("token")
	userModel := new(models.Users)
	var userInfo models.Users
	err := userModel.Redis.Map(coll, token+"_2", &userData)
	if err != nil {
		statuCode = 401
		data["code"] = 0
		data["msg"] = "登录失效"
		data["data"] = nil
		return
	}
	json.Unmarshal([]byte(userData.Data), &userInfo)
	usersInfo, err := new(models.Users).GetUsersByOpenId(userInfo.UserAppOpenId)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = usersInfo
	return
}

func (this *UsersViewForApp) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	OpenId := ctx.FormValue("openid")
	if OpenId == "" {
		data["code"] = 0
		data["msg"] = "参数缺失"
		data["data"] = nil
		return
	}
	AccessToken := config.Redis.HGet("wechatuser", OpenId).Val()
	phone := ctx.FormValue("phone")
	isTrue := models.CheckPhone(phone)
	if !isTrue {
		data["code"] = 0
		data["msg"] = "请输入正确的手机号"
		data["data"] = nil
		return
	}
	code := ctx.FormValue("code")
	if code == "" {
		data["code"] = 0
		data["msg"] = "请输入验证码"
		data["data"] = nil
		return
	}
	user := new(models.Users)
	codeIsTrue := user.CheckCode(phone, code)
	if codeIsTrue {
		_, err := user.GetUsersByPhone(phone)
		if err == nil {
			data["code"] = 0
			data["msg"] = "手机号已绑定"
			data["data"] = nil
			return
		}
		fmt.Println("acc", AccessToken)
		rep, err := new(models.Users).GetWxUserInfo(AccessToken, OpenId)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("微信获取授权失败(无法获取昵称及头像)%s", err.Error())
			data["data"] = nil
			return
		}

		user.UserFname = rep.Nickname
		user.UserAvatar = rep.Headimgurl
		user.UserAppOpenId = rep.OpenId
		user.UserMobile = phone
		userToken := bson.NewObjectId().Hex()
		user.UserToken = userToken
		user.UserUnionId = rep.UnionId
		userInfo, err := user.Insert()
		if err != nil {
			data["code"] = 0
			data["msg"] = "绑定失败"
			data["data"] = nil
			return
		}
		var userData = struct {
			Type int8   `json:"type"` //用户类型 0=manager 1=member 2=C端用户
			Data string `json:"data"` //用户内容json
		}{}
		userStr, _ := json.Marshal(userInfo)
		userData.Type = 2
		userData.Data = string(userStr)
		user.Redis.Save(coll, userToken+"_2", userData)
		statuCode = 200
		data["code"] = 1
		data["data"] = userInfo
		data["msg"] = "OK"
		return
	} else {
		data["code"] = 0
		data["msg"] = "验证码错误"
		data["data"] = nil
		return
	}
}

func (this *UsersViewForApp) Put(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.PostValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["msg"] = "参数deviceId缺失"
		data["data"] = nil
		return
	}
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	deviceData, err := new(models.Devices).GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["msg"] = "设备不存在"
		data["data"] = nil
		return
	}
	if deviceData.DeviceUser.UserId != Session.Customer.UserId {
		data["code"] = 0
		data["msg"] = "你无权限操作该设备"
		data["data"] = nil
		return
	}
	types := ctx.PostValueIntDefault("type", 0)
	userInfo, _ := new(models.Users).GetUsersByUserId(Session.Customer.UserId)
	var index = -1
	for key, travel := range userInfo.UserTravel {
		var id = travel.TravelDeviceId
		if travel.TravelDeviceId == 0 {
			id = travel.TravelDevice.DeviceId
		}
		if id == devId {
			index = key
			break
		}
	}
	userTravels := userInfo.UserTravel
	var latlng models.Latlng
	radius := 200
	if types == 0 { //开启设防
		deviceInfo := new(models.Devices).GetDeviceInfo(devId)
		latlng = deviceInfo.Device_latlng
		userTravels[index].TravelPen.PenType = 0
		userTravels[index].TravelPen.PenStatus = 1
		userTravels[index].TravelPen.PenRadius = uint32(radius)
		userTravels[index].TravelPen.PenPoint = latlng
		userTravels[index].TravelPen.PenDate = uint32(time.Now().Unix())
	} else { //关闭设防
		userTravels[index].TravelPen.PenStatus = 0
	}
	userInfo.UserTravel = userTravels
	err = userInfo.Update()
	if err != nil {
		data["code"] = 0
		data["msg"] = "开启设防失败"
		data["data"] = nil
		return
	} else {
		if types == 0 {
			var penInfo models.Pens
			penInfo.Pen_inout = 1
			penInfo.Pen_point = latlng
			penInfo.Pen_radius = uint32(radius)
			penInfo.Pen_type = 0
			penInfo.Pen_date = uint32(time.Now().Unix())
			penData, _ := json.Marshal(penInfo)
			config.Redis.HSet("pens_user_"+userInfo.UserId.Hex(), deviceId, string(penData))
		} else {
			config.Redis.HDel("pens_user_"+userInfo.UserId.Hex(), deviceId)
		}
	}
	statuCode = 200
	data["code"] = 1
	data["data"] = nil
	data["msg"] = "OK"
	return
}

//APP登录
func (this *UsersViewForApp) Login(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	code := ctx.FormValue("vcode")  //手机验证码
	phone := ctx.FormValue("phone") //手机号
	err, userInfo := new(models.Users).GetUserByPhone(phone, code)

	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}

	if userInfo.UserStatus == 0 {
		statuCode = 400
		data["code"] = 0
		data["msg"] = "该用户已被禁用"
		data["data"] = nil
		return
	}

	statuCode = 200
	data["code"] = 1
	data["data"] = userInfo
	data["msg"] = "OK"
	return
}
