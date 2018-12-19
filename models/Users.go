package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"risk-ext/app"
	"risk-ext/config"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Users struct {
	Model       `bson:"-" json:"-"` //model基类
	Redis       `bson:"-" json:"-"` //model基类
	UserId      bson.ObjectId       `bson:"_id,omitempty" json:"user_id"`     //id
	UserFname   string              `bson:"user_fname" json:"user_fname"`     //姓名
	UserUname   string              `bson:"user_uname" json:"user_uname"`     //登录名
	UserPasswd  string              `bson:"user_passwd" json:"user_passwd"`   //密码
	UserAvatar  string              `bson:"user_avatar" json:"user_avatar"`   //头像
	UserTravel  []Travel            `bson:"user_travel" json:"user_travel"`   //交通工具
	UserOpenId  string              `bson:"user_open_id" json:"user_open_id"` //微信openId
	UserMobile  string              `bson:"user_mobile" json:"user_mobile"`   //登录手机号码
	UserStatus  uint8               `bson:"user_status" json:"user_status"`   //用户状态0禁用 1启用 2未注册
	UserToken   string              `bson:"user_token" json:"user_token"`     //登录token
	UserLogin   uint32              `bson:"user_login" json:"user_login"`     //最后登录时间
	UserRead    uint32              `bson:"user_read" json:"user_read"`       //阅读报警的时间
	UserDeleted uint32              `bson:"user_deleted" json:"user_deleted"` //删除时间
	UserDate    uint32              `bson:"user_date" json:"user_date"`       //创建时间
}

const APPID = "wx1e72aeeba77e0307"
const APPSECRET = "70fed4b77c2a2b0f2a9bbaa8d36e5e1b"
const WECHATAPIURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

type WxResponse struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

type Travel struct {
	TravelName       string      `bson:"travel_name" json:"travel_name"`               //交通工具名称
	TravelType       uint8       `bson:"travel_type" json:"travel_type"`               //交通工具类型0=电动车 1=自行车 2=汽车
	TravelDevice     DevInfo     `bson:"travel_device" json:"travel_device"`           //绑定的设备信息
	TravelDeviceInfo *DeviceInfo `bson:"travel_device_info" json:"travel_device_info"` //绑定的设备实时数据
	TravelShare      string      `bson:"travel_share" json:"travel_share"`             //共享用户ID 为空则不是共享设备 共享设备只有查看权
	TravelDate       int64       `bson:"travel_date" json:"travel_date"`               //绑定时间
}

type DevInfo struct {
	DeviceId    uint64 `bson:"device_id" json:"device_id"`         //设备号
	DeviceIdStr string `bson:"device_id_str" json:"device_id_str"` //设备号
	DeviceSim   uint64 `bson:"device_sim" json:"device_sim"`       //设备sim卡号
	DeviceModel Models `bson:"device_model" json:"device_model"`   //设备类型
}

//type Models struct {
//	Model_name string `json:"model_name"` //型号名
//}

func (this *Users) GetUsersByOpenId(openId string) (rs Users, err error) {
	err = this.Collection(this).Find(bson.M{"user_open_id": openId, "user_deleted": 0}).One(&rs)
	if err == nil {
		for key, val := range rs.UserTravel {
			deviceInfo := new(Devices).GetDeviceInfo(val.TravelDevice.DeviceId)
			rs.UserTravel[key].TravelDeviceInfo = deviceInfo
		}
	}
	return
}

func (this *Users) GetUsersByPhone(phone string) (rs Users, err error) {
	err = this.Collection(this).Find(bson.M{"user_mobile": phone, "user_deleted": 0}).One(&rs)
	return
}

func (this *Users) Insert() (rs *Users, err error) {
	this.UserId = bson.NewObjectId()
	this.UserStatus = 1
	this.UserDate = uint32(time.Now().Unix())
	err = this.Collection(this).Insert(*this)
	rs = this
	return
}

//检测验证码
func (this *Users) CheckCode(phone string, code string) bool {
	value, err := this.Get(phone)
	if err != nil {
		return false
	}
	if value == "" {
		return false
	} else if code != value {
		return false
	}
	//	Redis.Del(phone)
	return true
}

func (this *Users) Update() (err error) {
	if this.UserId != EmptyId {
		update := bson.M{"$set": *this}
		err = this.Collection(this).UpdateId(this.UserId, update)
	}
	return
}

func (this *Users) GetOpenIdFromWechat(code string) (rep WxResponse, err error) {
	url := fmt.Sprintf(WECHATAPIURL, APPID, APPSECRET, code)
	err, jsonStr := app.HttpClient(url, "", "GET", "", "application/json", "")
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(jsonStr), &rep)
	if rep.Errcode == 0 {
		config.Redis.Set("wx_session_key", rep.SessionKey, 0)
	} else {
		err = errors.New(rep.Errmsg)
	}
	return
}
