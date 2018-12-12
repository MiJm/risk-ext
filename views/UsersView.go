package views

import (
	"risk-ext/models"

	"github.com/kataras/iris"
)

type UsersView struct {
	Views
}

func (this *UsersView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *UsersView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	openId := ctx.FormValue("openId")
	if openId == "" {
		data["code"] = 0
		data["error"] = "openId参数缺失"
		return
	}
	userInfo, err := new(models.Users).GetUsersByOpenId(openId)
	statuCode = 200
	if err != nil {
		data["code"] = 0
		return
	}
	data["code"] = 1
	data["userInfo"] = userInfo
	return
}
