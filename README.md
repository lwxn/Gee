# Gee
这是一个用golang写的一个仿造Gin的web框架
手工实现了动态路由，分组，统一的HTML机制

### chapter 1 基本框架编写
1. gee.go
   1. Engine的编写，Engine拦截了所有的HTTP请求，并且转向自己的处理逻辑
   2. 在Engine中添加路由的映射表：```router *Router```
      1. key由method-path组成（GET-/hello/lwxn）
   3. GET方法：添加路由
   4. ServeHTTP方法：解析请求中的method和path，并且在路由表中找到相对应的handler function，否则return null.
    ```
    func (engine *Engine) GET(pattern string,handler HandlerFunc){
       engine.router.addRoute("GET",pattern,handler)
    }
    ```

2. go.mod
   1. replace: 需要将整个本地目录映射一下
    
   ```go
    replace github.com/lwxn/gee => ./gee
    ```
    
     

### chapter 2 添加context的处理

#### necessary：

1. web服务需要根据请求req http.Request来构造w http.ResponseWriter，因此对HTTP的响应信息进行完整的构造是非常有必要的。
2. Context相当于每一次请求，我们都会把需要的信息，中间产物都会放在里面，每一次请求都会产生一个context，当使用完毕之后它就会进行销毁。

#### code：

1. context.go:

   1. 实现对context的结构的编写，以及构造函数。

   2. 构造String响应：这里要注意的是，必须先header().set(),再writerHeader,Write，否则会出错

      ```
      c.Writer.Header().Set(key,value)
      c.Writer.WriteHeader(code)
      c.Writer.Write([]byte(fmt.Sprintf(format,values...)))
      ```

   3. 构造其他形式的响应

2. router.go:

   1. 将路由的方法提出来进行封装，把addRouter写到router.go里面，并且将handle处理路由映射信息的功能也移进去，然后将参数改为c *context，而不是原来的r,w

      ```
      func (r *Router) handle(c *Context){}
      ```

3. 修改gee.go

   ​	1.将router的方法替换原来的addRoute，并且在ServeHTTP中，将处理路由的逻辑也转向router.handle



### Chapter 3 添加路由的模糊处理

#### 前缀树路由（动态路由匹配）：

1. 之前用的是映射来保存路由表，但是它无法处理如同/hello/:name, /hello/*filename这样带有模糊字段的路径。

2. 前缀树（字典树）：

   1. 每一个节点的子节点都拥有相同的前缀，如同：

      ```
      /home/:name/io:（只可匹配当前的位置）
      	/home/p/io
      	/home/lwxn/io
      	
      /home/*filename:(可匹配多段)
      	/home/lucky/po/ui
      ```

#### Code：

 1. 添加trie.go

 2. 定义树节点的数据结构：

    ```
    type node struct {
    	pattern string
    	part string
    	children []*node
    	isWild bool
    }
    ```

    1. pattern表示的是全部的路径，part表示当前节点的孩子节点对应的一截路径（/home）,以及当前的pattern是否是模糊匹配的，孩子节点。

	3. match单个的子节点（根据子节点的part来确定）

    ```
    //查找单个满足条件的node
    func (n *node)matchChild(part string) *node{
    	for _,node := range n.children{
    		if node.part == part || node.isWild{
    			return node
    		}
    	}
    	return nil
    }
    ```

	4. match所有满足part的子节点们

    ```
    //查找所有满足条件的node
    func (n *node)matchChildren(part string) []*node{
    	nodes := make([]*node,0)
    	for _,node := range n.children{
    		if(node.part == part || node.isWild){
    			nodes = append(nodes, node)
    		}
    	}
    	return nodes
    }
    ```

5. 插入子节点，比较麻烦的是/home/:name/po这种多层的路由，如果只匹配到/home/lang，当然也是可以匹配上的，但是这并不是这条路由的终点，因此需要为每条路由设置终点，子节点的pattern一直到最后才会有值，不到终点都设为空值。------dfs

   ```
   func (n *node) insert(pattern string,parts []string,height int)  {
   	//如果是最后一层,那么pattern才会有值，如果中间的路径有模糊匹配，就可以方便判断是否是终点
   	if height == len(parts){
   		n.pattern = pattern
   		return
   	}
   
   	part := parts[height]
   	child := n.matchChild(part)
   	fmt.Println("part:",part)
   	if child == nil{
   		child = &node{
   			part: part,
   			isWild: part[0] == '*' || part[0] == ':',
   		}
   		n.children = append(n.children, child)
   	}
   	child.insert(pattern,parts,height+1)
   }
   
   ```

6. 搜索子节点：

   DFS，如果一直搜索到了parts的终点，或者是匹配到了*，就返回该节点

   ```
   func (n *node) search(parts []string,height int)*node{
      if(len(parts) == height || strings.HasPrefix(n.part,"*")){
         if n.pattern == ""{
            return nil
         } else{
            return n
         }
      }
   
      part := parts[height]
      children := n.matchChildren(part)
      fmt.Println(len(children))
      for _,child := range children{
         result := child.search(parts,height+1)
         if result != nil{
            return result
         }
      }
      return nil
   }
   ```

7. 修改router，将前缀树应用到router之中：

   1. 拆分pattern为多个parts

   ```
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
   ```

    2. 添加路由进行修改，修改为添加树的子节点

       ```
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
       ```

   3. 匹配路由：如果路径里面有:,*，就要把它们换成是普通值， *的话就要把路径后面的一大段全部截掉接上去。

      ```
      func (r *router) getRoute(method string, path string) (*node, map[string]string) {
      	searchParts := parsePattern(path)
      	params := make(map[string]string)
      	root, ok := r.roots[method]
      
      	if !ok {
      		return nil, nil
      	}
      
      	n := root.search(searchParts, 0)
      
      	if n != nil {
      		parts := parsePattern(n.pattern)
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
      ```

   4. 在context.go之中，定义一个可以根据:name（name）来返回映射值(Param[key])的map[string]string.

### Chapter 4 分组控制

	#### 根据前缀分组

1. /post/a和/post/b应该都是归属于/post下面的。

#### code:

 1. Engine: embedded了RouterGroup，相当于继承

    ```
    Engine struct {
    	*RouterGroup
    	router *router
    	groups []*RouterGroup // store all groups
    }
    ```

2. RouterGroup: 因为要调用addRoute方法，所以要有一个指向engine的指针：

   ```
   type RouterGroup struct {
      prefix string
      middlewares []HandlerFunc
      parent *RouterGroup
      engine *Engine
   }
   ```

3. 根据前缀，产生一个新的group，注意父亲节点的指向

   ```
   func (group *RouterGroup) Group(prefix string) *RouterGroup {
   	engine := group.engine
   	newGroup := &RouterGroup{
   		prefix: group.prefix + prefix,
   		parent: group,
   		engine: engine,
   	}
   	engine.groups = append(engine.groups, newGroup)
   	return newGroup
   }
   ```

4. 调用engine的addRoute方法

   ```
   func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
   	pattern := group.prefix + comp
   	log.Printf("Route %4s - %s", method, pattern)
   	group.engine.router.addRoute(method, pattern, handler)
   }
   ```