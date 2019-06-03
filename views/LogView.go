package views

import (
	"risk-ext/models"
	"strconv"
	"strings"

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
	companyId := ctx.FormValueDefault("com_id", "")
	typ := ctx.FormValue("type")
	if typ == "" {
		data["error"] = "类型不正确"
		return
	}
	typeArr := strings.Split(typ, ",")
	typeCondition := make([]int, 0)
	for _, v := range typeArr {
		t, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			t = 0
		}
		typeCondition = append(typeCondition, int(t))
	}

	logs := new(models.Logs)
	query := bson.M{}
	if bson.IsObjectIdHex(companyId) {
		query = bson.M{"log_company_id": companyId, "log_type": bson.M{"$in": typeCondition}}
	} else {
		query = bson.M{"log_type": bson.M{"$in": typeCondition}}
	}

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
	data["ai_amount_time"] = amount.QueryAiCarTime
	data["dianhua_amount_time"] = amount.QueryDianHuaTime
	data["credit_amount_time"] = amount.QueryCreditTime
	statuCode = 200
	return
}

//添加操作待用
func (this *LogView) Post(ctx iris.Context) (statuCode int, data M) {
	return
}

//更新操作待用
func (this *LogView) Put(ctx iris.Context) (statuCode int, data M) {
	return
}

//删除操作待用
func (this *LogView) Delete(ctx iris.Context) (statuCode int, data M) {
	return
}
