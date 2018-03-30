package models

import (
	"gopkg.in/mgo.v2/bson"
)

var Extra = []string{"智能追车", "电话邦", "违章查询"}

type Logs struct {
	Model         `bson:"-" json:"-"` //model基类
	LogId         bson.ObjectId       `bson:"_id,omitempty" json:"log_id"`            //id
	LogDate       int64               `bson:"log_date" json:"log_date"`               //日期
	LogItem       string              `bson:"log_item" json:"log_item"`               //查询的项目
	LogCompany    string              `bson:"log_company" json:"log_company"`         //客户名称
	LogCompanyId  string              `bson:"log_company_id" json:"log_company_id"`   //客户ID
	LogDetail     string              `bson:"log_detail" json:"log_detail"`           //描述
	LogOperator   string              `bson:"log_operator" json:"log_operator"`       //操作人
	LogOperatorId string              `bson:"log_operator_id" json:"log_operator_id"` //操作人ID
}

func (this *Logs) Insert() error {
	return this.Collection(this).Insert(*this)
}
