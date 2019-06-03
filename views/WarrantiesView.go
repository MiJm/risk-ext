package views

import (
	"encoding/json"
	"fmt"
	"risk-ext/app"
	"risk-ext/models"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type WarrantiesView struct {
	Views
}

func (this *WarrantiesView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//分步提交保单信息
func (this *WarrantiesView) Put(ctx iris.Context) (statuCode int, data interface{}) {
	step := ctx.Params().Get("step") //分步提交第几步
	if step == "2" {
		statuCode, data = this.AddOwnerInfo(ctx)
	} else {
		statuCode, data = this.AddCarInfo(ctx)
	}
	return
}

//添加保单投保人信息(第二步)
func (this *WarrantiesView) AddOwnerInfo(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	defer func() {
		result = data
	}()
	statuCode = 400
	user := Session.Customer
	id := ctx.FormValue("id")
	rs, err := new(models.Warranty).One(user.UserId.Hex(), id, []uint8{})
	if err != nil {
		data["code"] = 0
		data["msg"] = "未查到该保单信息,请核实后再填写"
		data["data"] = nil
		return
	}
	if rs.WarrantyStatus != 0 && rs.WarrantyStatus != 2 {
		data["code"] = 0
		data["msg"] = "该保单已填写完整信息无法再次提交"
		data["data"] = nil
		return
	}
	name := ctx.FormValue("name")
	idcard := ctx.FormValueDefault("idcard", "")
	phone := ctx.FormValue("phone")
	if idcard == "" || name == "" || phone == "" {
		data["code"] = 0
		data["msg"] = "请填写完整投保人信息"
		data["data"] = nil
		return
	}

	ctx.SetMaxRequestBodySize(2 << 30)

	//身份证正面
	front := ctx.FormValue("front")
	if front == "" && (rs.WarrantyOwnerInfo.OwnerIDcardFront == "" || rs.WarrantyOwnerInfo.OwnerThumbIDcardFront == "") {
		data["code"] = 0
		data["msg"] = "请上传身份证正面照片"
		data["data"] = nil
		return
	}
	if front != "" {
		var frontPath models.ImgPath
		err = json.Unmarshal([]byte(front), &frontPath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传身份证正面失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if frontPath.Path == "" || frontPath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传身份证正面照片"
			data["data"] = nil
			return
		}
		rs.WarrantyOwnerInfo.OwnerIDcardFront = frontPath.Path
		rs.WarrantyOwnerInfo.OwnerThumbIDcardFront = frontPath.ThumbPath
	}

	rs.WarrantyOwnerInfo.OwnerMobile = phone

	//身份证背面
	back := ctx.FormValue("back")
	if back == "" && (rs.WarrantyOwnerInfo.OwnerIDcardBack == "" || rs.WarrantyOwnerInfo.OwnerThumbIDcardBack == "") {
		data["code"] = 0
		data["msg"] = "请上传身份证背面照片"
		data["data"] = nil
		return
	}
	if back != "" {
		var backPath models.ImgPath
		err = json.Unmarshal([]byte(back), &backPath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传身份证背面失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if backPath.Path == "" || backPath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传身份证背面照片"
			data["data"] = nil
			return
		}
		rs.WarrantyOwnerInfo.OwnerIDcardBack = backPath.Path
		rs.WarrantyOwnerInfo.OwnerThumbIDcardBack = backPath.ThumbPath
	}

	//手持身份证
	ownerIDcard := ctx.FormValue("owner_idcard")
	if ownerIDcard == "" && (rs.WarrantyOwnerInfo.OwnerIDcardImg == "" || rs.WarrantyOwnerInfo.OwnerThumbIDcardImg == "") {
		data["code"] = 0
		data["msg"] = "请上传手持身份证照片"
		data["data"] = nil
		return
	}
	if ownerIDcard != "" {
		var ownerIDcardPath models.ImgPath
		err = json.Unmarshal([]byte(ownerIDcard), &ownerIDcardPath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传手持身份证照片失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if ownerIDcardPath.Path == "" || ownerIDcardPath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传手持身份证照片"
			data["data"] = nil
			return
		}
		rs.WarrantyOwnerInfo.OwnerIDcardImg = ownerIDcardPath.Path
		rs.WarrantyOwnerInfo.OwnerThumbIDcardImg = ownerIDcardPath.ThumbPath
	}

	rs.WarrantyOwnerInfo.OwnerIDcard = idcard
	rs.WarrantyOwnerInfo.OwnerName = name
	rs.WarrantyStatus = 1
	err = rs.Update(false)
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传投保人信息失败"
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = rs
	return
}

//添加保单车辆信息
func (this *WarrantiesView) AddCarInfo(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	defer func() {
		result = data
	}()
	statuCode = 400
	user := Session.Customer
	id := ctx.FormValue("id")
	rs, err := new(models.Warranty).One(user.UserId.Hex(), id, []uint8{})
	if err != nil {
		data["code"] = 0
		data["msg"] = "未查到该保单信息,请核实后再填写"
		data["data"] = nil
		return
	}
	if rs.WarrantyStatus != 0 && rs.WarrantyStatus != 2 {
		data["code"] = 0
		data["msg"] = "该保单已填写完整信息无法再次提交"
		data["data"] = nil
		return
	}

	brand := ctx.FormValueDefault("brand", "")           //保单车辆品牌
	series := ctx.FormValueDefault("series", "")         //保单车辆型号
	vin := ctx.FormValueDefault("vin", "")               //保单车辆车架号
	purchase := ctx.PostValueInt64Default("purchase", 0) //保单车辆购买日期
	value := ctx.PostValueFloat64Default("value", 0.0)   //保单车辆购买时发票金额
	engine := ctx.FormValueDefault("engine", "")         //保单车辆电机号
	if brand == "" || series == "" || vin == "" || purchase == 0 || value == 0.0 || engine == "" {
		data["code"] = 0
		data["msg"] = "请填写完整保障物信息"
		data["data"] = nil
		return
	}
	carModel := rs.WarrantyCarModel
	carModel.CarBrand = brand
	carModel.CarSeries = series
	carModel.CarVin = vin
	carModel.CarPurchaseDate = uint32(purchase)
	carModel.CarValue = value
	carModel.CarEngine = engine
	rs.WarrantyCarModel = carModel
	ctx.SetMaxRequestBodySize(2 << 31)

	//车辆正面照片
	front := ctx.FormValue("front")
	if front == "" && rs.WarrantyCarModel.CarFrontImg == "" {
		data["code"] = 0
		data["msg"] = "请上传车辆正面照片"
		data["data"] = nil
		return
	}
	if front != "" {
		var frontPath models.ImgPath
		err = json.Unmarshal([]byte(front), &frontPath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆正面文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if frontPath.Path == "" || frontPath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆正面照片"
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarFrontImg = frontPath.Path
		rs.WarrantyCarModel.CarThumbFrontImg = frontPath.ThumbPath
	}

	//车辆侧面
	side := ctx.FormValue("side")
	if side == "" && (rs.WarrantyCarModel.CarSideImg == "" || rs.WarrantyCarModel.CarThumbSideImg == "") {
		data["code"] = 0
		data["msg"] = "请上传车辆侧面照片"
		data["data"] = nil
		return
	}
	if side != "" {
		var sidePath models.ImgPath
		err = json.Unmarshal([]byte(side), &sidePath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆侧面文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if sidePath.Path == "" || sidePath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆侧面照片"
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarSideImg = sidePath.Path
		rs.WarrantyCarModel.CarThumbSideImg = sidePath.ThumbPath
	}

	//车辆合格证
	certificate := ctx.FormValue("certificate")
	if certificate == "" && (rs.WarrantyCarModel.CarCertificateImg == "" || rs.WarrantyCarModel.CarThumbCertificateImg == "") {
		data["code"] = 0
		data["msg"] = "请上传车辆合格证照片"
		data["data"] = nil
		return
	}
	if certificate != "" {
		var certificatePath models.ImgPath
		err = json.Unmarshal([]byte(certificate), &certificatePath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆合格证文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if certificatePath.Path == "" || certificatePath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆合格证照片"
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarCertificateImg = certificatePath.Path
		rs.WarrantyCarModel.CarThumbCertificateImg = certificatePath.ThumbPath
	}

	//车辆购车发票
	receipt := ctx.FormValue("receipt")
	if receipt == "" && (rs.WarrantyCarModel.CarReceiptImg == "" || rs.WarrantyCarModel.CarThumbReceiptImg == "") {
		data["code"] = 0
		data["msg"] = "请上传车辆购车发票照片"
		data["data"] = nil
		return
	}
	if receipt != "" {
		var receiptPath models.ImgPath
		err = json.Unmarshal([]byte(receipt), &receiptPath)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆购车发票失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		if receiptPath.Path == "" || receiptPath.ThumbPath == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆购车发票照片"
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarReceiptImg = receiptPath.Path
		rs.WarrantyCarModel.CarThumbReceiptImg = receiptPath.ThumbPath
	}

	err = rs.Update(false)
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传投保车辆信息信息失败"
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = rs
	return
}

//查询单个保单信息
func (this *WarrantiesView) Detail(ctx iris.Context, id string) (statuCode int, data app.M) {
	data = make(app.M)
	statuCode = 400
	rs, err := new(models.Warranty).One(Session.Customer.UserId.Hex(), id, []uint8{})
	if err != nil {
		data["code"] = 0
		data["msg"] = fmt.Sprintf("获取保单信息失败")
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = rs
	return
}

//查询用户下已激活或待审核保单列表
func (this *WarrantiesView) List(ctx iris.Context) (statuCode int, data app.M) {
	data = make(app.M)

	statuCode = 400
	rs, err := new(models.Warranty).ListByUserId(Session.Customer.UserId.Hex(), []int{1, 2, 3})
	if err != nil || len(rs) == 0 {
		rs = make([]models.Warranty, 0)
	}
	rs1, err := new(models.Warranty).ListByUserId(Session.Customer.UserId.Hex(), []int{0})
	if err != nil || len(rs1) == 0 {
		rs1 = make([]models.Warranty, 0)
	}

	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = bson.M{"list": rs, "disactive": rs1}
	return
}

//获取账户下待激活的保单列表
func (this *WarrantiesView) GetDisActiveList(ctx iris.Context) (statuCode int, data app.M) {
	statuCode = 400
	data = make(app.M)
	rs, err := new(models.Warranty).ListByUserId(Session.Customer.UserId.Hex(), []int{0})
	if err != nil {
		data["code"] = 0
		data["msg"] = fmt.Sprintf("获取未激活保单列表失败")
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = rs
	return
}

//获取保单单个信息或保单列表()
func (this *WarrantiesView) Get(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	defer func() {
		result = data
	}()
	statuCode = 400
	id := ctx.Params().Get("id")
	if id != "" {
		statuCode, data = this.Detail(ctx, id)
	} else {
		statuCode, data = this.List(ctx)
	}
	return
}

//添加操作待用
func (this *WarrantiesView) Post(ctx iris.Context) (statuCode int, data interface{}) {
	return
}

//删除操作待用
func (this *WarrantiesView) Delete(ctx iris.Context) (statuCode int, data interface{}) {
	return
}
