/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @version V1.0
 * Description: 
 */

package statement

import (
    "github.com/xfali/gobatis"
    "github.com/xfali/gobatis/handler"
)

type Statement interface {
    Query(handler handler.ResultHandler, iterFunc gobatis.IterFunc, params ...interface{}) error
    Exec(params ...interface{}) (int64, error)
    Close()
}
