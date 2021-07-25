/*
 * @Author:RockyHoo
 * @Date: 2021-07-25 12:20:11
 * @LastEditTime: 2021-07-25 12:28:25
 * @LastEditors: Please set LastEditors
 * @Description: 测试reactloop使用(ping-pong)
 * @FilePath: /ReactLoop/main.go
 */
package main

import (
	"fmt"
	"main/EventLoop"
	"main/Socket"
)

func whenServing(el *EventLoop.EventLoop, _ *interface{}) {
	fmt.Println("Server start...")
}

func whenAccept(el *EventLoop.EventLoop, dataPtr *interface{}) {
	data := (*dataPtr).([]string)
	fmt.Println("Accept: ", data)
}

func echo(el *EventLoop.EventLoop, connPtr *interface{}) {
	conn := (*connPtr).(*Socket.Conn)
	msg := conn.Read()
	msgstr := string(msg)
	fmt.Println("Recv: ", msgstr)
	msgstr += " pong"
	conn.Write([]byte(msgstr))
}

func main() {
	litener, _ := Socket.NewListener("tcp4", "127.0.0.1:9090")
	server := NewServer()
	events := &EventLoop.Event{
		Open:    whenAccept,
		Serving: whenServing,
		Data:    echo,
	}
	server.AddListener(litener)
	server.AddSystemEvent(events)
	server.StartServe()
}
