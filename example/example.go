package main

import (
	"fmt"
	"log"
	"time"

	//"github.com/bean0815/zzserver"
	"gitee.com/douyaye/zzserver"
)

func main() {

	//开启服务
	srv := zzserver.NewZZServer()
	srv.SetRouter(&P{}) //绑定路由接口
	srv.WsPort = 9999   //websocket端口
	//srv.TcpPort = 9988  //不设置就不会启动监听
	srv.Start()
}

type P struct {
	zzserver.BaseRouter
}

// OnMessage 接收客户端发送的消息
func (p *P) OnMessage(c *zzserver.Client, message []byte) {
	if string(message) == "close" {
		c.Server.SendToAll([]byte(fmt.Sprintf("user%d closed", c.ConnectionIndex)))
		c.Close()
	} else {
		c.Server.SendToAll([]byte(fmt.Sprintf("user%d say: %s", c.ConnectionIndex, string(message))))
	}
}

// OnDisconnect 客户端断开
func (p *P) OnDisconnect(c *zzserver.Client) {
}
func (p *P) OnServerClose() {
	log.Println("OnServerClose.....")
	time.Sleep(3 * time.Second)
	log.Println("OnServerClose.....")
}
