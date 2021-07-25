/*
 * @Author: Rocky Hoo
 * @Date: 2021-07-25 11:25:59
 * @LastEditTime: 2021-07-25 17:30:37
 * @LastEditors: Please set LastEditors
 * @Description: 启动reactloop服务器
 * @FilePath: /ReactLoop/ReactLoop.go
 */
package Reactloop

import (
	"Reactloop/EventLoop"
	"Reactloop/Socket"
)

type Server struct {
	el        *EventLoop.EventLoop
	listeners []*Socket.Listener
}

func NewServer() *Server {
	return &Server{
		el:        EventLoop.New(),
		listeners: make([]*Socket.Listener, 0),
	}
}

func (s *Server) AddListener(l *Socket.Listener) {
	s.listeners = append(s.listeners, l)
}

/**
 * @description:添加可以通过系统epoll触发执行的事件
 * @param {*EventLoop.Event} event
 * @return {*}
 */
func (s *Server) AddSystemEvent(event *EventLoop.Event) {
	s.el.AddSystemEvent(event)
}

/**
 * @description:添加用户自定义的定时执行任务
 * @param {*EventLoop.UserEvent} user_event
 * @return {*}
 */
func (s *Server) AddUserEvent(user_event *EventLoop.UserEvent) {
	s.el.AddUserEvent(user_event)
}

func (s *Server) CloseAllListener() {
	for _, listener := range s.listeners {
		listener.Close()
	}
}

/**
 * @description:启动服务器,如果有一个监听socket报错则全部关闭
 * @param {*}
 * @return {*}
 */
func (s *Server) StartServe() error {
	for _, l := range s.listeners {
		if err := l.BindAndListen(); err != nil {
			s.CloseAllListener()
			return err
		}
		if err := l.RegisterAccept(s.el); err != nil {
			s.CloseAllListener()
			return err
		}
	}
	s.el.Run()
	return nil
}
