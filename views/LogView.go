package views

import (
	"risk-ext/models"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	"github.com/kataras/iris"
)

type LogView struct {
	Views
}

func (this *LogView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{}},
		"GET": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//日志列表
func (this *LogView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	page := ctx.FormValue("page")
	size := ctx.FormValue("size")
	p, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		p = 1
	}
	s, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		s = 30
	}
	companyId := ctx.FormValue("com_id")
	if !bson.IsObjectIdHex(companyId) {
		data["error"] = "企业ID不正确"
		return
	}
	typ := ctx.FormValue("type")
	if typ == "" {
		data["error"] = "类型不正确"
		return
	}
	t, err := strconv.ParseInt(typ, 10, 64)
	if err != nil {
		t = 0
	}
	logs := new(models.Logs)
	query := bson.M{"log_company_id": companyId, "log_type": t}
	rs, num, err := logs.List(query, int(p), int(s))
	if err != nil {
		data["list"] = "无数据"
		statuCode = 400
		return
	}
	if len(rs) == 0 {
		rs = []*models.Logs{}
	}
	amount := models.Amounts{}
	Session.Amount(companyId, &amount)
	data = make(M)
	data["list"] = rs
	data["num"] = num
	data["ai_amount"] = amount.QueryAiCar
	data["dianhua_amount"] = amount.QueryDianHua
	data["credit_amount"] = amount.QueryCredit
	statuCode = 200
	return
}
