package app

import (
	"log"
	"plugin"

	//socketgo "github.com/nulijiabei/socketgo"
)

type CarLocal struct {
	CarPlate     string `json:"car_plate"`      //车牌号
	CarOwner     string `json:"car_owner"`      //车主
	CarVoltage   uint8  `json:"car_voltage"`    //电量最大100
	CarLatlng    Latlng `json:"car_latlng"`     //坐标
	CarSpeed     uint8  `json:"car_speed"`      //车速
	CarDirection uint16 `json:"car_direction"`  //方向最大360
	CarCity      string `json:"car_city"`       //城市
	CarLoctype   uint8  `json:"car_loctype"`    //定位类型 0gps 1基站 2WiFi
	CarState     uint8  `json:"car_state"`      //0静止 1运行 2离线
	CarGroupId   string `json:"car_group_id"`   //组织ID
	CarPgroupId  string `json:"car_pgroup_id"`  //父级组织ID 默认空
	CarCompanyId string `json:"car_company_id"` //企业ID
	CarAddress   string `json:"car_address"`    //最后定位地址
	CarLoctime   uint32 `json:"car_loctime"`    //最后定位时间
}

type Latlng struct {
	Type        string    `json:"type"`        //Point
	Coordinates []float64 `json:"coordinates"` //lng lat
}

var (
	CarDataChan = make(chan CarLocal, 100)
)

//TCP
func StartUdp(port string) {
	p, err := plugin.Open("libs/libs.so")
	if err != nil {
		log.Println("插件加载失败", err)
	}
	su, _ := p.Lookup("StartUdp")
	su.(func(addr, port string, cchan chan CarLocal))("", port, CarDataChan)
}
