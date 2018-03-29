package models

import (
	"gopkg.in/mgo.v2/bson"
)

type CheckLog struct {
	CheckId         bson.ObjectId `bson:"_id,omitempty" json:"check_id"`              //id
	CheckDate       uint32        `bson:"check_date" json:"check_date"`               //日期
	CheckItem       string        `bson:"check_item" json:"check_item"`               //查询的项目
	CheckCustomer   string        `bson:"check_customer" json:"check_customer"`       //客户名称
	CheckCustomerId string        `bson:"check_customer_id" json:"check_customer_id"` //客户ID
	CheckCount      int16         `bson:"check_count" json:"check_count"`             //查询次数
	CheckEditor     string        `bson:"check_editor" json:"check_editor"`           //操作人
	CheckEditorId   string        `bson:"check_editor_id" json:"check_editor_id"`     //操作人ID
}
