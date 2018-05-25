package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var coll = "loginuser"

type Session struct {
	Redis   `bson:"-" json:"-"` //model基类
	Type    int8                //用户类型 0=manager 1=member
	User    users               //前端用户
	Manager managers            //管理员
}

type users struct {
	Redis            `bson:"-" json:"-"` //model基类
	UserId           bson.ObjectId       `json:"member_id"`            //id
	UserFname        string              `json:"member_fname"`         //姓名
	UserUname        string              `json:"member_uname"`         //登录名
	UserPasswd       string              `json:"member_passwd"`        //密码
	UserCompany_id   string              `json:"member_company_id"`    //客户ID
	UserCompanyName  string              `json:"member_company_name"`  //企业名
	UserCompanyFname string              `json:"member_company_fname"` //企业名
	UserCompanyLevel uint8               `json:"member_company_level"` //企业等级 0普通 1试用 2重要客户
	UserGroupId      string              `json:"member_group_id"`      //组织ID
	UserGroupName    string              `json:"member_group_name"`    //组织名
	UserMobile       string              `json:"member_mobile"`        //登录手机号码
	UserLevel        uint8               `json:"member_level"`         //用户等级0普通 1管理 2超级管理
	UserStatus       uint8               `json:"member_status"`        //用户状态0禁用 1启用 2未注册
	UserToken        string              `json:"member_token"`         //登录token
	UserLogin        uint32              `json:"member_login"`         //最后登录时间
	UserRead         uint32              `json:"member_read"`          //阅读报警的时间
	UserDeleted      uint32              `json:"member_deleted"`       //删除时间
	UserDate         uint32              `json:"member_date"`          //创建时间
	Amount           Amounts             `json:"-"`                    //余量
}

type managers struct {
	Redis          `bson:"-" json:"-"` //model基类
	Manager_id     bson.ObjectId       `json:"manager_id"`
	Manager_fname  string              `json:"manager_fname"`  //姓名
	Manager_mobile string              `json:"manager_mobile"` //手机号码
	Manager_level  uint8               `json:"manager_level"`  //0管理员 1客服 2仓库 3销售助理
	Manager_passwd string              `json:"manager_passwd"` //登录密码
	Manager_token  string              `json:"manager_token"`  //登录token
	Manager_enable uint8               `json:"manager_enable"` //0禁用 1启用
	Manager_login  uint32              `json:"manager_login"`  //最后登录时间
	Manager_date   uint32              `json:"manager_date"`   //注册时间
}

/**
 * 余量redis表（目前是智能追车，以后会有征信 违章等）
 */
type Amounts struct {
	Redis            `bson:"-" json:"-"` //model基类
	CompanyId        string              `json:"company_id"`         //id
	QueryAiCar       int                 `json:"query_ai_car"`       //只能追车查询数量
	QueryCredit      int                 `json:"query_credit"`       //征信查询数量
	QueryDianHua     int                 `json:"query_dianhua"`      //电话帮查询数量
	QueryAiCarTime   int64               `json:"query_ai_car_time"`  //只能追车增加时间
	QueryCreditTime  int64               `json:"query_credit_time"`  //征信增加时间
	QueryDianHuaTime int64               `json:"query_dianhua_time"` //电话帮增加时间
}

//任务信息
type Task struct {
	ReportId  string //报表ID
	CompanyId string //企业ID
	Type      int8   //类型0=追车，1=电话 2=违章 3=征信
	Sid       string //Sid 会话ID
	CarNum    string //车牌号
	Path      string //分析数据文件路径
	Name      string //姓名
	IdCard    string //身份证号
}

//获取当前登录用户
func (this *Session) Data(token string) *Session {

	var user = struct {
		Type int8   `json:"type"` //用户类型 0=manager 1=member
		Data string `json:"data"` //用户内容json
	}{}

	key := fmt.Sprintf("%s_%d", token, 1)
	err := this.Map(coll, key, &user)
	if err != nil {
		key = fmt.Sprintf("%s_%d", token, 0)
		err := this.Map(coll, key, &user)
		if err != nil {
			return nil
		}
	}

	this.Type = user.Type
	if user.Type == 0 {
		json.Unmarshal([]byte(user.Data), &this.Manager)
	} else if user.Type == 1 {
		err = json.Unmarshal([]byte(user.Data), &this.User)
		if err == nil {
			this.Map("amounts", this.User.UserCompany_id, &this.User.Amount)
		}
	}
	return this
}

//修改只能追车数量 erp使用
func (this *Session) ChangeAmount(company_id string, aiCarAmount, changeType int) error {
	if this.Type == 0 {
		var amount = new(Amounts)
		this.Map("amounts", company_id, amount)
		if changeType == 0 { //智能追车
			amount.QueryAiCar += aiCarAmount
			amount.QueryAiCarTime = time.Now().Unix()
		} else if changeType == 1 { //电话
			amount.QueryDianHua += aiCarAmount
			amount.QueryDianHuaTime = time.Now().Unix()
		} else if changeType == 2 { //违章

		} else if changeType == 3 { //征信
			amount.QueryCredit += aiCarAmount
			amount.QueryCreditTime = time.Now().Unix()
		}

		amount.CompanyId = company_id
		if amount.QueryAiCar < 0 {
			return errors.New("减扣次数不能超过当前次数")
		}
		return amount.Save()
	} else {
		return errors.New("没权限")
	}
}

//保存
func (this *Amounts) Save() error {
	return this.Redis.Save("amounts", this.CompanyId, *this)
}

//根据企业ID获取追车数量 erp使用
func (this *Session) Amount(company_id string, amount *Amounts) error {
	if this.Type == 0 {
		return this.Map("amounts", company_id, amount)
	} else {
		return errors.New("没权限")
	}
}
