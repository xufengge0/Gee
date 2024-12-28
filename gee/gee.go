package gee

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context)

type Engine struct {
	route *router // 路由表

	*RouterGroup
	groups []*RouterGroup

	htmlTemplates *template.Template // 将所有的模板加载进内存
	funcMap       template.FuncMap   // 所有的自定义模板渲染函数
}

// 路由分组
type RouterGroup struct {
	prefix      string        // 路由组名称，例如 /v1
	middlewares []HandlerFunc // 组内使用的中间件
	engine      *Engine
	parent      *RouterGroup // 父路由组
}

func New() *Engine {
	engine := &Engine{route: newrouter()}
	engine.RouterGroup = &RouterGroup{engine: engine} // 初始化根路由组
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}
func Default()*Engine{
	r := New()
	r.Use(Logger(),Recovery())
	return r
}
// 为Engine实现http.Handler接口,拦截所有http请求
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		// 根据路径前缀匹配中间件
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.route.handle(c)
}
// 设置自定义模板渲染函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// 路由表绑定
func (engine *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	engine.route.addRouter(method, pattern, handler)
}
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRouter("GET", pattern, handler)
}
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRouter("POST", pattern, handler)
}
func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

// 增加路由组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: group.engine,
		parent: group,
	}
	group.engine.groups = append(group.engine.groups, newGroup)
	return newGroup
}
func (group *RouterGroup) addRouter(method, pattern string, handler HandlerFunc) {
	newPattern := group.prefix + pattern
	group.engine.addRouter(method, newPattern, handler)
}
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRouter("GET", pattern, handler)
}
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRouter("POST", pattern, handler)
}
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// 服务端渲染静态文件
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root)) // 将 root 转换为 http.FileSystem 接口的类型。
	pattern := path.Join(relativePath, "/*filepath")
	group.addRouter("GET", pattern, handler)
}
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)

	// 使用 http.StripPrefix 移除 /static 前缀，并将剩余路径交给 http.FileServer(fs) 处理
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}
