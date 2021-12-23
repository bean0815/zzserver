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
	ActionAll(c *Client, message []byte) // ActionAll 用户协程中调用
	Disconnect(c *Client)                //Disconnect hub协成中调用
	BeforeServerClose()                  //关闭服务器执行一次
	UpdateServer()
}

var Router IRouter

func AddRouter(r IRouter) {
	Router = r
}

//BaseRouter 用于重写, 这样就不需要写出所有方法
type BaseRouter struct{}

func (b *BaseRouter) ActionAll(c *Client, message []byte) {}
func (b *BaseRouter) Disconnect(c *Client)                {}
func (b *BaseRouter) BeforeServerClose()                  {}
func (b *BaseRouter) UpdateServer()                       {}
