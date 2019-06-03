package views

import (
	"risk-ext/app"
	"risk-ext/models"
	"strconv"

	"github.com/kataras/iris"
)

type StayViewForApp struct {
	Views
}

func (this *StayViewForApp) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *StayViewForApp) Get(ctx iris.Context) (statuCode int, data app.M) {
	data = make(app.M)
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["msg"] = "参数deviceId缺失"
		data["data"] = nil
		return
	}
	var startTime, endTime, stayTime uint64
	var err error
	startTimeStr := ctx.FormValueDefault("startTime", "")
	if startTimeStr != "" {
		startTime, err = strconv.ParseUint(startTimeStr, 10, 32)
		if err != nil {
			data["code"] = 0
			data["msg"] = "请选择正确的时间"
			data["data"] = nil
			return
		}
	}
	endTimeStr := ctx.FormValueDefault("endTime", "")
	if endTimeStr != "" {
		endTime, err = strconv.ParseUint(endTimeStr, 10, 32)
		if err != nil {
			data["code"] = 0
			data["msg"] = "请选择正确的时间"
			data["data"] = nil
			return
		}
	}
	stayTimeStr := ctx.FormValueDefault("stayTime", "")
	if stayTimeStr != "" {
		stayTime, err = strconv.ParseUint(stayTimeStr, 10, 32)
		if err != nil || stayTime < 30*60 {
			data["code"] = 0
			data["msg"] = "请选择正确的停留时间"
			data["data"] = nil
			return
		}
	} else {
		data["code"] = 0
		data["msg"] = "请选择停留时间"
		data["data"] = nil
		return
	}
	stayList, err := new(models.Route).GetStayList(uint32(startTime), uint32(endTime), uint32(stayTime), deviceId)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = stayList
	return
}

//添加操作待用
func (this *StayViewForApp) Post(ctx iris.Context) (statuCode int, data app.M) {
	return
}

//更新操作待用
func (this *StayViewForApp) Put(ctx iris.Context) (statuCode int, data app.M) {
	return
}

//删除操作待用
func (this *StayViewForApp) Delete(ctx iris.Context) (statuCode int, data app.M) {
	return
}
