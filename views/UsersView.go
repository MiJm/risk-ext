package views

import (
	"risk-ext/models"

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
		statuCode = 403
		data["code"] = -1
		data["error"] = "登录失效"
		return
	}
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
