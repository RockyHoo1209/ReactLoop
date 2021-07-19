/*
 * @Description: Socket的封装类(模式perthred perloop)
 * @Author: Rocky Hoo
 * @Date: 2021-07-15 12:48:10
 * @LastEditTime: 2021-07-18 08:18:38
 * @LastEditors: Please set LastEditors
 * @CopyRight: XiaoPeng Studio
 * Copyright (c) 2021 XiaoPeng Studio
 */
package socket

import (
	"main/EventLoop"
	enum "main/Utils/Enum"
	err "main/Utils/Error"
	"net"
	"strconv"
	"syscall"
)

type Socket struct {
	network, address string //network:tcp/udp address:ip adress
	port             int
	sa               syscall.Sockaddr //特定格式的ip地址格式?
	in, out          []byte           //输入输出队列
	closedCount      int              //socket关闭连接计数器,shutdown只关闭一端,而close关闭读写两端
	fd               int              //socket对应的文件描述符
}

/**
 * @description:解析格式形如(host:port)的地址
 * @param {string} addr
 * @return {*}
 */
func parseIpv4Addr(addr string) (net.IP, int, error) {
	ipStr, portStr, errs := net.SplitHostPort(addr)
	if errs != nil {
		return nil, -1, errs
	}
	// convert ip to 4-bytes
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return nil, -1, &err.IP_FORMAT_ERR{
			IP: ipStr,
		}
	}
	port, errs := strconv.Atoi(portStr)
	if errs != nil {
		return nil, -1, errs
	}
	return ip, port, nil
}

/**
 * @description:根据sockaddr反向解析出socket的ip和地址
 * @param {syscall.Sockaddr} sa
 * @return {*}
 */
func resolveSockaddrInfo(sa syscall.Sockaddr) (string, string, int, error) {
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		return "tcp4", net.IP(v.Addr[:]).String(), v.Port, nil
	}
	return "", "", -1, &err.UNKNOW_NETWORK_ERR{
		Network: "unknown",
	}
}

/**
 * @description:将ip地址转换成对应的网络协议的Sockaddr对象
 * @param  {*}
 * @return {*}
 * @param {*} network
 * @param {string} addr
 */
func getSockAddr(network, addr string) (int, error) {
	return 0, nil
}

/**
 * @description: Socket构造函数
 * @param  {*}
 * @return {*}
 * @param {*} network
 * @param {string} addr(format:ip:port)
 * @param {int} port
 */
func NewSocket(network, addr string) (*Socket, error) {
	// AF_INET Socket地址族;proto设置为0，选择系统默认协议族
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	sa, err := getSockAddr(network, addr)
	var portStr string
	addr, portStr, err = net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	return &Socket{
		network:     network,
		address:     addr,
		port:        port,
		sa:          sa,
		in:          []byte{},
		out:         []byte{},
		closedCount: 0,
		fd:          fd,
	}, nil
}

/**
 * @description: 关闭socket(读写端都关闭),socket引用数减1，直到减到0才关闭socket
 * @param  {*}
 * @return {*}
 */
func (s *Socket) Close() error {
	s.closedCount = 2
	return syscall.Close(s.fd)
}

/**
 * @description: 可以关闭读写中的一端,并且不用等待引用计数减到0
 * @param  {*}
 * @return {*}
 */
func (s *Socket) Shutdown(how int) error {
	s.closedCount += 1
	if s.closedCount == 2 {
		s.Close()
	}
	return syscall.Shutdown(s.fd, how)
}

// Listener是Socket的一个装饰器m主要负责连接创立过程的响应处理
type Listener struct {
	*Socket
}

/**
 * @description: Listener构造函数
 * @param  {*}
 * @return {*}
 * @param {*} network
 * @param {string} addr
 */
func NewListener(network, addr string) (*Listener, error) {
	sock, err := NewSocket(network, addr)
	if err != nil {
		return nil, err
	}
	return &Listener{sock}, nil
}

/**
 * @description: 对应服务器建立socket连接后的bind\listen
 * @param  {*}
 * @return {*}
 */
func (l *Listener) BindAndListen() error {
	err := syscall.Bind(l.fd, l.sa)
	if err != nil {
		_ = l.Close()
		return err
	}
	// 第二个参数(backlog)为max(未完成连接队列容量，已完成连接队列容量)
	err = syscall.Listen(l.fd, 1024)
	if err != nil {
		l.Close()
		return err
	}
	return nil
}

/**
 * @description:处理accept事件后的操作，并返回对应的事件
 * @param  {*}
 * @return {*}
 * @param {*EventLoop.EventLoop} el
 * @param {interface{}} data
 */
func (l *Listener) acceptEvent(el *EventLoop.EventLoop, data interface{}) enum.Action {
	// l.fd为socket的监听套接字，整个服务器socket运行时只有一份,nfd为已连接套接字，即每次accept取出一个可用连接后都会返回一个nfdnfd对应的是
	nfd, sa, err := syscall.Accept(l.fd)
	if err != nil {
		return enum.CONTINUE
	}
	if err = syscall.SetNonblock(nfd, true); err != nil {
		syscall.Close(nfd)
		return enum.CONTINUE
	}
	c, err := NewConn(nfd, sa, el)
	if err != nil {
		return enum.CONTINUE
	}
	el.RegisterEvent(c.fd, enum.EVENT_READABLE, c.readEvent, nil)
	el.SetTrigerDataPtr([]string{c.network, c.addr, strconv.Itoa(c.port)})
	return enum.TRIGGER_OPEN_EVENT
}

/**
 * @description:将accept事件注册到事件循环中 每次新建一个连接accept 满足per-thread per-loop
 * @param  {*}
 * @return {*}
 */
func (l *Listener) RegisterAccept() error {
	event_loop := EventLoop.New()
	return event_loop.RegisterEvent(l.fd, enum.EVENT_READABLE, l.acceptEvent, nil)
}

// socket的装饰器,主要负责数据读写的工作
type Conn struct {
	*Socket
}

/**
 * @description: Conn构造函数
 * @param  {*}
 * @return {*}
 * @param {int} fd
 * @param {syscall.Sockaddr} sa
 * @param {*EventLoop.EventLoop} event_loop(与accept操作共享一个eventloop)
 */
func NewConn(fd int, sa syscall.Sockaddr, event_loop *EventLoop.EventLoop) (*Conn, error) {
	network, addr, port, err := resolveSockaddrInfo(sa)
	if err != nil {
		return nil, err
	}
	conn := &Conn{&Socket{
		network:     network,
		address:     addr,
		port:        port,
		sa:          sa,
		in:          []byte{},
		out:         []byte{},
		closedCount: 0,
		fd:          fd,
	}}
	return conn, nil
}
