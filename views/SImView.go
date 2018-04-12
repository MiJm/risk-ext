package views

import "github.com/kataras/iris"

type SimView struct {
	Views
}

func (this *SimView) Auth(ctx iris.Context) bool {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *SimView) Get(ctx iris.Context) {

}
