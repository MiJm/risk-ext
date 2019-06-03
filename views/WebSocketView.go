package views

import (
	"log"

	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

type WebSocketView struct {
	Views
	ws *websocket.Server
}

func NewWs() *WebSocketView {
	ws := new(WebSocketView)
	ws.InitSocket()
	return ws
}
func (this *WebSocketView) Get(ctx iris.Context) (int, interface{}) {
	return 200, this.ws.Handler()
}
func (this *WebSocketView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"GET": MA{"NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *WebSocketView) InitSocket() {
	this.ws = websocket.New(websocket.Config{
	// to enable binary messages (useful for protobuf):
	// BinaryMessages: true,
	})

	this.ws.OnConnection(func(c websocket.Connection) {

		c.OnMessage(func(data []byte) {
			this.OnMessage(data, c) // writes to itself
		})

		c.OnDisconnect(func() {
			this.OnDisconnect(c)
		})

	})
}

func (this *WebSocketView) OnMessage(data []byte, c websocket.Connection) {
	message := string(data)
	c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message)) // broadcast to all clients except this
	c.EmitMessage([]byte("Me: " + message))
}

func (this *WebSocketView) OnDisconnect(c websocket.Connection) {
	log.Printf("\nConnection with ID: %s has been disconnected!", c.ID())
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
