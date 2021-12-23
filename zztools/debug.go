package zztools

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"
)

// Debug_CallInfo 文件执行行号
func Debug_CallInfo() string {
	pc, file, line, ok := runtime.Caller(4)
	if !ok {
		return ""
	}
	//[文件 函数 行号]
	funcname := path.Base(runtime.FuncForPC(pc).Name())
	filename := path.Base(file)
	return fmt.Sprintf("[%s %d] [%s]", filename, line, funcname)
}

// PrintServerIps 打印服务端IP
func PrintServerIps() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ipads := "服务器IP:"
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipads += (ipnet.IP.String() + "  ")
			}
		}
	}
	fmt.Println(ipads)
}
