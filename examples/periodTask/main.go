/*
 * @Author: RockyHoo
 * @Date: 2021-07-25 12:32:00
 * @LastEditTime: 2021-07-25 17:22:20
 * @LastEditors: Please set LastEditors
 * @Description: 测试用户定义的周期函数
 * @FilePath: /ReactLoop/reactloop.go
 */
package main

import (
	"Reactloop"
	"Reactloop/EventLoop"
	"fmt"
	"time"
)

var count = 0

func periodTask(el *EventLoop.EventLoop, _ *interface{}) {
	fmt.Printf("%d seconds passed\n", count)
	count += 30
}

func main() {
	t := EventLoop.UserEvent{
		Task:     periodTask,
		Interval: 3 * time.Second,
	}
	server := Reactloop.NewServer()
	server.AddUserEvent(&t)
	if err := server.StartServe(); err != nil {
		panic(err)
	}
}
