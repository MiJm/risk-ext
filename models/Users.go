package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"risk-ext/app"
	"risk-ext/config"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Users struct {
	Model         `bson:"-" json:"-"` //model基类
	Redis         `bson:"-" json:"-"` //model基类
	UserId        bson.ObjectId       `bson:"_id,omitempty" json:"user_id"`           //id
	UserFname     string              `bson:"user_fname" json:"user_fname"`           //姓名
	UserUname     string              `bson:"user_uname" json:"user_uname"`           //登录名
	UserPasswd    string              `bson:"user_passwd" json:"user_passwd"`         //密码
	UserAvatar    string              `bson:"user_avatar" json:"user_avatar"`         //头像
	UserTravel    []Travel            `bson:"user_travel" json:"user_travel"`         //交通工具
	UserOpenId    string              `bson:"user_open_id" json:"user_open_id"`       //微信openId
	UserWxOpenId  string              `bson:"user_wxopen_id" json:"user_wxopen_id"`   //微信公众号openId
	UserAppOpenId string              `bson:"user_appopen_id" json:"user_appopen_id"` //微信App登录openId
	UserUnionId   string              `bson:"user_union_id" json:"user_union_id"`     //微信唯一ID
	UserMobile    string              `bson:"user_mobile" json:"user_mobile"`         //登录手机号码
	UserStatus    uint8               `bson:"user_status" json:"user_status"`         //用户状态0禁用 1启用 2未注册
	UserToken     string              `bson:"user_token" json:"user_token"`           //登录token
	UserLogin     uint32              `bson:"user_login" json:"user_login"`           //最后登录时间
	UserRead      uint32              `bson:"user_read" json:"user_read"`             //阅读报警的时间
	UserDeleted   uint32              `bson:"user_deleted" json:"user_deleted"`       //删除时间
	UserDate      uint32              `bson:"user_date" json:"user_date"`             //创建时间
}

const APPID = "wx1e72aeeba77e0307"
const APPSECRET = "70fed4b77c2a2b0f2a9bbaa8d36e5e1b"
const WECHATAPIURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

//手机端APP微信授权登录相关信息
const APPID2 = "wxaf20cb020743a753"
const APPSECRET2 = "5b4b4fd781dc78df66d9a2b389bac3e3"
const WECHATAPIURL2 = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
const GETWXUSERINFO = "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s"

type WxResponse struct {
	OpenId      string `json:"openid"`
	SessionKey  string `json:"session_key"`
	AccessToken string `json:"access_token"`
	UnionId     string `json:"unionid"`
	Headimgurl  string `json:"headimgurl"`
	Nickname    string `json:"nickname"`
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
}

type WeChat struct {
	AccessToken string `json:"access_token"`
	Openid      string `json:"openid"`
	UnionId     string `json:"unionid"`
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
}
type WXUserInfo struct {
	OpenId    string `json:"openId"`
	NickName  string `json:"nickName"`
	avatarUrl string `json:"acatarUrl"`
	UnionId   string `json:"unionId"`
}
type Travel struct {
	TravelName     string `bson:"travel_name" json:"travel_name"`           //交通工具名称
	TravelType     uint8  `bson:"travel_type" json:"travel_type"`           //交通工具类型0=电动车 1=自行车 2=汽车
	TravelDeviceId uint64 `bson:"travel_device_id" json:"travel_device_id"` //绑定的设备id
	TravelDevice   struct {
		DeviceId uint64 `bson:"device_id" json:"device_id"` //绑定的设备id
	} `bson:"travel_device" json:"travel_device"` //绑定的设备信息
	TravelPen struct {
		PenPoint  Latlng `bson:"pen_point" json:"pen_point"`   //中心坐标
		PenRadius uint32 `bson:"pen_radius" json:"pen_radius"` //半径 米
		PenType   uint8  `bson:"pen_type" json:"pen_type"`     //围栏类型0=圆 1=多边形 2区域
		PenStatus uint8  `bson:"pen_status" json:"pen_status"` //围栏状态 0已关闭 1已开启
		PenDate   uint32 `bson:"pen_date" json:"pen_date"`     //创建时间
	} `bson:"travel_pen" json:"travel_pen"` //车辆围栏信息
	TravelDeviceInfo *DeviceInfo `bson:"travel_device_info" json:"travel_device_info"` //绑定的设备实时数据
	TravelShare      string      `bson:"travel_share" json:"travel_share"`             //共享用户ID 为空则不是共享设备 共享设备只有查看权
	TravelDate       int64       `bson:"travel_date" json:"travel_date"`               //绑定时间
	TravleAlarmNum   int         `bson:"travel_alarm_num" json:"travel_alarm_num"`     //未读事件数量
}

type Pens struct {
	Pen_id         bson.ObjectId `bson:"_id" json:"pen_id"`
	Pen_name       string        `json:"pen_name"`       //围栏名称
	Pen_type       uint8         `json:"pen_type"`       //围栏类型0=圆 1=多边形 2区域
	Pen_inout      uint8         `json:"pen_inout"`      //围栏出入0=入围 1=出围
	Pen_company    string        `json:"pen_company"`    //企业客户ID
	Pen_group      string        `json:"pen_group"`      //组织ID
	Pen_area       string        `json:"pen_area"`       //区域围栏省/市
	Pen_child_area string        `json:"pen_child_area"` //区域围栏区/县
	Pen_citycode   string        `json:"pen_citycode"`   //区域围栏 城市 编号 例如：010（北京）
	Pen_area_type  uint8         `json:"pen_area_type"`  //区域围栏类型0=市 1=省 2=区县
	Pen_polygon    Polygon       `json:"pen_polygon"`    //多边形围栏
	Pen_point      Latlng        `json:"pen_point"`      //中心坐标
	Pen_radius     uint32        `json:"pen_radius"`     //半径 米
	Pen_date       uint32        `json:"pen_date"`       //创建时间
	Pen_alarm      bool          `json:"pen_alarm"`      //是否有警报
	Pen_alarm_time uint32        `json:"pen_alarm_time"` //警报时间
}

type Polygon struct {
	Type        string        `json:"type"`        //Polygon
	Coordinates [][][]float64 `json:"coordinates"` //lng lat[[[ 89.8496, 14.093 ], [ 90.3933, 14.004 ]]]
}

func (this *Users) GetUsersByUnionId(unionId string, flag ...bool) (rs Users, err error) {
	err = this.Collection(this).Find(bson.M{"user_union_id": unionId, "user_deleted": 0}).One(&rs)
	if err == nil {
		if len(flag) > 0 && flag[0] {
			return
		}
		deviceModel := new(Devices)
		alarmModel := new(Alarms)
		for key, val := range rs.UserTravel {
			var devId uint64
			if val.TravelDeviceId != 0 {
				devId = val.TravelDeviceId
			} else {
				devId = val.TravelDevice.DeviceId
			}
			deviceInfo := deviceModel.GetDeviceInfo(devId)
			rs.UserTravel[key].TravelDeviceInfo = deviceInfo
			unReadAlarmNum, _ := alarmModel.GetUnReadAlarmNums(strconv.FormatUint(devId, 10), rs.UserId.Hex())
			rs.UserTravel[key].TravleAlarmNum = unReadAlarmNum
		}
	}
	return
}

func (this *Users) GetUsersByUserId(userId bson.ObjectId) (rs Users, err error) {
	err = this.Collection(this).Find(bson.M{"_id": userId}).One(&rs)
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
	this.UserLogin = uint32(time.Now().Unix())
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

//校验手机号是否存在
func (this *Users) GetUserByPhone(phone, code string) (err error, userInfo Users) {
	flag := this.CheckCode(phone, code)
	if !flag {
		err = errors.New("验证码错误")
		return
	}
	err = this.Collection(this).Find(bson.M{"user_mobile": phone}).One(&userInfo)
	if err != nil || userInfo.UserMobile == "" { //查询不到该手机号做新增操作
		userInfo.UserMobile = phone
		userInfo.UserId = bson.NewObjectId()
		userInfo.UserStatus = 1
		userInfo.UserDate = uint32(time.Now().Unix())
		userToken := bson.NewObjectId().Hex()
		userInfo.UserToken = userToken
		_, err = userInfo.Insert()
		if err != nil {
			err = errors.New("登录失败(新增账户失败)")
			return
		}
	} else { //存在该用户信息处理
		oldToken := userInfo.UserToken
		if oldToken != "" {
			config.Redis.HDel("loginuser", oldToken+"_1")
		}
		userInfo.UserToken = bson.NewObjectId().Hex()
		err = userInfo.Update()
		if err != nil {
			err = errors.New("登录失败(后台程序)")
		}
	}
	var userData = struct {
		Type int8   `json:"type"` //用户类型 0=manager 1=member 2=C端用户
		Data string `json:"data"` //用户内容json
	}{}
	userStr, _ := json.Marshal(userInfo)
	userData.Type = 2
	userData.Data = string(userStr)
	this.Redis.Save("loginuser", userInfo.UserToken+"_1", userData)
	// err = this.Redis.Delete(phone)
	return
}

//根据用户ID获取用户下车辆列表
func (this *Users) TravelList(userId string) (Travels []Travel, err error) {
	if !bson.IsObjectIdHex(userId) {
		err = errors.New("请求参数有误")
		return
	}
	var userInfo Users
	err = this.Collection(this).FindId(bson.ObjectIdHex(userId)).Select(bson.M{"user_travel": 1, "_id": 0}).One(&userInfo)

	Travels = userInfo.UserTravel
	for k, v := range Travels {
		var devId uint64
		if v.TravelDeviceId != 0 {
			devId = v.TravelDeviceId
		} else {
			devId = v.TravelDevice.DeviceId
		}
		deviceInfo := new(Devices).GetDeviceInfo(devId)
		Travels[k].TravelDeviceInfo = deviceInfo
		unReadAlarmNum, _ := new(Alarms).GetUnReadAlarmNums(strconv.FormatUint(devId, 10), userId)
		Travels[k].TravleAlarmNum = unReadAlarmNum
	}

	return
}

//APP获取AccessToken方法
func (this *Users) GetAccessToken(code string) (rep WxResponse, err error) {
	url := fmt.Sprintf(WECHATAPIURL2, APPID2, APPSECRET2, code)
	err, jsonStr := app.HttpClient(url, "", "GET", "", "application/json", "")
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(jsonStr), &rep)
	if rep.Errcode == 0 {
		config.Redis.Set("wx_app_token", rep.AccessToken, 7200)
	} else {
		err = errors.New(rep.Errmsg)
	}
	return
}

//获取微信用户信息
func (this *Users) GetWxUserInfo(access_token, openid string) (rep WxResponse, err error) {
	url := fmt.Sprintf(GETWXUSERINFO, access_token, openid)
	err, jsonStr := app.HttpClient(url, "", "GET", "", "application/json", "")
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(jsonStr), &rep)
	if rep.Errcode != 0 {
		err = errors.New(rep.Errmsg)
	}
	return
}

//app微信授权登录时通过openID查询数据库是否存在
func (this *Users) GetUsers(unionId string, flag ...bool) (rs Users, err error) {
	err = this.Collection(this).Find(bson.M{"user_union_id": unionId, "user_deleted": 0}).One(&rs)
	return
}
