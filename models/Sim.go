package models

import (
	"fmt"
	"risk-ext/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

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
		if info.Device_running_params.Mod == 0 || len(info.Device_running_params.Timer) == 0 {
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
