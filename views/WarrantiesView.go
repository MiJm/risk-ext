package views

import (
	"fmt"
	"io"
	"os"
	"risk-ext/config"
	"risk-ext/models"
	"risk-ext/utils"
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
	if step == "1" {
		statuCode, data = this.AddOwnerInfo(ctx)
	} else {
		statuCode, data = this.AddCarInfo(ctx)
	}
	return
}

//添加保单投保人信息(第一步)
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

	name := ctx.FormValue("name")
	idcard := ctx.FormValueDefault("idcard", "")
	if idcard == "" || name == "" {
		data["code"] = 0
		data["msg"] = "请填写完整投保人信息"
		data["data"] = nil
		return
	}

	ctx.SetMaxRequestBodySize(2 << 31)

	//身份证正面
	front, _, err := ctx.FormFile("front")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传身份证正面文件失败"
		data["data"] = nil
		return
	}
	defer front.Close()
	frontTitle := fmt.Sprintf("%s_front.png", idcard)
	frontSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/"
	frontOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/"
	err = utils.IsFile(frontSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	f, err := os.OpenFile(frontSaveUrl+frontTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer f.Close()
	_, err = io.Copy(f, front)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	frontThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb"
	frontThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/thumb"
	err = utils.IsFile(frontThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.ImageResize(frontSaveUrl+frontTitle, frontThumbUrl+frontTitle)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}
	rs.WarrantyOwnerInfo.OwnerIDcardFront = frontOpenUrl + frontTitle
	rs.WarrantyOwnerInfo.OwnerThumbIDcardFront = frontThumbOpenUrl + frontTitle

	//身份证背面
	back, _, err := ctx.FormFile("back")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传身份证背面文件失败"
		data["data"] = nil
		return
	}
	defer back.Close()
	backTitle := fmt.Sprintf("%s_back.png", idcard)
	backSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/"
	backOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/"
	err = utils.IsFile(backSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	f1, err := os.OpenFile(backSaveUrl+backTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer f1.Close()
	_, err = io.Copy(f1, back)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	backThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb"
	backThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/thumb"
	err = utils.IsFile(backThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.ImageResize(backSaveUrl+backTitle, backThumbUrl+backTitle)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}
	rs.WarrantyOwnerInfo.OwnerIDcardBack = backOpenUrl + backTitle
	rs.WarrantyOwnerInfo.OwnerThumbIDcardBack = backThumbOpenUrl + backTitle

	//手持身份证
	ownerIDcard, _, err := ctx.FormFile("owner_idcard")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传手持身份证文件失败"
		data["data"] = nil
		return
	}
	defer ownerIDcard.Close()
	ownerTitle := fmt.Sprintf("%s_owner.png", idcard)
	ownerSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/"
	ownerOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/"
	err = utils.IsFile(ownerSaveUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	f2, err := os.OpenFile(ownerSaveUrl+ownerTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	defer f2.Close()
	_, err = io.Copy(f2, ownerIDcard)
	if err != nil {
		data["code"] = 0
		data["msg"] = "保存文件失败"
		data["data"] = nil
		return
	}
	ownerThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb"
	ownerThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/thumb"
	err = utils.IsFile(ownerThumbUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	err = utils.ImageResize(ownerSaveUrl+ownerTitle, ownerThumbUrl+ownerTitle)
	if err != nil {
		data["code"] = 0
		data["msg"] = "生成文件失败"
		data["data"] = nil
		return
	}
	rs.WarrantyOwnerInfo.OwnerIDcard = idcard
	rs.WarrantyOwnerInfo.OwnerName = name
	rs.WarrantyOwnerInfo.OwnerIDcardImg = ownerOpenUrl + ownerTitle
	rs.WarrantyOwnerInfo.OwnerThumbIDcardImg = ownerThumbOpenUrl + ownerTitle
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

	ctx.SetMaxRequestBodySize(2 << 31)

	//车辆正面照片
	front, _, err := ctx.FormFile("front")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传车辆正面文件失败"
		data["data"] = nil
		return
	}
	defer front.Close()
	frontTitle := fmt.Sprintf("%s_front.png", vin)
	frontSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	frontOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	err = utils.IsFile(frontSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	f, err := os.OpenFile(frontSaveUrl+frontTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer f.Close()
	_, err = io.Copy(f, front)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	frontThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	frontThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	err = utils.IsFile(frontThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.ImageResize(frontSaveUrl+frontTitle, frontThumbUrl+frontTitle)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}
	rs.WarrantyCarModel.CarFrontImg = frontOpenUrl + frontTitle
	rs.WarrantyCarModel.CarThumbFrontImg = frontThumbOpenUrl + frontTitle

	//车辆侧面
	side, _, err := ctx.FormFile("side")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传车辆侧面文件失败"
		data["data"] = nil
		return
	}
	defer side.Close()
	sideTitle := fmt.Sprintf("%s_side.png", vin)
	sideSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	sideOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	err = utils.IsFile(sideSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	f1, err := os.OpenFile(sideSaveUrl+sideTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer f1.Close()
	_, err = io.Copy(f1, side)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	sideThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	sideThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	err = utils.IsFile(sideThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.ImageResize(sideSaveUrl+sideTitle, sideThumbUrl+sideTitle)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}
	rs.WarrantyCarModel.CarSideImg = sideOpenUrl + sideTitle
	rs.WarrantyCarModel.CarThumbSideImg = sideThumbOpenUrl + sideTitle

	//车辆合格证
	certificate, _, err := ctx.FormFile("certificate")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传车辆合格证失败"
		data["data"] = nil
		return
	}
	defer certificate.Close()
	certifyTitle := fmt.Sprintf("%s_certificate.png", vin)
	certifySaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	certifyOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	err = utils.IsFile(certifySaveUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	f2, err := os.OpenFile(certifySaveUrl+certifyTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	defer f2.Close()
	_, err = io.Copy(f2, certificate)
	if err != nil {
		data["code"] = 0
		data["msg"] = "保存文件失败"
		data["data"] = nil
		return
	}
	certifyThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	certifyThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	err = utils.IsFile(certifyThumbUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	err = utils.ImageResize(certifySaveUrl+certifyTitle, certifyThumbUrl+certifyTitle)
	if err != nil {
		data["code"] = 0
		data["msg"] = "生成文件失败"
		data["data"] = nil
		return
	}
	rs.WarrantyCarModel.CarCertificateImg = certifyOpenUrl + certifyTitle
	rs.WarrantyCarModel.CarThumbCertificateImg = certifyThumbOpenUrl + certifyTitle

	//车辆购车发票
	receipt, _, err := ctx.FormFile("receipt")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传车辆购车发票失败"
		data["data"] = nil
		return
	}
	defer receipt.Close()
	receiptTitle := fmt.Sprintf("%s_receipt.png", vin)
	receiptSaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	receiptOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/"
	err = utils.IsFile(receiptSaveUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	f3, err := os.OpenFile(receiptSaveUrl+receiptTitle, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	defer f3.Close()
	_, err = io.Copy(f3, receipt)
	if err != nil {
		data["code"] = 0
		data["msg"] = "保存文件失败"
		data["data"] = nil
		return
	}
	receiptThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	receiptThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/" + vin + "/thumb"
	err = utils.IsFile(receiptThumbUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	err = utils.ImageResize(receiptSaveUrl+receiptTitle, receiptThumbUrl+receiptTitle)
	if err != nil {
		data["code"] = 0
		data["msg"] = "生成文件失败"
		data["data"] = nil
		return
	}
	rs.WarrantyCarModel.CarReceiptImg = receiptOpenUrl + receiptTitle
	rs.WarrantyCarModel.CarThumbReceiptImg = receiptThumbOpenUrl + receiptTitle
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
	rs, err := new(models.Warranty).ListByUserId(Session.Customer.UserId.Hex(), []uint8{1, 2})
	if err != nil {
		rs = make([]models.Warranty, 0)
	}
	rs1, err := new(models.Warranty).ListByUserId(Session.Customer.UserId.Hex(), []uint8{0})
	if err != nil {
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
	rs, err := new(models.Warranty).ListByUserId(Session.Customer.UserId.Hex(), []uint8{0})
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
