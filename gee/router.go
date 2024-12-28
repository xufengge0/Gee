package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node       // 前缀树
	handlers map[string]HandlerFunc // 存储key：handler
}

func newrouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 解析路由，将路由中的动态参数和通配符参数解析
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, part := range vs {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 注册路由和前缀树
func (r *router) addRouter(method, pattern string, handler HandlerFunc) {
	// 注册路由和handler
	key := method + "-" + pattern // GET-/hello
	r.handlers[key] = handler     // 绑定key-handler

	// 向前缀树中插入
	parts := parsePattern(pattern)
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{} // 第一个节点为根节点
	}
	r.roots[method].insert(pattern, parts, 0)
}

// 获取node和请求路径中的所有动态参数
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path) // /p/123/doc
	params := make(map[string]string)

	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern) // /p/:lang/doc
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index] // 将动态参数（例如 123）存入map
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

// 处理请求
func (r *router) handle(c *Context) {
	node, params := r.getRoute(c.Method, c.Path)
	if node != nil {
		c.Params = params
		key := c.Method + "-" + node.pattern
		c.handlers = append(c.handlers, r.handlers[key]) // 处理请求的handler
	} else {
		c.handlers = append(c.handlers, func(ctx *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next() // 进入第一个中间件
}
