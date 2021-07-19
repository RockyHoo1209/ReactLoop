/*
 * @Author: Rocky Hoo
 * @Date: 2021-07-17 08:51:49
 * @LastEditTime: 2021-07-18 07:58:53
 * @LastEditors: Please set LastEditors
 * @Description: Ip地址格式错误
 * @FilePath: /ReactLoop/Utils/Error/IP_FORMAT_ERR.go
 */
package err

import "fmt"

type IP_FORMAT_ERR struct {
	IP string
}

func (e *IP_FORMAT_ERR) Error() string {
	return fmt.Sprintf("iP %s format error\n", e.IP)
}
