package views

import (
	"fmt"
	"net/url"
	"risk-ext/models"
	"risk-ext/utils"
	"strconv"
	"time"

	"github.com/kataras/iris"
)

type DevicesViewForApp struct {
	Views
}

func (this *DevicesViewForApp) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *DevicesViewForApp) Get(ctx iris.Context) (statuCode int, data M) {
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
			data["msg"] = "无效的二维码"
			data["data"] = nil
			return
		}
		devId, _ := strconv.ParseUint(deviceId, 10, 64)
		deviceData, err := deviceModel.GetDeviceByDevId(devId)
		if err != nil {
			data["code"] = 0
			data["msg"] = "该设备不存在"
			data["data"] = nil
			return
		}
		if deviceData.DeviceOutType != 2 {
			data["code"] = 0
			data["msg"] = "该设备未出库"
			data["data"] = nil
			return
		}
		if deviceData.DeviceUser != nil {
			if deviceData.DeviceUser.UserId != models.EmptyId {
				data["code"] = 0
				data["msg"] = "该设备已激活"
				data["data"] = nil
				return
			}
		}
		statuCode = 200
		data["code"] = 1
		result := make(M)
		result["deviceId"] = deviceId
		data["data"] = result
		data["msg"] = "OK"
		return
	}
	if deviceId == "" {
		data["code"] = 0
		data["msg"] = "deviceId参数缺失"
		data["data"] = nil
		return
	}
	devId, err := strconv.ParseUint(deviceId, 10, 64)
	if err != nil {
		data["code"] = 0
		data["msg"] = "无效的deviceId"
		data["data"] = nil
		return
	}

	deviceData, err := deviceModel.GetDeviceByDevId(devId)
	if err != nil {
		data["code"] = 0
		data["msg"] = "该设备不存在"
		data["data"] = nil
		return
	}
	if deviceData.DeviceUser.UserId != userData.UserId {
		data["code"] = 0
		data["msg"] = "该设备您无权限查看"
		data["data"] = nil
		return
	}
	deviceInfo := deviceModel.GetDeviceInfo(devId)
	deviceData.Device_info = deviceInfo
	userInfo, _ := new(models.Users).GetUsersByUserId(Session.Customer.UserId)
	var deviceTravel models.Travel
	for _, travel := range userInfo.UserTravel {
		var id = travel.TravelDeviceId
		if travel.TravelDeviceId == 0 {
			id = travel.TravelDevice.DeviceId
		}
		if id == devId {
			deviceTravel = travel
		}
	}
	statuCode = 200
	data["code"] = 1
	data["data"] = map[string]interface{}{"device": deviceData, "travelInfo": deviceTravel}
	data["msg"] = "OK"
	return
}

func (this *DevicesViewForApp) Put(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	statuCode = 400
	userModel := new(models.Users)
	var userData = Session.Customer
	qrcodeStr := ctx.FormValue("qrcodeStr")
	qrcodeStr, err := url.QueryUnescape(qrcodeStr)
	//deviceId := ctx.FormValue("deviceId")
	if qrcodeStr == "" || err != nil {
		data["code"] = 0
		data["msg"] = "二维码失效"
		data["data"] = nil
		return
	}
	deviceId, err := utils.AesDecode(qrcodeStr)
	if err != nil {
		data["code"] = 0
		data["msg"] = "无效的二维码"
		data["data"] = nil
		return
	}
	travelName := ctx.FormValue("travelName")
	if travelName == "" {
		data["code"] = 0
		data["msg"] = "请输入名称"
		data["data"] = nil
		return
	}
	travelType, _ := ctx.PostValueInt("travelType")
	userInfo, err := userModel.GetUsersByUnionId(userData.UserUnionId)
	if err != nil {
		data["code"] = 0
		data["msg"] = "用户已被注销"
		data["data"] = nil
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
		data["msg"] = "该设备不存在"
		data["data"] = nil
		return
	}
	if deviceData.DeviceOutType != 2 {
		data["code"] = 0
		data["msg"] = "该设备未出库"
		data["data"] = nil
		return
	}
	if deviceData.DeviceUser != nil {
		if deviceData.DeviceUser.UserId != models.EmptyId {
			data["code"] = 0
			data["msg"] = "该设备已激活"
			data["data"] = nil
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
		data["msg"] = "激活失败"
		data["data"] = nil
		return
	}
	device.Device_id = deviceData.Device_id
	device.DeviceUser = &userInfo
	device.DeviceActivateTime = uint32(time.Now().Unix())
	err = device.Update(false)
	if err != nil {
		data["code"] = 0
		data["msg"] = "激活失败"
		data["data"] = nil
		return
	}
	var warrnty = new(models.Warranty)
	rs, err := warrnty.GetWarrantyByDeviceId(deviceData.Device_id)
	if err != nil || rs.WarrantyId == models.EmptyId { //不存在保单直接创建一个新保单
		warrnty.WarrantyUserId = userData.UserId.Hex()
		warrnty.WarrantyDeviceId = deviceData.Device_id
		warrnty.WarrantyServer = "一年"
		warrnty.WarrantyServerStart = device.DeviceActivateTime
		warrnty.WarrantyServerEnd = device.DeviceActivateTime + uint32(365*86400)
		warrnty.WarrantyName = "电动自行车盗抢保障"
		warrnty.WarrantyService = "久劲"
		warrnty.WarrantyCarModel.CarName = travelName
		warrnty.WarrantyDeviceIdStr = fmt.Sprintf("%d", deviceData.Device_id)
		warrnty.Add()
	}
	var deviceInfo models.DeviceInfo
	err = userModel.Map("devices", deviceId, &deviceInfo)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	deviceInfo.Device_activity_time = uint32(travelInfo.TravelDate)
	deviceInfo.DeviceUserId = userInfo.UserId.Hex()
	userModel.Save("devices", deviceId, deviceInfo)
	statuCode = 200
	data["code"] = 1
	data["data"] = nil
	data["msg"] = "OK"
	return
}
