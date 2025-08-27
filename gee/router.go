package gee

import (
	"net/http"
	"strings"
)

/* roots key eg, roots['GET'] roots['POST']
每种方法有一颗自己的前缀树，根据method找到对应的node，然后进行路由注册匹配
*/
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 将url字符串解析为数组方便匹配
func parsePattern(pattern string) []string {
	parts := make([]string, 0)
	//要对解析后的字符串做一个筛选
	s := strings.Split(pattern, "/")
	for _, st := range s {
		if st != "" {
			parts = append(parts, st)
			if st[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	s := method + "-" + pattern
	r.handlers[s] = handler
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
}

// 找到node和其中的动态匹配参数
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path) //用户使用的url
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern) //定义的模版路由
		//查找动态匹配字符串的映射
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		//注意这里使用的是n.pattern，而不是c.Path，因为c.Path是用户输入的，而n.pattern是路由树中匹配到的
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context){
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}

// 前缀树实现动态路由解析
// 支持两种模式:name和*filepath
type node struct {
	pattern  string //待匹配路由（完整），该字段在查找匹配中没什么作用，但是作为处理器的key等情况很关键
	part     string //路由的一部分，可理解为/.../中间的...，该字段在下面的查找匹配中很重要
	children []*node
	isWild   bool
}

// 子方法：查找第一个匹配到的路由片段，用于实现注册；
// 因为注册只要保证当条路由在路由树中即可，不需要遍历所有符合的路由都注册上
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if part == child.part || child.isWild {
			return child
		}
	}
	return nil
}

// 子方法：查找所有匹配的路由，用于实现路由查找
func (n *node) matchChildren(part string) []*node {
	children := make([]*node, 0)
	for _, child := range n.children {
		if part == child.part || child.isWild {
			children = append(children, child)
		}
	}
	return children
}

// 核心操作：注册路由
func (n *node) insert(pattern string, parts []string, height int) {
	//只在路由的终点标记，中间的node不算匹配成功；实现精确匹配
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	//得到当前的路由片段和匹配的node，如果node不存在则需要注册
	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// 查找url对应的node
func (n *node) search(parts []string, height int) *node {
	//递归终止条件
	// 注意终止是发生在当前节点n的n.search()，所以调用点还需根据返回值判断是否成功。并且调用时不知道当前node是否已经有匹配的路由，所以要对pattern进行判空
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		if node := child.search(parts, height+1); node != nil {
			return node
		}
	}
	return nil
}
