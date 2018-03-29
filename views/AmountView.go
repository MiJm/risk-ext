package views

import (
	"github.com/kataras/iris"
    "risk-ext/models"
)

type UsersView struct {
}

func (this *UsersView) Put(ctx iris.Context) (statuCode int, data interface{}) {
	statuCode = 200
	company_id := ctx.FormValue("company_id")
	company_name := ctx.FormValue("company_name")
    type := ctx.FormValue("type") //0=追车，1=电话 2=违章
	changNum, err := ctx.PostValueInt("change_num")
	if err != nil {
		statuCode = 400
		data = "修改数量参数不正确"
		return
	}
    new(models.Amounts)
	data = iris.Map{"xx": d, "err": err}
	//	report := new(models.Reports)
	//	report.ReportId = bson.ObjectIdHex("5ab8e2eca9d63f1b7f69a568")
	//	report.One(report)
	//	report.ReportPlate = "xx"
	//	report.Update()
	//	data = *report
	return
}
