package views

import (
	"risk-ext/models"

	"github.com/kataras/iris"
)

type CommonsView struct {
	Views
}

func (this *CommonsView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"GET": MA{"NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *CommonsView) Get(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	mobile := ctx.FormValue("mobile")
	flg := models.CheckPhone(mobile)
	if !flg {
		data["error"] = "手机号码有误"
		data["code"] = 0
		return
	}
	reportId := ctx.Params().Get("report_id")
	res, err := new(models.Reports).CheckPhone(mobile, reportId)
	if err != nil || res == nil {
		data["error"] = "该手机号无授权"
		data["code"] = 0
		return
	}
	codes := this.GetCode(mobile)
	rs := this.SendMsg(mobile, "报表授权验证码:"+codes)
	if rs > 0 {
		data["code"] = 1
		statusCode = 200
		return
	} else {
		data["code"] = 0
		data["error"] = "验证码发送失败"
		return
	}
}
