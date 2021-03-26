package gee

import (
	"fmt"
	"strings"
)

type node struct {
	pattern string
	part string
	children []*node
	isWild bool
}

//查找单个满足条件的node
func (n *node)matchChild(part string) *node{
	for _,node := range n.children{
		if node.part == part || node.isWild{
			return node
		}
	}
	return nil
}

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


func (n *node) search(parts []string,height int)*node{
	if(len(parts) == height || strings.HasPrefix(n.part,"*")){
		fmt.Println("height:",height)
		fmt.Println("pattern:",n.pattern)
		if n.pattern == ""{
			return nil
		} else{
			return n
		}
	}

	part := parts[height]
	children := n.matchChildren(part)
	for _,child := range children{
		result := child.search(parts,height+1)
		if result != nil{
			return result
		}
	}
	return nil
}