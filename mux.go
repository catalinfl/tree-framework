package tree

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Method int

const (
	GET Method = iota
	POST
	PUT
	DELETE
	PATCH
	HEAD
	OPTIONS
	GROUP
	USE
)

var methodToString = map[Method]string{
	GET:     "GET",
	POST:    "POST",
	PUT:     "PUT",
	DELETE:  "DELETE",
	PATCH:   "PATCH",
	HEAD:    "HEAD",
	OPTIONS: "OPTIONS",
	GROUP:   "GROUP",
	USE:     "USE",
}

func (m Method) String() string {
	return methodToString[m]
}

type Route struct {
	h          *http.HandlerFunc
	CtxHandler CtxFunc
	path       string
}

type MiddlewareFunc func(*Ctx) error

type Middleware struct {
	Path    string
	Handler MiddlewareFunc
}

type Mux struct {
	handlers    map[Method]*[]Route
	middlewares []Middleware
	handlerMap  map[*http.HandlerFunc]CtxFunc
	automatic   bool
	trees       map[Method]*Tree
}

func InitMux() *Mux {
	return &Mux{
		handlers:    make(map[Method]*[]Route),
		middlewares: make([]Middleware, 0),
		handlerMap:  make(map[*http.HandlerFunc]CtxFunc),
		automatic:   false,
		trees:       nil,
	}
}

func (r *Mux) addMutationRoute(url string, method Method, t CtxFunc) {

	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t(NewCtx(w, req, url, r.middlewares, t, r.automatic))
	})

	r.handlerMap[&wrappedHandler] = t

	mType := Route{
		h:          &wrappedHandler,
		CtxHandler: t,
		path:       url,
	}
	if _, ok := r.handlers[method]; !ok {
		r.handlers[method] = &[]Route{mType}
	} else {
		*r.handlers[method] = append(*r.handlers[method], mType)
	}
}

func pathMatch(middlewarePath, requestPath string) bool {
	if middlewarePath == "" {
		return true
	}

	if requestPath == middlewarePath {
		return true
	}

	if strings.HasPrefix(requestPath, middlewarePath) && len(requestPath) > len(middlewarePath) && requestPath[len(middlewarePath)] == '/' {
		return true
	}

	return false
}

func getParams(clientPath string, serverPath string) map[string]string {
	params := make(map[string]string)

	clientPathParts := strings.Split(clientPath, "/")
	serverPathParts := strings.Split(serverPath, "/")

	for i, part := range serverPathParts {
		if strings.HasPrefix(part, ":") {
			if len(clientPathParts) > i {
				if params[part[1:]] == "" {
					params[part[1:]] = clientPathParts[i]
				} else if _, exists := params[part[1:]]; exists {
					suffix := 1
					newParamName := part[1:]
					for {
						strSuffix := strconv.Itoa(suffix)
						newParamName = newParamName + "_" + strSuffix
						if _, exists := params[newParamName]; !exists {
							params[newParamName] = clientPathParts[i]
							break
						} else {
							suffix++
						}
					}
				}
			}
			// this is for non-regex paths
		} else if part != clientPathParts[i] {
			return nil
		}
	}

	return params
}

func (r *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.trees == nil {
		r.trees = r.buildTrees()
	}

	for method := GET; method <= OPTIONS; method++ {
		if r.trees[method] != nil {
			if methodToString[method] == req.Method {
				treeMethod := r.trees[method]
				treeStartNode := treeMethod.startNode
				handler, ctxHandler, params := treeStartNode.InDepthSearch(req.URL.Path)

				if handler != nil {
					middlewares := findMatchingMiddleware(r.middlewares, req.URL.Path)
					var handlerFunc CtxFunc

					ctx := &Ctx{
						w:               w,
						r:               req,
						keys:            make(map[string]any),
						params:          params,
						routerPath:      req.URL.Path,
						middlewareIndex: -1,
						middlewares:     middlewares,
						handler:         handlerFunc,
						automatic:       r.automatic,
						maxMemory:       10 << 20, // 10 MB
						formParsed:      false,
					}

					if r.automatic {
						handlerFunc = ctxHandler

						for _, middleware := range middlewares {
							if pathMatch(middleware.Path, req.URL.Path) {
								if err := middleware.Handler(ctx); err != nil {
									http.Error(w, err.Error(), http.StatusInternalServerError)
									return
								}
							}
						}
					} else {
						handlerFunc = func(c *Ctx) error {
							c.middlewares = middlewares
							c.middlewareIndex = -1
							c.handler = ctxHandler
							c.automatic = r.automatic
							return c.Next()
						}

						handlerFunc(ctx)
					}
					handler(w, req)
					return
				}
			}
		}
	}
	http.NotFound(w, req)
}

// ...existing code...

func findMatchingMiddleware(middlewares []Middleware, path string) []Middleware {
	var matchingMiddleware []Middleware

	for _, middleware := range middlewares {
		if pathMatch(middleware.Path, path) {
			matchingMiddleware = append(matchingMiddleware, middleware)
		}
	}

	return matchingMiddleware
}

// buildTrees() must be done after init of the tree

func (r *Mux) setMiddlewareAutomatically(automatic bool) {
	r.automatic = automatic
}

func (r *Mux) buildTrees() map[Method]*Tree {
	var method Method

	trees := initTreesMap()
	if trees == nil {
		log.Println("[ERROR] Failed to initialize trees")
		return nil
	}
	for method = GET; method <= OPTIONS; method++ {
		if handlers, ok := r.handlers[method]; ok {
			for _, handler := range *handlers {
				t := trees[method]
				var err error
				t, err = r.createTreeIdea(handler, t, method)
				if err != nil {
					log.Println("[ERROR] Duplicate routes:", err)
					log.Printf("[WARNING] Skipping route: %s, the second one wrote \n", handler.path)
					continue
				}
				trees[method] = t
			}
		}
	}

	return trees
}
