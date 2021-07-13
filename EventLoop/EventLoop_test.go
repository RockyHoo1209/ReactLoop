/*
 * @Description: EventLoop测试模块
 * @Author: Rocky Hoo
 * @Date: 2021-07-12 15:59:49
 * @LastEditTime: 2021-07-13 16:10:24
 * @LastEditors: Please set LastEditors
 * @CopyRight: XiaoPeng Studio
 * Copyright (c) 2021 XiaoPeng Studio
 */
package EventLoop

import (
	"fmt"
	enum "main/Utils/Enum"
	"testing"
	"time"
)

func Open(el *EventLoop, data interface{}) {
	fmt.Println("Opening...", data)
}

func Closed(el *EventLoop, data interface{}) {
	fmt.Println("Closing...", data)
}

func Data(el *EventLoop, data interface{}) {
	fmt.Println("Data:", data)
}

func Serving(el *EventLoop, data interface{}) {
	fmt.Println("Serving")
}

var cnt=1
func EventProcess(el *EventLoop, data interface{}) enum.Action{
	cnt+=1
	fmt.Println("EventProcess",data,"cnt:",cnt)
	return enum.CONTINUE
}

func TestEventLoop(*testing.T) {
	eventLoop1 := New()
	eventLoop2 := New()
	event := &Event{
		Open:    Open,
		Close:   Closed,
		Serving: Serving,
		Data:    Data,
	}
	eventLoop1.RegisterEvent(1,enum.EVENT_READABLE,EventProcess,"hello1")
	eventLoop1.system_events = append(eventLoop1.system_events, event)
	go eventLoop1.Run()

	eventLoop2.RegisterEvent(1,enum.EVENT_READABLE,EventProcess,"hello2")
	eventLoop2.system_events = append(eventLoop2.system_events, event)
	go eventLoop2.Run()
	time.Sleep(10*time.Second)
	eventLoop1.Done()
	eventLoop2.Done()
	eventLoop1.UnRegister(1,enum.EVENT_READABLE)
	eventLoop2.UnRegister(1,enum.EVENT_READABLE)
}
