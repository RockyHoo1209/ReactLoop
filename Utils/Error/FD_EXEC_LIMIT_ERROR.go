/*
 * @Description: 文件描述符超出限制
 * @Author: Rocky Hoo
 * @Date: 2021-07-01 13:06:44
 * @LastEditTime: 2021-07-02 10:57:53
 * @LastEditors: Please set LastEditors
 * @CopyRight:
 * Copyright (c) 2021 XiaoPeng Studio
 */
 package err

 import "fmt"
 
 type FD_EXEC_LIMIT_ERROR struct {
	 FD int
 }
 
 func (e *FD_EXEC_LIMIT_ERROR) Error() string {
	 return fmt.Sprintf("fd %d exceed the limit", e.FD)
 }
 