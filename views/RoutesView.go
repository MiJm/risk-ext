package views

import (
	"risk-ext/app"
	"risk-ext/models"
	"time"

	"github.com/kataras/iris"
)

type RoutesView struct {
	Views
}

func (this *RoutesView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *RoutesView) Post(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	defer func() {
		result = data
	}()
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "参数deviceId缺失"
		return
	}
	startTime := ctx.PostValueInt64Default("startTime", time.Now().Unix())
	endTime := ctx.PostValueInt64Default("endTime", time.Now().Unix()+86400)
	page := ctx.PostValueIntDefault("page", 1)
	pageSize := ctx.PostValueIntDefault("pageSize", 15)
	types := ctx.PostValueIntDefault("type", 0)
	routeList, count, err := new(models.Route).NewGetRoutesByPaging(deviceId, uint32(startTime), uint32(endTime), page, pageSize, types)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	statuCode = 200
	data["code"] = 1
	data["list"] = routeList
	data["count"] = count
	return
}

//获取详情或列表待用
func (this *RoutesView) Get(ctx iris.Context) (statuCode int, data interface{}) {
	return
}

//更新操作待用
func (this *RoutesView) Put(ctx iris.Context) (statuCode int, data interface{}) {
	return
}

//删除操作待用
func (this *RoutesView) Delete(ctx iris.Context) (statuCode int, data interface{}) {
	return
}
