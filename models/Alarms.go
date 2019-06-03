package models

import (
	"fmt"
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
	Alarm_type         uint8               `json:"alarm_type"`                         //报警类型 0 断电报警（有线） 1见光报警（无线）  2围栏（出围） 3围栏（入围）  4关机报警 5开机报警 6低电报警 7通电报警 8见光恢复 9低电恢复 10解除出围 11解除入围,12风险点预警,13离线报警，14异动报警(离线后上线位置与上次上报差距大)，15震动报警，16拆卸报警，17acc关报警，18acc开报警 19停车超时
	Alarm_loctype      uint8               `json:"alarm_loctype"`                      //报警定位类型 0gps 1基站 2WiFi
	Alarm_read         uint8               `json:"alarm_read"`                         //0未读；1已读
	AlarmUserId        string              `bson:"alarm_user_id" json:"alarm_user_id"` //报警用户ID
	Alarm_date         uint32              `json:"alarm_date"`                         //报警时间
	Alarm_created      uint32              `json:"alarm_created"`                      //创建时间
	Alarm_desc         string              `json:"alarm_desc"`                         //报警描述()
}

var Alarm_type_string = []string{"断电/关机预警", "通电/开机预警", "光感预警", "出围预警", "入围预警", "低电量预警", "风险点预警", "离线预警", "异动报警", "停车超时", "超速预警", "常驻点预警", "设防预警"}
var Alarm_type_color = []string{"#F0644E", "#3ED080", "#F8AD2E", "#886EFF", "#886EFF", "#F78C75", " #F66A00", "#F53E35", "#3A9CFF", "#32C5FF", "#F53E35", "#F66A00", "#F8AD2E"}

var Alarm_descrip = []bson.M{{"type": "断电", "color": "#F0644E", "discription": "电源被切断", "args": 0, "isShow": true}, {"type": "见光", "color": "#F8AD2E", "discription": "检测到光感报警", "args": 0, "isShow": true}, {"type": "出围", "color": "#886EFF", "discription": "驶出围栏[%s]", "args": 1, "isShow": true}, {"type": "入围", "color": "#886EFF", "discription": "驶入围栏[%s]", "args": 1, "isShow": true},
	{"type": "关机", "color": "#F0644E", "discription": "已关机", "args": 0, "isShow": true}, {"type": "开机", "color": "#3ED080", "discription": "关机报警解除，已重新开机", "args": 0, "isShow": true}, {"type": "低电量", "color": "#F78C75", "discription": "电量少于20%", "args": 0, "isShow": true}, {"type": "通电", "color": "#3ED080", "discription": "断电报警解除，已恢复通电", "args": 0, "isShow": true},
	{"type": "见光恢复", "color": "#F8AD2E", "discription": "光感报警解除，已复位", "args": 0, "isShow": true}, {"type": "低电恢复", "color": "#F78C75", "discription": "低电报警解除，已复位", "args": 0, "isShow": true}, {"type": "出围解除", "color": "#886EFF", "discription": "围栏报警解除", "args": 0, "isShow": true}, {"type": "入围解除", "color": "#886EFF", "discription": "围栏报警解除", "args": 0, "isShow": true},
	{"type": "风险点", "color": "#F66A00", "discription": "进入风险点[%s]", "args": -1, "isShow": true}, {"type": "离线", "color": "#F53E35", "discription": "离线报警", "args": 0, "isShow": true}, {"type": "异动", "color": "#3A9CFF", "discription": "异动报警", "args": 0, "isShow": true}, {"type": "震动", "color": "#3A9CFF", "discription": "震动报警", "args": 0, "isShow": false}, {"type": "拆卸", "color": "#3A9CFF", "discription": "拆卸报警", "args": 0, "isShow": false},
	{"type": "ACC关", "color": "#3A9CFF", "discription": "ACC关报警", "args": 0, "isShow": false}, {"type": "ACC开", "color": "#3A9CFF", "discription": "ACC开报警", "args": 0, "isShow": false}, {"type": "停车超时", "color": "#32C5FF", "discription": "车辆停车已超过24小时", "args": 0, "isShow": true}, {"type": "超速预警", "color": "#F53E35", "discription": "车辆时速超过120公里", "args": 0, "isShow": true}, {"type": "常驻点预警", "color": "#F66A00", "discription": "车辆未进入常驻点[%s]", "args": 1, "isShow": true},
	{"type": "设防报警", "color": "#F8AD2E", "discription": "车辆设防中上传定位", "args": 0, "isShow": true}}

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
	} else {
		for k, v := range rs {
			if int(v.Alarm_type) > len(Alarm_descrip)-1 {
				v.Alarm_desc = "未知警报"
			} else if Alarm_descrip[v.Alarm_type]["args"].(int) > 0 {
				v.Alarm_desc = fmt.Sprintf(Alarm_descrip[v.Alarm_type]["discription"].(string), v.Alarm_pen_name)
			} else if Alarm_descrip[v.Alarm_type]["args"].(int) == 0 {
				v.Alarm_desc = Alarm_descrip[v.Alarm_type]["discription"].(string)
			}
			rs[k] = v
		}
	}
	return
}

func (this *Alarms) GetAlarmInfo(alarmId string) (rs Alarms, err error) {
	err = this.Collection(this).FindId(bson.ObjectIdHex(alarmId)).One(&rs)
	if err == nil {
		if int(rs.Alarm_type) > len(Alarm_descrip)-1 {
			rs.Alarm_desc = "未知警报"
		} else if Alarm_descrip[rs.Alarm_type]["args"].(int) > 0 {
			rs.Alarm_desc = fmt.Sprintf(Alarm_descrip[rs.Alarm_type]["discription"].(string), rs.Alarm_pen_name)
		} else if Alarm_descrip[rs.Alarm_type]["args"].(int) == 0 {
			rs.Alarm_desc = Alarm_descrip[rs.Alarm_type]["discription"].(string)
		}
	}
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
