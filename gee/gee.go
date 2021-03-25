package gee

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(c *Context)

type RouterGroup struct {
	prefix string
	middlewares []HandlerFunc
	parent *RouterGroup
	engine *Engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (group *RouterGroup) Group(prefix string)*RouterGroup{
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups,newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string,comp string,handler HandlerFunc){
	pattern := group.prefix + comp
	group.engine.router.addRoute(method,pattern,handler)
}

func (group *RouterGroup) GET(pattern string,handler HandlerFunc){
	group.addRoute("GET",pattern,handler)
}

func (group *RouterGroup)POST(pattern string,handler HandlerFunc){
	group.addRoute("POST",pattern,handler)
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc){
	group.middlewares = append(group.middlewares,middlewares...);
}

func (group *RouterGroup) createStaticHandler(relativePath string,fs http.FileSystem) HandlerFunc{
	absolutePath := path.Join(group.prefix,relativePath)
	fileServer := http.StripPrefix(absolutePath,http.FileServer(fs))
	return func(c *Context) {
		file := c.Params["filepath"]

		if _,err := fs.Open(file);err != nil{
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer,c.Req)
	}
}

func (group *RouterGroup) Static (relativePath string,root string){
	handler := group.createStaticHandler(relativePath,http.Dir(root))
	urlPattern := path.Join(relativePath,"/*filepath")

	group.GET(urlPattern,handler)
}





type Engine struct {
	*RouterGroup
	router *Router
	groups []*RouterGroup
	htmlTemplates *template.Template
	funcMap template.FuncMap
}

func New() *Engine{
	engine := &Engine{
		router: NewRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}


func (engine *Engine) addRoute(method string,pattern string,handler HandlerFunc){
	engine.router.addRoute(method,pattern,handler)
}

func (engine *Engine) GET(pattern string,handler HandlerFunc){
	engine.router.addRoute("GET",pattern,handler)
}

func (engine *Engine) POST(pattern string, handler HandlerFunc){
	engine.router.addRoute("POST",pattern,handler)
}

func (engine *Engine) RUN(port string)(err error){
	return http.ListenAndServe(port,engine)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap)  {
	engine.funcMap = funcMap
}

func (engine *Engine)LoadHTMLGlob(pattern string){
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter,req *http.Request){
	var middlewares []HandlerFunc
	for _,group := range engine.groups{
		if strings.HasPrefix(req.URL.Path,group.prefix){
			middlewares = append(middlewares,group.middlewares...)
		}
	}

	c := NewContext(w,req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}