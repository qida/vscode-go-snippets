/*
 * @Author: sunqida
 * @Date: 2019-06-14 13:12:45
 * @LastEditors: sunqida
 * @LastEditTime: 2019-10-08 17:20:01
 * @Description:
 */
package logs

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/axgle/mahonia"
	"github.com/qida/tcp_server"
)

var (
	DebugList map[string]*tcp_server.Client
)
var (
	LogConn = logs.NewLogger(1000)
	LogMail = logs.NewLogger(1000)
	Enc     = mahonia.NewEncoder("gb18030")
)

func Server(port int) {
	go ServerTcp(port)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.Async(1e3)
	LogConn.SetLevel(logs.LevelDebug)
	LogConn.SetLogger(logs.AdapterConn, fmt.Sprintf(`{"net":"tcp","addr":"0.0.0.0:%d","reconnect":true}`, port))
}

func Email() {
	LogMail.Async()
	LogMail.EnableFuncCallDepth(true)
	err := LogMail.SetLogger(logs.AdapterMail, `{"level":7,"username":"sunqida@126.com","password":"","fromAddress":"sunqida@126.com","subject":"", "host":"smtp.126.com:994","sendTos":["sunqida@foxmail.com"]}`) //654/994
	if err != nil {
		panic(err.Error())
	}
	if beego.BConfig.RunMode == "dev" {
		LogMail.Notice("Api Test系统开始运行：%v", time.Now())
	} else {
		LogMail.Notice("Api Prod系统开始运行：%v", time.Now())
	}
}

func ServerTcp(port int) {
	fmt.Printf("调试 在 %d 监听...\r\n", port)
	DebugList = make(map[string]*tcp_server.Client)
	server := tcp_server.New(fmt.Sprintf("0.0.0.0:%d", port))
	// utf-8=>gb18030
	//dec := mahonia.NewDecoder("GB18030")
	// gb18030=>utf-8
	//enc := mahonia.Newutil.Encoder("GB18030")
	server.OnNewClient(func(c *tcp_server.Client) {
		c.Send(fmt.Sprintf("Welcome %s \n", c.GetConn().RemoteAddr().String()))
	})
	server.OnNewMessage(func(c *tcp_server.Client, message string) {
		if message == "debug\r\n" {
			DebugList[c.GetConn().RemoteAddr().String()] = c
			c.Send("Welcome Debugger\r\n")
			return
		}
		// 中文处理 //
		for _, v := range DebugList {
			v.Send(message)
		}
	})
	server.OnClientConnectionClosed(func(c *tcp_server.Client, err error) {
		fmt.Printf("调试端断开\r\n")
		delete(DebugList, c.GetConn().RemoteAddr().String())
	})
	server.Listen()
}
