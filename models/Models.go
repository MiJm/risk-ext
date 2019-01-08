package models

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

type Models struct {
	Model                `bson:"-" json:"-"` //model基类
	Model_id             bson.ObjectId       `bson:"_id" json:"model_id"`
	Model_name           string              `json:"model_name"`           //型号名
	Model_type           uint8               `json:"model_type"`           //设备类型0无线 1有线
	Model_works          [][]string          `json:"model_works"`          //可用工作模式 0闹钟模式 1定时循环模式 2星期模式 3追踪模式 工作参数 如果是数字则单位分钟
	Model_works_mode     []string            `json:"model_works_mode"`     //选择的设备工作模式(无线设备) 0闹钟模式 1定时模式 2星期模式 3追踪模式
	Model_current_work   uint8               `json:"model_current_work"`   //当前工作模式 0闹钟模式 1定时循环模式 2星期模式 3追踪模式
	Model_correspondence string              `json:"model_correspondence"` //对应设备型号
	Model_supplier_model string              `json:"model_supplier_model"` //供应商型号
	Model_status         uint8               `json:"model_status"`         //设备状态 0正在销售 1停止销售 2测试
	Model_loc_model      []string            `json:"model_loc_model"`      //型号支持的定位模式  0GPS 1基站 2WIFI 3北斗
	Model_command        []string            `json:"model_command"`        //型号支持的指令 0=重启设备 1=断油电 2=通油电 3开追踪 4设置闹钟 5设置星期
	Model_alarm          []string            `json:"model_wireless_alarm"` //型号支持的预警 当无线设备时 0光感 1翻转 2低电量; 当有线设备时 0断电
	Model_state_method   []string            `json:"model_state_method"`   //型号支持的状态检测（有线设备） 0 ACC
	Model_control        []string            `json:"model_control"`        //型号支持的控制功能（有线设备） 0断油电
	Model_stop_date      uint32              `json:"model_stop_date"`      //停止销售时间
	Model_date           uint32              `json:"model_date"`           //添加日期
	Model_port           uint32              `json:"model_port"`           //端口号
}

//根据类型Id获取模式信息
func (this *Models) One(id string) (mod Models, err error) {
	if !bson.IsObjectIdHex(id) {
		errors.New("不是BsonId")
		return
	}
	err = this.Collection(this).FindId(bson.ObjectIdHex(id)).One(&mod)
	return
}
