package views

import (
	"risk-ext/models"
	"strconv"

	"github.com/kataras/iris"
)

type StayView struct {
	Views
}

func (this *StayView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *StayView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "参数deviceId缺失"
		return
	}
	var startTime, endTime, stayTime uint64
	var err error
	startTimeStr := ctx.FormValueDefault("startTime", "")
	if startTimeStr != "" {
		startTime, err = strconv.ParseUint(startTimeStr, 10, 32)
		if err != nil {
			data["code"] = 0
			data["error"] = "请选择正确的时间"
			return
		}
	}
	endTimeStr := ctx.FormValueDefault("endTime", "")
	if endTimeStr != "" {
		endTime, err = strconv.ParseUint(endTimeStr, 10, 32)
		if err != nil {
			data["code"] = 0
			data["error"] = "请选择正确的时间"
			return
		}
	}
	stayTimeStr := ctx.FormValueDefault("stayTime", "")
	if stayTimeStr != "" {
		stayTime, err = strconv.ParseUint(stayTimeStr, 10, 32)
		if err != nil || stayTime < 30*60 {
			data["code"] = 0
			data["error"] = "请选择正确的停留时间"
			return
		}
	}
	stayList, err := new(models.Route).GetStayList(uint32(startTime), uint32(endTime), uint32(stayTime), deviceId)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	statuCode = 200
	data["code"] = 1
	data["list"] = stayList
	return
}
