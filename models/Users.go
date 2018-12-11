package models

import "gopkg.in/mgo.v2/bson"

type Users struct {
	UserId      bson.ObjectId `bson:"_id,omitempty" json:"user_id"`     //id
	UserFname   string        `bson:"user_fname" json:"user_fname"`     //姓名
	UserUname   string        `bson:"user_uname" json:"user_uname"`     //登录名
	UserPasswd  string        `bson:"user_passwd" json:"user_passwd"`   //密码
	UserAvatar  string        `bson:"user_avatar" json:"user_avatar"`   //头像
	UserTravel  []Travel      `bson:"user_travel" json:"user_travel"`   //交通工具
	UserOpenId  string        `bson:"user_open_id" json:"user_open_id"` //微信openId
	UserMobile  string        `bson:"user_mobile" json:"user_mobile"`   //登录手机号码
	UserStatus  uint8         `bson:"user_status" json:"user_status"`   //用户状态0禁用 1启用 2未注册
	UserToken   string        `bson:"user_token" json:"user_token"`     //登录token
	UserLogin   uint32        `bson:"user_login" json:"user_login"`     //最后登录时间
	UserRead    uint32        `bson:"user_read" json:"user_read"`       //阅读报警的时间
	UserDeleted uint32        `bson:"user_deleted" json:"user_deleted"` //删除时间
	UserDate    uint32        `bson:"user_date" json:"user_date"`       //创建时间
}

type Travel struct {
	TravelName   string     `bson:"travel_name" json:"travel_name"`     //交通工具名称
	TravelType   uint8      `bson:"travel_type" json:"travel_type"`     //交通工具类型0=电动车 1=自行车 2=汽车
	TravelDevice DeviceInfo `bson:"travel_device" json:"travel_device"` //绑定的设备号
	TravelShare  string     `bson:"travel_share" json:"travel_share"`   //共享用户ID 为空则不是共享设备 共享设备只有查看权
	TravelDate   int64      `bson:"travel_date" json:"travel_date"`     //绑定时间
}

/*******************实时数据字段****************************/
type DeviceInfo struct {
	Device_id              uint64   `json:"device_id"`              //设备号
	Device_voltage         uint8    `json:"device_voltage"`         //设备电量百分比
	Device_car_id          string   `json:"device_car_id"`          //车辆ID
	Device_speed           uint8    `json:"device_speed"`           //速度
	Device_direction       uint16   `json:"device_direction"`       //方向360度
	Device_mileage         uint32   `json:"device_mileage"`         //里程
	Device_latlng          Latlng   `json:"device_latlng"`          //经纬度
	Device_slatlng         Latlng   `json:"device_slatlng"`         //原坐标
	Device_address         string   `json:"device_address"`         //地址
	Device_name            string   `json:"device_name"`            //设备名
	Device_loctype         uint8    `json:"device_loctype"`         //定位类型0=GPS 1=基站 2=WiFi
	Device_status          uint8    `json:"device_status"`          //状态0=静止 1=运行 2=离线
	Device_tracking        uint8    `json:"device_tracking"`        //状态0=未追踪  1=准备追踪 2=正在追踪 3=在等待关闭追踪 4=等待更改工作模式 5=追踪期间更改频率
	Device_tracking_params uint16   `json:"device_tracking_params"` //状态 1=准备追踪 2=正在追踪时有效 追踪参数 单位分钟
	Device_running_params  struct { //正在运行的参数
		Mod      uint8    //1=闹钟 2=定时 3=星期追踪
		Timer    []string //mod 1时 唤醒时间可多个,2时开始时间只有一个,3时星期中每天唤醒时间
		Selector string   //mod=1时无效，mod=2时 是间隔时间 单位小时， mod=3时星期数 例如1,3,7代表周一 周三 周日
	} `json:"device_running_params"` //状态 3=准备恢复正常模式 时有效  常规模式参数 根据不同设备组合
	Device_will_params struct { //将要执行的参数
		Mod      uint8    //1=闹钟 2=定时 3=星期追踪
		Timer    []string //mod 1时 唤醒时间可多个,2时开始时间只有一个,3时星期中每天唤醒时间
		Selector string   //mod=1时无效，mod=2时 是间隔时间 单位小时， mod=3时星期数 例如1,3,7代表周一 周三 周日
	} `json:"device_will_params"` //状态 3=准备恢复正常模式 时有效  常规模式参数 根据不同设备组合
	Device_offtime         uint32 `json:"device_offtime"`         //离线时间
	Device_statictime      uint32 `json:"device_statictime"`      //静止时间
	Device_staticlen       uint32 `json:"device_staticlen"`       //停留时间
	Device_timertime       uint32 `json:"device_timertime"`       //心跳包最新时间
	Device_acttime         uint32 `json:"device_acttime"`         //设备通信最新时间
	Device_loctime         uint32 `json:"device_loctime"`         //定位时间
	Device_server_endtime  uint32 `json:"device_server_endtime"`  //服务结束时间
	Device_type            uint8  `json:"device_type"`            //最后定位设备类型 0无线 1有线
	Device_last_tracking   uint32 `json:"device_last_tracking"`   //最后一次开启追踪时间
	Device_alarm           int8   `json:"device_alarm"`           //警报类型 0 断电报警（有线） 1见光报警（无线）  2围栏（出围） 3围栏（入围）  4关机报警 5开机报警 6低电报警 7通电报警 8见光恢复 9低电恢复 10解除出围 11解除入围
	Device_next_time       string `json:"device_next_time"`       //距离下次的时间
	Device_activity_time   uint32 `json:"device_activity_time"`   //激活时间（第一次绑车时间）
	Device_activity_latlng Latlng `json:"device_activity_latlng"` //激活经纬度
	Device_acc_state       uint8  `json:"device_acc_state"`       //点火状态 0=未知 1=点火 2=熄火
	Device_power_state     uint8  `json:"device_power_state"`     //主电状态 0=未知 1=断电 2=通电 3=欠压
	Device_oil_status      uint8  `json:"device_oil_status"`      //设备油电状态 0=未知 1=断油电 2=通油电
	Device_adcode          string `json:"device_adcode"`          //地址代码
	DeviceShipping         uint8  `json:"device_shipping"`        //是否已出库0=未出库 1=已出库
	DeviceOnline           uint8  `json:"device_online"`          //状态 1=在线 0=离线
	DeviceStoreState       uint8  `json:"device_store_state"`     //设备库存状态 0=正常 1=待检测 2=报废 主要对入库
	DeviceCmdNum           uint8  `json:"device_cmd_num"`         //设备当前正在执行的指令个数
}