package views

import (
	"risk-ext/models"

	"github.com/kataras/iris"
)

type AlarmsViewForApp struct {
	Views
}

func (this *AlarmsViewForApp) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *AlarmsViewForApp) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	alarmId := ctx.FormValue("alarmId")
	if alarmId == "" {
		data["code"] = 0
		data["msg"] = "参数alarmId缺失"
		data["data"] = nil
		return
	}
	alarmInfo, err := new(models.Alarms).GetAlarmInfo(alarmId)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	if alarmInfo.Alarm_read == 0 {
		alarmInfo.Alarm_read = 1
		alarmInfo.Update()
	}
	statuCode = 200
	data["code"] = 1
	data["data"] = map[string]interface{}{"alarmInfo": alarmInfo}
	data["msg"] = "OK"
	return
}

func (this *AlarmsViewForApp) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	page := ctx.PostValueIntDefault("page", 1)
	pageSize := ctx.PostValueIntDefault("pageSize", 15)
	alarmList, count, err := new(models.Alarms).GetAlarmListByUserId(Session.Customer.UserId.Hex(), deviceId, page, pageSize)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	unReadAlarmNum, err := new(models.Alarms).GetUnReadAlarmNums(deviceId, Session.Customer.UserId.Hex())
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = map[string]interface{}{"count": count, "unreadCount": unReadAlarmNum, "alarmList": alarmList}
	return
}
