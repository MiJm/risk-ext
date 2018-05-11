package models

import (
	"errors"
	"risk-ext/utils"

	"gopkg.in/mgo.v2/bson"
)

type Devices struct {
	Model                   `bson:"-" json:"-"` //model基类
	Device_id               uint64              `json:"device_id"`                                            //设备号
	Device_name             string              `json:"device_name"`                                          //设备名
	Device_sim              uint64              `json:"device_sim"`                                           //sim卡
	Device_sim_info         SimInfo             `bson:"device_sim_info" json:"device_sim_info"`               //SIM卡详细信息
	Device_car_id           bson.ObjectId       `bson:"device_car_id,omitempty" json:"device_car_id"`         //车辆ID
	Device_car_plate        string              `json:"device_car_plate"`                                     //车牌号
	Device_company_id       bson.ObjectId       `bson:"device_company_id,omitempty" json:"device_company_id"` //企业ID
	Device_company_name     string              `json:"device_company_name"`                                  //企业名
	Device_group_id         bson.ObjectId       `bson:"device_group_id,omitempty" json:"device_group_id"`     //组织ID
	Device_group_name       string              `json:"device_group_name"`                                    //组织名
	Device_server_starttime uint32              `json:"device_server_starttime"`                              //服务开始时间
	Device_server_endtime   uint32              `json:"device_server_endtime"`                                //服务结束时间
	Device_due              uint8               `json:"device_due"`                                           //是否到期0未到期 1到期
	Device_remark           string              `json:"device_remark"`                                        //备注
	Device_sim_starttime    uint32              `json:"device_sim_starttime"`                                 //sim开始时间
	Device_sim_endtime      uint32              `json:"device_sim_endtime"`                                   //sim结束时间
	Device_install          int8                `json:"device_install"`                                       //安装位置
	Device_outtime          uint32              `json:"device_outtime"`                                       //出库时间
	Device_batch_id         uint32              `json:"device_batch_id"`                                      //入库批次号
	Device_serial_id        uint32              `json:"device_serial_id"`                                     //出库流水号
	Device_intime           uint32              `json:"device_intime"`                                        //入库时间
	Device_rectime          uint32              `json:"device_rectime"`                                       //回收时间
	Device_deleted          uint8               `json:"device_deleted"`                                       //是否已删除0未 1已
	Device_tracking         uint8               `json:"device_tracking"`                                      //状态0=未追踪 1=准备追踪 2=正在追踪 3=准备恢复正常模式
	Device_bind_time        uint32              `json:"device_bind_time"`                                     //绑车时间
	Device_unbind_time      uint32              `json:"device_unbind_time"`                                   //解绑时间
}

type SimInfo struct {
	Msisdn               string `json:"msisdn"`               //SIM卡号
	Iccid                string `json:"iccid"`                //iccid
	Imsi                 string `json:"imsi"`                 //imsi
	Sp_code              string `json:"sp_code"`              //短信端口
	Carrier              string `json:"carrier"`              //运营商
	Data_plan            int    `json:"data_plan"`            //套餐大小
	Data_usage           string `json:"data_usage"`           //当月用量
	Account_status       string `json:"account_status"`       //卡状态
	Expiry_date          string `json:"expiry_date"`          //计费结束日期
	Active               bool   `json:"active"`               //激活/未激活
	Test_valid_date      string `json:"test_valid_date"`      //测试期起始日期
	Silent_valid_date    string `json:"silent_valid_date"`    //沉默期起始日期
	Test_used_data_usage string `json:"test_used_data_usage"` //测试期已用流量
	Active_date          string `json:"active_date"`          //激活日期
	Data_balance         string `json:"data_balance"`         //剩余流量
	Outbound_date        string `json:"outbound_date"`        //出库日期
	Support_sms          bool   `json:"support_sms"`          //是否支持短信
}

//根据SIM卡号搜索
func (this *Devices) OneBySim(sim uint64) (dev Devices, err error) {
	err = this.Collection(this).Find(bson.M{"device_sim": sim}).One(&dev)
	return
}

//更新本对象
func (this *Devices) Update(flag bool, unsetfiled ...string) error {
	if this.Device_id == 0 {
		return errors.New("无效的设备ID")
	}
	query := bson.M{}
	data := utils.Struct2Map(*this, flag)
	query["$set"] = data

	if len(unsetfiled) > 0 {
		unsetData := bson.M{}
		for _, ud := range unsetfiled {
			unsetData[ud] = 1
		}
		query["$unset"] = unsetData
	}
	err := this.Collection(this).Update(bson.M{"device_id": this.Device_id}, query)
	return err
}
