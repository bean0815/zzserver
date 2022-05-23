package zzserver

/*
//方法1 回调
type CallBack func(c *Client, message []byte)

var callbacck CallBack

func actionAll(c *Client, message []byte) {
	callbacck(c, message)
}
*/

//IRouter 方法2: 接口
type IRouter interface {
	OnMessage(c *Client, message []byte) // OnMessage 用户协程中调用
	OnDisconnect(c *Client)              // OnDisconnect hub协成中调用
	OnServerClose()                      // 关闭服务器执行一次
	OnServerUpdate()
	OnConnected(c *Client)
}

var Router IRouter

func AddRouter(r IRouter) {
	Router = r
}

//BaseRouter 用于重写, 这样就不需要写出所有方法
type BaseRouter struct{}

func (b *BaseRouter) OnMessage(c *Client, message []byte) {}
func (b *BaseRouter) OnDisconnect(c *Client)              {}
func (b *BaseRouter) OnServerClose()                      {}
func (b *BaseRouter) OnServerUpdate()                     {}
func (b *BaseRouter) OnConnected(c *Client)               {}
