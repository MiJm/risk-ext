package models

import (
	"gopkg.in/mgo.v2/bson"
)

var Extra = []string{"智能追车", "电话邦", "违章查询", "征信防欺诈"}

type Logs struct {
	Model         `bson:"-" json:"-"` //model基类
	LogId         bson.ObjectId       `bson:"_id,omitempty" json:"log_id"`            //id
	LogDate       int64               `bson:"log_date" json:"log_date"`               //日期
	LogItem       string              `bson:"log_item" json:"log_item"`               //查询的项目
	LogType       int                 `bson:"log_type" json:"log_type"`               //查询项目 0=追车，1=电话 2=违章 3=征信
	LogCompany    string              `bson:"log_company" json:"log_company"`         //客户名称
	LogCompanyId  string              `bson:"log_company_id" json:"log_company_id"`   //客户ID
	LogDetail     string              `bson:"log_detail" json:"log_detail"`           //描述
	LogOperator   string              `bson:"log_operator" json:"log_operator"`       //操作人
	LogOperatorId string              `bson:"log_operator_id" json:"log_operator_id"` //操作人ID
	LogOperateIp  string              `bson:"log_operate_ip" json:"log_operate_ip"`   //操作ip地址
}

func (this *Logs) Insert() error {
	return this.Collection(this).Insert(*this)
}

func (this *Logs) List(query interface{}, page, size int) (rs []*Logs, num int, err error) {
	if this.LogId == EmptyId {
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * size
		find := this.Collection(this).Find(query)
		num, _ = find.Count()
		err = find.Sort("-log_date").Skip(offset).Limit(size).All(&rs)
	}
	return
}
