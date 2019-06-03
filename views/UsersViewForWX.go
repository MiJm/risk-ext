package views

import (
	"encoding/json"
	"fmt"
	"risk-ext/config"
	"risk-ext/models"
	"risk-ext/utils"
	"time"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type UsersViewForWX struct {
	Views
}

func (this *UsersViewForWX) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"NOLOGIN": A{1}},
		"POST":   MA{"NOLOGIN": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//小程序登录
func (this *UsersViewForWX) Get(ctx iris.Context) (statuCode int, data M) {
	// isApp := ctx.FormValueDefault("app", "")
	// if isApp != "" {
	// 	statuCode, data = this.Login(ctx)
	// 	return
	// }
	data = make(M)
	statuCode = 400
	//openId := ctx.FormValue("openId")
	code := ctx.Params().Get("code") //微信code
	datas := ctx.FormValue("data")

	iv := ctx.FormValue("iv")
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
		if reponse.UnionId == "" {
			dataStr, err := utils.PswDecrypt(datas, reponse.SessionKey, iv)
			if err != nil {
				data["code"] = 0
				data["error"] = fmt.Sprintf("用户授权数据不合法(%s)", err.Error())
				return
			}
			if dataStr == "" {
				data["code"] = 0
				data["error"] = "获取用户授权信息失败"
				return
			}
			dataByte := []byte(dataStr)
			var wxUserInfo models.WXUserInfo
			err = json.Unmarshal(dataByte, &wxUserInfo)
			if err != nil {
				data["code"] = 0
				data["error"] = fmt.Sprintf("用户授权失败(%s)", err.Error())
				return
			}
			if wxUserInfo.UnionId == "" {
				data["code"] = 0
				data["error"] = "请重新授权登录"
				return
			}
			reponse.UnionId = wxUserInfo.UnionId
			reponse.Headimgurl = wxUserInfo.AvatarUrl
			reponse.Nickname = wxUserInfo.NickName
		}
		statuCode = 200
		userInfos, err := new(models.Users).GetUsersByUnionId(reponse.UnionId, true)
		if err != nil {
			userInfos, err = new(models.Users).GetUsersByOpenId(reponse.OpenId, true)
			if err != nil {
				if reponse.UnionId != "" {
					reponseByte, _ := json.Marshal(reponse)
					config.Redis.HSet("wechatuser", reponse.OpenId, reponseByte)
				}
				data["code"] = -1
				data["openId"] = reponse.OpenId
				return
			}
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
		userInfos.UserToken = userInfos.UserToken + "_1"
		data["code"] = 1
		data["userInfo"] = userInfos
		return
	}
	token := ctx.FormValue("token")
	userModel := new(models.Users)
	var userInfo models.Users
	err := userModel.Redis.Map(coll, token, &userData)
	if err != nil {
		statuCode = 401
		data["code"] = 0
		data["error"] = "登录失效"
		return
	}
	json.Unmarshal([]byte(userData.Data), &userInfo)
	usersInfo, err := new(models.Users).GetUsersByUnionId(userInfo.UserUnionId)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	usersInfo.UserToken = usersInfo.UserToken + "_1"
	statuCode = 200
	data["code"] = 1
	data["userInfo"] = usersInfo
	return
}

//添加操作待用
func (this *UsersViewForWX) Post(ctx iris.Context) (statuCode int, data M) {
	return
}

//更新操作待用
func (this *UsersViewForWX) Put(ctx iris.Context) (statuCode int, data M) {
	return
}

//删除操作待用
func (this *UsersViewForWX) Delete(ctx iris.Context) (statuCode int, data M) {
	return
}
