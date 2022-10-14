package zzserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 51200
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	//socket类型
	typeWebsocket int = 1
	typeSocket    int = 2
)

var (
	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 5) / 10

	heartbeatMsg []byte = []byte(`{"t":"heartbeat"}`)
)

// var (
// 	newline = []byte{'\n'}
// 	space   = []byte{' '}
// )

var upgrader = websocket.Upgrader{
	ReadBufferSize:  51200,
	WriteBufferSize: 51200,

	// 解决跨域问题
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 保存客户端连接和用户
type Client struct {
	socketType      int
	connWebsocket   *websocket.Conn
	connSocket      net.Conn
	bufSend         chan []byte
	chanClose       chan bool
	Server          *Server     //server对线
	User            interface{} //用户对象
	UserData        sync.Map    //保存一些用户自定义内容
	ConnectionIndex int         //
	LastMsgTime     time.Time   //最后收到消息时间
	Ip              string      //登录IP
	//closed          bool        //已经关闭状态
}

//处理消息
func actionAll(c *Client, message []byte) {
	//if ZZServer.IsClosing {
	//	message = []byte("-1")
	//}
	//if ZZServer.IsUpdate {
	//	message = []byte("-2")
	//}
	if len(message) == 0 {
		fmt.Println("len(message) == 0")
		return
	}

	//处理消息
	defer func() {
		if r := recover(); r != nil {
			log.Println("actionAll 发生了panic错误, 已经recover()错误内容:", r)
			log.Println("堆栈:", string(debug.Stack()))
		}
	}()

	c.LastMsgTime = time.Now()
	c.Server.router.OnMessage(c, message)
}

func disconnect(c *Client) {
	//处理消息
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnect 发生了panic错误, 已经recover()错误内容:", r)
			log.Println("堆栈:", string(debug.Stack()))
		}
	}()
	c.Server.router.OnDisconnect(c)
}

//ServeWs handles websocket requests from the peer.
func serveWs(hub *Server, w http.ResponseWriter, r *http.Request) {
	connWebsocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// log.Println("新用户连接websocket错误!", err)
		return
	}

	forwardedIp := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	realIp := r.Header.Get("X-real-ip")
	remoteIp, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	ip := forwardedIp
	if strings.Contains(ip, "127.0.0.1") || ip == "" {
		ip = realIp
		if strings.Contains(ip, "127.0.0.1") || ip == "" {
			ip = remoteIp
		}
		if ip == "" {
			ip = "127.0.0.1"
		}
	}

	//log.Println("r.RemoteAddr:", r.RemoteAddr)
	//log.Println("X-Forwarded-For:", r.Header.Get("X-Forwarded-For"))
	//log.Println("X-real-ip:", r.Header.Get("X-real-ip"))

	//log.Println("新websocket用户连接", r.RemoteAddr)
	//ip := strings.Split(r.RemoteAddr, ":")[0]

	cIndex := atomic.AddInt64(&hub.ConnectionIndex, 1)

	client := &Client{
		Server:          hub,
		connWebsocket:   connWebsocket,
		bufSend:         make(chan []byte, 256),
		chanClose:       make(chan bool, 1),
		socketType:      typeWebsocket,
		ConnectionIndex: int(cIndex),
		LastMsgTime:     time.Now(),
		Ip:              ip,
	}
	client.Server.connected <- client

	go client.websocketWrite()
	go client.websocketRead()
}

// websocketWrite
func (c *Client) websocketWrite() {
	ticker1 := time.NewTicker(pingPeriod)
	defer func() {
		ticker1.Stop()
		_ = c.connWebsocket.Close()
	}()
	for {
		select {
		case message, ok := <-c.bufSend:
			c.connWebsocket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.connWebsocket.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("c.send_websocket close")
				return
			}

			w, err := c.connWebsocket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker1.C: // 心跳包
			c.connWebsocket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.connWebsocket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.chanClose:
			return
		}
	}
}

// websocketRead pumps messages from the websocket connection to the hub.
//
// The application runs websocketRead in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) websocketRead() {
	c.Server.router.OnConnected(c)
	ticker1 := time.NewTicker(pingPeriod)
	defer func() {
		//log.Println("websocketRead close", c.GetRemoteAddr())
		ticker1.Stop()
		disconnect(c)
		c.Server.disconnected <- c
		_ = c.connWebsocket.Close()
		//通知write
		select {
		case c.chanClose <- true:
		default:
		}
	}()
	c.connWebsocket.SetReadLimit(maxMessageSize)
	c.connWebsocket.SetReadDeadline(time.Now().Add(pongWait))
	c.connWebsocket.SetPongHandler(func(string) error { c.connWebsocket.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.connWebsocket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("server断开")
			}
			break
		}
		actionAll(c, message)
	}
}

//ServerSocket 连接
func serverSocket(hub *Server, socketaddr string) {
	// processor.ServerSocket(hub)
	netListen, err := net.Listen("tcp", socketaddr)
	if err != nil {
		log.Println("socket error!", err)
	}
	defer func() {
		netListen.Close()
		panic("ServerSocket 程序退出!")
	}()
	for {
		connTCP, err := netListen.Accept()
		if err != nil {
			continue
		}

		// fmt.Println(conn_socket.RemoteAddr().String(), " 连接成功")
		ip := strings.Split(connTCP.RemoteAddr().String(), ":")[0]
		//log.Println("新serverSocket用户连接:", connTCP.RemoteAddr().String())

		cIndex := atomic.AddInt64(&hub.ConnectionIndex, 1)

		//uuid, _ := uuid.NewV4()
		client := &Client{
			Server:     hub,
			connSocket: connTCP,
			//bufSocket:       make(chan []byte, 256),
			bufSend:         make(chan []byte, 256),
			chanClose:       make(chan bool, 1),
			socketType:      typeSocket,
			LastMsgTime:     time.Now(),
			Ip:              ip,
			ConnectionIndex: int(cIndex),
			//Uuid:        uuid,
		}
		client.Server.connected <- client
		go client.socketWrite()
		go client.socketRead()
	}
}

//socket读取数据
func (c *Client) socketRead() {
	//readerChannel := make(chan []byte, 8)//老的写法
	c.Server.router.OnConnected(c)
	defer func() {
		c.Server.disconnected <- c
		_ = c.connSocket.Close()
		disconnect(c)
		//通知write
		select {
		case c.chanClose <- true:
		default:
		}
	}()
	//新写法
	reader := bufio.NewReader(c.connSocket)
	for {
		data, err := unpack2(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}
		//fmt.Println("socket receive:", string(data))
		actionAll(c, data)
	}
}
func (c *Client) reader(readerChannel chan []byte) {
	//defer func() {
	//	log.Println("defer reader Close")
	//}()
	for {
		select {
		case data, ok := <-readerChannel:
			if !ok {
				return
			}
			// log.Println(string(data))
			actionAll(c, data)
		}
	}
}

//socket发送数据
func (c *Client) socketWrite() {
	ticker1 := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("socketWrite close")
		ticker1.Stop()
		c.connSocket.Close()
	}()
	for {
		select {
		case message, ok := <-c.bufSend:
			if !ok {
				// The hub closed the channel.
				return
			}
			c.connSocket.SetWriteDeadline(time.Now().Add(writeWait))
			_, err := c.connSocket.Write(message)
			if err != nil {
				log.Println("返回消息失败!", err)
				return
			}

		case t := <-ticker1.C: //心跳包
			c.connSocket.SetWriteDeadline(time.Now().Add(writeWait))
			_, err := c.connSocket.Write(packet([]byte(heartbeatMsg)))
			if err != nil {
				return
			}

			if t.Sub(c.LastMsgTime) > pongWait {
				// log.Print("心跳包断开")
				return
			}
		case <-c.chanClose:
			return
		}
	}
}

func (c *Client) Close() {
	// log.Println(" (c *Client) Close()")
	select {
	case c.chanClose <- true:
	default:

	}

	//if c.closed {
	//	return
	//}
	//c.closed = true

	//if c.connWebsocket != nil {
	//	c.connWebsocket.Close()
	//}
	//if c.connSocket != nil {
	//	c.connSocket.Close()
	//}
}

//GetRemoteAddr 获取地址
func (c *Client) GetRemoteAddr() string {
	if c.socketType == typeWebsocket {
		return c.connWebsocket.RemoteAddr().String()
	}
	return c.connSocket.RemoteAddr().String()
}

//SendText 发送消息 封装
func (c *Client) SendText(msg string) {
	// log.Println("返回消息:", msg)
	c.SendByte([]byte(msg))
}

//SendByte 发送byte, socket会进行编码后再发送
func (c *Client) SendByte(bytes []byte) {
	//c.sendByteWithNoPacket(bytes, packet(bytes))
	if c.socketType == typeSocket {
		bytes = packet(bytes)
	}
	select {
	case c.bufSend <- bytes:
	default:
		c.Close()
	}
}

//SendJson 先把obj转为json对象再发送
func (c *Client) SendJson(obj interface{}) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	//fmt.Println("#", string(bytes))
	//c.sendByteWithNoPacket(bytes, packet(bytes))
	c.SendByte(bytes)
	return nil
}

// sendByteWithNoPacket socket发送 不封装
//func (c *Client) sendByteWithNoPacket(bytesWs []byte, bytesSocket []byte) {
//	//一个Client中 send_websocket send_socket 只会存在一个
//	select {
//	case c.bufSend <- bytesWs:
//	case c.bufSocket <- bytesSocket:
//	default:
//		c.Close()
//	}
//}
