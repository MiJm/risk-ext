package views

import (
	"risk-ext/models"
	"strconv"
	"time"

	"risk-ext/config"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type UsersView struct {
	Views
}

func (this *UsersView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"NOLOGIN": A{1}},
		"GET":    MA{"NOLOGIN": A{1}},
		"POST":   MA{"NOLOGIN": A{1}},
		"DELETE": MA{"NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *UsersView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	//openId := ctx.FormValue("openId")
	token := ctx.FormValue("token")
	code := ctx.Params().Get("code")
	if code != "" {
		reponse, err := new(models.Users).GetOpenIdFromWechat(code)
		if err != nil {
			data["code"] = 0
			data["error"] = err.Error()
			return
		}
		statuCode = 200
		userInfos, err := new(models.Users).GetUsersByOpenId(reponse.OpenId)
		if err != nil {
			data["code"] = -1
			data["openId"] = reponse.OpenId
			return
		}
		oldToken := userInfos.UserToken
		userToken := bson.NewObjectId().Hex()
		userInfos.UserToken = userToken
		err = userInfos.Update()
		if err == nil {
			config.Redis.HDel("logincustomer", oldToken)
			userInfos.Redis.Save("logincustomer", userToken, userInfos)
		}
		data["code"] = 1
		data["userInfo"] = userInfos
		return
	}
	if token == "" {
		data["code"] = 0
		data["error"] = "token参数缺失"
		return
	}
	userModel := new(models.Users)
	var userInfo models.Users
	err := userModel.Redis.Map("logincustomer", token, &userInfo)
	if err != nil {
		data["code"] = -1
		data["error"] = "登录失效"
		return
	}

	//userInfo, err := new(models.Users).GetUsersByOpenId(openId)

	//	statuCode = 200
	//	if err != nil {
	//		data["code"] = -1
	//		return
	//	}
	data["code"] = 1
	data["userInfo"] = userInfo
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
		userInfo, err := user.Insert()
		if err != nil {
			data["code"] = 0
			data["error"] = "绑定失败"
			return
		}
		user.Redis.Save("logincustomer", userToken, userInfo)
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
	token := ctx.FormValue("token")
	if token == "" {
		statuCode = 403
		data["code"] = 0
		data["error"] = "token参数缺失"
		return
	}
	userModel := new(models.Users)
	var userData models.Users
	err := userModel.Redis.Map("logincustomer", token, &userData)
	if err != nil {
		statuCode = 403
		data["code"] = -1
		data["error"] = "登录失效"
		return
	}
	deviceId := ctx.FormValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "deviceId参数缺失"
		return
	}
	//	openId := ctx.FormValue("openId")
	//	if openId == "" {
	//		data["code"] = 0
	//		data["error"] = "openId参数缺失"
	//		return
	//	}
	travelName := ctx.FormValue("travelName")
	if travelName == "" {
		data["code"] = 0
		data["error"] = "请输入名称"
		return
	}
	travelType, _ := ctx.PostValueInt("travelType")
	userInfo, err := userModel.GetUsersByOpenId(userData.UserOpenId)
	if err != nil {
		data["code"] = 0
		data["error"] = "用户已被注销"
		return
	}
	var travelInfo models.Travel
	var userTravel []models.Travel
	travelInfo.TravelName = travelName
	travelInfo.TravelType = uint8(travelType)
	travelInfo.TravelDate = time.Now().Unix()
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	device := new(models.Devices)
	deviceData, err := device.GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["error"] = "设备不存在"
		return
	}
	if deviceData.DeviceUser != nil {
		if deviceData.DeviceUser.UserId != models.EmptyId {
			data["code"] = 0
			data["error"] = "设备已激活"
			return
		}
	}
	if deviceData.DeviceOutType != 2 {
		data["code"] = 0
		data["error"] = "设备未出库"
		return
	}
	travelInfo.TravelDevice = deviceData
	userTravel = append(userInfo.UserTravel, travelInfo)
	userInfo.UserTravel = userTravel
	err = userInfo.Update()
	if err != nil {
		data["code"] = 0
		data["error"] = "激活失败"
		return
	}
	device.Device_id = deviceData.DeviceId
	device.DeviceUser = &userInfo
	err = device.Update(false)
	if err != nil {
		data["code"] = 0
		data["error"] = "激活失败"
		return
	}
	var deviceInfo models.DeviceInfo
	err = userModel.Map("devices", deviceId, &deviceInfo)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	deviceInfo.Device_activity_time = uint32(travelInfo.TravelDate)
	userModel.Save("devices", deviceId, deviceInfo)
	statuCode = 200
	data["code"] = 1
	return
}
