package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/satriajidam/go-echo-skeleton/pkg/log"
	"github.com/satriajidam/go-echo-skeleton/pkg/server/http/middleware/logger"
	"github.com/satriajidam/go-echo-skeleton/pkg/server/http/middleware/requestid"
)

var (
	CORSDefaultAllowOrigins = []string{"*"}
	CORSDefaultAllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	// CORS safelisted-request-header: https://fetch.spec.whatwg.org/#cors-safelisted-request-header
	// CORS forbidden-header-name: https://fetch.spec.whatwg.org/#forbidden-header-name
	CORSDefaultAllowHeaders = []string{
		"Accept",
		"Accept-Charset",
		"Accept-Encoding",
		"Accept-Language",
		"Content-Language",
		"Content-Length",
		"Content-Type",
		"Host",
		"Origin",
	}
	CORSDefaultAllowCredentials = true
	CORSDefaultMaxAge           = 0
)

// Server represents the implementation of HTTP server object.
type Server struct {
	RouterGroup
	router       *echo.Echo
	loggerConfig *logger.Config
	middlewares  []echo.MiddlewareFunc
	routes       []route
	enableCORS   bool
	CORS         *middleware.CORSConfig
	Port         string
}

func NewServer(port string, enableCORS bool, enablePredefinedRoutes bool) *Server {
	routes := []route{}

	server := &Server{
		router: echo.New(),
		middlewares: []echo.MiddlewareFunc{
			middleware.Recover(),
			requestid.New(),
		},
		loggerConfig: &logger.Config{
			Stdout:   log.Stdout(),
			Stderr:   log.Stderr(),
			Routes:   []logger.Route{},
			SkipPath: []string{},
		},
		enableCORS: enableCORS,
		CORS: &middleware.CORSConfig{
			AllowCredentials: CORSDefaultAllowCredentials,
		},
		routes: routes,
		Port:   port,
	}

	server.RouterGroup = RouterGroup{
		server: server,
	}

	return server
}

// AddMiddleware adds a gin middleware the HTTP server.
func (s *Server) AddMiddleware(m echo.MiddlewareFunc) {
	s.middlewares = append(s.middlewares, m)
}

// LoggerSkipPaths registers endpoint paths that you want to skip from being logged
// by the logger middleware.
func (s *Server) LoggerSkipPaths(paths ...string) {
	s.loggerConfig.SkipPath = append(s.loggerConfig.SkipPath, paths...)
}

// GetRoutePaths retrieves all route paths registerd to this HTTP server.
func (s *Server) GetRoutePaths() []string {
	paths := []string{}
	for _, r := range s.routes {
		paths = append(paths, r.relativePath)
	}
	return paths
}

func (s *Server) loadLoggerRoutes() {
	for _, route := range s.routes {
		s.loggerConfig.Routes = append(
			s.loggerConfig.Routes,
			logger.Route{
				Method:       route.method,
				RelativePath: route.relativePath,
				LogPayload:   route.logPayload,
			},
		)
	}
}

func (s *Server) loadRoutes() {
	for _, route := range s.routes {
		switch route.method {
		case http.MethodGet:
			s.router.GET(route.relativePath, route.handler, route.middlewares...)
		case http.MethodHead:
			s.router.HEAD(route.relativePath, route.handler, route.middlewares...)
		case http.MethodPost:
			s.router.POST(route.relativePath, route.handler, route.middlewares...)
		case http.MethodPut:
			s.router.PUT(route.relativePath, route.handler, route.middlewares...)
		case http.MethodPatch:
			s.router.PATCH(route.relativePath, route.handler, route.middlewares...)
		case http.MethodDelete:
			s.router.DELETE(route.relativePath, route.handler, route.middlewares...)
		case http.MethodOptions:
			s.router.OPTIONS(route.relativePath, route.handler, route.middlewares...)
		}
	}
}

func (s *Server) setupCORS() {
	if s.enableCORS {
		if len(s.CORS.AllowOrigins) <= 0 {
			s.CORS.AllowOrigins = CORSDefaultAllowOrigins
		}
		if len(s.CORS.AllowMethods) <= 0 {
			s.CORS.AllowMethods = CORSDefaultAllowMethods
		}
		if len(s.CORS.AllowHeaders) <= 0 {
			s.CORS.AllowHeaders = CORSDefaultAllowHeaders
		}
		if s.CORS.MaxAge <= 0 {
			s.CORS.MaxAge = CORSDefaultMaxAge
		}
		s.AddMiddleware(middleware.CORSWithConfig(*s.CORS))
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	log.Info(fmt.Sprintf("Start HTTP server on port %s", s.Port))
	s.loadLoggerRoutes()
	s.AddMiddleware(logger.New(s.Port, *s.loggerConfig))
	s.setupCORS()
	s.router.Use(s.middlewares...)
	s.loadRoutes()
	if err := s.router.Start(fmt.Sprintf(":%s", s.Port)); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	log.Info(fmt.Sprintf("Stop HTTP server on port %s", s.Port))
	if err := s.router.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

type RouteOption struct {
	Path        string
	Handler     echo.HandlerFunc
	Middlewares []echo.MiddlewareFunc
	LogPayload  bool
}

type route struct {
	method       string
	relativePath string
	handler      echo.HandlerFunc
	middlewares  []echo.MiddlewareFunc
	logPayload   bool
}

func (s *Server) appendRoute(
	method, relativePath string,
	handler echo.HandlerFunc,
	middlewares []echo.MiddlewareFunc,
	logPayload bool,
) {
	s.routes = append(s.routes, route{
		method:       method,
		relativePath: relativePath,
		handler:      handler,
		middlewares:  middlewares,
		logPayload:   logPayload,
	})
}

// POST registers HTTP server endpoint with Post method.
func (s *Server) POST(opt RouteOption) {
	s.appendRoute(http.MethodPost, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// GET registers HTTP server endpoint with Get method.
func (s *Server) GET(opt RouteOption) {
	s.appendRoute(http.MethodGet, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// DELETE registers HTTP server endpoint with Delete method.
func (s *Server) DELETE(opt RouteOption) {
	s.appendRoute(http.MethodDelete, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// PATCH registers HTTP server endpoint with Patch method.
func (s *Server) PATCH(opt RouteOption) {
	s.appendRoute(http.MethodPatch, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// PUT registers HTTP server endpoint with Put method.
func (s *Server) PUT(opt RouteOption) {
	s.appendRoute(http.MethodPut, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// OPTIONS registers HTTP server endpoint with Options method.
func (s *Server) OPTIONS(opt RouteOption) {
	s.appendRoute(http.MethodOptions, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// HEAD registers HTTP server endpoint with Head method.
func (s *Server) HEAD(opt RouteOption) {
	s.appendRoute(http.MethodHead, opt.Path, opt.Handler, opt.Middlewares, opt.LogPayload)
}

// RouterGroup groups path under one path prefix.
type RouterGroup struct {
	prefix string
	server *Server
}

// Group creates new RouterGroup with the given path prefix.
func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	return &RouterGroup{
		prefix: prefix,
		server: rg.server,
	}
}

// POST registers HTTP server endpoint with Post method.
func (rg *RouterGroup) POST(opt RouteOption) {
	opt.Path = rg.prefix + opt.Path
	rg.server.POST(opt)
}

// GET registers HTTP server endpoint with Get method.
func (rg *RouterGroup) GET(opt RouteOption) {
	opt.Path = rg.prefix + opt.Path
	rg.server.GET(opt)
}

// DELETE registers HTTP server endpoint with Delete method.
func (rg *RouterGroup) DELETE(opt RouteOption) {
	opt.Path = rg.prefix + opt.Path
	rg.server.DELETE(opt)
}

// PATCH registers HTTP server endpoint with Patch method.
func (rg *RouterGroup) PATCH(opt RouteOption) {
	opt.Path = rg.prefix + opt.Path
	rg.server.PATCH(opt)
}

// PUT registers HTTP server endpoint with Put method.
func (rg *RouterGroup) PUT(opt RouteOption) {
	opt.Path = rg.prefix + opt.Path
	rg.server.PUT(opt)
}

// OPTIONS registers HTTP server endpoint with Options method.
func (rg *RouterGroup) OPTIONS(opt RouteOption) {
	rg.server.OPTIONS(opt)
}

// HEAD registers HTTP server endpoint with Head method.
func (rg *RouterGroup) HEAD(opt RouteOption) {
	opt.Path = rg.prefix + opt.Path
	rg.server.HEAD(opt)
}
