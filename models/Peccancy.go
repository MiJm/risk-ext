package models

import (
	"crypto/tls"

	"github.com/astaxie/beego/httplib"
	"gopkg.in/mgo.v2/bson"
)

type Peccancies struct {
	PeccancyId     bson.ObjectId `bson:"_id,omitempty" json:"peccancy_id"`                   //id
	TotalFine      float32       `bson:"peccancy_total_fine" json:"peccancy_total_fine"`     //未处理违章总罚款
	TotalPoints    uint8         `bson:"peccancy_total_points" json:"peccancy_total_points"` //未处理违章总扣分
	Untreated      uint8         `bson:"peccancy_untreated" json:"peccancy_untreated"`       //未处理违章条数
	PeccanciesInfo []Info        `bson:"peccancy_info" json:"peccancy_info"`                 //违章时间
	QueryDate      uint32        `bson:"query_date" json:"query_date"`                       //查询日期
	CarPlate       string        `bson:"car_plate" json:"car_plate"`                         //车牌号

}

type Info struct {
	PeccancyTime          uint32  `bson:"peccancy_time" json:"peccancy_time"`                     //违章时间
	PeccancyFine          float32 `bson:"peccancy_fine" json:"peccancy_fine"`                     //违章罚款总额
	PeccancyAddress       string  `bson:"peccancy_address" json:"peccancy_address"`               //违章地址
	PeccancyReason        string  `bson:"peccancy_reason" json:"peccancy_reason"`                 //违章原因
	PeccancyPoint         uint8   `bson:"peccancy_point" json:"peccancy_point"`                   //违章扣分
	PeccancyViolationCity string  `bson:"peccancy_violation_city" json:"peccancy_violation_city"` //违章发生城市
	PeccancyProvince      string  `bson:"peccancy_province" json:"peccancy_province"`             //省份
	PeccancyCity          string  `bson:"peccancy_city" json:"peccancy_city"`                     //城市
	PeccancyViolationNum  string  `bson:"peccancy_violation_num" json:"peccancy_time"`            //违章官方条码
	PeccancyStatus        int8    `bson:"peccancy_status" json:"peccancy_status"`                 //违章缴费状态
}

//阿里云获取违章支持的城市接口
func GetPeccancyCity() (rs string) {
	appcode := "5ba0770c939048b98fed07083106988a"
	url := "http://ddycapi.market.alicloudapi.com/violation/condition"
	req := httplib.Get(url)
	req.Header("Authorization", "APPCODE "+appcode)
	rs, err := req.String()
	if err != nil {
		rs = ""
	}
	return
}

//阿里云违章查询的接口
func GetPeccancy(carNum, vin, engineNo, city, carType string) string {
	appcode := "5ba0770c939048b98fed07083106988a"
	url := "http://ddycapi.market.alicloudapi.com/violation/query"
	req := httplib.Post(url)
	req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	req.Header("Content-Type", "application/json;charset=UTF-8")
	req.Header("Authorization", "APPCODE "+appcode)
	req.Param("plateNumber", carNum)
	req.Param("vin", vin)
	req.Param("engineNo", engineNo)
	req.Param("carType", carType)
	req.Param("city", city)
	rs, err := req.String()
	if err != nil {
		return ""
	}
	return rs
}
