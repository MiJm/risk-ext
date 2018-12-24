package views

import (
	"net/url"
	"risk-ext/models"
	"risk-ext/utils"
	"strconv"
	"time"

	"github.com/kataras/iris"
)

type DevicesView struct {
	Views
}

func (this *DevicesView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *DevicesView) Get(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	qrcodeStr := ctx.Params().Get("qrcodeStr")
	qrcodeStr, _ = url.QueryUnescape(qrcodeStr)
	deviceId := ctx.FormValue("deviceId")
	var userData = Session.Customer
	deviceModel := new(models.Devices)
	if qrcodeStr != "" {
		deviceId, err := utils.AesDecode(qrcodeStr)
		if err != nil {
			data["code"] = 0
			data["error"] = "无效的二维码"
			return
		}
		devId, _ := strconv.ParseUint(deviceId, 10, 64)
		deviceData, err := deviceModel.GetDeviceByDevId(devId)
		if err != nil {
			data["code"] = 0
			data["error"] = "该设备不存在"
			return
		}
		if deviceData.DeviceOutType != 2 {
			data["code"] = 0
			data["error"] = "该设备未出库"
			return
		}
		if deviceData.DeviceUser != nil {
			if deviceData.DeviceUser.UserId != models.EmptyId {
				data["code"] = 0
				data["error"] = "该设备已激活"
				return
			}
		}
		statuCode = 200
		data["code"] = 1
		data["deviceId"] = deviceId
		return
	}
	if deviceId == "" {
		data["code"] = 0
		data["error"] = "deviceId参数缺失"
		return
	}
	devId, err := strconv.ParseUint(deviceId, 10, 64)
	if err != nil {
		data["code"] = 0
		data["error"] = "无效的deviceId"
		return
	}

	deviceData, err := deviceModel.GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["error"] = "该设备不存在"
		return
	}
	if deviceData.DeviceUser.UserId != userData.UserId {
		data["code"] = 0
		data["error"] = "该设备您无权限查看"
		return
	}
	deviceInfo := deviceModel.GetDeviceInfo(devId)
	deviceData.Device_info = deviceInfo
	statuCode = 200
	data["code"] = 1
	data["device"] = deviceData
	return
}

func (this *DevicesView) Put(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	userModel := new(models.Users)
	var userData = Session.Customer
	qrcodeStr := ctx.FormValue("qrcodeStr")
	qrcodeStr, err := url.QueryUnescape(qrcodeStr)
	//deviceId := ctx.FormValue("deviceId")
	if qrcodeStr == "" || err != nil {
		data["code"] = 0
		data["error"] = "二维码失效"
		return
	}
	deviceId, err := utils.AesDecode(qrcodeStr)
	if err != nil {
		data["code"] = 0
		data["error"] = "无效的二维码"
		return
	}
	travelName := ctx.FormValue("travelName")
	if travelName == "" {
		data["code"] = 0
		data["error"] = "请输入名称"
		return
	}
	travelType, _ := ctx.PostValueInt("travelType")
	userInfo, err := userModel.GetUsersByOpenId(userData.UserOpenId)
	if err != nil {
		data["code"] = 0
		data["error"] = "用户已被注销"
		return
	}
	var travelInfo models.Travel
	var userTravel []models.Travel
	travelInfo.TravelName = travelName
	travelInfo.TravelType = uint8(travelType)
	travelInfo.TravelDate = time.Now().Unix()
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	device := new(models.Devices)
	deviceData, err := device.GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["error"] = "该设备不存在"
		return
	}
	if deviceData.DeviceOutType != 2 {
		data["code"] = 0
		data["error"] = "该设备未出库"
		return
	}
	if deviceData.DeviceUser != nil {
		if deviceData.DeviceUser.UserId != models.EmptyId {
			data["code"] = 0
			data["error"] = "该设备已激活"
			return
		}
	}
	travelInfo.TravelDeviceId = devId
	var travels = []models.Travel{travelInfo}
	//userTravel = append(userInfo.UserTravel, travelInfo)
	userTravel = append(travels, userInfo.UserTravel...)
	userInfo.UserTravel = userTravel
	err = userInfo.Update()
	if err != nil {
		data["code"] = 0
		data["error"] = "激活失败"
		return
	}
	device.Device_id = deviceData.Device_id
	device.DeviceUser = &userInfo
	device.DeviceActivateTime = uint32(time.Now().Unix())
	err = device.Update(false)
	if err != nil {
		data["code"] = 0
		data["error"] = "激活失败"
		return
	}
	var deviceInfo models.DeviceInfo
	err = userModel.Map("devices", deviceId, &deviceInfo)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	deviceInfo.Device_activity_time = uint32(travelInfo.TravelDate)
	deviceInfo.DeviceUserId = userInfo.UserId.Hex()
	userModel.Save("devices", deviceId, deviceInfo)
	statuCode = 200
	data["code"] = 1
	return
}
