/*
 * @Author: RockyHoo
 * @Date: 2021-07-25 12:32:00
 * @LastEditTime: 2021-07-25 12:53:59
 * @LastEditors: Please set LastEditors
 * @Description: 测试用户定义的周期函数
 * @FilePath: /ReactLoop/main.go
 */
package main

import (
	"fmt"
	"main/EventLoop"
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
	server := NewServer()
	server.AddUserEvent(&t)
	if err := server.StartServe(); err != nil {
		panic(err)
	}
}
