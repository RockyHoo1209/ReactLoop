/*
 * @Author: your name
 * @Date: 2021-07-20 15:31:58
 * @LastEditTime: 2021-07-20 15:58:44
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /ReactLoop/Socket/Socket_test.go
 */
package socket

import (
	"main/EventLoop"
	"testing"
)

func TestSocket(*testing.T) {
	listener, err := NewListener("tcp4", "localhost")
	if err != nil {
		println("error:", err)
		return
	}
	listener.BindAndListen()
	el := EventLoop.New()
	listener.acceptEvent(el, nil)
}
