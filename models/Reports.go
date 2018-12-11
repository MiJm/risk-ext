package models

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Reports struct {
	Model           `bson:"-" json:"-"` //model基类
	ReportId        bson.ObjectId       `bson:"_id,omitempty" json:"report_id"`             //id
	ReportType      uint8               `bson:"report_type" json:"report_type"`             //报表类型 0=追车，1=电话 2=违章
	ReportPlate     string              `bson:"report_plate" json:"report_plate"`           //报表车牌号
	ReportDataFrom  uint8               `bson:"report_data_from" json:"report_data_from"`   //报表源数据来源0=内部数据库 1=外部导入 @ReportType=0 仅追车有效
	ReportStatus    int8                `bson:"report_status" json:"report_status"`         //报表状态0=查询中 1=查询成功 -1=查询失败
	ReportOpenId    string              `bson:"report_open_id" json:"report_open_id"`       //报表关联第三方ID
	ReportData      interface{}         `bson:"report_data" json:"report_data"`             //报表关联第三方结果
	ReportShares    map[string]Shares   `bson:"report_shares" json:"report_shares"`         //报表分享列表  share_id作为键值
	ReportCreateAt  int64               `bson:"report_createat" json:"report_createat"`     //报表查询时间
	ReportDoneAt    int64               `bson:"report_doneat" json:"report_doneat"`         //报表查询完成时间
	ReportDeleteAt  int64               `bson:"report_deleteat" json:"report_deleteat"`     //报表删除时间 0=未删除
	ReportCompanyId string              `bson:"report_company_id" json:"report_company_id"` //报表关联企业ID
}

type Shares struct {
	ShareId       string `bson:"share_id" json:"share_id"`             //id bson.ObjectId hex
	ShareMobile   string `bson:"share_mobile" json:"share_mobile"`     //报表分享所属手机
	ShareFname    string `bson:"share_fname" json:"share_fname"`       //分享人姓名
	ShareViews    uint32 `bson:"share_views" json:"share_views"`       //报表分享人查看次数
	ShareCreateAt int64  `bson:"share_createat" json:"share_createat"` //报表分享时间
}

type Routes struct {
	Device_speed   uint8  `json:"device_speed"`   //速度
	Device_latlng  Latlng `json:"device_latlng"`  //经纬度
	Device_slatlng Latlng `json:"device_slatlng"` //原坐标
	Device_address string `json:"device_address"` //地址
	Device_loctype uint8  `json:"device_loctype"` //定位类型0=GPS 1=基站 2=WiFi
	Device_loctime uint32 `json:"device_loctime"` //定位时间}
}
type Latlng struct {
	Type        string    `json:"type"`        //Point
	Coordinates []float64 `json:"coordinates"` //lng lat
}

func (this *Reports) RemoveShare(shareId string) (err error) {
	if this.ReportShares != nil {
		delete(this.ReportShares, shareId)
	}
	err = this.Update()
	return
}

func (this *Reports) Delete() error {
	if this.ReportId == EmptyId {
		return errors.New("id be empty")
	}
	this.ReportDeleteAt = time.Now().Unix()
	return this.Update()
}

func (this *Reports) Insert() (err error) {
	this.ReportId = bson.NewObjectId()
	this.ReportCreateAt = time.Now().Unix()
	err = this.Collection(this).Insert(*this)
	return
}

func (this *Reports) Update() (err error) {
	if this.ReportId != EmptyId && this.ReportShares != nil {
		update := bson.M{"$set": *this}
		err = this.Collection(this).UpdateId(this.ReportId, update)
	}
	return
}

func (this *Reports) List(query interface{}, page, size int) (rs []*Reports, num int) {
	if this.ReportId == EmptyId {
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * size
		find := this.Collection(this).Find(query)
		num, _ = find.Count()
		find.Sort("-report_createat").Skip(offset).Limit(size).All(&rs)
	}
	return
}

func (this *Reports) Lists(query interface{}, page, size int) (rs []*Reports, num int, err error) {
	find := this.Collection(this).Find(query)
	if this.ReportId == EmptyId {
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * size
		num, _ = find.Count()
		err = find.Sort("-report_createat").Skip(offset).Limit(size).All(&rs)
	}
	return
}

func (this *Reports) One(id ...string) (rs Reports, err error) {
	var gid bson.ObjectId
	if len(id) == 1 {
		if !bson.IsObjectIdHex(id[0]) {
			err = errors.New("报表ID有误")
			return
		} else {
			gid = bson.ObjectIdHex(id[0])
		}
	} else {
		if this.ReportId == EmptyId {
			err = errors.New("报表ID错误")
			return
		} else {
			gid = this.ReportId
		}
	}

	err = this.Collection(this).FindId(gid).One(&rs)

	if err == nil && rs.ReportDeleteAt > 0 {
		err = errors.New("该报表已被删除了")
	}
	return
}
