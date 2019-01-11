package views

import (
	"risk-ext/models"
	"strconv"

	"risk-ext/config"

	"github.com/kataras/iris"
)

type TravelView struct {
	Views
}

func (this *TravelView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *TravelView) Delete(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "参数deviceId缺失"
		return
	}
	var deviceModel = new(models.Devices)
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	deviceData, err := deviceModel.GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["error"] = "设备不存在"
		return
	}
	if deviceData.DeviceUser.UserId != Session.Customer.UserId {
		data["code"] = 0
		data["error"] = "你无权限操作该设备"
		return
	}
	userInfo, _ := new(models.Users).GetUsersByUserId(Session.Customer.UserId)
	var index = -1
	for key, travel := range userInfo.UserTravel {
		var id = travel.TravelDeviceId
		if travel.TravelDeviceId == 0 {
			id = travel.TravelDevice.DeviceId
		}
		if id == devId {
			index = key
			break
		}
	}
	userTravels := append(userInfo.UserTravel[:index], userInfo.UserTravel[index+1:]...)
	userInfo.UserTravel = userTravels
	err = userInfo.Update()
	if err == nil {
		var device models.Devices
		device.Device_id = deviceData.Device_id
		device.DeviceActivateTime = 1
		device.Update(false, "device_user")
		var deviceInfo models.DeviceInfo
		deviceModel.Map("devices", deviceId, &deviceInfo)
		deviceInfo.Device_activity_time = 0
		deviceInfo.DeviceUserId = ""
		deviceModel.Save("devices", deviceId, deviceInfo)
		config.Redis.HDel("pens_user_"+userInfo.UserId.Hex(), deviceId)
	}
	statuCode = 200
	data["code"] = 1
	return
}

func (this *TravelView) Put(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.PostValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "参数deviceId缺失"
		return
	}
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	deviceData, err := new(models.Devices).GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["error"] = "设备不存在"
		return
	}
	if deviceData.DeviceUser.UserId != Session.Customer.UserId {
		data["code"] = 0
		data["error"] = "你无权限操作该设备"
		return
	}
	travelName := ctx.PostValue("travelName")
	if travelName == "" {
		data["code"] = 0
		data["error"] = "请输入名称"
		return
	}
	travelType, _ := ctx.PostValueInt("travelType")
	userInfo, _ := new(models.Users).GetUsersByUserId(Session.Customer.UserId)
	var index = -1
	for key, travel := range userInfo.UserTravel {
		var id = travel.TravelDeviceId
		if travel.TravelDeviceId == 0 {
			id = travel.TravelDevice.DeviceId
		}
		if id == devId {
			index = key
			break
		}
	}
	userTravels := userInfo.UserTravel
	userTravels[index].TravelName = travelName
	userTravels[index].TravelType = uint8(travelType)
	userInfo.UserTravel = userTravels
	err = userInfo.Update()
	if err != nil {
		data["code"] = 0
		data["error"] = "修改失败"
		return
	}
	statuCode = 200
	data["code"] = 1
	return
}
