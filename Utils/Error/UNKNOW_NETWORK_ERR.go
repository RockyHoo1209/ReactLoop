/*
 * @Author: Rocky Hoo
 * @Date: 2021-07-18 08:09:32
 * @LastEditTime: 2021-07-18 08:12:02
 * @LastEditors: Please set LastEditors
 * @Description:network无法解析报错
 * @FilePath: /ReactLoop/Utils/Error/UNKNOW_NETWORK_ERR.go
 */
package err

import "fmt"

type UNKNOW_NETWORK_ERR struct {
	Network string
}

func (e *UNKNOW_NETWORK_ERR) Error() string {
	return fmt.Sprintf("Network %s is unknown\n", e.Network)
}
