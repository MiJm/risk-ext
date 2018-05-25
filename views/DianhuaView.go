package views

import (
	"risk-ext/models"
	"risk-ext/utils"

	"gopkg.in/mgo.v2/bson"

	"github.com/kataras/iris"
)

type DianhuaView struct {
	Views
}

func (this *DianhuaView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//获取登录所需输入信息
func (this *DianhuaView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	tel := ctx.FormValue("tel")
	if tel == "" {
		data["code"] = 0
		data["error"] = "请输入完整手机号"
		return
	}
	res, err := models.GetLoginMethod(tel)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}

	if res.Status != 0 {
		data["code"] = 0
		data["error"] = res.Msg
		return
	}
	statuCode = 200
	data["code"] = 1
	data["result"] = res.Data
	return
}

//登录电话邦
func (this *DianhuaView) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	amount := Session.User.Amount.QueryDianHua
	comId := Session.User.UserCompany_id
	if amount <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}
	tel := ctx.FormValue("tel")
	sid := ctx.FormValue("sid")
	fname := ctx.FormValue("fname")
	idcard := ctx.FormValue("idcard")
	pwd := ctx.FormValue("pwd")
	sms := ctx.FormValue("sms")
	captcha := ctx.FormValue("captcha")
	t := ctx.FormValueDefault("type", "1")

	//二次校验
	if t != "1" {
		result, err := models.LoginVerify(sid, sms, captcha)
		if err != nil {
			data["code"] = 0
			data["error"] = err.Error()
			return
		}
		data["data"] = nil
		if result.Status == 0 {
			if result.CommonData.Action == "processing" {
				statuCode = 200
				data["code"] = 1
				data["data"] = result.Data
				return
			}
		} else {
			data["code"] = 0
			data["error"] = result.Msg
			return
		}
	} else {

		res, err := models.Login(sid, tel, pwd, fname, idcard, sms, captcha)
		if err != nil {
			data["code"] = 0
			data["error"] = err.Error()
			return
		}
		data["data"] = nil
		if res.Status == 0 {
			if res.CommonData.Action == "processing" {
				statuCode = 200
				data["code"] = 1
				data["data"] = res.Data
				return
			}
		} else {
			data["code"] = 0
			data["error"] = res.Msg
			return
		}
	}
	report := new(models.Reports)
	report.ReportType = 1
	report.ReportCompanyId = comId
	report.ReportMobile = tel
	report.ReportName = fname
	str := bson.NewObjectId().Hex()
	numStr := utils.SubString(str, 12)
	report.ReportNumber = numStr
	err := report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	task := models.Task{}
	reportId := report.ReportId.Hex()
	task.ReportId = reportId
	task.CompanyId = comId
	task.Type = int8(1)
	task.Sid = sid

	err = new(models.Redis).ListPush("analysis_tasks", task)
	if err != nil {
		data["error"] = "建立任务失败"
		data["code"] = 0
		return
	}
	amount--
	am := Session.User.Amount
	am.QueryDianHua = amount
	new(models.Redis).Save("amounts", comId, am)
	data["code"] = 1
	statuCode = 200
	return
	return
}
