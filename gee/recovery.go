package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

//全局中间件，在panic时会逐层向上传递panic，回退到该函数时捕获panic并打印错误
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				msg := fmt.Sprintf("%s", err)
				//trace函数用来获取触发panic的堆栈信息
				log.Printf("%s\n", trace(msg))
				c.Fail(http.StatusInternalServerError, "Server Eorror")
			}
		}()
		c.Next()
	}
}

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}