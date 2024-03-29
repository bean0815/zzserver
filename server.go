package zzserver

import (
	"encoding/json"
	"fmt"
	"time"

	//"gitee.com/douyaye/zzserver/zztools"
	"log"
	"net/http"
	"os"
	"os/signal"
)

// Server maintains the set of active clients and broadcasts messages to the
// clients.
type Server struct {
	router IRouter

	//监听的端口
	WsPort  int
	TcpPort int

	connected       chan *Client              //刚进入的,放入管道
	disconnected    chan *Client              //退出的,放入管道
	clients         map[int]*Client           // [登陆序号]客户端
	onlineNumber    int                       //在线人数
	broadcast       chan []byte               //广播消息
	ConnectionIndex int64                     //连接计数
	chanRange       chan func(c *Client) bool //
}

//ZZServer 单例
//var ZZServer *Server

func NewZZServer() *Server {
	s := &Server{
		clients:      make(map[int]*Client),
		broadcast:    make(chan []byte, 64),
		connected:    make(chan *Client, 64),
		disconnected: make(chan *Client, 64),
		chanRange:    make(chan func(c *Client) bool, 64),
	}
	go s.run()
	return s
}

//func init() {
//	ZZServer = &Server{
//		broadcast:          make(chan []byte, 64),
//		login:              make(chan *Client, 64),
//		connected:           make(chan *Client, 64),
//		disconnected:         make(chan *Client, 64),
//		ClientsNotloggedin: make(map[int]*Client),
//	}
//	go ZZServer.run()
//
//	InitLogs()
//}

// run 开始
func (h *Server) run() {
	defer func() {
		panic("hub run 方法运行终止 , 程序结束退出!")
	}()

	for {
		select {
		case client := <-h.connected: //用户连接
			h.clients[client.ConnectionIndex] = client
			//log.Println("用户 进入:"+client.GetRemoteAddr()+",当前登录:", h.ClientsNum, ",未登录:", len(h.ClientsNotloggedin))
			h.onlineNumber++
		case client := <-h.disconnected: //用户断开连接
			client.Close()
			if _, ok := h.clients[client.ConnectionIndex]; ok {
				delete(h.clients, client.ConnectionIndex)
				h.onlineNumber--
			}
			// log.Println("当前人数:", h.ClientsNum)
			// log.Print("用户 断开:"+client.GetRemoteAddr()+",当前登录:", len(h.clients), "未登录:", len(h.ClientsNotloggedin), "\r\n")
		case message := <-h.broadcast: //广播消息
			//var socketMsg = packet(message) //socket先封装, 再发送
			for _, client := range h.clients {
				//client.sendByteWithNoPacket(message, socketMsg)
				client.SendByte(message)
			}
		case f := <-h.chanRange:
			for _, client := range h.clients {
				if !f(client) {
					break
				}
			}
		}
	}
}

func (h *Server) SetRouter(r IRouter) {
	h.router = r
}

//SendToAll 广播消息
func (h *Server) SendToAll(message []byte) {
	h.broadcast <- message
}

//SendToAllJson 广播消息
func (h *Server) SendToAllJson(obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		return
	}
	h.broadcast <- b
}

//InitLogs  初始化日志
//func InitLogs() {
//	logpath := zzcfg.GoFileDir + "logs/test.log"
//	fmt.Println("日志路径:", logpath)
//	// logs.SetLogger("console")
//	// logs.SetLogger(logs.AdapterFile, `{"filename":"logs/test.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"color":true}`)
//	// logs.SetLogger(logs.AdapterMultiFile, `{"filename":"logs/log.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}`)
//	_ = logs.SetLogger(logs.AdapterMultiFile, `{"filename":"`+logpath+`","separate":["error", "info", "debug"]}`)
//	logs.Async()
//	logs.EnableFuncCallDepth(true)
//	logs.SetLogFuncCallDepth(3)
//
//	// l := logs.GetLogger()
//	// l.Println("this is a message of http")
//	//an official log.Logger with prefix ORM
//	// logs.GetLogger("ORM").Println("this is a message of orm")
//
//	// logs.Debug("my book is bought in the year of ", 2016)
//	// logs.Info("this %s cat is %v years old", "yellow", 3)
//	// logs.Warn("json is a type of kv like", map[string]int{"key": 2016})
//	// logs.Error(1024, "is a very", "good game")
//	// logs.Critical("oh,crash")
//
//	// beego.SetLogger("file", `{"filename":"logs/test.log"}`)
//	// logs.SetLogger(logs.AdapterFile, `{"filename":"project.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"color":true}`)
//
//	// beego.SetLevel(beego.LevelInformational)
//	// beego.SetLogFuncCall(true)
//	// beego.Info("this is test info")
//}

// startTcpSocket 开启socket监听
func (h *Server) startTcpSocket() {
	//开启普通socket监听
	serverSocket(h, fmt.Sprintf(":%d", h.TcpPort))
}

// startWebsocketAndHttp 开启websocket和http监听
func (h *Server) startWebsocketAndHttp() {
	//开启普通websocket监听
	defer func() {
		log.Fatalln("websocket监听 已经退出!")
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			return
		}
		serveWs(h, w, r)
	})
	err := http.ListenAndServe(fmt.Sprintf(":%d", h.WsPort), nil)
	if err != nil {
		log.Fatal(err)
	}

	//var err error
	//if !zzcfg.Wss {
	//	err = http.ListenAndServe(websocketaddr, nil)
	//} else {
	//	//err := http.ListenAndServeTLS(":443", "server.crt", "server.pem", nil)
	//	log.Println("server run WSS")
	//	err = http.ListenAndServeTLS(websocketaddr, zzcfg.Wss_cert, zzcfg.Wss_key, nil)
	//}
}

func (h *Server) Start() {
	if h.router == nil {
		log.Fatalln("please set router use server.SetRouter()")
		return
	}
	//defer func() {
	//	h.Close()
	//	time.Sleep(time.Second)
	//}()
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	//zztools.PrintServerIps()
	if h.WsPort > 0 {
		log.Println("Websocket server at:", h.WsPort)
		go h.startWebsocketAndHttp() //开启普通websocket监听
	}
	if h.TcpPort > 0 {
		fmt.Println("监听TCP：", h.TcpPort)
		go h.startTcpSocket()
	}
	log.Println("zzServer start success")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	log.Println("Got signal", s, "程序执行结束 执行Close()")
	h.Close()
	time.Sleep(time.Second)
	log.Println("服务器关闭, 再见~ ")
}

func (h *Server) Close() {
	if h.router != nil {
		h.router.OnServerClose()
	}

}

func (h *Server) Online() int {
	return h.onlineNumber
}

//Range 在同一个go中执行, 不要执行很久的操作
func (h *Server) Range(f func(c *Client) bool) {
	select {
	case h.chanRange <- f:
	default:
		log.Println("(h *Server) Range ERROR:  chan is full")
		//return errors.New("chan is full")
	}
}

//关闭服务器
func closeserver() {
	//ZZServer.IsClosing = true
	//router.OnServerClose()
	//log.Println("服务端 等待数据库队列完成!!")
	//time.Sleep(1 * time.Second)
	//zzdbhelp.Workor_end()
	//ZZServer.wg_dbwork.Wait() //等待数据库执行完成
	//log.Println("服务端 执行完毕退出!!")
	//os.Exit(0)
}

//func (h *Server) WaiteToClose() {
//	log.Println("WaiteToClose 等待程序执行完毕后退出")
//	if ZZServer.CanCloseImmediatelyGame {
//		ZZServer.NeedClose = true
//		ZZServer.Wg_Close.Done()
//	} else {
//		ZZServer.NeedClose = true
//	}
//}
