package gee

import "net/http"

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next() // 可能发生panic
	}
	
}
