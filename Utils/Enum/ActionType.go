/*
 * @Description: Action类型定义
 * @Author: Rocky Hoo
 * @Date: 2021-07-11 00:20:03
 * @LastEditTime: 2021-07-11 00:44:02
 * @LastEditors: Please set LastEditors
 * @CopyRight: XiaoPeng Studio
 * Copyright (c) 2021 XiaoPeng Studio
 */
package enum

type Action int

const (
	CONTINUE Action = iota
	SHUTDOWN_RD
	SHUTDOWN_WR
	SHUTDOWN_RDWR
	TRIGGER_OPEN_EVENT
	TRIGGER_DATA_EVENT
	TRIGGER_CLOSE_EVENT
)
