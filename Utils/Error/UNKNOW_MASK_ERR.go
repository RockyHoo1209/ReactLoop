/*
 * @Description: 返回UNKNOW_MASK_ERR错误
 * @Author: Rocky Hoo
 * @Date: 2021-07-01 11:23:17
 * @LastEditTime: 2021-07-17 09:02:06
 * @LastEditors: Please set LastEditors
 * @CopyRight:
 * Copyright (c) 2021 XiaoPeng Studio
 */
package err

import "fmt"

type UNKNOW_MASK_ERR struct {
	Mask uint32
}

func (e *UNKNOW_MASK_ERR) Error() string {
	return fmt.Sprintf("unknow event mask: %d", e.Mask)
}
