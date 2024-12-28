package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request
	Path   string
	Method string

	Params     map[string]string // 存放请求路径中的动态部分或参数
	StatusCode int

	handlers []HandlerFunc // 中间件
	index    int           // 记录中间件的执行顺序

	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// 控制权交给下一个中间件
func (c *Context) Next() {
	c.index++
	for ; c.index < len(c.handlers); c.index++ {
		c.handlers[c.index](c)
	}
}

// 获取请求body中表单数据
func (c *Context) PostForm(key string) string {
	return c.Req.PostFormValue(key)
}

// 获取请求URL中的key
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}
func (c *Context) Param(key string) string {
	value := c.Params[key]
	return value
}
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// 设置返回的状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(c.StatusCode)
}

// 设置返回的header
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// 返回字符串
func (c *Context) String(code int, formate string, values ...any) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(formate, values...)))
}

// 返回HTML
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err :=c.engine.htmlTemplates.ExecuteTemplate(c.Writer,name,data); err!=nil{
		c.Fail(500,err.Error())
	}
	// c.Writer.Write([]byte(html))
}

// 返回JSON
func (c *Context) JSON(code int, obj any) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500) // 不会执行
	}

	// data, err := json.Marshal(obj)
	// if err != nil {
	// 	http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	// }
	// c.Writer.Write(data)
}

// 返回DATA
func (c *Context) DATA(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}
