package models

import (
	"errors"
	"fmt"
	"risk-ext/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Devices struct {
	Model                   `bson:"-" json:"-"` //model基类
	Redis                   `bson:"-" json:"-"` //redis基类
	Device_id               uint64              `json:"device_id"`                                            //设备号
	Device_id_str           string              `json:"device_id_str"`                                        //字符串设备号
	Device_name             string              `json:"device_name"`                                          //设备名
	Device_model            Models              `json:"device_model"`                                         //设备型号
	Device_sim              uint64              `json:"device_sim"`                                           //sim卡
	Device_sim_info         SimInfo             `bson:"device_sim_info" json:"device_sim_info"`               //SIM卡详细信息
	Device_car_id           bson.ObjectId       `bson:"device_car_id,omitempty" json:"device_car_id"`         //车辆ID
	Device_car_plate        string              `json:"device_car_plate"`                                     //车牌号
	Device_company_id       bson.ObjectId       `bson:"device_company_id,omitempty" json:"device_company_id"` //企业ID
	Device_company_name     string              `json:"device_company_name"`                                  //企业名
	Device_group_id         bson.ObjectId       `bson:"device_group_id,omitempty" json:"device_group_id"`     //组织ID
	Device_group_name       string              `json:"device_group_name"`                                    //组织名
	Device_server_starttime uint32              `json:"device_server_starttime"`                              //服务开始时间
	Device_server_endtime   uint32              `json:"device_server_endtime"`                                //服务结束时间
	Device_due              uint8               `json:"device_due"`                                           //是否到期0未到期 1到期
	Device_remark           string              `json:"device_remark"`                                        //备注
	Device_sim_starttime    uint32              `json:"device_sim_starttime"`                                 //sim开始时间
	Device_sim_endtime      uint32              `json:"device_sim_endtime"`                                   //sim结束时间
	Device_install          uint8               `json:"device_install"`                                       //安装位置
	Device_outtime          uint32              `json:"device_outtime"`                                       //出库时间
	Device_batch_id         uint32              `json:"device_batch_id"`                                      //入库批次号
	Device_batch_id_str     string              `json:"device_batch_id_str"`                                  //入库批次号(ObjectIdStr)
	Device_batch_no         string              `json:"device_batch_no"`                                      //入库批次号(模糊查询显示使用，入库objectId后6位)
	Device_serial_id        uint32              `json:"device_serial_id"`                                     //出库流水号
	Device_serial_objid     string              `json:"device_serial_objid"`                                  //出库流水号ObjectId
	Device_intime           uint32              `json:"device_intime"`                                        //入库时间
	Device_rectime          uint32              `json:"device_rectime"`                                       //回收时间
	Device_deleted          uint8               `json:"device_deleted"`                                       //是否已删除0未 1已
	Device_info             *DeviceInfo         `json:"device_info"`                                          //设备实时数据
	Device_tracking         uint8               `json:"device_tracking"`                                      //状态0=未追踪 1=准备追踪 2=正在追踪 3=准备恢复正常模式
	Device_bind_time        uint32              `json:"device_bind_time"`                                     //绑车时间
	Device_unbind_time      uint32              `json:"device_unbind_time"`                                   //解绑时间
	Device_tracking_time    int64               `json:"device_tracking_time"`                                 //开启追踪的时间
	DeviceStoreState        uint8               `bson:"device_store_state" json:"device_store_state"`         //设备库存状态 0=正常 1=待检测 2=报废 主要对入库
	DeviceStoreType         uint8               `bson:"device_store_type" json:"device_store_type"`           //设备设备来源 0=采购 1=试用 2=回收
	DeviceState             uint8               `bson:"device_state" json:"device_state"`                     //设备调度类型 0=正常 1=调度中（锁定）
	DeviceOutType           uint8               `bson:"device_out_type" json:"device_out_type"`               //设备出库类型  0正式出库 1测试出库 2To customer 出库
	DeviceBatcheNo          string              `bson:"device_batche_no" json:"device_batche_no"`             //所属批次号
	DeviceReposId           string              `bson:"device_repos_id" json:"device_repos_id"`               //所属仓库ID
	DeviceCarOwner          string              `bson:"device_car_owner" json:"device_car_owner"`             //车主
	DeviceCarVin            string              `bson:"device_car_vin" json:"device_car_vin"`                 //VIN车架号
	DeviceOnline            uint8               `bson:"device_online" json:"device_online"`                   //状态 1=在线 0=离线
	DeviceUser              *Users              `bson:"device_user" json:"device_user"`                       //C端客户信息
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

type Models struct {
	Model_id             bson.ObjectId `bson:"_id" json:"model_id"`
	Model_name           string        `json:"model_name"`           //型号名
	Model_type           uint8         `json:"model_type"`           //设备类型0无线 1有线
	Model_works          [][]string    `json:"model_works"`          //可用工作模式 0闹钟模式 1定时循环模式 2星期模式 3追踪模式 工作参数 如果是数字则单位分钟
	Model_works_mode     []string      `json:"model_works_mode"`     //选择的设备工作模式(无线设备) 0闹钟模式 1定时模式 2星期模式 3追踪模式
	Model_current_work   uint8         `json:"model_current_work"`   //当前工作模式 0闹钟模式 1定时循环模式 2星期模式 3追踪模式
	Model_correspondence string        `json:"model_correspondence"` //对应设备型号
	Model_supplier_model string        `json:"model_supplier_model"` //供应商型号
	Model_status         uint8         `json:"model_status"`         //设备状态 0正在销售 1停止销售 2测试
	Model_loc_model      []string      `json:"model_loc_model"`      //型号支持的定位模式  0GPS 1基站 2WIFI 3北斗
	Model_command        []string      `json:"model_command"`        //型号支持的指令 0=重启设备 1=断油电 2=通油电 3开追踪 4设置闹钟 5设置星期
	Model_alarm          []string      `json:"model_wireless_alarm"` //型号支持的预警 当无线设备时 0光感 1翻转 2低电量; 当有线设备时 0断电
	Model_state_method   []string      `json:"model_state_method"`   //型号支持的状态检测（有线设备） 0 ACC
	Model_control        []string      `json:"model_control"`        //型号支持的控制功能（有线设备） 0断油电
	Model_stop_date      uint32        `json:"model_stop_date"`      //停止销售时间
	Model_date           uint32        `json:"model_date"`           //添加日期
	Model_port           uint32        `json:"model_port"`           //端口号
}

type SimInfo struct {
	Msisdn               string `json:"msisdn"`               //SIM卡号
	Iccid                string `json:"iccid"`                //iccid
	Imsi                 string `json:"imsi"`                 //imsi
	Sp_code              string `json:"sp_code"`              //短信端口
	Carrier              string `json:"carrier"`              //运营商
	Data_plan            int    `json:"data_plan"`            //套餐大小
	Data_usage           string `json:"data_usage"`           //当月用量
	Account_status       string `json:"account_status"`       //卡状态
	Expiry_date          string `json:"expiry_date"`          //计费结束日期
	Active               bool   `json:"active"`               //激活/未激活
	Test_valid_date      string `json:"test_valid_date"`      //测试期起始日期
	Silent_valid_date    string `json:"silent_valid_date"`    //沉默期起始日期
	Test_used_data_usage string `json:"test_used_data_usage"` //测试期已用流量
	Active_date          string `json:"active_date"`          //激活日期
	Data_balance         string `json:"data_balance"`         //剩余流量
	Outbound_date        string `json:"outbound_date"`        //出库日期
	Support_sms          bool   `json:"support_sms"`          //是否支持短信
}

//根据SIM卡号搜索
func (this *Devices) OneBySim(sim uint64) (dev Devices, err error) {
	err = this.Collection(this).Find(bson.M{"device_sim": sim}).One(&dev)
	return
}

//更新本对象
func (this *Devices) Update(flag bool, unsetfiled ...string) error {
	if this.Device_id == 0 {
		return errors.New("无效的设备ID")
	}
	query := bson.M{}
	data := utils.Struct2Map(*this, flag)
	query["$set"] = data

	if len(unsetfiled) > 0 {
		unsetData := bson.M{}
		for _, ud := range unsetfiled {
			unsetData[ud] = 1
		}
		query["$unset"] = unsetData
	}
	err := this.Collection(this).Update(bson.M{"device_id": this.Device_id}, query)
	return err
}

func (this *Devices) GetDeviceByDevId(deviceId uint64) (dev Devices, err error) {
	err = this.Collection(this).Find(bson.M{"device_id": deviceId}).One(&dev)
	return
}

func (this *Devices) GetDeviceInfo(deviceId uint64) (info *DeviceInfo) {
	err := this.Map("devices", fmt.Sprintf("%d", deviceId), &info)
	if err != nil {
		return
	}
	nextTime := ""
	if info != nil {
		if info.Device_will_params.Mod == 0 || len(info.Device_will_params.Timer) == 0 {
			nextTime = ""
		} else {
			nextTime = this.GetTrackInterval(*info)
		}
		now := time.Now().Unix()
		acttime := info.Device_acttime
		len := now - int64(acttime)
		lenloc := now - int64(info.Device_loctime)
		if info.Device_type == 1 {

			if len > 1800 {
				info.Device_status = 2
			} else {
				if info.Device_speed > 5 {
					if lenloc > 300 {
						info.Device_status = 0
					} else {
						info.Device_status = 1
					}
				} else {
					info.Device_status = 0
				}
			}
		} else {
			if info.Device_tracking == 2 || info.Device_tracking == 3 {
				if len > int64(info.Device_tracking_params)*3*60 {
					info.Device_status = 2
				} else {
					if info.Device_speed > 5 {
						info.Device_status = 1
					} else {
						info.Device_status = 0
					}
				}
			} else {
				runMod := info.Device_running_params
				if runMod.Mod == 1 { //闹钟模式下计算离线
					if len > 86400 {
						info.Device_status = 2
					} else {
						flag := this.CheckStatus(*info)
						if flag {
							info.Device_status = 2
						} else {
							info.Device_status = 0
						}
					}
				} else if runMod.Mod == 3 { //星期模式下计算离线
					if len > 86400*7 {
						info.Device_status = 2
					} else {
						flag := this.CheckStatus(*info)
						if flag {
							info.Device_status = 2
						} else {
							info.Device_status = 0
						}
					}
				} else if runMod.Mod == 2 { // 定时模式下计算离线
					hour, err := strconv.Atoi(info.Device_running_params.Selector)
					if err == nil {
						if len > int64(hour)*3*60*60 {
							info.Device_status = 2
						} else {
							info.Device_status = 0
						}
					}

				}
			}
		}
		if info.Device_tracking == 5 {
			info.Device_tracking = 2
		}
		if info.Device_tracking == 2 {
			nowtime := time.Now().Unix()
			startTime := info.Device_last_tracking
			if startTime == 0 {
				startTime = uint32(time.Now().Unix())
				info.Device_last_tracking = startTime
				err = this.Save("devices", strconv.Itoa(int(deviceId)), info)
				if err != nil {
					return
				}
			}
			lens := nowtime - int64(startTime)
			nextTime = utils.Timelen(int(lens))
		}
		info.Device_next_time = nextTime
	}
	return
}

//获取追踪时间间隔
func (this *Devices) GetTrackInterval(deviceInfo DeviceInfo) string {
	var timer uint32
	nowTime := uint32(time.Now().Unix())
	nowDate := time.Now().Format("2006-01-02")
	running_params := deviceInfo.Device_running_params
	if running_params.Mod == 1 {
		var sortClockTime []int
		for _, val := range running_params.Timer {
			clockDate := nowDate + " " + val + ":00"
			clockTime := utils.Str2Time(clockDate)
			sortClockTime = append(sortClockTime, int(clockTime))
		}
		if len(sortClockTime) > 1 {
			sortClockTime = append(sortClockTime, int(nowTime))
		}
		sort.Ints(sortClockTime)
		if len(sortClockTime) == 1 {
			if uint32(sortClockTime[0]) > nowTime {
				timer = uint32(sortClockTime[0]) - nowTime
			} else {
				timer = 24*3600 - (nowTime - uint32(sortClockTime[0]))
			}

		} else {
			var index int
			for key, val := range sortClockTime {
				if val == int(nowTime) {
					index = key
					break
				}
			}
			if index == 0 {
				timer = uint32(sortClockTime[1]) - nowTime
			} else if index == 1 {
				timer = uint32(sortClockTime[2]) - nowTime
			}
			if index == len(sortClockTime)-1 {
				timer = 24*3600 - (nowTime - uint32(sortClockTime[0]))
			} else {
				timer = uint32(sortClockTime[index+1]) - nowTime
			}
		}

	} else if running_params.Mod == 2 {
		spacing_time, _ := strconv.Atoi(running_params.Selector)
		tlen := uint32(time.Now().Unix()) - deviceInfo.Device_loctime
		if spacing_time == 0 {
			timer = uint32(spacing_time * 60 * 60)
		} else {
			timer = uint32(spacing_time*60*60) - (tlen % uint32(spacing_time*60*60))
		}

	} else if running_params.Mod == 3 {
		nowWeek := time.Now().Weekday().String()
		var nowWeekNum int
		switch nowWeek {
		case "Monday":
			nowWeekNum = 1
		case "Tuesday":
			nowWeekNum = 2
		case "Wednesday":
			nowWeekNum = 3
		case "Thursday":
			nowWeekNum = 4
		case "Friday":
			nowWeekNum = 5
		case "Saturday":
			nowWeekNum = 6
		case "Sunday":
			nowWeekNum = 7
		}
		weeks := strings.Split(running_params.Selector, ",")
		weekDate := nowDate + " " + running_params.Timer[0] + ":00"
		weekTime := utils.Str2Time(weekDate)
		firstWeek, _ := strconv.Atoi(weeks[0])
		for key, val := range weeks {
			intWeek, _ := strconv.Atoi(val)
			if intWeek > nowWeekNum {
				timer = uint32((intWeek-nowWeekNum)*24*3600) - (nowTime - weekTime)
				break
			} else if intWeek == nowWeekNum {
				if nowTime >= weekTime {
					if key == len(weeks)-1 {
						nextWeek, _ := strconv.Atoi(weeks[0])
						timer = weekTime + uint32(86400*(7+nextWeek-nowWeekNum)) - nowTime
					} else {
						nextWeek, _ := strconv.Atoi(weeks[key+1])
						timer = weekTime + uint32(86400*(nextWeek-nowWeekNum)) - nowTime
					}
					//					timer = uint32(7*24*3600) - (nowTime - weekTime)
				} else {
					timer = weekTime - nowTime
				}
				break
			} else {
				if len(weeks)-1 == key {
					timer = uint32((7-nowWeekNum+firstWeek)*24*3600) - (nowTime - weekTime)
					break
				}
			}
		}

	} else {
		timer = 0
	}
	intTimer := int(timer)
	timeStr := utils.Timelen(intTimer)
	return timeStr
}

//获取追踪时间间隔(发送指令)
func (this *Devices) GetIntTrackInterval(deviceInfo *DeviceInfo) int {
	var timer uint32
	nowTime := uint32(time.Now().Unix())
	nowDate := time.Now().Format("2006-01-02")
	running_params := deviceInfo.Device_running_params
	if running_params.Mod == 1 {
		var sortClockTime []int
		for _, val := range running_params.Timer {
			clockDate := nowDate + " " + val + ":00"
			clockTime := utils.Str2Time(clockDate)
			sortClockTime = append(sortClockTime, int(clockTime))
		}
		if len(sortClockTime) > 1 {
			sortClockTime = append(sortClockTime, int(nowTime))
		}
		sort.Ints(sortClockTime)
		if len(sortClockTime) == 1 {
			if uint32(sortClockTime[0]) > nowTime {
				timer = uint32(sortClockTime[0]) - nowTime
			} else {
				timer = 24*3600 - (nowTime - uint32(sortClockTime[0]))
			}

		} else {
			var index int
			for key, val := range sortClockTime {
				if val == int(nowTime) {
					index = key
					break
				}
			}
			if index == 0 {
				timer = uint32(sortClockTime[1]) - nowTime
			} else if index == 1 {
				timer = uint32(sortClockTime[2]) - nowTime
			}
			if index == len(sortClockTime)-1 {
				timer = 24*3600 - (nowTime - uint32(sortClockTime[0]))
			} else {
				timer = uint32(sortClockTime[index+1]) - nowTime
			}
		}

	} else if running_params.Mod == 2 {
		spacing_time, _ := strconv.Atoi(running_params.Selector)
		tlen := uint32(time.Now().Unix()) - deviceInfo.Device_loctime
		if spacing_time == 0 {
			timer = uint32(spacing_time * 60 * 60)
		} else {
			timer = uint32(spacing_time*60*60) - (tlen % uint32(spacing_time*60*60))
		}

	} else {
		nowWeek := time.Now().Weekday().String()
		var nowWeekNum int
		switch nowWeek {
		case "Monday":
			nowWeekNum = 1
		case "Tuesday":
			nowWeekNum = 2
		case "Wednesday":
			nowWeekNum = 3
		case "Thursday":
			nowWeekNum = 4
		case "Friday":
			nowWeekNum = 5
		case "Saturday":
			nowWeekNum = 6
		case "Sunday":
			nowWeekNum = 7
		}
		weeks := strings.Split(running_params.Selector, ",")
		weekDate := nowDate + " " + running_params.Timer[0] + ":00"
		weekTime := utils.Str2Time(weekDate)
		firstWeek, _ := strconv.Atoi(weeks[0])
		for key, val := range weeks {
			intWeek, _ := strconv.Atoi(val)
			if intWeek > nowWeekNum {
				timer = uint32((intWeek-nowWeekNum)*24*3600) - (nowTime - weekTime)
				break
			} else if intWeek == nowWeekNum {
				if nowTime >= weekTime {
					timer = uint32(7*24*3600) - (nowTime - weekTime)
				} else {
					timer = weekTime - nowTime
				}
				break
			} else {
				if len(weeks)-1 == key {
					timer = uint32((7-nowWeekNum+firstWeek)*24*3600) - (nowTime - weekTime)
					break
				}
			}
		}
	}
	intTimer := int(timer)
	return intTimer
}

func (this *Devices) CheckStatus(deviceInfo DeviceInfo) bool {
	var flag = false
	nowTime := uint32(time.Now().Unix())
	nowDate := time.Now().Format("2006-01-02")
	locDate := utils.Time2Str1(deviceInfo.Device_acttime)
	running_params := deviceInfo.Device_running_params
	if running_params.Mod == 1 {
		var sortClockTime []int
		for _, val := range running_params.Timer {
			clockDate := nowDate + " " + val + ":00"
			clockTime := utils.Str2Time(clockDate)
			sortClockTime = append(sortClockTime, int(clockTime))
		}
		if locDate != nowDate {
			for _, val := range running_params.Timer {
				clockDate := locDate + " " + val + ":00"
				clockTime := utils.Str2Time(clockDate)
				sortClockTime = append(sortClockTime, int(clockTime))
			}
		}

		if len(sortClockTime) > 1 {
			sortClockTime = append(sortClockTime, int(nowTime))
			sortClockTime = append(sortClockTime, int(deviceInfo.Device_acttime))
		}
		sort.Ints(sortClockTime)
		if len(sortClockTime) > 1 {
			var end int
			var begin int
			for key, val := range sortClockTime {
				if val == int(nowTime) {
					end = key
				} else if val == int(deviceInfo.Device_acttime) {
					begin = key
				}
			}
			if (end - begin) > 1 {
				flag = true
			}

		}

	} else if running_params.Mod == 3 {
		tm := time.Unix(int64(deviceInfo.Device_acttime), 0)
		locWeek := tm.Weekday().String()

		var locWeekNum int
		switch locWeek {
		case "Monday":
			locWeekNum = 1
		case "Tuesday":
			locWeekNum = 2
		case "Wednesday":
			locWeekNum = 3
		case "Thursday":
			locWeekNum = 4
		case "Friday":
			locWeekNum = 5
		case "Saturday":
			locWeekNum = 6
		case "Sunday":
			locWeekNum = 7
		}
		weeks := strings.Split(running_params.Selector, ",")
		var start = -1
		var startIndex int
		for k, v := range weeks {
			i, err := strconv.Atoi(v)
			if err == nil {
				if i <= locWeekNum {
					start = i
					startIndex = k
				}
			}
		}
		clockDate := locDate + " " + running_params.Timer[0] + ":00"
		clockTime := utils.Str2Time(clockDate)
		if start == locWeekNum {
			if clockTime > deviceInfo.Device_acttime {
				if (int(nowTime) - int(clockTime)) > 900 {
					flag = true
				}
			} else {
				var nextTime uint32
				if startIndex < (len(weeks) - 1) {
					nextWeek := weeks[startIndex+1]
					num1, _ := strconv.Atoi(nextWeek)
					nextTime = clockTime + uint32((num1-locWeekNum)*86400)
				} else {
					nextWeek := weeks[0]
					num1, _ := strconv.Atoi(nextWeek)
					nextTime = clockTime + uint32((num1+7-locWeekNum)*86400)
				}

				if (int(nowTime) - int(nextTime)) > 900 {
					flag = true
				}
			}
		}

	}
	return flag
}
