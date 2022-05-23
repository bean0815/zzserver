# zzServer

该项目是一个简易的socket服务端程序，支持webSocket以及普通tcpSocket

``` 

//开启服务
srv := zzserver.NewZZServer()
srv.SetRouter(&P{}) //绑定路由接口
srv.WsPort = 9999   //websocket端口
//srv.TcpPort = 9988  //不设置就不会启动监听
srv.Start()
```


请参考 `example.go`



by. `douya`

有疑问请联系QQ：`330496906`