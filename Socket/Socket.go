/*
 * @Description: Socket的封装类(模式perthred perloop)
 * @Author: Rocky Hoo
 * @Date: 2021-07-15 12:48:10
 * @LastEditTime: 2021-07-25 11:32:43
 * @LastEditors: Please set LastEditors
 * @CopyRight: XiaoPeng Studio
 * Copyright (c) 2021 XiaoPeng Studio
 */
package Socket

import (
	"main/EventLoop"
	enum "main/Utils/Enum"
	err "main/Utils/Error"
	"net"
	"strconv"
	"syscall"
)

/**
 * @description:连接套接字和监听套接字都复用此结构体(注意连接套接字和监听套接字的区别)
 * @param {*}
 * @return {*}
 */
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
func getSockAddr(network, addr string) (syscall.Sockaddr, error) {
	switch network {
	case "tcp4":
		ip, port, errors := parseIpv4Addr(addr)
		if errors != nil {
			return nil, errors
		}
		sa := &syscall.SockaddrInet4{Port: port}
		copy(sa.Addr[:], ip[:4])
		return sa, nil
	}
	return nil, &err.UNKNOW_NETWORK_ERR{
		Network: network,
	}
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
	if err := syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	sa, err := getSockAddr(network, addr)
	if err != nil {
		return nil, err
	}
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

// Listener是Socket的一个装饰器m主要负责连接创立过程的响应处理(监听套接字)
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
	confd, sa, err := syscall.Accept(l.fd)
	if err != nil {
		return enum.CONTINUE
	}
	if err = syscall.SetNonblock(confd, true); err != nil {
		syscall.Close(confd)
		return enum.CONTINUE
	}
	c, err := NewConn(confd, sa, el)
	if err != nil {
		return enum.CONTINUE
	}
	el.RegisterEvent(c.fd, enum.EVENT_READABLE, c.readEvent, nil)
	// accept的时候通过修改trigerPtr输出对应信息
	el.SetTrigerDataPtr([]string{c.network, c.address, strconv.Itoa(c.port)})
	return enum.TRIGGER_OPEN_EVENT
}

/**
 * @description:将accept事件注册到事件循环中 每次新建一个连接accept 满足per-thread per-loop
 * @param  {*}
 * @return {*}
 */
func (l *Listener) RegisterAccept(event_loop *EventLoop.EventLoop) error {
	return event_loop.RegisterEvent(l.fd, enum.EVENT_READABLE, l.acceptEvent, nil)
}

// socket的装饰器,主要负责数据读写的工作(此为连接套接字,即其中维护的是连接描述符,每与一个客户端建立连接就会创建一个连接套接字)
type Conn struct {
	*Socket
}

/**
 * @description: Conn构造函数
 * @param  {*}
 * @return {*}
 * @param {int} confd socket与客户端的一个连接描述符
 * @param {syscall.Sockaddr} sa
 * @param {*EventLoop.EventLoop} event_loop(与accept操作共享一个eventloop)
 */
func NewConn(confd int, sa syscall.Sockaddr, event_loop *EventLoop.EventLoop) (*Conn, error) {
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
		fd:          confd,
	}}
	return conn, nil
}

/**
 * @description:提供给上层调用的api;通过readEvent将数据读到Conn.in后,从Conn.in中读出去
 * @param {*}
 * @return {*}
 */
func (c *Conn) Read() []byte {
	res := c.in
	c.in = []byte{}
	return res
}

/**
 * @description:将数据写入到c.out，供socket发出
 * @param {[]byte} data
 * @return {*}
 */
func (c *Conn) Write(data []byte) {
	c.out = append(c.out, data...)
}

/**
 * @description:执行一次读操作
 * @param {*EventLoop.EventLoop} el
 * @param {interface{}} data
 * @return {*}
 */
func (c *Conn) readEvent(el *EventLoop.EventLoop, _ interface{}) enum.Action {
	var (
		action enum.Action
		inBuf  = [1024]byte{}
	)
	n, err := syscall.Read(c.fd, inBuf[:])
	//以下报错都是非阻塞操作中可以忽略的错误,参考:https://www.cnblogs.com/bastard/archive/2013/04/10/3012724.html
	if err == syscall.EINTR || err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
		action = enum.CONTINUE
	} else if n <= 0 {
		// n小于0,说明此时收到对端发来的关闭信号
		c.Shutdown(syscall.SHUT_RD)
		action = enum.SHUTDOWN_RD
	} else {
		// inBuf切片被打散传入
		c.in = append(c.in, inBuf[:n]...)
		// 将连接socket的指针存入eventloop,则可以通过这个指针访问conn(委托模式)
		el.SetTrigerDataPtr(c)
		action = enum.TRIGGER_DATA_EVENT
	}
	// 读时间注册完后注册监听写事件
	if c.closedCount == 0 {
		el.RegisterEvent(c.fd, enum.EVENT_WRITABLE, c.writeEvent, nil)
	}
	return action
}

/**
 * @description:执行一次socket写事件
 * @param {*EventLoop.EventLoop} el
 * @param {interface{}} _
 * @return {*}
 */
func (c *Conn) writeEvent(el *EventLoop.EventLoop, _ interface{}) enum.Action {
	var action enum.Action
	if len(c.out) == 0 {
		return enum.CONTINUE
	}
	n, err := syscall.Write(c.fd, c.out)
	if err != nil || n <= 0 {
		c.Shutdown(syscall.SHUT_WR)
		action = enum.SHUTDOWN_WR
	} else {
		//读了前面部分数据,剩下的数据从n开始读
		c.out = c.out[n:]
	}
	// 需要再用读事件覆盖写事件
	if c.closedCount == 0 {
		el.RegisterEvent(c.fd, enum.EVENT_READABLE, c.readEvent, nil)
	}
	return action
}
