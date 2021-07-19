/*
 * @Description: 事件管理器
 * @Author: Rocky Hoo
 * @Date: 2021-07-01 10:21:37
 * @LastEditTime: 2021-07-17 09:01:10
 * @LastEditors: Please set LastEditors
 * @CopyRight:
 * Copyright (c) 2021 XiaoPeng Studio
 */
package EventManager

import (
	"log"
	enum "main/Utils/Enum"
	err "main/Utils/Error"
	"syscall"
)

type Selector struct {
	epfd          int //epoll返回的fd
	selectorykeys []*SelectorKey
}

// 存放保存的事件的fd(如socket的fd)
type SelectorKey struct {
	Fd         int    //socket返回的fd
	event_mask uint32 //epoll监听事件的类型
	Data       interface{}
}

/**
 * @description: 创建一个Selector监控所有socket
 * @param  {*}
 * @return {*}
 * @param {int} size epoll监听队列的大小
 */
func New(size int) *Selector {
	epfd, err := syscall.EpollCreate(size)
	if err != nil {
		log.Panic(err.Error())
		return nil
	}
	return &Selector{
		epfd:          epfd,
		selectorykeys: make([]*SelectorKey, size),
	}
}

/**
 * @description:关闭epoll
 * @param  {*}
 * @return {*}
 */
func (p *Selector) Close() {
	if err := syscall.Close(p.epfd); err != nil {
		log.Panic(err)
	}
	p.selectorykeys = nil
}

/**
 * @description: 插入对应事件
 * @param  {*}
 * @return {*}
 * @param {*syscall.EpollEvent} epollevent
 * @param {uint32} mask
 * @param {uint32} op epoll需要监听的操作事件
 */
func InitEpollEvent(selectorkey *SelectorKey, event_mask uint32) (*syscall.EpollEvent, error) {
	// 为epoll事件注册进对应事件的fd
	epollevent := &syscall.EpollEvent{
		Fd: int32(selectorkey.Fd),
	}
	switch {
	case event_mask&enum.EVENT_READABLE != 0:
		epollevent.Events |= syscall.EPOLLIN
		fallthrough
	case event_mask&enum.EVENT_WRITABLE != 0:
		epollevent.Events |= syscall.EPOLLOUT
	default:
		return nil, &err.UNKNOW_MASK_ERR{
			Mask: selectorkey.event_mask,
		}
	}
	return epollevent, nil
}

/**
 * @description: 1.将socket事件保存到自定义数据结构selector.selectorkeys 2.将socket事件注册成epollevent，被epoll监听
 * @param  {*}
 * @return {*}
 * @param {int} fd
 * @param {uint32} mask 需要被epoll监听的事件选项
 * @param {interface{}} Data
 */
func (p *Selector) Register(fd int, event_mask uint32, Data interface{}) error {
	if fd > len(p.selectorykeys) {
		return &err.FD_EXEC_LIMIT_ERROR{
			FD: fd,
		}
	}
	if p.selectorykeys[fd] == nil {
		p.selectorykeys[fd] = &SelectorKey{
			Fd:         fd,
			event_mask: enum.EVENT_NONE,
			Data:       Data,
		}
	}
	selectorkey := p.selectorykeys[fd]

	var op int
	// 如果epoll中没有注册事件泽注册事件，否则修改对应事件
	if selectorkey.event_mask == enum.EVENT_NONE {
		op = syscall.EPOLL_CTL_ADD
	} else {
		op = syscall.EPOLL_CTL_MOD
	}
	selectorkey.event_mask = event_mask
	epollevent, err := InitEpollEvent(selectorkey, event_mask)
	if err != nil {
		log.Panic(err)
		return err
	}
	// 将epoll事件注册到内核
	if err := syscall.EpollCtl(p.epfd, op, fd, epollevent); err != nil {
		log.Panic(err)
		return err
	}
	return nil
}

/**
 * @description: 反注册epoll
 * @param  {*}
 * @return {*}
 * @param {int} fd socket's fd
 * @param {uint32} event_mask
 */
func (p *Selector) UnRegister(fd int, event_mask uint32) (*SelectorKey, error) {
	if fd > len(p.selectorykeys) {
		return nil, &err.FD_EXEC_LIMIT_ERROR{
			FD: fd,
		}
	}
	selectorkey := p.selectorykeys[fd]
	selectorkey.Data = nil
	p.selectorykeys[fd] = nil
	if selectorkey == nil || selectorkey.event_mask&event_mask == 0 {
		log.Printf("EventManager.UnRegister:UnRegister failed,selectorkey or event do not exist!\n")
		return nil, nil
	}
	epollevent, err := InitEpollEvent(selectorkey, event_mask)
	if err != nil {
		return nil, err
	}
	op := syscall.EPOLL_CTL_DEL
	if err := syscall.EpollCtl(p.epfd, op, fd, epollevent); err != nil {
		return nil, err
	}
	return selectorkey, nil
}

/**
 * @description: 根据fd获取对应socket事件的data
 * @param  {*}
 * @return {*}
 * @param {int} fd socket对应的fd
 */
func (p *Selector) GetData(fd int) interface{} {
	return p.selectorykeys[fd].Data
}

/**
* @description:
1.使用EpollWait监听事件在超时时间内是否有响应
2.通过p.selectorkey[fd]得到对应的event;记录下来响应的事件和响应的事件类型，最终返回出去
* @param  {*}
* @return {*}
* @param {int} time 超时等待时间
*/
func (p *Selector) Poll(time int) ([]*SelectorKey, []uint32, error) {
	events := make([]syscall.EpollEvent, len(p.selectorykeys))
	n, err := syscall.EpollWait(p.epfd, events, time)
	if err != nil {
		return nil, nil, err
	}
	var awake_event, mask = make([]*SelectorKey, n), make([]uint32, n)
	for i := 0; i < n; i++ {
		epoll_event := &events[i]
		awake_event[i] = p.selectorykeys[epoll_event.Fd]

		if (epoll_event.Events&syscall.EPOLLERR != 0) && (epoll_event.Events&syscall.EPOLLRDHUP != 0) {
			continue
		}
		if epoll_event.Events&syscall.EPOLLIN != 0 {
			mask[i] |= syscall.EPOLLIN
		}
		if epoll_event.Events&syscall.EPOLLOUT != 0 {
			mask[i] |= syscall.EPOLLOUT
		}
	}
	return awake_event, mask, nil
}
