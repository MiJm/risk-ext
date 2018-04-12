package views

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"risk-ext/app"
	"risk-ext/config"
	"risk-ext/models"
	"strconv"

	"github.com/kataras/iris"
)

const (
	MANAGER_ADMIN     = 0
	MANAGER_SERVICE   = 1
	MANAGER_STORE     = 2
	MANAGER_ASSISTANT = 3

	MEMBER_SUPER   = 2
	MEMBER_ADMIN   = 1
	MEMBER_GENERAL = 0

	HTTP_OK_200                  = 200
	HTTP_100_Continue            = 100
	HTTP_101_Switching_Protocols = 101
	HTTP_102_Processing          = 102
)

type (
	A   []uint8
	M   map[string]interface{}
	MA  map[string]A
	PMS map[string]MA
)

var Session *models.Session

type Views struct {
}

func (this *Views) Auth(ctx iris.Context) bool {
	token := ctx.GetHeader("token")
	if token == "" {
		token = ctx.FormValue("token")
	}
	Session = new(models.Session).Data(token)
	return true
}

func (this *Views) CheckPerms(perm MA) bool {
	if perm == nil {
		return false //方法为授权默认无权限访问
	}

	if perm["NOLOGIN"] != nil {
		return true //无需登录
	}
	if Session == nil {
		return false //session不存在无权限访问
	}
	if Session.Type == 1 { //当前登录用户为普通用户
		if perm["USER"] == nil {
			return false //普通用户无权限
		} else {
			for _, p := range perm["USER"] {
				if p == Session.User.UserLevel {
					return true //找到了相应的权限
				}
			}
			return false //未找到相应权限
		}
	} else { //后台管理员
		if perm["ADMIN"] == nil {
			return false //管理员无权限
		} else {
			for _, p := range perm["ADMIN"] {
				if p == Session.Manager.Manager_level {
					return true //找到了相应的权限
				}
			}
			return false //未找到相应权限
		}
	}

}

func (this *Views) GetMainData(path, params string, result interface{}) error {
	url := config.GetString("main_url") + path
	if Session.Type == 1 {
		if params == "" {
			params = "token=" + Session.User.UserToken
		} else {
			params += "&token=" + Session.User.UserToken
		}
	}
	return app.HttpClient(url, params, "POST", result)
}

func (this *Views) GetAnalysisData(path string, params interface{}, result interface{}, method ...string) error {
	m5 := md5.New()
	m5.Write([]byte(config.GetString("analysis_pwd")))
	loginParams := M{"username": config.GetString("analysis_name"), "password": hex.EncodeToString(m5.Sum(nil))}
	loginUrl := config.GetString("analysis_url") + "authorize"

	method_type := "POST"
	if len(method) != 0 {
		method_type = method[0]
	}

	loginData := struct {
		Code       int
		Expires_in int
		Msg        string
		Token      string
	}{}

	err := app.HttpClient(loginUrl, loginParams, "POST", &loginData)

	if err != nil || loginData.Code != 0 {
		return errors.New("认证失败")
	}
	url := config.GetString("analysis_url") + path
	err = app.HttpClient(url, params, method_type, result, loginData.Token)
	return err
}

//发送短信
func (this *Views) SendMsg(phone, msg string, result interface{}, method ...string) int64 {
	url := "http://www.jianzhou.sh.cn/JianzhouSMSWSServer/http/sendBatchMessage"
	method_type := "POST"
	if len(method) != 0 {
		method_type = method[0]
	}
	params := M{"account": "sdk_jiujin", "destmobile": phone, "msgText": msg + " 【风控一号】", "password": "joy1101gin"}
	err := app.HttpClient(url, params, method_type, result)
	if err != nil {
		return 0
	}
	code, err := strconv.ParseInt(result.(string), 10, 64)
	if err != nil {
		return 0
	}
	return code
}

//量讯平台登录
func (this *Views) SimLogin() (token string) {
	url := "http://120.26.213.169/api/access_token/"
	method_type := "POST"
	username := config.GetString("upiot_name")
	passwd := config.GetString("upiot_pwd")
	params := M{"username": username, "password": passwd}
	result := struct {
		Code  int
		Token string
	}{}
	err := app.HttpClient(url, params, method_type, result)
	if err == nil {
		token = result.Token
	}
	return
}

//量讯获取卡号信息
func (this *Views) SimInfo(simCard string, result interface{}, token ...string) {
	url := "http://120.26.213.169/api/card/" + simCard
	method_type := "GET"
	params := M{}
	err := app.HttpClient(url, params, method_type, token[0])
	if err == nil {

	}
}
