package gee

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type HandlerFunc func(*Context)

type Engine struct {
	//匿名嵌套
	//把路由相关的全部交给RouterGroup管理
	*RouterGroup
	Router *router
	groups []*RouterGroup
}

func New() *Engine {
	engine := &Engine{Router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func Default() *Engine{
	engine := New()
	engine.Use(logger(), Recovery())
	return engine
}

func (g *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, g)
}

func (g *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range g.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	g.Router.handle(c)
}

// 对底层的路由的管理，统一管理某个前缀的所有路由，方便实现中间件
type RouterGroup struct {
	prefix      string
	engine      *Engine
	middlewares []HandlerFunc
	parent      *RouterGroup
}

// 创建一个新的group
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: group.engine,
		parent: group,
	}
	group.engine.groups = append(group.engine.groups, newGroup)
	return newGroup
}

// 路由操作相关函数时对router的方法的封装
// RouterGroup中有一个指向engine的指针，所以可以通过engine调用框架内的所有方法
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s-%s", method, pattern)
	group.engine.Router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc){
	group.middlewares = append(group.middlewares, middlewares...)
}

//默认添加的中间件logger
func logger() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}