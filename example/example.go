package main

import (
	"fmt"
	server "gitee.com/douyaye/zz-server/zzserver"
)

func main() {

	//设置接收路由
	server.AddRouter(&P{})

	//开启服务
	srv := server.NewZZServer()
	srv.WsPort = 9999 //
	//srv.TcpPort = 9988  //不设置就不会启动监听
	srv.Start()
}

type P struct {
	server.BaseRouter
}

//ActionAll 接收客户端发送的消息
func (p *P) ActionAll(c *server.Client, message []byte) {
	if string(message) == "close" {
		c.Close()
	} else {
		//c.SendJson(map[string]interface{}{"code": 0, "Message": "I received your message! thanks!"})
		c.SendText(fmt.Sprintf("当前在线人数：%d", c.Server.Online()))
	}
}

// Disconnect 客户端断开
func (p *P) Disconnect(c *server.Client) {

}

// BeforeServerClose 服务器要关闭的时候
func (p *P) BeforeServerClose() {

}
