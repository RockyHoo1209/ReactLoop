/*
 * @Description: EventLoop事件循环模块(目前一个server只有一个eventloop,通过不断修改triger_ptr来实现不同的操作)
 * @Author: Rocky Hoo
 * @Date: 2021-07-09 12:44:10
 * @LastEditTime: 2021-07-24 14:23:58
 * @LastEditors: Please set LastEditors
 * @CopyRight: XiaoPeng Studio
 * Copyright (c) 2021 XiaoPeng Studio
 */
package EventLoop

import (
	"log"
	"main/EventManager"
	enum "main/Utils/Enum"
	"time"
)

/**
 * @description:定义事件循环结构体
 * @param  {*}
 * @return {*}
 */
type EventLoop struct {
	*EventManager.Selector               //指向一个Selector
	system_events          []*Event      //系统事件
	user_events            []*UserEvent  //用户定义事件
	interval               time.Duration //定义事件循环的轮询周期
	done                   bool          //事件是否完成的标志位,也是是否退出循环的标志位
	triger_data_ptr        *interface{}  //指定触发器特定数据的指针(通过委托指针实现不同eventloop的功能)
}

/**
 * @description: EventLoop构造函数
 * @param  {*}
 * @return {*}
 */
func New() *EventLoop {
	return &EventLoop{
		Selector:      EventManager.New(1024), //调用EventManager初始化一个事件管理器
		system_events: []*Event{},
		user_events:   []*UserEvent{},
		interval:      100 * time.Millisecond,
	}
}

/**
 * @description: 设置该事件已经完成
 * @param  {*}
 * @return {*}
 */
func (el *EventLoop) Done() {
	el.done = true
}

/**
 * @description: 启动事件循环系统
 * @param  {*}
 * @return {*}
 */
func (el *EventLoop) Run() {
	for _, system_event := range el.system_events {
		if system_event.Serving != nil {
			system_event.Serving(el, nil)
		}
	}
	for _, user_event := range el.user_events {
		user_event.setNextTrigerTime()
	}
	// 只要事件循环没有设置为结束就一直执行
	for !el.done {
		el.TikTok()
	}
}

/**
 * @description: 设置特定指针，通过指针触发不同的操作
 * @param  {*}
 * @return {*}
 * @param {interface{}} data
 */
func (el *EventLoop) SetTrigerDataPtr(data interface{}) {
	el.triger_data_ptr = &data
}

/**
 * @description: 添加一个系统事件
 * @param  {*}
 * @return {*}
 * @param {Event} event
 */
func (el *EventLoop) AddSystemEvent(event *Event) {
	el.system_events = append(el.system_events, event)
}

/**
 * @description: 添加一个用户事件
 * @param  {*}
 * @return {*}
 * @param {*UserEvent} task
 */
func (el *EventLoop) AddUserEvent(task *UserEvent) {
	el.user_events = append(el.user_events, task)
}

/**
 * @description:遍历用户事件,找到最接近现在的触发事件
 * @param  {*}
 * @return {*}
 */
func (el *EventLoop) FindNearestTask() *UserEvent {
	for _, user_event := range el.user_events {
		if user_event.nexttriggerTime.After(time.Now()) {
			return user_event
		}
	}
	return nil
}

/**
 * @description:根据传入的action类型进行相应的操作,此函数对应着发生系统事件后，客户需要进行的后续操作如Read()
 * @param  {*}
 * @return {*}
 * @param {Action} action
 */
func (el *EventLoop) processAction(action enum.Action, fd int) {
	switch action {
	case enum.SHUTDOWN_RD:
		if _, err := el.UnRegister(fd, enum.EVENT_READABLE); err != nil {
			log.Printf("EventLoop-processAction:%s", "尝试执行任务失败")
		}
	case enum.SHUTDOWN_WR:
		if _, err := el.UnRegister(fd, enum.EVENT_WRITABLE); err != nil {
			log.Printf("EventLoop-processAction:%s", "尝试执行任务失败")
		}
	case enum.SHUTDOWN_RDWR:
		if _, err := el.UnRegister(fd, enum.EVENT_WRITABLE|enum.EVENT_READABLE); err != nil {
			log.Printf("EventLoop-processAction:%s", "尝试执行任务失败")
		}
	case enum.TRIGGER_OPEN_EVENT:
		for _, event := range el.system_events {
			if event.Close != nil {
				event.Open(el, el.triger_data_ptr)
			}
		}
	case enum.TRIGGER_DATA_EVENT:
		for _, event := range el.system_events {
			if event.Data != nil {
				event.Data(el, el.triger_data_ptr)
			}
		}
	case enum.TRIGGER_CLOSE_EVENT:
		for _, event := range el.system_events {
			if event.Close != nil {
				event.Close(el, el.triger_data_ptr)
			}
		}
	case enum.CONTINUE:
	}
	// 将数据与操作进行一次绑定后，需要清空数据,下一次到来的事件的触发指针可能不一样
	el.triger_data_ptr = nil
}

/**
 * @description:通过阻塞去轮询事件是否完成(系统时间调用epoll,用户每隔interval轮询调用一次)
 * @param  {*}
 * @return {*}
 */
func (el *EventLoop) TikTok() {
	sleepTime := el.interval
	nearestTask := el.FindNearestTask()
	if nearestTask != nil {
		sleepTime = nearestTask.nexttriggerTime.Sub(time.Now())
		if sleepTime < 0 {
			sleepTime = 0
		}
	}
	selectorkeys, _, _ := el.Poll(int(sleepTime / time.Millisecond))
	for _, selectorkey := range selectorkeys {
		ed := selectorkey.Data.(EventData)
		action := ed.e(el, ed)
		el.processAction(action, selectorkey.Fd)
	}
	if nearestTask != nil {
		nearestTask.Task(el, nil)
		nearestTask.setNextTrigerTime()
	}
}

/**
 * @description: 将事件注册进事件管理器
 * @param  {*}
 * @return {*}
 * @param {int} fd
 * @param {uint32} mask
 * @param {EventProc} event_proc
 * @param {interface{}} event_data
 */
func (el *EventLoop) RegisterEvent(fd int, mask uint32, event_proc EventProc, event_data interface{}) error {
	return el.Register(fd, mask, EventData{
		e:    event_proc,
		data: event_data,
	})
}

/**
 * @description: 从事件管理器上注销事件
 * @param  {*}
 * @return {*}
 * @param {int} fd
 * @param {uint32} mask
 */
func (el *EventLoop) UnRegisterEvent(fd int, mask uint32) {
	el.UnRegister(fd, mask)
}

/**
 * @description:
 * @param {*EventLoop} el
 * @param {*interface{}} triger_data_ptr 触发操作相对应的指针
 * @return {*}
 */
type TrigerProcess func(el *EventLoop, triger_data_ptr *interface{})

/**
 * @description:定义一个事件，
 *  包括Serving\Open\Closed\Data(传送数据)
 *  三个状态，均需要用户传入 (需要暴露给用户传入的属性可以这么定义)
 * @param  {*}
 * @return {*}
 */
type Event struct {
	Open, Close, Serving, Data TrigerProcess
}

/**
 * @description:用户定义的事件类型
 * @param  {*}
 * @return {*}
 */
type UserEvent struct {
	nexttriggerTime time.Time     //下一次需要执行的具体时间
	Task            TrigerProcess //执行事件的函数
	interval        time.Duration //运行的时间周期间隔
}

/**
 * @description:设置一个UserEvent的下一个触发时间
 * @param  {*}
 * @return {*}
 */
func (ue *UserEvent) setNextTrigerTime() {
	ue.nexttriggerTime = ue.nexttriggerTime.Add(ue.interval)
}

type EventProc func(el *EventLoop, data interface{}) enum.Action
type EventData struct {
	e    EventProc
	data interface{}
}
