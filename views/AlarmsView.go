package views

import (
	"risk-ext/app"
	"risk-ext/models"

	"github.com/kataras/iris"
)

type AlarmsView struct {
	Views
}

func (this *AlarmsView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *AlarmsView) Get(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	statuCode = 400
	alarmId := ctx.FormValue("alarmId")
	if alarmId == "" {
		data["code"] = 0
		data["error"] = "参数alarmId缺失"
		result = data
		return
	}
	alarmInfo, err := new(models.Alarms).GetAlarmInfo(alarmId)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		result = data
		return
	}
	if alarmInfo.Alarm_read == 0 {
		alarmInfo.Alarm_read = 1
		alarmInfo.Update()
	}
	statuCode = 200
	data["code"] = 1
	data["alarmInfo"] = alarmInfo
	result = data
	return
}

func (this *AlarmsView) Post(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	page := ctx.PostValueIntDefault("page", 1)
	pageSize := ctx.PostValueIntDefault("pageSize", 15)
	alarmList, count, err := new(models.Alarms).GetAlarmListByUserId(Session.Customer.UserId.Hex(), deviceId, page, pageSize)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		result = data
		return
	}
	unReadAlarmNum, err := new(models.Alarms).GetUnReadAlarmNums(deviceId, Session.Customer.UserId.Hex())
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		result = data
		return
	}
	statuCode = 200
	data["code"] = 1
	data["count"] = count
	data["unreadCount"] = unReadAlarmNum
	data["alarmList"] = alarmList
	result = data
	return
}

//更新操作
func (this *AlarmsView) Put(ctx iris.Context) (statuCode int, data interface{}) {
	return
}

//删除操作待用
func (this *AlarmsView) Delete(ctx iris.Context) (statuCode int, data interface{}) {
	return
}
