package main

import (
	"fmt"
	"gitee.com/douyaye/zz-server"
)

func main() {

	//设置接收路由
	zzServer.AddRouter(&P{})

	//开启服务
	srv := zzServer.NewZZServer()
	srv.WsPort = 9999 // 
	//srv.TcpPort = 9988  //不设置就不会启动监听
	srv.Start()
}

type P struct {
	zzServer.BaseRouter
}

//ActionAll 接收客户端发送的消息
func (p *P) ActionAll(c *zzServer.Client, message []byte) {
	if string(message) == "close" {
		c.Close()
	} else {
		//c.SendJson(map[string]interface{}{"code": 0, "Message": "I received your message! thanks!"})
		c.SendText(fmt.Sprintf("当前在线人数：%d", c.Server.Online()))
	}
}

// Disconnect 客户端断开
func (p *P) Disconnect(c *zzServer.Client) {

}

// BeforeServerClose 服务器要关闭的时候
func (p *P) BeforeServerClose() {

}
