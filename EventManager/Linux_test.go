/*
 * @Description:
 * @Author: Rocky Hoo
 * @Date: 2021-07-02 12:40:48
 * @LastEditTime: 2021-07-07 08:13:04
 * @LastEditors: Please set LastEditors
 * @CopyRight:
 * Copyright (c) 2021 XiaoPeng Studio
 */
package EventManager
import(
	enum "main/Utils/Enum"	
	"testing"
)

  
func TestLinux(*testing.T){
	selector:=New(100)	  
	selector.Register(1,enum.EVENT_READABLE,"hello")
	selector.Poll(1)
	selector.UnRegister(1,enum.EVENT_READABLE|enum.EVENT_WRITABLE)
	selector.Close()
 }