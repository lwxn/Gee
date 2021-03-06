package gee

import (
	"fmt"
	"net/http"
	"strings"
)

type Router struct {
	roots map[string]*node
	handlers map[string]HandlerFunc
}

//roots eg: roots['GET'] roots['POST']
//handlers eg: handlers['GET-/p/:lang/doc'],handlers['POST-/p/book']

func NewRouter() *Router {
	return &Router{
		roots: make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

//   /p/:name    /p/*   /p/*name/*
func parsePattern(pattern string) ([]string){
	vs := strings.Split(pattern,"/")

	parts := make([]string,0)
	for _, v := range vs{
		if v != ""{
			parts = append(parts,v)
			if v[0] == '*'{
				break
			}
		}
	}
	return parts
}

func (r *Router) addRoute(method string,pattern string,handler HandlerFunc){
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	fmt.Println(key)
	_,ok := r.roots[method]
	if !ok{
		r.roots[method] = &node{} //!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	}
	r.roots[method].insert(pattern,parts,0)
	r.handlers[key] = handler
}

func (r *Router) getRoute(method string,path string)(*node,map[string]string){
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root,ok := r.roots[method]

	if !ok{
		return nil,nil
	}

	fmt.Println("---------",searchParts)
	n := root.search(searchParts,0)
	fmt.Println(n.pattern,n.part,n.isWild)
	if n != nil{
		parts := parsePattern(n.pattern)
		for i,part := range parts{
			if part[0] == ':'{
				params[part[1:]] = searchParts[i]
			}else if part[0] == '*' && len(part) >1{
				params[part[1:]] = strings.Join(searchParts[i:],"/")
				break
			}
		}
		return n,params
	}
	return nil,nil
}

func (r *Router) handle(c *Context){
	n,params := r.getRoute(c.Method,c.Path)
	if n != nil{
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers,r.handlers[key])
	}else{
		c.handlers = append(c.handlers,func(c *Context){
			c.String(http.StatusNotFound,"404 not found: %s\n",c.Path)
		})

	}
	c.Next()
}