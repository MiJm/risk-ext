package models

import (
	"errors"
	"fmt"
	"risk-ext/utils"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//平台成员结构体
type Members struct {
	Model                `bson:"-" json:"-"` //model基类
	Redis                `bson:"-" json:"-"` //redis基类
	Member_id            bson.ObjectId       `bson:"_id,omitempty" json:"member_id"`                 //id
	Member_fname         string              `json:"member_fname"`                                   //姓名
	Member_uname         string              `json:"member_uname"`                                   //登录名
	Member_passwd        string              `json:"member_passwd"`                                  //密码
	Member_open_id       string              `json:"member_open_id"`                                 //微信openId
	Member_company_id    string              `json:"member_company_id"`                              //客户ID
	Member_company_name  string              `json:"member_company_name"`                            //企业名
	Member_company_fname string              `json:"member_company_fname"`                           //企业名
	Member_company_level uint8               `json:"member_company_level"`                           //企业等级 0普通 1试用 2重要客户
	Member_group_id      string              `json:"member_group_id"`                                //组织ID
	Member_group_name    string              `json:"member_group_name"`                              //组织名
	Member_mobile        string              `json:"member_mobile"`                                  //登录手机号码
	Member_level         uint8               `json:"member_level"`                                   //用户等级0普通 1管理 2超级管理 3库管
	Member_status        uint8               `json:"member_status"`                                  //用户状态0禁用 1启用 2未注册
	Member_token         string              `json:"member_token"`                                   //登录token
	Member_login         uint32              `json:"member_login"`                                   //最后登录时间
	Member_read          uint32              `json:"member_read"`                                    //阅读报警的时间
	Member_deleted       uint32              `json:"member_deleted"`                                 //删除时间
	Member_date          uint32              `json:"member_date"`                                    //创建时间
	Member_is_erp        uint8               `json:"member_is_erp"`                                  //判断是否是erp管理员登录0不是，1是erp登录
	MemberCompanyLogo    string              `bson:"member_company_logo" json:"member_company_logo"` //企业LOGO
	MemberPages          []Pages             `bson:"member_pages" json:"member_pages"`               //当前用户可视页面
}

type Pages struct {
	PageId        bson.ObjectId `bson:"_id,omitempty" json:"page_id"` //id
	PageTitle     string        `json:"page_title"`                   //页面标题
	PagePath      string        `json:"page_path"`                    //页面前端路径（vue）
	PageApi       string        `json:"page_api"`                     //主接口地址
	PageSort      uint8         `json:"page_sort"`                    //排序
	PageCreatedAt int64         `json:"page_created_at"`              //创建时间
}

//获取账户下车辆数，设备数，即将过期设备数，报警数
func GetNums(loginMem users) (statistic bson.M, err error) {
	statistic = bson.M{}
	level := loginMem.UserLevel
	comId := loginMem.UserCompany_id
	groId := loginMem.UserGroupId
	carWhere := bson.M{}
	devWhere := bson.M{}
	exdevWhere := bson.M{}
	carWhere["car_deleted"] = 0

	if level == MEMBER_SUPER || level == MEMBER_STORE {
		carWhere["car_company_id"] = comId
		devWhere["device_company_id"] = bson.ObjectIdHex(comId)
		exdevWhere["device_company_id"] = bson.ObjectIdHex(comId)
	} else {
		var orcarwhere = make([]bson.M, 0)
		var ordevwhere = make([]bson.M, 0)
		gro, err := new(Groups).One(groId)
		if err != nil {
			err = errors.New("不存在该组织")
		}
		grosub := gro.Group_sub
		if len(grosub) > 0 {
			for _, v := range grosub {
				orcarwhere = append(orcarwhere, bson.M{"car_group_id": v.Group_id.Hex()})
				ordevwhere = append(ordevwhere, bson.M{"device_group_id": v.Group_id})
			}

			orcarwhere = append(orcarwhere, bson.M{"car_group_id": groId})
			ordevwhere = append(ordevwhere, bson.M{"device_group_id": bson.ObjectIdHex(groId)})
			carWhere["$or"] = orcarwhere
			devWhere["$or"] = ordevwhere
			exdevWhere["$or"] = ordevwhere
		} else {
			carWhere["car_group_id"] = groId
			devWhere["device_group_id"] = bson.ObjectIdHex(groId)
			exdevWhere["device_group_id"] = bson.ObjectIdHex(groId)
		}
	}

	devdate := bson.M{}
	now := utils.Time2Str1(uint32(time.Now().Unix()))
	end := utils.Str2Time(fmt.Sprintf("%s 23:59:59", now))

	devdate["$gte"] = now
	devdate["$lte"] = end
	exdevWhere["device_server_endtime"] = bson.M{"$lte": time.Now().Unix()}

	//开始计算监管车辆
	carWhere["car_devices"] = bson.M{"$elemMatch": bson.M{"$ne": nil}} //绑设备的车辆
	statistic["carNum"], _ = new(Cars).GetCarNumsByCondition(carWhere) //获取监管车辆数
	delete(carWhere, "car_devices")
	//计算完成监管车辆

	//>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	//计算在线设备数：onlineNum ；离线设备数:offlineNum；服务期已到期的设备数:expireDevNum
	devWhere["device_online"] = 1 //在线
	statistic["onlineNum"], _ = new(Devices).GetDevNumsByCondition(devWhere)
	devWhere["device_online"] = 0 //离线
	statistic["offlineNum"], _ = new(Devices).GetDevNumsByCondition(devWhere)
	delete(devWhere, "device_online")
	statistic["expireDevNum"], _ = new(Devices).GetDevNumsByCondition(exdevWhere) //设备服务期到期
	//<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	//>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	//计算正常金融状态的车辆：normalNum ;已结清金融状态的车辆:squareNum;已逾期金融状态的车辆:overdueNum
	carWhere["car_finance_status"] = 1 //正常
	statistic["normalNum"], _ = new(Cars).GetCarNumsByCondition(carWhere)
	carWhere["car_finance_status"] = 2 //已结清
	statistic["squareNum"], _ = new(Cars).GetCarNumsByCondition(carWhere)
	carWhere["car_finance_status"] = 3 //逾期
	statistic["overdueNum"], _ = new(Cars).GetCarNumsByCondition(carWhere)
	delete(devWhere, "car_finance_status")
	return
}

//获取预警数量
func GetAlarmNum(loginMem users) (statistic bson.M, err error) {
	statistic = bson.M{}
	level := loginMem.UserLevel
	comId := loginMem.UserCompany_id
	groId := loginMem.UserGroupId
	alaWhere := bson.M{}

	if level == MEMBER_SUPER || level == MEMBER_STORE {
		alaWhere["alarm_company_id"] = comId
	} else {
		var oralawhere = make([]bson.M, 0)
		gro, err := new(Groups).One(groId)
		if err != nil {
			err = errors.New("不存在该组织")
		}
		grosub := gro.Group_sub
		if len(grosub) > 0 {
			for _, v := range grosub {
				oralawhere = append(oralawhere, bson.M{"alarm_group_id": v.Group_id.Hex()})
			}

			oralawhere = append(oralawhere, bson.M{"alarm_group_id": groId})
			alaWhere["$or"] = oralawhere
		} else {
			alaWhere["alarm_group_id"] = groId
		}
	}

	aladate := bson.M{}
	now := utils.Time2Str1(uint32(time.Now().Unix()))
	start := utils.Str2Time(fmt.Sprintf("%s 00:00:00", now))
	aladate["$gte"] = start
	alaWhere["alarm_date"] = aladate

	statistic["alaNum"], _ = new(Alarms).GetAlarmNumsByCondition(alaWhere) //当日预警数

	return

}
