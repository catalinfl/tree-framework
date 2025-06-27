package tree

func (r *Mux) GET(url string, t CtxFunc) {
	r.addMutationRoute(url, GET, t)
}

func (r *Mux) POST(url string, t CtxFunc) {
	r.addMutationRoute(url, POST, t)
}

func (r *Mux) PUT(url string, t CtxFunc) {
	r.addMutationRoute(url, PUT, t)
}

func (r *Mux) DELETE(url string, t CtxFunc) {
	r.addMutationRoute(url, DELETE, t)
}

func (r *Mux) PATCH(url string, t CtxFunc) {
	r.addMutationRoute(url, PATCH, t)
}

func (r *Mux) HEAD(url string, t CtxFunc) {
	r.addMutationRoute(url, HEAD, t)
}

func (r *Mux) OPTIONS(url string, t CtxFunc) {
	r.addMutationRoute(url, OPTIONS, t)
}

func (r *Mux) USE(url string, t CtxFunc) {

	if url == "" {
		url = "/"
	}

	middlewareFunc := MiddlewareFunc(t)
	middleware := Middleware{
		Path:    url,
		Handler: middlewareFunc,
	}
	r.middlewares = append(r.middlewares, middleware)

}
