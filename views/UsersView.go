package views

import (
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
	return
}
