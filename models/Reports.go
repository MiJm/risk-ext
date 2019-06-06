package models

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Reports struct {
	Model             `bson:"-" json:"-"` //model基类
	ReportId          bson.ObjectId       `bson:"_id,omitempty" json:"report_id"`                 //id
	ReportType        uint8               `bson:"report_type" json:"report_type"`                 //报表类型 0=追车，1=电话 2=违章 3=征信
	ReportPlate       string              `bson:"report_plate" json:"report_plate"`               //报表车牌号
	ReportName        string              `bson:"report_name" json:"report_name"`                 //姓名
	ReportMobile      string              `bson:"report_mobile" json:"report_mobile"`             //手机号
	ReportDataFrom    uint8               `bson:"report_data_from" json:"report_data_from"`       //报表源数据来源0=内部数据库 1=外部导入 @ReportType=0 仅追车有效
	ReportStatus      int8                `bson:"report_status" json:"report_status"`             //报表状态0=查询中 1=查询成功 -1=查询失败 2=待审核 3=拒绝
	ReportCompanyName string              `bson:"report_company_name" json:"report_company_name"` //企业名称
	ReportOpenId      string              `bson:"report_open_id" json:"report_open_id"`           //报表关联第三方ID
	ReportData        interface{}         `bson:"report_data" json:"report_data"`                 //报表关联第三方结果
	ReportShares      map[string]Shares   `bson:"report_shares" json:"report_shares"`             //报表分享列表  share_id作为键值
	ReportCreateAt    int64               `bson:"report_createat" json:"report_createat"`         //报表查询时间
	ReportDoneAt      int64               `bson:"report_doneat" json:"report_doneat"`             //报表查询完成时间
	ReportDeleteAt    int64               `bson:"report_deleteat" json:"report_deleteat"`         //报表删除时间 0=未删除
	ReportCompanyId   string              `bson:"report_company_id" json:"report_company_id"`     //报表关联企业ID
	ReportImage       *Image              `bson:"report_image" json:"report_image"`               //征信相关文件图片
	ReportIdCard      string              `bson:"report_idcard" json:"report_idcard"`             //征信查询证件号码
	ReportAuditName   string              `bson:"report_audit_name" json:"report_audit_name"`     //审核人
	ReportAuditTime   int64               `bson:"report_audit_time" json:"report_audit_time"`     //审核时间
	ReportCheckName   string              `bson:"report_check_name" json:"report_check_name"`     //查询人
	ReportNumber      string              `bson:"report_number" json:"report_num"`                //报告编号
}

type Shares struct {
	ShareId       string `bson:"share_id" json:"share_id"`             //id bson.ObjectId hex
	ShareMobile   string `bson:"share_mobile" json:"share_mobile"`     //报表分享所属手机
	ShareFname    string `bson:"share_fname" json:"share_fname"`       //分享人姓名
	ShareViews    uint32 `bson:"share_views" json:"share_views"`       //报表分享人查看次数
	ShareCreateAt int64  `bson:"share_createat" json:"share_createat"` //报表分享时间
}

type Image struct {
	AuthImage       string `bson:"auth_image" json:"auth_image"`               //认证文件原图
	AuthImageThumb  string `bson:"auth_image_thumb" json:"auth_image_thumb"`   //认证文件缩略图
	FrontImageUrl   string `bson:"front_image_url" json:"front_image_url"`     //证件图片地址
	FrontImageThumb string `bson:"front_image_thumb" json:"front_image_thumb"` //缩略图地址
	BackImageUrl    string `bson:"back_image_url" json:"back_image_url"`       //证件图片地址
	BackImageThumb  string `bson:"back_image_thumb" json:"back_image_thumb"`   //缩略图地址
}

type Routes struct {
	Device_latlng  Latlng `json:"device_latlng"`  //经纬度
	Device_loctime uint32 `json:"device_loctime"` //定位时间
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

//type Devices struct {
//	Device_id           uint64        `json:"device_id"`                                            //设备号
//	Device_car_id       bson.ObjectId `bson:"device_car_id,omitempty" json:"device_car_id"`         //车辆ID
//	Device_car_plate    string        `json:"device_car_plate"`                                     //车牌号
//	Device_company_id   bson.ObjectId `bson:"device_company_id,omitempty" json:"device_company_id"` //企业ID
//	Device_company_name string        `json:"device_company_name"`                                  //企业名
//	Device_group_id     bson.ObjectId `bson:"device_group_id,omitempty" json:"device_group_id"`     //组织ID
//	Device_group_name   string        `json:"device_group_name"`                                    //组织名
//	Device_deleted      uint8         `json:"device_deleted"`                                       //是否已删除0未 1已
//	Device_bind_time    uint32        `json:"device_bind_time"`                                     //绑车时间
//	Device_unbind_time  uint32        `json:"device_unbind_time"`                                   //解绑时间
//}

func (this *Cars) OneCar(carNum string) (count int, err error, ca Cars) {
	count, err = this.Collection(this).Find(bson.M{"car_plate": carNum, "car_deleted": 0}).Count()
	err = this.Collection(this).Find(bson.M{"car_plate": carNum, "car_deleted": 0}).One(&ca)
	return
}

func (this *Cars) GetCarNumsByCondition(condition bson.M) (count int, err error) {
	count, err = this.Collection(this).Find(condition).Count()
	return
}

type Groups struct {
	Model         `bson:"-" json:"-"` //model基类
	Group_id      bson.ObjectId       `bson:"_id,omitempty" json:"group_id"` //id
	Group_name    string              `json:"group_name"`                    //组织名
	Group_comid   string              `json:"group_comid"`                   //所属企业ID
	Group_carnum  int32               `json:"group_carnum"`                  //组织车辆数量
	Group_devnum  int32               `json:"group_devnum"`                  //组织设备数量
	Group_sub     []Groups            `json:"group_sub"`                     //子级组织
	Group_parent  *Sgroups            `json:"group_parent"`                  //组织父节点
	Group_deleted uint32              `json:"group_deleted"`                 //是否已删除0未 删除时间
	Group_date    uint32              `json:"group_date"`                    //组织创建时间
}

type Sgroups struct {
	Group_id   bson.ObjectId `json:"group_id"`   //id
	Group_name string        `json:"group_name"` //组织名
}

func (this *Groups) One(id ...string) (gval Groups, err error) {

	var gid bson.ObjectId
	if len(id) == 1 {
		if !bson.IsObjectIdHex(id[0]) {
			err = errors.New("组织ID错误")
			return
		} else {
			gid = bson.ObjectIdHex(id[0])
		}
	} else {
		if this.Group_id == EmptyId {
			err = errors.New("组织ID错误")
			return
		} else {
			gid = this.Group_id
		}
	}

	err = this.Collection(this).FindId(gid).One(&gval)

	if err == nil && gval.Group_deleted == 1 {
		err = errors.New("组织已被删除了")
	}

	return
}

func (this *Reports) CheckPhone(phone, reportId string) (res *Reports, err error) {
	if !bson.IsObjectIdHex(reportId) {
		err = errors.New("查看报告路径有误")
		return
	}
	query := "report_shares." + phone
	where := bson.M{}
	where[query] = bson.M{"$ne": nil}
	where["report_deleteat"] = 0
	where["report_status"] = 1
	where["_id"] = bson.ObjectIdHex(reportId)
	err = this.Collection(this).Find(where).One(&res)
	return
}

//判断用户是否有权限查看该车辆
func IsCanCheck(groupId, companyId string, mem users) bool {
	var flag = false
	if mem.UserLevel == MEMBER_SUPER {
		if mem.UserCompany_id == companyId {
			flag = true
		}
	} else {
		gro, err := new(Groups).One(groupId)
		if err == nil {
			grosub := gro.Group_sub
			if len(grosub) > 0 {
				for _, v := range grosub {
					if v.Group_id.Hex() == mem.UserGroupId {
						flag = true
					}
				}
			} else {
				if mem.UserGroupId == groupId {
					flag = true
				}
			}
		}
	}
	return flag
}
