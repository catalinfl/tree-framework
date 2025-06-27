package tree

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Tree struct {
	method    Method
	startNode *Node
}

type Node struct {
	path       string
	children   []*Node
	Handler    http.HandlerFunc
	CtxHandler CtxFunc
}

func initTreesMap() map[Method]*Tree {
	return make(map[Method]*Tree)
}

/*
	  handler.path - current path of server router
		t - tree which needs to be modified
		createTreeIdea - used for every route and method in buildTrees()
*/
func (r *Mux) createTreeIdea(handler Route, t *Tree, m Method) (*Tree, error) {
	if t == nil {
		t = &Tree{method: m,
			startNode: &Node{
				path:       "/",
				children:   nil,
				Handler:    nil,
				CtxHandler: nil,
			}}
	} // 1. tree inexistent, creates one

	if t.startNode == nil {
		t.startNode = &Node{path: "/",
			children:   nil,
			Handler:    nil,
			CtxHandler: nil,
		}
	} // 2. startNode inexistent, creates one

	if handler.path == "" || handler.path == "/" {
		// 3. check if server's route path is first - empty one
		if t.startNode.Handler != nil {
			return nil, errors.New(handler.path)
		}
		return t, nil
	}

	// 4. verify all paths
	startNode := t.startNode
	if len(strings.Split(handler.path, "/")) >= 2 {
		paths := strings.Split(strings.TrimSpace(handler.path), "/") // paths of server route

		for i := range paths {
			paths[i] = strings.TrimSpace(paths[i])
		}

		// s := fmt.Sprintf("Method: %s, paths %s", m, paths)
		// fmt.Println(s)

		node := t.startNode

		for i, path := range paths {
			if foundNode := node.FindNodePath(path); foundNode != nil {
				node = foundNode
				// 5. check if the final path node already has a handler
				if i == len(paths)-1 {
					// if already has a handler, return error
					if node.Handler != nil {
						return nil, errors.New(handler.path)
					}
					// if not change it to true
					node.ChangeNodeState(handler, true)
				}

			} else {
				// path not found in tree, creating new node
				// 6. check if the path node already exists, if not create it
				if i == len(paths)-1 {
					node.AddNode(handler, true, path)
				} else {
					// add the path into tree parent node
					node = node.AddNode(handler, false, path)
				}
			}
		}

	} else {
		if foundNode := t.startNode.FindNodePath(handler.path); foundNode != nil {
			// Check if the node already has a handler
			if foundNode.Handler != nil {
				return nil, errors.New(handler.path)
			}
			foundNode.ChangeNodeState(handler, true)
		} else {
			t.startNode.AddNode(handler, true, handler.path)
		}
	}

	t.startNode = startNode
	return t, nil
}

func (n *Node) ChangeNodeState(handler Route, isFinalRoute bool) {
	if isFinalRoute {
		n.Handler = *handler.h
	}
}

func (n *Node) AddNode(handler Route, isFinalRoute bool, segment string) *Node {
	var child *Node

	if segment == "" || segment == "/" {
		return n
	}
	if isFinalRoute {
		child = &Node{
			path:       segment,
			Handler:    *handler.h,
			CtxHandler: handler.CtxHandler,
			children:   nil,
		}
	} else {
		child = &Node{
			path:       segment,
			Handler:    nil,
			children:   nil,
			CtxHandler: nil,
		}
	}
	n.children = append(n.children, child)

	return child
}

func (n *Node) InDepthSearch(path string) (http.HandlerFunc, CtxFunc, map[string]string) {
	params := make(map[string]string)
	requestPathSegments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	currentNode := n

	for _, segment := range requestPathSegments {
		if segment == "" {
			continue
		}

		found := false
		for _, child := range currentNode.children {
			if child.path == segment {
				currentNode = child
				found = true
				break
			}

			if strings.HasPrefix(child.path, ":") {
				paramName := strings.TrimPrefix(child.path, ":")
				params[paramName] = segment
				currentNode = child
				found = true
				break
			}
		}

		if !found {
			return nil, nil, nil
		}
	}

	return currentNode.Handler, currentNode.CtxHandler, params
}

func (n *Node) FindNodePath(path string) *Node {
	for _, node := range n.children {
		if node.path == path {
			return node
		}
	}

	return nil
}

func (n *Node) FindNodePathBool(path string) bool {
	for _, node := range n.children {
		if node.path == path {
			return true
		}
	}
	return false
}

func (t *Tree) PrintTree() {
	if t == nil {
		fmt.Println("Tree is nil")
		return
	}
	if t.startNode == nil {
		fmt.Println("Tree is empty")
		return
	}

	printNode(t.startNode, 0)

}

func printNode(n *Node, level int) {
	indent := strings.Repeat(" ", level)
	fmt.Printf("%s%s (handler=%v)\n", indent, n.path, n.Handler != nil)

	for _, child := range n.children {
		printNode(child, level+1)
	}
}

func (r *Mux) TestRoute() map[Method]*Tree {
	return r.buildTrees()
}
