package views

import (
	"encoding/json"
	"risk-ext/models"
	"strconv"
	"strings"
	"time"

	"risk-ext/config"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type UsersView struct {
	Views
}

var coll = "loginuser"

func (this *UsersView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"NOLOGIN": A{1}},
		"POST":   MA{"NOLOGIN": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *UsersView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	//openId := ctx.FormValue("openId")
	code := ctx.Params().Get("code")
	var userData = struct {
		Type int8   `json:"type"` //用户类型 0=manager 1=member 2=C端用户
		Data string `json:"data"` //用户内容json
	}{}
	if code != "" {
		reponse, err := new(models.Users).GetOpenIdFromWechat(code)
		if err != nil {
			data["code"] = 0
			data["error"] = err.Error()
			return
		}
		statuCode = 200
		userInfos, err := new(models.Users).GetUsersByOpenId(reponse.OpenId, true)
		if err != nil {
			if reponse.UnionId != "" {
				config.Redis.HSet("wechatuser", reponse.OpenId, reponse.UnionId)
			}
			data["code"] = -1
			data["openId"] = reponse.OpenId
			return
		}
		if userInfos.UserStatus == 0 {
			statuCode = 400
			data["code"] = 0
			data["error"] = "该用户已被禁用"
			return
		}
		oldToken := userInfos.UserToken
		userToken := bson.NewObjectId().Hex()
		userInfos.UserLogin = uint32(time.Now().Unix())
		userInfos.UserToken = userToken
		if userInfos.UserUnionId == "" {
			userInfos.UserUnionId = reponse.UnionId
		}
		err = userInfos.Update()
		if err == nil {
			config.Redis.HDel(coll, oldToken+"_1")
			userStr, _ := json.Marshal(userInfos)
			userData.Type = 2
			userData.Data = string(userStr)
			userInfos.Redis.Save(coll, userToken+"_1", userData)
		}
		data["code"] = 1
		data["userInfo"] = userInfos
		return
	}
	token := ctx.FormValue("token")
	userModel := new(models.Users)
	var userInfo models.Users
	err := userModel.Redis.Map(coll, token+"_1", &userData)
	if err != nil {
		statuCode = 401
		data["code"] = 0
		data["error"] = "登录失效"
		return
	}
	json.Unmarshal([]byte(userData.Data), &userInfo)
	usersInfo, err := new(models.Users).GetUsersByOpenId(userInfo.UserOpenId)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	statuCode = 200
	data["code"] = 1
	data["userInfo"] = usersInfo
	return
}

func (this *UsersView) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	openId := ctx.FormValue("openId")
	if openId == "" {
		data["code"] = 0
		data["error"] = "openId参数缺失"
		return
	}
	phone := ctx.FormValue("phone")
	isTrue := models.CheckPhone(phone)
	if !isTrue {
		data["code"] = 0
		data["error"] = "请输入正确的手机号"
		return
	}
	code := ctx.FormValue("code")
	if code == "" {
		data["code"] = 0
		data["error"] = "请输入验证码"
		return
	}
	user := new(models.Users)
	codeIsTrue := user.CheckCode(phone, code)
	if codeIsTrue {
		_, err := user.GetUsersByPhone(phone)
		if err == nil {
			data["code"] = 0
			data["error"] = "手机号已绑定"
			return
		}
		userName := ctx.FormValue("userName")
		userAvatar := ctx.FormValue("userAvatar")
		user.UserFname = userName
		user.UserAvatar = userAvatar
		user.UserOpenId = openId
		user.UserMobile = phone
		userToken := bson.NewObjectId().Hex()
		user.UserToken = userToken
		userUnionId := config.Redis.HGet("wechatuser", openId).Val()
		if userUnionId != "" {
			user.UserUnionId = userUnionId
		}
		userInfo, err := user.Insert()
		if err != nil {
			data["code"] = 0
			data["error"] = "绑定失败"
			return
		}
		var userData = struct {
			Type int8   `json:"type"` //用户类型 0=manager 1=member 2=C端用户
			Data string `json:"data"` //用户内容json
		}{}
		userStr, _ := json.Marshal(userInfo)
		userData.Type = 2
		userData.Data = string(userStr)
		user.Redis.Save(coll, userToken+"_1", userData)
		statuCode = 200
		data["code"] = 1
		data["userInfo"] = userInfo
		return
	} else {
		data["code"] = 0
		data["error"] = "验证码错误"
		return
	}
}

func (this *UsersView) Put(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.PostValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "参数deviceId缺失"
		return
	}
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	deviceData, err := new(models.Devices).GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["error"] = "设备不存在"
		return
	}
	if deviceData.DeviceUser.UserId != Session.Customer.UserId {
		data["code"] = 0
		data["error"] = "你无权限操作该设备"
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
		positionStr := ctx.PostValue("latlngs")
		if positionStr == "" {
			data["code"] = 0
			data["error"] = "参数latlng缺失"
			return
		}
		positionArr := strings.Split(positionStr, ",")
		longitude, _ := strconv.ParseFloat(positionArr[0], 64) //经度
		latitude, _ := strconv.ParseFloat(positionArr[1], 64)  //纬度
		latlng.Type = "Point"
		latlng.Coordinates = []float64{longitude, latitude}
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
		data["error"] = "开启设防失败"
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
	return
}
