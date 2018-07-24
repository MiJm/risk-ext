package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"risk-ext/app"
	"risk-ext/config"
	"risk-ext/utils"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type CommonData struct {
	Status uint   `json:"status"`
	Action string `json:"action"`
	Msg    string `json:"msg"`
}

type LoginType struct {
	CommonData
	Data LoginMethod
}

type LoginResult struct {
	CommonData
	Data LoginData
}

type Login2Result struct {
	CommonData
	Data LoginVerifyData
}

type TokenData struct {
	Token   string `json:"token"`
	Expires string `json:"expires"`
}

type LoginMethod struct {
	Sid          string `json:"sid"`            //会话标识
	NeedPinPwd   uint8  `json:"need_pin_pwd"`   //是否需要输入服务密码:1 是;0 否
	NeedFullName uint8  `json:"need_full_name"` //是否需要输入姓名:1 是;0 否
	NeedIdCard   uint8  `json:"need_id_card"`   //登入页面是否需用户身份证号 1:是, 0:否
	SmsDuration  int    `json:"sms_duration"`   //短信验证码提示有效时间
	CaptchaImage string `json:"captcha_image"`  //图形验证码 base64 码
	VerifyNotes  string `json:"verify_notes"`   // 短信提示(只针对于吉林电信并 sms_duration 不为空)
}

type LoginData struct {
	CaptchaImage string `json:"captcha_image"` //图形验证码 base64码
	SmsDuration  int    `json:"sms_duration"`  //短信验证码提示有效时间
	VerifyNotes  string `json:"verify_notes"`  // 短信提示(只针对于吉林电信并 sms_duration 不为空)
}

type LoginVerifyData struct {
	CaptchaImage string `json:"captcha_image"` //图形验证码 base64码
	SmsDuration  int    `json:"sms_duration"`  //短信验证码提示有效时间
	Content      string `json:"content"`       //说明
}

//获取token
func GetAccessToken() (token string, err error) {
	token = config.Redis.Get("dianhua_token").Val()
	if token != "" {
		return
	}
	appId := config.Get("appId").(string)
	appSecret := config.Get("appSecret").(string)
	nowTime := time.Now().Unix()
	str := appId + appSecret + fmt.Sprintf("%d", nowTime)
	sign := utils.String2Md5(str)
	data := struct {
		CommonData
		Data TokenData
	}{}
	url := fmt.Sprintf("https://crs-api.dianhua.cn/token?appid=%s&sign=%s&time=%s", appId, sign, fmt.Sprintf("%d", nowTime))
	err, jsonStr := app.HttpClient(url, "", "GET", nil, "application/json")
	if err != nil {
		return
	}
	json.Unmarshal([]byte(jsonStr), &data)
	if data.CommonData.Status != 0 {
		err = errors.New(data.CommonData.Msg)
		return
	}
	token = data.Data.Token
	err = config.Redis.Set("dianhua_token", token, 7200000000000).Err()
	return
}

//获取电话邦登录方式

func GetLoginMethod(tel string) (data LoginType, err error) {

	token, err := GetAccessToken()
	if err != nil {
		return
	}

	url := fmt.Sprintf("https://crs-api.dianhua.cn/calls/flow?token=%s", token)
	param := bson.M{"tel": tel}
	err, jsonStr := app.HttpClient(url, param, "POST", nil, "application/json")
	if err != nil {
		return
	}
	fmt.Println(jsonStr)
	json.Unmarshal([]byte(jsonStr), &data)

	return
}

//获取图形验证码
func GetCaptcha(sid string) (captcha_image string, err error) {
	data := struct {
		CommonData
		Data struct {
			CaptchaImage string
		} `json:"data"`
	}{}
	token, err := GetAccessToken()
	if err != nil {
		return
	}
	url := fmt.Sprintf("https://crs-api.dianhua.cn/calls/verify/captcha?token=%s&sid=%s", token, sid)
	err, jsonStr := app.HttpClient(url, "", "GET", nil, "application/json")
	if err != nil {
		return
	}
	json.Unmarshal([]byte(jsonStr), &data)
	if data.CommonData.Status != 0 {
		err = errors.New(data.CommonData.Msg)
		return
	}
	captcha_image = data.Data.CaptchaImage
	return
}

//登录电话邦
func Login(sid, tel, pin_pwd, full_name, id_card, sms_code, captcha_code string) (data LoginResult, err error) {
	token, err := GetAccessToken()
	if err != nil {
		return
	}

	url := fmt.Sprintf("https://crs-api.dianhua.cn/calls/login?token=%s", token)
	var params = bson.M{}
	params["sid"] = sid
	params["tel"] = tel
	if pin_pwd != "" {
		params["pin_pwd"] = pin_pwd
	}
	if full_name != "" {
		params["full_name"] = full_name
	}
	if id_card != "" {
		params["id_card"] = id_card
	}
	if sms_code != "" {
		params["sms_code"] = sms_code
	}
	if captcha_code != "" {
		params["captcha_code"] = captcha_code
	}
	err, jsonStr := app.HttpClient(url, params, "POST", nil, "application/json")
	if err != nil {
		return
	}
	json.Unmarshal([]byte(jsonStr), &data)
	return
}

//登录后二次验证
func LoginVerify(sid, sms_code, captcha_code string) (data Login2Result, err error) {
	token, err := GetAccessToken()
	if err != nil {
		return
	}

	url := fmt.Sprintf("https://crs-api.dianhua.cn/calls/verify?token=%s", token)
	var params = bson.M{}
	params["sid"] = sid
	if sms_code != "" {
		params["sms_code"] = sms_code
	}
	if captcha_code != "" {
		params["captcha_code"] = captcha_code
	}
	err, jsonStr := app.HttpClient(url, params, "POST", nil, "application/json")
	if err != nil {
		return
	}
	json.Unmarshal([]byte(jsonStr), &data)

	return
}
