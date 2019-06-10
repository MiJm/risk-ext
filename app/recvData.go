package app

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net"

	socketgo "github.com/nulijiabei/socketgo"
)

type CarLocal struct {
	CarPlate     string `json:"car_plate"`      //车牌号
	CarOwner     string `json:"car_owner"`      //车主
	CarVoltage   uint8  `json:"car_voltage"`    //电量最大100
	CarLatlng    Latlng `json:"car_latlng"`     //坐标
	CarSpeed     uint8  `json:"car_speed"`      //车速
	CarDirection uint16 `json:"car_direction"`  //方向最大360
	CarCity      string `json:"car_city"`       //城市
	CarLoctype   uint8  `json:"car_loctype"`    //定位类型 0gps 1基站 2WiFi
	CarState     uint8  `json:"car_state"`      //0静止 1运行 2离线
	CarGroupId   string `json:"car_group_id"`   //组织ID
	CarPgroupId  string `json:"car_pgroup_id"`  //父级组织ID 默认空
	CarCompanyId string `json:"car_company_id"` //企业ID
	CarAddress   string `json:"car_address"`    //最后定位地址
	CarLoctime   uint32 `json:"car_loctime"`    //最后定位时间
}

type AlarmNoty struct {
	AnPlate     string    `json:"an_plate"`      //车牌号
	AnLatlng    []float64 `json:"an_latlng"`     //经纬度
	AnType      uint8     `json:"an_type"`       //警报类型
	AnName      string    `json:"an_name"`       //警报类型名
	AnGroupId   string    `json:"an_group_id"`   //组织ID
	AnCompanyId string    `json:"an_company_id"` //企业ID
	AnDate      int64     `json:"an_date"`       //警报时间
}

type Latlng struct {
	Type        string    `json:"type"`        //Point
	Coordinates []float64 `json:"coordinates"` //lng lat
}

var (
	CarDataChan   = make(chan CarLocal, 100)
	AlarmDataChan = make(chan AlarmNoty, 100)
	alarmType     = [...]string{"断电", "见光", "出围", "入围", "关机", "开机", "低电量", "通电", "见光恢复", "低电恢复", "出围解除", "入围解除", "风险点", "离线", "异动", "震动", "拆卸", "ACC关", "ACC开", "停车超时", "超速预警", "常驻点预警", "设防报警", "设备分离预警", "二押点预警"}
)

//TCP
func StartUdp(port string) {

	listener, err := socketgo.NewListen("", port, 3).ListenUDP()
	//listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("UDP错误：", err.Error())
		return
	}
	log.Println("启动UDP端口" + port + "成功")
	//defer listener.Close()

	go func(conn *net.UDPConn) {
		defer conn.Close()
		reader := bufio.NewReader(conn)
		//writer := bufio.NewWriter(conn)
		for {
			buf, err := reader.ReadBytes(byte('$'))
			//this.SetTimeout(conn, buf)

			if err != nil && err != io.EOF {
				return
			} else if err == io.EOF {
				log.Println("设备已断开连接")
				return
			}

			if len(buf) <= 4 {
				log.Println("数据太少")
				continue
			}
			var length = len(buf) - 1
			buf = buf[:length]
			ptype := string(buf[:1])
			buf = buf[1:]
			maxLen := base64.RawStdEncoding.DecodedLen(len(buf))
			dst := make([]byte, maxLen)
			_, err = base64.RawStdEncoding.Decode(dst, buf)
			if err != nil {
				log.Println("非法数据格式", err)
				continue
			}
			//log.Println(string(dst))
			switch ptype {
			case "0": //车辆定位
				var carData = CarLocal{}
				//log.Println(string(dst))
				if err = json.Unmarshal(dst, &carData); err != nil {
					log.Println("非法数据格式,不是地址json")
					continue
				}
				CarDataChan <- carData
			case "1": //警报
				var almData = AlarmNoty{}
				if err = json.Unmarshal(dst, &almData); err != nil {
					log.Println("非法数据格式,不是地址json")
					continue
				}
				if int(almData.AnType) > len(alarmType)-1 {
					almData.AnName = "未知预警"
				} else {
					almData.AnName = alarmType[almData.AnType]
				}
				AlarmDataChan <- almData
			}

		}
	}(listener)

}
