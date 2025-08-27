package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

// 中间件信息也保存在
type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int
	//middleware
	handlers []HandlerFunc
	index    int
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
		index:  -1,
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// 与从url中取值不同，这是返回动态匹配的参数
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// 查询url和body中的键的值
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 只查询url中键对应的值
// 使用get方法返回的是key对应的第一个value，相当于map[0]
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 写入状态码
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
	c.StatusCode = code
}

// set方法会覆盖旧值，如果要追加值应该用add
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// format 是一个格式化字符串，用来指定输出内容的模板。它的作用是定义文本的结构，并通过占位符（如 %s、%d 等）来表示需要插入的动态数据。
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)

	//先解析再写入，一般不这样使用：创建json解析器并写入writer，再用解析器解析obj同时自动写入相应
	jsonData, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println("json解析失败")
	}
	c.Writer.Write(jsonData)
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

// Next方法用于等待剩下的handler处理完，index记录当然执行的层数
// 放在Context中是因为handler的参数是c，自然也需要c中的方法来调用实现
func (c *Context) Next() {
	c.index++
	for ; c.index < len(c.handlers);c.index++ {
		c.handlers[c.index](c)
	}
}
