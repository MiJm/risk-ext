package views

import (
	"risk-ext/models"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type UsersView struct {
}

func (this *UsersView) Get(ctx iris.Context) (statuCode int, data interface{}) {
	statuCode = 200
	data = iris.Map{"xx": "错错错"}
	report := new(models.Reports)
	report.ReportId = bson.ObjectIdHex("5ab8e2eca9d63f1b7f69a568")
	report.One(report)
	report.ReportPlate = "xx"
	report.Update()
	data = *report
	return
}
