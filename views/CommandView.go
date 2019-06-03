package views

import (
	"fmt"
	"regexp"
	"risk-ext/models"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris"
)

type CommandView struct {
	Views
}

func (this *CommandView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//设备下发指令
//cmd_type int 指令类型：1 开追踪 2关闭追踪 3 设置工作模式
//set_work int cmd_type 为3 传值 0:闹钟模式 2:星期
//cmd_param string 设置的参数 cmd_type=1(5分钟) cmd_type=2(可空) cmd_type=3(set_work=0>>[10:00;13:00;18:00;20:00] set_work=2>>[10:00;1,2,3,4,5,6,7])
//deviceId int 设备号
func (this *CommandView) Post(ctx iris.Context) (statuCode int, data M) {
	statuCode = 400
	data = make(M)
	mem := Session.Customer
	deviceModel := new(models.Devices)
	deviceId := ctx.PostValueInt64Default("deviceId", 0)
	cmd_type := ctx.PostValueInt64Default("cmd_type", 0)
	fmt.Println(deviceId, cmd_type)
	if deviceId == 0 || cmd_type == 0 {
		data["code"] = 0
		data["error"] = "下发指令失败,请联系工作人员"
		return
	}
	cmd_param := ctx.PostValue("cmd_param")
	deviceData, err := deviceModel.GetDeviceByDevId(uint64(deviceId))
	if err != nil {
		data["code"] = 0
		data["error"] = "不存在该设备,请联系工作人员"
		return
	}
	if deviceData.DeviceUser.UserId != mem.UserId {
		data["code"] = 0
		data["error"] = "发送指令受限(该设备不属于此操作人)"
		return
	}
	mod, err := new(models.Models).One(deviceData.Device_model.Model_id.Hex())
	if err != nil {
		data["code"] = 0
		data["error"] = "该设备工作模式有误,请联系工作人员"
		return
	}
	deviceInfo := new(models.Devices).GetDeviceInfo(deviceData.Device_id)
	var arg string
	if cmd_type == 1 { //设备开启追踪
		if mod.Model_type == 0 {
			trackParam := mod.Model_works[3]
			max, err := strconv.Atoi(trackParam[0])
			if err != nil {
				data["code"] = 0
				data["error"] = "该设备工作模式有误,请联系工作人员"
				return

			}
			track_min, err := strconv.Atoi(cmd_param)
			if err != nil {
				data["code"] = 0
				data["error"] = "追踪开启失败,请联系工作人员"
				return
			}
			if track_min < 1 || track_min > max {
				data["code"] = 0
				data["error"] = "追踪时间超过上限,请重新设置"
				return
			}
			deviceInfo.Device_tracking_params = uint16(track_min)
			arg = fmt.Sprintf("%d", track_min)
			if deviceInfo.Device_tracking == 3 || deviceInfo.Device_tracking == 2 || deviceInfo.Device_tracking == 5 {
				deviceInfo.Device_tracking = 5
			} else {
				deviceInfo.Device_tracking = 1 //准备追踪
			}
		} else {
			deviceInfo.Device_tracking = 2
			deviceInfo.Device_last_tracking = uint32(time.Now().Unix())
		}

		cmd_type = 3 //开追踪
		deviceData.Device_tracking = deviceInfo.Device_tracking
	} else if cmd_type == 2 { //设备关闭追踪
		if mod.Model_type == 0 { //无线设备
			deviceInfo.Device_tracking = 3 //准备恢复正常模式
			params := deviceInfo.Device_running_params
			if params.Mod == 1 { //闹钟
				for k, v := range params.Timer {
					if k == 0 {
						arg = fmt.Sprintf("%s", v)
					} else {
						arg = fmt.Sprintf("%s,%s", arg, v)
					}

				}
			} else if params.Mod == 2 || params.Mod == 3 { //定时 星期
				arg = fmt.Sprintf("%s,%s", params.Timer[0], params.Selector)
			}
			cmd_type = int64(3) + int64(params.Mod)

			deviceInfo.Device_tracking = 4
			deviceData.Device_tracking = 4
		} else { //有线设备
			deviceInfo.Device_tracking = 0
			deviceData.Device_tracking = 0
		}
	} else if cmd_type == 3 { //设备设置工作模式
		set_work := ctx.PostValueInt64Default("set_work", -1)
		Mar := strings.Split(cmd_param, ";")
		if set_work == 0 { //闹钟模式
			flag, _ := regexp.MatchString("((([0-1]\\d)|(2[0-3])):[0-5]\\d)(;(([0-1]\\d)|(2[0-3])):[0-5]\\d){0,3}\\z", cmd_param)
			if !flag {
				data["code"] = 0
				data["error"] = "闹钟模式设置有误(参数有误)"
				return
			}
			deviceData.Device_model.Model_works[0] = Mar
			deviceInfo.Device_will_params.Mod = 1
			deviceInfo.Device_will_params.Timer = Mar
			for k, v := range Mar {
				if k == 0 {
					arg = fmt.Sprintf("%s", v)
				} else {
					arg = fmt.Sprintf("%s,%s", arg, v)
				}

			}
		} else if set_work == 2 { //星期模式
			flag, _ := regexp.MatchString("(((^[0-1]{1}[0-9]{1})|(^2[0-3]{1})):[0-5]\\d;)[1-7](,[1-7]){0,7}\\z", cmd_param)
			if !flag {
				data["code"] = 0
				data["error"] = "星期模式设置有误(参数有误)"
				return
			}
			deviceData.Device_model.Model_works[2] = Mar
			if len(Mar) <= 1 {
				data["code"] = 0
				data["error"] = "星期模式设置有误(参数有误)"
				return
			}
			deviceInfo.Device_will_params.Mod = 3
			deviceInfo.Device_will_params.Timer = []string{Mar[0]}
			deviceInfo.Device_will_params.Selector = Mar[1]
			deviceInfo.Device_tracking = 4
			deviceData.Device_tracking = 4
			arg = fmt.Sprintf("%s,%s", Mar[0], Mar[1])
		}
	}
	trackInterval, err := new(models.Commands).Command(fmt.Sprintf("%d", deviceId), uint8(cmd_type), deviceInfo, mod, deviceData, arg)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	statuCode = 200
	data["result"] = trackInterval
	data["code"] = 1
	data["error"] = "OK"
	return
}

//获取详情或列表待用
func (this *CommandView) Get(ctx iris.Context) (statuCode int, data M) {
	return
}

//更新操作待用
func (this *CommandView) Put(ctx iris.Context) (statuCode int, data M) {
	return
}

//删除操作待用
func (this *CommandView) Delete(ctx iris.Context) (statuCode int, data M) {
	return
}
