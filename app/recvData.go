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

type Latlng struct {
	Type        string    `json:"type"`        //Point
	Coordinates []float64 `json:"coordinates"` //lng lat
}

var (
	CarDataChan = make(chan CarLocal, 100)
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
			//log.Println(string(buf))
			maxLen := base64.RawStdEncoding.DecodedLen(len(buf))
			dst := make([]byte, maxLen)
			_, err = base64.RawStdEncoding.Decode(dst, buf)
			if err != nil {
				log.Println("非法数据格式", err)
				continue
			}
			var carData = CarLocal{}
			if err = json.Unmarshal([]byte(string(dst)), &carData); err != nil {
				log.Println("非法数据格式,不是地址json", string(dst))
				continue
			}
			CarDataChan <- carData
		}
	}(listener)

}
