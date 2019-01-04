package models

import (
	"errors"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

type Route struct {
	Model                  `bson:"-" json:"-"` //model基类
	Device_id              uint64              `json:"device_id"`              //设备号
	Device_voltage         uint8               `json:"device_voltage"`         //设备电量百分比
	Device_car_id          string              `json:"device_car_id"`          //车辆ID
	Device_speed           uint8               `json:"device_speed"`           //速度
	Device_direction       uint16              `json:"device_direction"`       //方向360度
	Device_mileage         uint32              `json:"device_mileage"`         //里程
	Device_latlng          Latlng              `json:"device_latlng"`          //经纬度
	Device_slatlng         Latlng              `json:"device_slatlng"`         //原坐标
	Device_address         string              `json:"device_address"`         //地址
	Device_name            string              `json:"device_name"`            //设备名
	Device_loctype         uint8               `json:"device_loctype"`         //定位类型0=GPS 1=基站 2=WiFi
	Device_status          uint8               `json:"device_status"`          //状态0=静止 1=运行 2=离线
	Device_tracking        uint8               `json:"device_tracking"`        //状态0=未追踪  1=准备追踪 2=正在追踪 3=在等待关闭追踪 4=等待更改工作模式
	Device_tracking_params uint16              `json:"device_tracking_params"` //状态 1=准备追踪 2=正在追踪时有效 追踪参数 单位分钟
	Device_running_params  struct {            //正在运行的参数
		Mod      uint8    //1=闹钟 2=定时 3=星期追踪
		Timer    []string //mod 1时 唤醒时间可多个,2时开始时间只有一个,3时星期中每天唤醒时间
		Selector string   //mod=1时无效，mod=2时 是间隔时间 单位小时， mod=3时星期数 例如1,3,7代表周一 周三 周日
	} `json:"device_running_params"` //状态 3=准备恢复正常模式 时有效  常规模式参数 根据不同设备组合
	Device_offtime        uint32 `json:"device_offtime"`        //离线时间
	Device_statictime     uint32 `json:"device_statictime"`     //静止时间
	Device_staticlen      uint32 `json:"device_staticlen"`      //停留时间
	Device_timertime      uint32 `json:"device_timertime"`      //心跳包最新时间
	Device_acttime        uint32 `json:"device_acttime"`        //设备通信最新时间
	Device_loctime        uint32 `json:"device_loctime"`        //定位时间
	Device_server_endtime uint32 `json:"device_server_endtime"` //服务结束时间
	Device_type           uint8  `json:"device_type"`           //最后定位设备类型 0无线 1有线
	Device_last_tracking  uint32 `json:"device_last_tracking"`  //最后一次开启追踪时间
	Device_alarm          int8   `json:"device_alarm"`          //警报类型 0 断电报警（有线） 1见光报警（无线）  2围栏（出围） 3围栏（入围）  4关机报警 5开机报警 6低电报警 7通电报警 8见光恢复 9低电恢复 10解除出围 11解除入围
	Device_acc_state      uint8  `json:"device_acc_state"`      //点火状态 0=未知 1=点火 2=熄火
	Device_power_state    uint8  `json:"device_power_state"`    //主电状态 0=未知 1=断电 2=通电 3=欠压
	Device_oil_status     uint8  `json:"device_oil_status"`     //设备油电状态 0=未知 1=通油电 2=断油电
}

type RoutesResult struct {
	Device_id        uint64 `json:"device_id"`        //设备号
	Device_latlng    Latlng `json:"device_latlng"`    //经纬度
	Device_voltage   uint8  `json:"device_voltage"`   //设备电量百分比
	Device_address   string `json:"device_address"`   //地址
	Device_loctype   uint8  `json:"device_loctype"`   //定位类型0=GPS 1=基站 2=WiFi
	Device_direction uint16 `json:"device_direction"` //方向360度
	Device_status    uint8  `json:"device_status"`    //状态0=静止 1=运行 2=离线
	Device_loctime   uint32 `json:"device_loctime"`   //定位时间
}

//分页获取轨迹列表
func (this *Route) NewGetRoutesByPaging(deviceId string, startTime, endTime uint32, page, pageSize, types int) (rou []RoutesResult, count int, err error) {
	if page < 1 {
		page = 1
	}
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	devInfo, err := new(Devices).GetDeviceByDevId(devId)
	if err != nil {
		err = errors.New("不存在该设备")
		return
	}

	var offset = (page - 1) * pageSize
	lentime := endTime - startTime
	if lentime > 30*24*60*60 {
		err = errors.New("查看轨迹超过30天，请分段查询")
		return
	}
	where := bson.M{}
	if endTime > startTime {
		if startTime < devInfo.DeviceActivateTime {
			err = errors.New("起止时间不能小于设备激活时间")
			return
		}
	} else {
		err = errors.New("开始时间不能小于结束时间")
		return
	}
	where["device_id"] = devInfo.Device_id
	where["device_loctime"] = bson.M{"$gte": startTime, "$lte": endTime}
	key := "locs"
	count, _ = this.RouteCollection(key).Find(where).Count()
	data := bson.M{"device_latlng": 1, "device_address": 1, "device_id": 1, "device_loctype": 1, "device_direction": 1, "device_status": 1, "device_loctime": 1, "device_voltage": 1}
	if types == 0 { //轨迹列表
		err = this.RouteCollection(key).Find(where).Sort("-device_loctime").Skip(offset).Limit(pageSize).Select(data).All(&rou)
	} else { //轨迹打点
		err = this.RouteCollection(key).Find(where).Sort("device_loctime").Skip(offset).Limit(pageSize).Select(data).All(&rou)
	}
	return
}

//获取设备停留列表
func (this *Route) GetStayList(startTime, endTime, stayTime uint32, deviceId string) (rou []Route, err error) {
	devId, _ := strconv.ParseUint(deviceId, 10, 64)
	devInfo, err := new(Devices).GetDeviceByDevId(devId)
	if err != nil {
		err = errors.New("不存在该设备")
		return
	}
	lentime := endTime - startTime
	if lentime > 30*24*60*60 {
		err = errors.New("查看轨迹超过30天，请分段查询")
		return
	}
	where := bson.M{}
	lent := bson.M{}
	if endTime > startTime {
		if startTime < devInfo.DeviceActivateTime {
			err = errors.New("起止时间不能小于设备激活时间")
			return
		}
	} else {
		err = errors.New("开始时间不能小于结束时间")
		return
	}
	lent["$gte"] = stayTime
	where["device_staticlen"] = lent
	where["device_id"] = devId
	where["device_loctime"] = bson.M{"$gte": startTime, "$lte": endTime}
	//	routes = system.RoutesMongo.C("locs_" + dev.Device_id_str[len(dev.Device_id_str)-3:])
	key := "locs"
	err = this.RouteCollection(key).Find(where).Sort("device_loctime").All(&rou)
	return
}
