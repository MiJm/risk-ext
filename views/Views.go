package views

import (
	"encoding/json"
	"errors"
	"risk-ext/app"
	"risk-ext/config"
	"risk-ext/models"
	"strconv"
	"time"

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
	M   app.M
	MA  map[string]A
	PMS map[string]MA
)

var Session *models.Session

type Views struct {
}

func (this *Views) Auth(ctx iris.Context) int64 {
	token := ctx.GetHeader("token")
	if token == "" {
		token = ctx.FormValue("token")
	}
	Session = new(models.Session).Data(token)
	return 1
}

func (this *Views) CheckPerms(perm MA) int {
	if perm == nil {
		return 403 //方法为授权默认无权限访问
	}

	if perm["NOLOGIN"] != nil {
		return 1 //无需登录
	}
	if Session == nil {
		return 401 //session不存在无权限访问
	}
	if Session.Type == 1 { //当前登录用户为普通用户
		if perm["USER"] == nil {
			return 403 //普通用户无权限
		} else {
			for _, p := range perm["USER"] {
				if p == Session.User.UserLevel {
					return 1 //找到了相应的权限
				}
			}
			return 403 //未找到相应权限
		}
	} else if Session.Type == 2 { //C端用户
		if perm["CUSTOMER"] == nil {
			return 403
		}
		return 1
	} else { //后台管理员
		if perm["ADMIN"] == nil {
			return 403 //管理员无权限
		} else {
			for _, p := range perm["ADMIN"] {
				if p == Session.Manager.Manager_level {
					return 1 //找到了相应的权限
				}
			}
			return 403 //未找到相应权限
		}
	}

}

//func (this *Views) GetMainData(path, params string, result interface{}) error {
//	url := config.GetString("main_url") + path
//	if Session.Type == 1 {
//		if params == "" {
//			params = "token=" + Session.User.UserToken
//		} else {
//			params += "&token=" + Session.User.UserToken
//		}
//	}
//	return app.HttpClient(url, params, "POST", result)
//}

//func (this *Views) GetAnalysisData(path string, params interface{}, result interface{}, method ...string) error {
//	m5 := md5.New()
//	m5.Write([]byte(config.GetString("analysis_pwd")))
//	loginParams := M{"username": config.GetString("analysis_name"), "password": hex.EncodeToString(m5.Sum(nil))}
//	loginUrl := config.GetString("analysis_url") + "authorize"

//	method_type := "POST"
//	if len(method) != 0 {
//		method_type = method[0]
//	}

//	loginData := struct {
//		Code       int
//		Expires_in int
//		Msg        string
//		Token      string
//	}{}

//	err := app.HttpClient(loginUrl, loginParams, "POST", &loginData)

//	if err != nil || loginData.Code != 0 {
//		return errors.New("认证失败")
//	}
//	url := config.GetString("analysis_url") + path
//	err = app.HttpClient(url, params, method_type, result, loginData.Token)
//	return err
//}

//发送短信
func (this *Views) SendMsg(phone, msg string, method ...string) int64 {
	var result interface{}
	url := "http://www.jianzhou.sh.cn/JianzhouSMSWSServer/http/sendBatchMessage"
	method_type := "POST"
	if len(method) != 0 {
		method_type = method[0]
	}
	params := "account=sdk_jiujin&destmobile=" + phone + "&msgText=" + msg + " 【风控一号】&password=joy1101gin"
	contentType := "application/x-www-form-urlencoded"
	err, result := app.HttpClient(url, params, method_type, result, contentType)
	if err != nil {
		return 0
	}
	code, err := strconv.ParseInt(result.(string), 10, 64)
	if err != nil {
		return 0
	}
	return code
}

//获取验证码
func (this *Views) GetCode(phone string) string {
	value, err := config.Redis.Get(phone).Result()
	if err == nil && value != "" {
		return value
	}
	code := models.GetRandCode()
	config.Redis.Set(phone, code, time.Minute*30).Err()
	return code
}

//检测验证码
func (this *Views) CheckCode(phone string, code string) bool {
	value, err := config.Redis.Get(phone).Result()

	if err != nil {
		return false
	}

	if value == "" {
		return false
	} else if code != value {
		return false
	}
	config.Redis.Del(phone)
	return true
}

//量讯平台登录
func (this *Views) SimLogin() (err error, token string) {
	url := "http://120.26.213.169/api/access_token/"
	method_type := "POST"
	username := config.GetString("upiot_name")
	passwd := config.GetString("upiot_pwd")
	params := "username=" + username + "&password=" + passwd
	result := struct {
		Token string `json:"token"`
		Code  int    `json:"code"`
	}{}
	contentType := "application/x-www-form-urlencoded"
	err, jsonStr := app.HttpClient(url, params, method_type, result, contentType)
	if err == nil {
		err = json.Unmarshal([]byte(jsonStr), &result)
		if result.Code == 200 {
			token = result.Token
		}
	}
	if token == "" {
		err = errors.New("请求失败")
		return
	}
	err = config.Redis.Set("SimToken", token, time.Minute*120).Err()
	return
}

//量讯获取卡号信息
func (this *Views) SimInfo(simCard string) (err error, simResult interface{}) {
	simToken, err := config.Redis.Get("SimToken").Result()
	if err != nil {
		err, simToken = this.SimLogin()
		if err != nil && simToken == "" {
			return
		}
	}
	simResult = struct {
		Code                 int
		Msisdn               string
		Iccid                string
		Imsi                 string
		Carrier              string
		Sp_code              string
		Expiry_date          string
		Data_plan            int
		Data_usage           string
		Account_status       string
		Active               bool
		Test_valid_date      string
		Silent_valid_date    string
		Test_used_data_usage string
		Active_date          string
		Data_balance         int
		Outbound_date        string
		Support_sms          bool
	}{}
	url := "http://120.26.213.169/api/card/" + simCard + "/"
	method_type := "GET"
	params := ""
	simToken = "JWT " + simToken
	contentType := "application/json"
	err, jsonStr := app.HttpClient(url, params, method_type, simResult, contentType, simToken)
	if err == nil {
		err = json.Unmarshal([]byte(jsonStr), &simResult)
	}
	return
}

type bill_group struct {
	Code int `json:"code"` //状态码
	Data []struct {
		Carrier   string `json:"carrier"`   //运营商
		Bg_code   string `json:"bg_code"`   //计费组代码
		Name      string `json:"name"`      //套餐名称
		Data_plan int    `json:"data_plan"` //套餐大小
	} `json:"data"`
}

//量讯平台获取计费组列表
func (this *Views) GetBillGroup() (err error, bill bill_group) {
	simToken, err := config.Redis.Get("SimToken").Result()
	if err != nil {
		err, simToken = this.SimLogin()
		if err != nil && simToken == "" {
			return
		}
	}
	url := "http://120.26.213.169/api/billing_group/"
	method_type := "GET"
	params := ""
	simToken = "JWT " + simToken
	contentType := "application/json"
	err, jsonStr := app.HttpClient(url, params, method_type, bill, contentType, simToken)
	if err == nil {
		err = json.Unmarshal([]byte(jsonStr), &bill)
	}
	return
}

type result struct {
	Code      int
	Data      []models.SimInfo
	Per_page  int    //每页数量
	Num_pages int    //总页数
	Page      string //当前页数

}

//计费组物联卡列表
func (this *Views) SimList(bdCode string, page, size int) (err error, res result) {
	simToken, err := config.Redis.Get("SimToken").Result()
	if err != nil {
		err, simToken = this.SimLogin()
		if err != nil && simToken == "" {
			return
		}
	}
	url := "http://120.26.213.169/api/card/" + "?bg_code=" + bdCode + "&page=" + strconv.Itoa(page) + "&per_page=" + strconv.Itoa(size)
	method_type := "GET"
	params := ""
	simToken = "JWT " + simToken
	contentType := "application/json"
	err, jsonStr := app.HttpClient(url, params, method_type, res, contentType, simToken)
	if err == nil {
		err = json.Unmarshal([]byte(jsonStr), &res)
	}
	return
}
