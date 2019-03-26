package views

import (
	"fmt"
	"risk-ext/config"
	"risk-ext/models"
	"time"

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
func (this *WarrantiesView) Put(ctx iris.Context) (statuCode int, data M) {
	step := ctx.Params().Get("step") //分步提交第几步
	if step == "2" {
		statuCode, data = this.AddOwnerInfo(ctx)
	} else {
		statuCode, data = this.AddCarInfo(ctx)
	}
	return
}

//添加保单投保人信息(第二步)
func (this *WarrantiesView) AddOwnerInfo(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	user := Session.Customer
	id := ctx.FormValue("id")
	rs, err := new(models.Warranty).One(user.UserId.Hex(), id, []uint8{0})
	if err != nil {
		data["code"] = 0
		data["msg"] = "未查到该保单信息,请核实后再填写"
		data["data"] = nil
		return
	}
	if rs.WarrantyStatus != 0 {
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

	if front != "" {
		frontTitle := fmt.Sprintf("%s_front.png", idcard)
		frontSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/"
		frontOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/"
		frontThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb/"
		frontThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/thumb/"
		err = models.SaveImg(front, frontSaveUrl, frontThumbUrl, frontTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传身份证正面照片失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyOwnerInfo.OwnerIDcardFront = frontOpenUrl + frontTitle
		rs.WarrantyOwnerInfo.OwnerThumbIDcardFront = frontThumbOpenUrl + frontTitle
		rs.WarrantyOwnerInfo.OwnerMobile = phone
	} else {
		if rs.WarrantyOwnerInfo.OwnerIDcardFront == "" {
			data["code"] = 0
			data["msg"] = "请上传身份证正面照"
			data["data"] = nil
			return
		}
	}

	//身份证背面
	back := ctx.FormValue("back")

	if back != "" {
		backTitle := fmt.Sprintf("%s_back.png", idcard)
		backSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/"
		backOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/"

		backThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb/"
		backThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/thumb/"
		err = models.SaveImg(back, backSaveUrl, backThumbUrl, backTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传身份证背面照片失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyOwnerInfo.OwnerIDcardBack = backOpenUrl + backTitle
		rs.WarrantyOwnerInfo.OwnerThumbIDcardBack = backThumbOpenUrl + backTitle
	} else {
		if rs.WarrantyOwnerInfo.OwnerIDcardBack == "" {
			data["code"] = 0
			data["msg"] = "请上传身份证背面照"
			data["data"] = nil
			return
		}
	}

	//手持身份证
	ownerIDcard := ctx.FormValue("owner_idcard")

	if ownerIDcard != "" {
		ownerTitle := fmt.Sprintf("%s_owner.png", idcard)
		ownerSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/"
		ownerOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/"

		ownerThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb/"
		ownerThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/thumb/"
		err = models.SaveImg(ownerIDcard, ownerSaveUrl, ownerThumbUrl, ownerTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传手持身份证照片失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyOwnerInfo.OwnerIDcardImg = ownerOpenUrl + ownerTitle
		rs.WarrantyOwnerInfo.OwnerThumbIDcardImg = ownerThumbOpenUrl + ownerTitle

	} else {
		if rs.WarrantyOwnerInfo.OwnerIDcardImg == "" {
			data["code"] = 0
			data["msg"] = "请上传手持身份证照"
			data["data"] = nil
			return
		}
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
func (this *WarrantiesView) AddCarInfo(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	user := Session.Customer
	id := ctx.FormValue("id")
	rs, err := new(models.Warranty).One(user.UserId.Hex(), id, []uint8{0})
	if err != nil {
		data["code"] = 0
		data["msg"] = "未查到该保单信息,请核实后再填写"
		data["data"] = nil
		return
	}
	if rs.WarrantyStatus != 0 {
		data["code"] = 0
		data["msg"] = "该保单已填写完整信息无法再次提交"
		data["data"] = nil
		return
	}

	idcard := rs.WarrantyOwnerInfo.OwnerIDcard
	brand := ctx.FormValueDefault("brand", "")           //保单车辆品牌
	series := ctx.FormValueDefault("series", "")         //保单车辆型号
	vin := ctx.FormValueDefault("vin", "")               //保单车辆车架号
	purchase := ctx.PostValueInt64Default("purchase", 0) //保单车辆购买日期
	value := ctx.PostValueFloat64Default("value", 0.0)   //保单车辆购买时发票金额
	if brand == "" || series == "" || vin == "" || purchase == 0 || value == 0.0 {
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
	rs.WarrantyCarModel = carModel
	ctx.SetMaxRequestBodySize(2 << 31)

	//车辆正面照片
	front := ctx.FormValue("front")

	if front != "" {
		frontTitle := fmt.Sprintf("%s_front.png", vin)
		frontSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		frontOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		frontThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		frontThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		err = models.SaveImg(front, frontSaveUrl, frontThumbUrl, frontTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆正面文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarFrontImg = frontOpenUrl + frontTitle
		rs.WarrantyCarModel.CarThumbFrontImg = frontThumbOpenUrl + frontTitle
	} else {
		if rs.WarrantyCarModel.CarFrontImg == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆正面照"
			data["data"] = nil
			return
		}
	}
	//车辆侧面
	side := ctx.FormValue("side")

	if side != "" {
		sideTitle := fmt.Sprintf("%s_side.png", vin)
		sideSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		sideOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		sideThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		sideThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		err = models.SaveImg(side, sideSaveUrl, sideThumbUrl, sideTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆侧面文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarSideImg = sideOpenUrl + sideTitle
		rs.WarrantyCarModel.CarThumbSideImg = sideThumbOpenUrl + sideTitle
	} else {
		if rs.WarrantyCarModel.CarSideImg == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆侧面照"
			data["data"] = nil
			return
		}
	}

	//车辆合格证
	certificate := ctx.FormValue("certificate")
	if certificate != "" {
		certifyTitle := fmt.Sprintf("%s_certificate.png", vin)
		certifySaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		certifyOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		certifyThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		certifyThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		err = models.SaveImg(certificate, certifySaveUrl, certifyThumbUrl, certifyTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆合格证文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarCertificateImg = certifyOpenUrl + certifyTitle
		rs.WarrantyCarModel.CarThumbCertificateImg = certifyThumbOpenUrl + certifyTitle
	} else {
		if rs.WarrantyCarModel.CarCertificateImg == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆合格证照"
			data["data"] = nil
			return
		}
	}

	//车辆购车发票
	receipt := ctx.FormValue("receipt")

	if receipt != "" {
		receiptTitle := fmt.Sprintf("%s_receipt.png", vin)
		receiptSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		receiptOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
		receiptThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		receiptThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb/"
		err = models.SaveImg(receipt, receiptSaveUrl, receiptThumbUrl, receiptTitle)
		if err != nil {
			data["code"] = 0
			data["msg"] = fmt.Sprintf("上传车辆购车发票文件失败(%s)", err.Error())
			data["data"] = nil
			return
		}
		rs.WarrantyCarModel.CarReceiptImg = receiptOpenUrl + receiptTitle
		rs.WarrantyCarModel.CarThumbReceiptImg = receiptThumbOpenUrl + receiptTitle
	} else {
		if rs.WarrantyCarModel.CarCertificateImg == "" {
			data["code"] = 0
			data["msg"] = "请上传车辆购车发票照片"
			data["data"] = nil
			return
		}
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
func (this *WarrantiesView) Detail(ctx iris.Context, id string) (statuCode int, data M) {
	statuCode = 400
	data = make(M)
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
func (this *WarrantiesView) List(ctx iris.Context) (statuCode int, data M) {
	statuCode = 400
	data = make(M)
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
func (this *WarrantiesView) GetDisActiveList(ctx iris.Context) (statuCode int, data M) {
	statuCode = 400
	data = make(M)
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
func (this *WarrantiesView) Get(ctx iris.Context) (statuCode int, data M) {
	statuCode = 400
	data = make(M)
	id := ctx.Params().Get("id")
	if id != "" {
		statuCode, data = this.Detail(ctx, id)
	} else {
		statuCode, data = this.List(ctx)
	}
	return
}
