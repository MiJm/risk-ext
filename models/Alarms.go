package models

import (
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

type Alarms struct {
	Model              `bson:"-" json:"-"` //model基类
	Alarm_id           bson.ObjectId       `bson:"_id,omitempty" json:"alarm_id"`
	Alarm_car_id       bson.ObjectId       `json:"alarm_car_id"`                       //报警车辆ID
	Alarm_device_id    uint64              `json:"alarm_device_id"`                    //报警设备ID
	Alarm_device_no    string              `json:"alarm_device_no"`                    //报警设备no. 字符串形式的设备号
	Alarm_device_name  string              `json:"alarm_device_name"`                  //报警设备名
	Alarm_carowner     string              `json:"alarm_carowner"`                     //报警车主
	Alarm_device_type  uint8               `json:"alarm_device_type"`                  //报警设备类型0无线 1有线
	Alarm_car_plate    string              `json:"alarm_car_plate"`                    //报警车牌号
	Alarm_group_id     string              `json:"alarm_group_id"`                     //报警组织ID
	Alarm_group_name   string              `json:"alarm_group_name"`                   //报警组织名
	Alarm_company_id   string              `json:"alarm_company_id"`                   //报警企业ID
	Alarm_company_name string              `json:"alarm_company_name"`                 //报警企业名
	Alarm_pen_name     string              `json:"alarm_pen_name"`                     //报警围栏名
	Alarm_pen_id       string              `json:"alarm_pen_id"`                       //报警围栏ID
	Alarm_latlng       Latlng              `json:"alarm_latlng"`                       //报警坐标
	Alarm_address      string              `json:"alarm_address"`                      //报警地址
	Alarm_type         uint8               `json:"alarm_type"`                         //警报类型 0 断电报警（有线） 1见光报警（无线）  2围栏（出围） 3围栏（入围）  4关机报警 5开机报警 6低电报警 7通电报警 8见光恢复 9低电恢复 10解除出围 11解除入围 12风险点预警 13离线预警
	Alarm_loctype      uint8               `json:"alarm_loctype"`                      //报警定位类型 0gps 1基站 2WiFi
	Alarm_read         uint8               `json:"alarm_read"`                         //0未读；1已读
	AlarmUserId        string              `bson:"alarm_user_id" json:"alarm_user_id"` //报警用户ID
	Alarm_date         uint32              `json:"alarm_date"`                         //报警时间
	Alarm_created      uint32              `json:"alarm_created"`                      //创建时间
	Alarm_desc         string              `json:"alarm_desc"`                         //报警描述()
}

func (this *Alarms) GetAlarmListByUserId(userId, deviceId string, page, pageSize int) (rs []Alarms, count int, err error) {
	if page < 1 {
		page = 1
	}
	var offset = (page - 1) * pageSize
	var where = bson.M{}
	where["alarm_user_id"] = userId
	if deviceId != "" {
		devId, err := strconv.ParseUint(deviceId, 10, 64)
		if err == nil {
			where["alarm_device_id"] = devId
		}
	}
	count, _ = this.Collection(this).Find(where).Count()
	err = this.Collection(this).Find(where).Sort("-alarm_date").Limit(pageSize).Skip(offset).All(&rs)
	if rs == nil {
		rs = make([]Alarms, 0)
	}
	return
}

func (this *Alarms) GetAlarmInfo(alarmId string) (rs Alarms, err error) {
	err = this.Collection(this).FindId(bson.ObjectIdHex(alarmId)).One(&rs)
	return
}

func (this *Alarms) Update() (err error) {
	if this.Alarm_id != EmptyId {
		update := bson.M{"$set": *this}
		err = this.Collection(this).UpdateId(this.Alarm_id, update)
	}
	return
}

//获取未读预警数量
func (this *Alarms) GetUnReadAlarmNums(deviceId, userId string) (num int, err error) {
	var where = bson.M{}
	where["alarm_read"] = 0
	where["alarm_user_id"] = userId
	if deviceId != "" {
		devId, err := strconv.ParseUint(deviceId, 10, 64)
		if err == nil {
			where["alarm_device_id"] = devId
		}
	}
	num, err = this.Collection(this).Find(where).Count()
	return
}
