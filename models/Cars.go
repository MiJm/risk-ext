package models

import "gopkg.in/mgo.v2/bson"

type Cars struct {
	Model            `bson:"-" json:"-"` //model基类
	Car_id           bson.ObjectId       `bson:"_id,omitempty" json:"car_id"`
	Car_plate        string              `json:"car_plate"`        //车牌号
	Car_company_id   string              `json:"car_company_id"`   //企业ID
	Car_company_name string              `json:"car_company_name"` //企业简称
	Car_group_id     string              `json:"car_group_id"`     //组织ID
	Car_group_name   string              `json:"car_group_name"`   //组织简称
	Car_devices      []Devices           `json:"car_devices"`      //绑定的设备
	Car_deleted      uint32              `json:"car_deleted"`      //删除时间
}

//获取全部车辆列表
func (this *Cars) GetAllCars(loginMember users) (interface{}, error) {
	result := []struct {
		Car_latlng Latlng `json:"car_latlng"` //坐标
	}{}
	var where = bson.M{}

	if loginMember.UserLevel == MEMBER_SUPER {
		where["car_company_id"] = loginMember.UserCompany_id
	} else {
		groupData, _ := new(Groups).One(loginMember.UserGroupId)
		var ids []string
		if len(groupData.Group_sub) > 0 {
			for _, val := range groupData.Group_sub {
				ids = append(ids, val.Group_id.Hex())
			}
		}
		ids = append(ids, loginMember.UserGroupId)
		where["car_group_id"] = bson.M{"$in": ids}
	}
	where["car_deleted"] = 0
	where["car_devices"] = bson.M{"$elemMatch": bson.M{"$ne": nil}}
	err := this.Collection(this).Find(where).All(&result)

	return result, err
}
