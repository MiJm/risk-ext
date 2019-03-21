package models

import (
	"errors"
	"risk-ext/utils"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//保单
type Warranty struct {
	Model               `bson:"-" json:"-"` //model基类
	Redis               `bson:"-" json:"-"` //model基类
	WarrantyId          bson.ObjectId       `bson:"_id,omitempty" json:"warranty_id"`                   //id
	WarrantyService     string              `bson:"warranty_service" json:"warranty_service"`           //保单提供方
	WarrantyName        string              `bson:"warranty_name" json:"warranty_name"`                 //保单名
	WarrantyServer      string              `bson:"warranty_server" json:"warranty_server"`             //服务有效期
	WarrantyServerStart uint32              `bson:"warranty_server_start" json:"warranty_server_start"` //服务开始时间
	WarrantyServerEnd   uint32              `bson:"warranty_server_end" json:"warranty_server_end"`     //服务结束时间
	WarrantyDeviceId    uint64              `bson:"warranty_device_id" json:"warranty_device_id"`       //绑定的设备号
	WarrantyOwnerInfo   OwnerInfo           `bson:"warranty_owner_info" json:"warranty_owner_info"`     //保单投保人
	WarrantyCarModel    CarInfo             `bson:"warranty_car_model" json:"warranty_car_model"`       //保单投保的车辆信息
	WarrantyUserId      string              `bson:"warranty_user_id" json:"warranty_user_id"`           //保单绑定的用户ID
	WarrantyStatus      uint8               `bson:"warranty_status" json:"warranty_status"`             //保单状态 0:未激活 1:待审核 2:已激活
	WarrantyCreted      uint32              `bson:"warranty_created" json:"warranty_created"`           //保单创建时间
}

type CarInfo struct {
	CarName                string  `bson:"car_name" json:"car_name"`                                   //车辆名称
	CarBrand               string  `bson:"car_brand" json:"car_brand"`                                 //车辆品牌
	CarSeries              string  `bson:"car_series" json:"car_series"`                               //车辆型号
	CarVin                 string  `bson:"car_vin" json:"car_vin"`                                     //车架号
	CarPurchaseDate        uint32  `bson:"car_purchase_date" json:"car_purchase_date"`                 //车辆购买日期
	CarValue               float64 `bson:"car_value" json:"car_value"`                                 //车辆发票金额
	CarFrontImg            string  `bson:"car_front_img" json:"car_front_img"`                         //车辆正面照片
	CarThumbFrontImg       string  `bson:"car_thumb_front_img" json:"car_thumb_front_img"`             //车辆正面照片缩略图
	CarSideImg             string  `bson:"car_side_img" json:"car_side_img"`                           //车辆侧面照片
	CarThumbSideImg        string  `bson:"car_thumb_side_img" json:"car_thumb_side_img"`               //车辆侧面照片缩略图
	CarCertificateImg      string  `bson:"car_certificate_img" json:"car_certificate_img"`             //车辆合格证照片
	CarThumbCertificateImg string  `bson:"car_thumb_certificate_img" json:"car_thumb_certificate_img"` //车辆合格证照片缩略图
	CarReceiptImg          string  `bson:"car_receipt_img" json:"car_receipt_img"`                     //车辆购车发票照片
	CarThumbReceiptImg     string  `bson:"car_thumb_receipt_img" json:"car_thumb_receipt_img"`         //车辆购车发票照片缩略图
}

type OwnerInfo struct {
	OwnerName             string `bson:"owner_name" json:"owner_name"`                             //投保人姓名
	OwnerIDcard           string `bson:"owner_IDcard" json:"owner_IDcard"`                         //投保人身份证号
	OwnerIDcardFront      string `bson:"owner_IDcard_front" json:"owner_IDcard_front"`             //身份证正面
	OwnerThumbIDcardFront string `bson:"owner_thumb_IDcard_front" json:"owner_thumb_IDcard_front"` //身份证正面缩略图
	OwnerIDcardBack       string `bson:"owner_IDcard_back" json:"owner_IDcard_back"`               //身份证背面
	OwnerThumbIDcardBack  string `bson:"owner_thumb_IDcard_back" json:"owner_thumb_IDcard_back"`   //身份证背面缩略图
	OwnerIDcardImg        string `bson:"owner_IDcard_img" json:"owner_IDcard_img"`                 //手持身份证照片
	OwnerThumbIDcardImg   string `bson:"owner_thumb_IDcard_img" json:"owner_thumb_IDcard_img"`     //手持身份证照片缩略图
}

//根据UserId和保单ID获取保单信息
func (this *Warranty) One(userId, id string, status []uint8) (rs Warranty, err error) {
	if !bson.IsObjectIdHex(id) {
		err = errors.New("ID有误")
		return
	}
	where := bson.M{}
	where["warranty_user_id"] = userId
	where["_id"] = bson.ObjectIdHex(id)
	if len(status) > 1 {
		where["warranty_status"] = bson.M{"$in": status}
	} else if len(status) == 1 {
		where["warranty_status"] = status[0]
	}
	err = this.Collection(this).Find(where).One(&rs)
	return
}

//根据UserId获取账户下保单列表
func (this *Warranty) ListByUserId(userId string, status []uint8) (rs []Warranty, err error) {
	where := bson.M{}
	where["warranty_user_id"] = userId
	if len(status) > 1 {
		where["warranty_status"] = bson.M{"$in": status}
	} else if len(status) == 1 {
		where["warranty_status"] = status[0]
	}
	err = this.Collection(this).Find(where).All(&rs)
	return
}

//根据UserId获取账户下保单列表
func (this *Warranty) GetCount(userId string) (count int) {
	where := bson.M{}
	where["warranty_user_id"] = userId
	where["warranty_status"] = 0
	count, _ = this.Collection(this).Find(where).Count()
	return
}

//第一次激活设备时新增保单
func (this *Warranty) Add() (err error) {
	if this.WarrantyId == EmptyId {
		this.WarrantyId = bson.NewObjectId()
	}
	this.WarrantyCreted = uint32(time.Now().Unix())
	err = this.Collection(this).Insert(*this)
	return
}

//更新本对象
func (this *Warranty) Update(flag bool, unsetfiled ...string) error {
	if this.WarrantyId == EmptyId {
		return errors.New("无效的保单ID")
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
	err := this.Collection(this).UpdateId(this.WarrantyId, query)
	return err
}

//根据设备号查询保单信息
func (this *Warranty) GetWarrantyByDeviceId(deviceId uint64) (rs Warranty, err error) {
	err = this.Collection(this).Find(bson.M{"warranty_device_id": deviceId}).One(&rs)
	return
}
