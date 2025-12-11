package echo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Conansgithub/due-private/v2/component"
	"github.com/Conansgithub/due-private/v2/core/info"
	xnet "github.com/Conansgithub/due-private/v2/core/net"
	"github.com/Conansgithub/due-private/v2/log"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	component.Base
	opts  *options
	app   *echo.Echo
	proxy *Proxy
}

func NewServer(opts ...Option) *Server {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	s := &Server{opts: o}
	s.app = echo.New()
	s.app.HideBanner = true
	s.app.HidePort = true
	s.proxy = newProxy(s)

	if o.console {
		s.app.Use(echomw.Logger())
	}

	s.app.Use(echomw.Recover())

	if o.bodyLimit > 0 {
		limit := fmt.Sprintf("%dB", o.bodyLimit)
		s.app.Use(echomw.BodyLimit(limit))
	}

	if o.corsOpts.Enable {
		cfg := echomw.CORSConfig{
			AllowOrigins:     o.corsOpts.AllowOrigins,
			AllowMethods:     o.corsOpts.AllowMethods,
			AllowHeaders:     o.corsOpts.AllowHeaders,
			ExposeHeaders:    o.corsOpts.ExposeHeaders,
			AllowCredentials: o.corsOpts.AllowCredentials,
			MaxAge:           o.corsOpts.MaxAge,
		}
		s.app.Use(echomw.CORSWithConfig(cfg))
	}

	for i := range o.middlewares {
		switch handler := o.middlewares[i].(type) {
		case Handler:
			s.app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					if err := handler(&context{Context: c, proxy: s.proxy}); err != nil {
						return err
					}
					return next(c)
				}
			})
		case echo.MiddlewareFunc:
			s.app.Use(handler)
		case echo.HandlerFunc:
			s.app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					if err := handler(c); err != nil {
						return err
					}
					return next(c)
				}
			})
		}
	}

	return s
}

// Name 组件名称
func (s *Server) Name() string {
	return s.opts.name
}

// Init 初始化组件
func (s *Server) Init() {}

// Proxy 获取HTTP代理API
func (s *Server) Proxy() *Proxy {
	return s.proxy
}

// Start 启动组件
func (s *Server) Start() {
	listenAddr, exposeAddr, err := xnet.ParseAddr(s.opts.addr)
	if err != nil {
		log.Fatalf("echo addr parse failed: %v", err)
	}

	if s.opts.transporter != nil && s.opts.registry != nil {
		s.opts.transporter.SetDefaultDiscovery(s.opts.registry)
	}

	s.printInfo(exposeAddr)

	go func() {
		var runErr error
		if s.opts.certFile != "" && s.opts.keyFile != "" {
			runErr = s.app.StartTLS(listenAddr, s.opts.certFile, s.opts.keyFile)
		} else {
			runErr = s.app.Start(listenAddr)
		}

		if runErr != nil && runErr != http.ErrServerClosed {
			log.Fatalf("echo server startup failed: %v", runErr)
		}
	}()
}

func (s *Server) printInfo(addr string) {
	infos := make([]string, 0, 3)
	infos = append(infos, fmt.Sprintf("Name: %s", s.Name()))

	var baseURL string
	if s.opts.certFile != "" && s.opts.keyFile != "" {
		baseURL = fmt.Sprintf("https://%s", addr)
	} else {
		baseURL = fmt.Sprintf("http://%s", addr)
	}

	infos = append(infos, fmt.Sprintf("Url: %s", baseURL))

	if s.opts.swagOpts.Enable {
		infos = append(infos, fmt.Sprintf("Swagger: %s/%s", baseURL, strings.TrimPrefix(s.opts.swagOpts.BasePath, "/")))
	} else {
		infos = append(infos, "Swagger: -")
	}

	if s.opts.registry != nil {
		infos = append(infos, fmt.Sprintf("Registry: %s", s.opts.registry.Name()))
	} else {
		infos = append(infos, "Registry: -")
	}

	if s.opts.transporter != nil {
		infos = append(infos, fmt.Sprintf("Transporter: %s", s.opts.transporter.Name()))
	} else {
		infos = append(infos, "Transporter: -")
	}

	info.PrintBoxInfo("Echo", infos...)
}
