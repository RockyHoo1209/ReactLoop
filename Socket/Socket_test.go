/*
 * @Author: your name
 * @Date: 2021-07-20 15:31:58
 * @LastEditTime: 2021-07-24 23:04:21
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /ReactLoop/Socket/Socket_test.go
 */
package socket

import (
	"fmt"
	"main/EventLoop"
	"testing"
)

func whenServing(el *EventLoop.EventLoop, _ *interface{}) {
	fmt.Println("Server start...")
}

func whenAccept(el *EventLoop.EventLoop, dataPtr *interface{}) {
	data := (*dataPtr).([]string)
	fmt.Println("Accept: ", data)
}

func echo(el *EventLoop.EventLoop, connPtr *interface{}) {
	conn := (*connPtr).(*Conn)
	msg := conn.Read()
	msgstr := string(msg)
	fmt.Println("Recv: ", msgstr)
	msgstr += " pong"
	conn.Write([]byte(msgstr))
}

func TestSocket(*testing.T) {
	event := EventLoop.Event{
		Serving: whenServing,
		Open:    whenAccept,
		Data:    echo,
	}
	listener, err := NewListener("tcp4", "127.0.0.1:9090")
	if err != nil {
		println("error:", err)
		return
	}
	listener.BindAndListen()
	el := EventLoop.New()
	el.AddSystemEvent(&event)
	listener.RegisterAccept(el)

	el.Run()
}
