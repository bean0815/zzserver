package main

import (
	"fmt"
	"github.com/bean0815/zzserver"
)

func main() {

	//设置接收路由
	zzserver.AddRouter(&P{})

	//开启服务
	srv := zzserver.NewZZServer()
	srv.WsPort = 9999 //
	//srv.TcpPort = 9988  //不设置就不会启动监听
	srv.Start()
}

type P struct {
	zzserver.BaseRouter
}

//ActionAll 接收客户端发送的消息
func (p *P) ActionAll(c *zzserver.Client, message []byte) {
	if string(message) == "close" {
		c.Close()
	} else {
		//c.SendJson(map[string]interface{}{"code": 0, "Message": "I received your message! thanks!"})
		c.SendText(fmt.Sprintf("当前在线人数：%d", c.Server.Online()))
	}
}

// Disconnect 客户端断开
func (p *P) Disconnect(c *zzserver.Client) {
}

// BeforeServerClose 服务器要关闭的时候
func (p *P) BeforeServerClose() {

}
