package views

import (
	"fmt"
	"risk-ext/models"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/kataras/iris"
)

type AmountView struct {
	Views
}

func (this *AmountView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{}},
		"GET": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{MEMBER_ADMIN}, "NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *AmountView) Put(ctx iris.Context) (statuCode int, data interface{}) {
	statuCode = 400
	companyId := ctx.FormValue("company_id")
	companyName := ctx.FormValue("company_name")
	changeType, err := ctx.PostValueInt("type") //0=追车，1=电话 2=违章
	if err != nil || changeType > 2 || changeType < 0 {
		data = "修改项目类型参数不正确"
		return
	}

	changNum, err := ctx.PostValueInt("change_num")
	if err != nil {
		data = "修改数量参数不正确"
		return
	}

	if !bson.IsObjectIdHex(companyId) {
		data = "企业ID不正确"
		return
	}

	if companyName == "" {
		data = "企业名不正确"
		return
	}
	err = Session.ChangeAmount(companyId, changNum)
	if err == nil {
		logs := new(models.Logs)
		logs.LogCompany = companyName
		logs.LogCompanyId = companyId
		logs.LogDate = time.Now().Unix()
		//logs.LogDetail = Session.Manager.Manager_fname + " 更改了企业：" + companyName + " 的"
		//logs.LogDetail += models.Extra[changeType] + "的查询次数。"
		logs.LogDetail = fmt.Sprintf("%d", changNum)
		logs.LogItem = models.Extra[changeType]
		logs.LogOperator = Session.Manager.Manager_fname
		logs.LogOperatorId = Session.Manager.Manager_id.Hex()
		logs.Insert()
		statuCode = 204
	} else {
		data = err.Error()
	}
	return
}
