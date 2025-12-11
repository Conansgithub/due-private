package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler describes a due HTTP handler implemented on top of Echo.
type Handler = func(ctx Context) error

// Router provides the same routing surface as the Fiber HTTP组件,但内部使用 Echo。
type Router interface {
	Get(path string, handler any, middlewares ...any) Router
	Post(path string, handler any, middlewares ...any) Router
	Head(path string, handler any, middlewares ...any) Router
	Put(path string, handler any, middlewares ...any) Router
	Delete(path string, handler any, middlewares ...any) Router
	Connect(path string, handler any, middlewares ...any) Router
	Options(path string, handler any, middlewares ...any) Router
	Trace(path string, handler any, middlewares ...any) Router
	Patch(path string, handler any, middlewares ...any) Router
	All(path string, handler any, middlewares ...any) Router
	Add(methods []string, path string, handler any, middlewares ...any) Router
	Group(prefix string, middlewares ...any) Router
}

var allHTTPMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodHead,
	http.MethodOptions,
	http.MethodConnect,
	http.MethodTrace,
}

type router struct {
	app   *echo.Echo
	proxy *Proxy
}

func (r *router) Get(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodGet}, path, handler, middlewares...)
}

func (r *router) Post(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodPost}, path, handler, middlewares...)
}

func (r *router) Head(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodHead}, path, handler, middlewares...)
}

func (r *router) Put(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodPut}, path, handler, middlewares...)
}

func (r *router) Delete(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodDelete}, path, handler, middlewares...)
}

func (r *router) Connect(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodConnect}, path, handler, middlewares...)
}

func (r *router) Options(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodOptions}, path, handler, middlewares...)
}

func (r *router) Trace(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodTrace}, path, handler, middlewares...)
}

func (r *router) Patch(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodPatch}, path, handler, middlewares...)
}

func (r *router) All(path string, handler any, middlewares ...any) Router {
	return r.Add(allHTTPMethods, path, handler, middlewares...)
}

func (r *router) Add(methods []string, path string, handler any, middlewares ...any) Router {
	echoMiddlewares := wrapEchoMiddlewares(r.proxy, middlewares...)
	echoHandler := wrapEchoHandler(r.proxy, handler)
	if echoHandler == nil {
		return r
	}

	for _, method := range methods {
		r.app.Add(method, path, echoHandler, echoMiddlewares...)
	}

	return r
}

func (r *router) Group(prefix string, middlewares ...any) Router {
	echoMiddlewares := wrapEchoMiddlewares(r.proxy, middlewares...)
	grp := r.app.Group(prefix, echoMiddlewares...)
	return &routeGroup{proxy: r.proxy, group: grp}
}

type routeGroup struct {
	proxy *Proxy
	group *echo.Group
}

func (r *routeGroup) Get(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodGet}, path, handler, middlewares...)
}

func (r *routeGroup) Post(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodPost}, path, handler, middlewares...)
}

func (r *routeGroup) Head(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodHead}, path, handler, middlewares...)
}

func (r *routeGroup) Put(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodPut}, path, handler, middlewares...)
}

func (r *routeGroup) Delete(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodDelete}, path, handler, middlewares...)
}

func (r *routeGroup) Connect(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodConnect}, path, handler, middlewares...)
}

func (r *routeGroup) Options(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodOptions}, path, handler, middlewares...)
}

func (r *routeGroup) Trace(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodTrace}, path, handler, middlewares...)
}

func (r *routeGroup) Patch(path string, handler any, middlewares ...any) Router {
	return r.Add([]string{http.MethodPatch}, path, handler, middlewares...)
}

func (r *routeGroup) All(path string, handler any, middlewares ...any) Router {
	return r.Add(allHTTPMethods, path, handler, middlewares...)
}

func (r *routeGroup) Add(methods []string, path string, handler any, middlewares ...any) Router {
	echoMiddlewares := wrapEchoMiddlewares(r.proxy, middlewares...)
	echoHandler := wrapEchoHandler(r.proxy, handler)
	if echoHandler == nil {
		return r
	}

	for _, method := range methods {
		r.group.Add(method, path, echoHandler, echoMiddlewares...)
	}

	return r
}

func (r *routeGroup) Group(prefix string, middlewares ...any) Router {
	echoMiddlewares := wrapEchoMiddlewares(r.proxy, middlewares...)
	grp := r.group.Group(prefix, echoMiddlewares...)
	return &routeGroup{proxy: r.proxy, group: grp}
}

func wrapEchoHandler(proxy *Proxy, handler any) echo.HandlerFunc {
	switch h := handler.(type) {
	case echo.HandlerFunc:
		return h
	case Handler:
		return func(c echo.Context) error {
			return h(&context{Context: c, proxy: proxy})
		}
	default:
		return nil
	}
}

func wrapEchoMiddlewares(proxy *Proxy, middlewares ...any) []echo.MiddlewareFunc {
	echoMiddlewares := make([]echo.MiddlewareFunc, 0, len(middlewares))
	for _, middleware := range middlewares {
		switch m := middleware.(type) {
		case echo.MiddlewareFunc:
			echoMiddlewares = append(echoMiddlewares, m)
		case Handler:
			echoMiddlewares = append(echoMiddlewares, func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					if err := m(&context{Context: c, proxy: proxy}); err != nil {
						return err
					}
					return next(c)
				}
			})
		}
	}
	return echoMiddlewares
}
