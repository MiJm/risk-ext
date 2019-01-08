package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"risk-ext/config"
	"risk-ext/utils"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

/**
 * 指令表结构
 **/
type Commands struct {
	Model         `bson:"-" json:"-"` //model基类
	Redis         `bson:"-" json:"-"` //redis基类
	CmdId         bson.ObjectId       `bson:"_id,omitempty" json:"cmd_id"`
	CmdDeviceId   uint64              `bson:"cmd_device_id" json:"cmd_device_id"`   //指令所属设备ID
	CmdModelePort uint32              `bson:"cmd_model_port" json:"cmd_model_port"` //指令所属设备类型 端口号
	CmdType       uint8               `bson:"cmd_type" json:"cmd_type"`             //指令类型  0=重启设备 1=断油电 2=通油电 3开追踪 4设置闹钟 5设置定时 6设置星期
	CmdStatus     uint8               `bson:"cmd_status" json:"cmd_status"`         //指令状态0=等待执行（无线设备） 1=正在执行 2=执行成功 3=执行失败 4=已取消
	CmdExcTime    string              `bson:"cmd_exc_time" json:"cmd_exc_time"`     //CmdStatus=0有效预计指令执行时间 ，如果当次闹钟未执行则执行时间更新到下次闹钟时间
	//参数 CmdType=3 为间隔时间 单位分钟 例：20。
	//CmdType=4 为闹钟点 最多四个 例：12:00,20:00,18:00,00:00。
	//CmdType=5 为定时起点时间+间隔时间(单位小时) 例：12:20,2。
	//CmdType=6 为定时起点时间+星期编号 例：12:20,1,3,7 表示周一周3周日 的12:20唤醒
	CmdArgs       string `bson:"cmd_args" json:"cmd_args"`
	CmdError      string `bson:"cmd_error" json:"cmd_error"`             //指令执行失败的原因
	CmdCreatedAt  uint32 `bson:"cmd_created_at" json:"cmd_created_at"`   //指令创建时间
	CmdFinishedAt uint32 `bson:"cmd_finished_at" json:"cmd_finished_at"` //指令完成时间
	CmdName       string `bson:"cmd_name" json:"cmd_name"`               //指令名
	CmdSendAt     uint32 `bson:"cmd_send_at" json:"cmd_send_at"`         //指令发送时间
}

var CmdMap = []string{"重启设备", "断开油电", "接通油电", "开启追踪模式", "设置为闹钟模式", "设置为定时模式", "设置为星期模式"}

//获取redis里面的cmd结构
func (this *Commands) CmdHGet(deviceId string) (command Commands, err error) {
	data, err := config.Redis.HGet("cmds", deviceId).Bytes()
	if err == nil {
		if json.Unmarshal(data, &command) == nil {
			return
		}
	}
	return
}

//添加一条指令
func (this *Commands) Add() (err error) {
	err = this.Collection(this).Insert(this)
	return
}

func (this *Commands) Update() (err error) {
	if this.CmdId == EmptyId {
		return errors.New("无效的指令ID")
	}
	var data bson.M
	data = utils.Struct2Map(*this)
	err = this.Collection(this).UpdateId(this.CmdId, bson.M{"$set": data})
	return
}

//redis添加指令
func (this *Commands) CmdHSet() (err error) {
	redisData, _ := json.Marshal(this)
	err = config.Redis.HSet("cmds", strconv.FormatUint(this.CmdDeviceId, 10), redisData).Err()
	if err == nil {
		this.Update()
	}
	return
}

//设备指令发送
func (this *Commands) Command(devIdStr string, typ uint8, deviceInfo *DeviceInfo, mod Models, deviceData Devices, args ...string) (trackInterval string, err error) {
	deviceId, err := strconv.Atoi(devIdStr)
	if err != nil {
		return
	}
	dev, _ := new(Devices).GetDeviceByDevId(uint64(deviceId))
	if typ != 3 && typ != 4 && typ != 5 && typ != 6 {
		flag := false
		for _, v := range mod.Model_command {
			if fmt.Sprintf("%d", typ) == v {
				flag = true
			}
		}
		if !flag {
			err = errors.New("该设备不支持该指令")
			return
		}
	}

	cmd, err := this.CmdHGet(devIdStr)
	if err == nil {
		if cmd.CmdStatus == 0 || cmd.CmdStatus == 1 {
			err = errors.New("存在未执行的指令，请前往指令控制台撤销后再执行")
			return
		}
	}
	oil_status := dev.Device_info.Device_oil_status
	if typ == 1 || typ == 2 {
		if oil_status == typ {
			if typ == 1 {
				err = errors.New("终端已处于断油电状态，本指令不再执行")
				return
			} else if typ == 2 {
				err = errors.New("终端已恢复油电成功，本指令不再执行")
				return
			}

		}
	}
	newCmd := new(Commands)
	newCmd.CmdId = bson.NewObjectId()
	newCmd.CmdDeviceId = dev.Device_id
	newCmd.CmdType = typ
	newCmd.CmdStatus = 0
	newCmd.CmdModelePort = mod.Model_port
	if len(args) == 2 {
		newCmd.CmdName = args[1]
	} else {
		newCmd.CmdName = CmdMap[typ]
	}
	if len(args) > 0 {
		newCmd.CmdArgs = args[0]
	}
	newCmd.CmdCreatedAt = uint32(time.Now().Unix())
	err = newCmd.Add()
	if err == nil {
		err = newCmd.CmdHSet()
		if err == nil {
			if deviceInfo.Device_id == 0 {
				var newDeviceInfo = new(DeviceInfo)
				data, err := config.Redis.HGet("devices", devIdStr).Bytes()
				if err == nil {
					json.Unmarshal(data, newDeviceInfo)
				}
				deviceInfo = newDeviceInfo
			}
			deviceInfo.DeviceCmdNum = deviceInfo.DeviceCmdNum + 1
			redisData, _ := json.Marshal(&deviceInfo)
			if config.Redis.HSet("devices", devIdStr, redisData).Err() == nil {
				deviceData.Update(false)
				trackInterval = new(Devices).GetTrackInterval(*deviceInfo)
			} else {
				err = errors.New("指令发送失败")
			}
		}
	}
	return
}
