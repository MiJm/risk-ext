package views

import (
	"encoding/json"
	"fmt"
	"log"
	"risk-ext/app"
	"risk-ext/models"
	"risk-ext/utils"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

var ws *websocket.Server

type WebSocketView struct {
	Views
}

func (this *WebSocketView) Delete(ctx iris.Context) (code int, rs interface{}) {
	return
}
func (this *WebSocketView) Post(ctx iris.Context) (code int, rs interface{}) {
	return
}
func (this *WebSocketView) Put(ctx iris.Context) (code int, rs interface{}) {
	return
}
func (this *WebSocketView) Get(ctx iris.Context) (int, interface{}) {
	return 200, ws.Handler()
}
func (this *WebSocketView) Auth(ctx iris.Context) int {
	return 0
}

func init() {
	ws = websocket.New(websocket.Config{
		EnableCompression: true,
		// to enable binary messages (useful for protobuf):
		// BinaryMessages: true,
	})

	ws.OnConnection(func(c websocket.Connection) {
		NewClient(c) //创建新客户连接通道
	})
}

func NewWs() *WebSocketView {
	wv := new(WebSocketView)
	return wv
}

////////////////////////////////////////////////////////////////////WsClient/////////////////////////////////////////
/**
 * 客户端处理类
 */
type WsClient struct {
	Views
	client       websocket.Connection
	disconnected bool //是否已断开连接
}

func NewClient(c websocket.Connection) *WsClient {
	log.Println("客户连接:", c.Context().Request().Host, c.ID())
	wc := new(WsClient)
	wc.client = c
	wc.Init() //初始化开始
	return wc
}

/**
 * 初始化客户端信息
 */
func (this *WsClient) Init() {
	if this.client == nil {
		return
	}
	code := this.Auth(this.client.Context())
	if code == 401 || code == 403 { //无权限关闭客户端连接
		this.client.Disconnect()
		return
	}
	this.client.OnMessage(this.OnMessage)
	this.client.OnDisconnect(this.OnDisconnect)
	return
}

//客户权限设置
func (this *WsClient) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"GET": MA{"NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

//消息处理
func (this *WsClient) OnMessage(data []byte) {

	go this.PushCarNum()    //推送车辆和设备数据
	go this.PushAlarmNum()  //推送当日预警数
	go this.PushAlarmList() //推送最新的预警列表
	go this.GetLastCarLoc() //推送车辆最新数据
	go this.PushCarList()   //推送车辆全部数据
	go this.GetLastAlarm()  //推送最新的警报
}

//推送车辆和设备数据
func (this *WsClient) PushCarNum() {
	for !this.disconnected {
		if this.Session == nil {
			log.Println("session不存在", this.client.ID(), this.disconnected, this.client.Context().FormValue("token"))
			break
		}
		result, err := models.GetNums(this.Session.User)
		if err != nil {
			return
		}

		// c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message)) // broadcast to all clients except this
		this.Result("car_dev_num", result)
		time.Sleep(15 * time.Second)
	}

}

//推送当日预警数据
func (this *WsClient) PushAlarmNum() {
	for !this.disconnected {
		if this.Session == nil {
			log.Println("session不存在", this.client.ID(), this.disconnected, this.client.Context().FormValue("token"))
			break
		}
		result, err := models.GetAlarmNum(this.Session.User)
		if err != nil {
			return
		}
		// c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message)) // broadcast to all clients except this
		this.Result("alarm_num", result)
		time.Sleep(10 * time.Second)
	}

}

//推送最新预警列表
func (this *WsClient) PushAlarmList() {
	start := 0
	now := utils.Time2Str1(uint32(time.Now().Unix()))
	startTime := utils.Str2Time(fmt.Sprintf("%s 00:00:00", now))
	for !this.disconnected {
		if this.Session == nil {
			log.Println("session不存在", this.client.ID(), this.disconnected, this.client.Context().FormValue("token"))
			break
		}

		result, err := new(models.Alarms).GetAlarmList(this.Session.User, start, 0, 5)
		if err != nil {
			return
		}

		rs, err := new(models.Alarms).GetNums(this.Session.User, startTime)
		if err != nil {
			return
		}
		allRes := make(map[string]interface{})
		allRes["alarm_list"] = result
		allRes["alarm_gather"] = rs

		// c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message)) // broadcast to all clients except this
		this.Result("alarm_list", allRes)
		start = int(time.Now().Unix())
		time.Sleep(10 * time.Second)
	}

}

//推送全部车辆数据
func (this *WsClient) PushCarList() {
	for !this.disconnected {
		if this.Session == nil {
			log.Println("session不存在", this.client.ID(), this.disconnected, this.client.Context().FormValue("token"))
			break
		}
		result, err := new(models.Cars).GetAllCars(this.Session.User)
		if err != nil {
			return
		}
		// c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message)) // broadcast to all clients except this
		this.Result("car_list", result)
		time.Sleep(10 * time.Second)
	}

}

func (this *WsClient) Result(msgType string, msgData interface{}) {
	if this.disconnected {
		log.Println("客户端断开连接")
		return
	}
	msg := struct {
		MsgType string      `json:"msg_type"` //消息类型car_num=车辆数，device_num=设备数，alarm=警报
		MsgData interface{} `json:"msg_data"` //消息实体数据
	}{msgType, msgData}
	jsonData, err := json.Marshal(msg)
	if err == nil {
		this.client.EmitMessage(jsonData) //发送客户端
	}
}

func (this *WsClient) OnDisconnect() {
	this.disconnected = true
	log.Printf("\n连接 ID: %s 已经断开 %s !", this.client.Context().Request().Host, this.client.ID())
}

func (this *WsClient) GetLastCarLoc() {
	for !this.disconnected {
		data, OK := <-app.CarDataChan
		if !OK {
			time.Sleep(1 * time.Second)
			continue
		}
		if this.Session == nil {
			break
		}
		flag := models.IsCanCheck(data.CarGroupId, data.CarCompanyId, this.Session.User)
		if flag {
			this.Result("last_car_loc", data)
		}
	}
}

func (this *WsClient) GetLastAlarm() {
	for !this.disconnected {
		data, OK := <-app.AlarmDataChan
		if !OK {
			time.Sleep(1 * time.Second)
			continue
		}
		if this.Session == nil {
			break
		}
		flag := models.IsCanCheck(data.AnGroupId, data.AnCompanyId, this.Session.User)
		if flag {
			this.Result("last_alarm", data)
		}
	}
}
