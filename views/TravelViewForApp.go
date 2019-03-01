package views

import (
	"risk-ext/models"
	"strconv"

	"risk-ext/config"

	"github.com/kataras/iris"
)

type TravelViewForApp struct {
	Views
}

func (this *TravelViewForApp) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *TravelViewForApp) Delete(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.FormValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["msg"] = "参数deviceId缺失"
		data["data"] = nil
		return
	}
	var deviceModel = new(models.Devices)
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	deviceData, err := deviceModel.GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["msg"] = "设备不存在"
		data["data"] = nil
		return
	}
	if deviceData.DeviceUser.UserId != Session.Customer.UserId {
		data["code"] = 0
		data["msg"] = "你无权限操作该设备"
		data["data"] = nil
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
	data["msg"] = "OK"
	data["data"] = nil
	return
}

func (this *TravelViewForApp) Put(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	deviceId := ctx.PostValue("deviceId")
	if deviceId == "" {
		data["code"] = 0
		data["msg"] = "参数deviceId缺失"
		data["data"] = nil
		return
	}
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	deviceData, err := new(models.Devices).GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["msg"] = "设备不存在"
		data["data"] = nil
		return
	}
	if deviceData.DeviceUser.UserId != Session.Customer.UserId {
		data["code"] = 0
		data["msg"] = "你无权限操作该设备"
		data["data"] = nil
		return
	}
	travelName := ctx.PostValue("travelName")
	if travelName == "" {
		data["code"] = 0
		data["msg"] = "请输入名称"
		data["data"] = nil
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
		data["msg"] = "修改失败"
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = nil
	return
}

//获取车辆列表
func (this *TravelViewForApp) Get(ctx iris.Context) (statuCode int, data M) {
	userId := Session.Customer.UserId
	data = make(M)
	statuCode = 400
	data["code"] = 0
	data["data"] = nil
	Travels, err := new(models.Users).TravelList(userId.Hex())
	if err != nil {
		data["msg"] = err.Error()
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	data["data"] = Travels
	return
}
