package views

import (
	"risk-ext/models"

	"gopkg.in/mgo.v2/bson"

	"github.com/kataras/iris"
)

type LogView struct {
	Views
}

func (this *LogView) Auth(ctx iris.Context) bool {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{}},
		"GET": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{MEMBER_SUPER, MEMBER_ADMIN}, "NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//日志列表
func (this *LogView) Get(ctx iris.Context) (statuCode int, data map[string]interface{}) {
	statuCode = 400
	page, err := ctx.PostValueInt("page")
	if err != nil {
		page = 1
	}
	pageSize, err := ctx.PostValueInt("size")
	if err != nil {
		pageSize = 30
	}
	logs := new(models.Logs)
	rs, num, err := logs.List(bson.M{}, page, pageSize)
	if err != nil {
		statuCode = 400
		return
	}
	data["num"] = num
	data["list"] = rs
	statuCode = 200
	return
}
