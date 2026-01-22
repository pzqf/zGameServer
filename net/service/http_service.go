package service

import (
	"fmt"
	"net/http"

	"github.com/pzqf/zEngine/zLog"
	"github.com/pzqf/zEngine/zObject"
	"github.com/pzqf/zGameServer/config"
	"go.uber.org/zap"
)

// HTTPHandlerFunc HTTP请求处理函数类型
type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

// RouteMap HTTP路由映射表
type RouteMap map[string]HTTPHandlerFunc

// HTTPService HTTP服务
type HTTPService struct {
	zObject.BaseObject
	server     *http.Server
	httpConfig *config.HTTPConfig
	routes     RouteMap
	mux        *http.ServeMux
}

// NewHTTPService 创建HTTP服务
func NewHTTPService() *HTTPService {
	hs := &HTTPService{
		routes: make(RouteMap),
		mux:    http.NewServeMux(),
	}
	hs.SetId(ServiceIdHttpServer)
	return hs
}

// Init 初始化HTTP服务
func (hs *HTTPService) Init() error {
	hs.httpConfig = config.GetHTTPConfig()

	// 如果HTTP服务未启用，直接返回
	if !hs.httpConfig.Enabled {
		zLog.Info("HTTP service is disabled")
		return nil
	}

	zLog.Info("Initializing HTTP service...", zap.String("listen_address", hs.httpConfig.ListenAddress))

	// 注册默认路由
	hs.registerDefaultRoutes()

	// 创建HTTP服务器
	hs.server = &http.Server{
		Addr:    hs.httpConfig.ListenAddress,
		Handler: hs.mux,
	}

	return nil
}

// Close 关闭HTTP服务
func (hs *HTTPService) Close() error {
	// 如果HTTP服务未启用，直接返回
	if !hs.httpConfig.Enabled {
		return nil
	}

	zLog.Info("Closing HTTP service...")
	if hs.server != nil {
		return hs.server.Close()
	}
	return nil
}

// Serve 启动HTTP服务
func (hs *HTTPService) Serve() {
	// 如果HTTP服务未启用，直接返回
	if !hs.httpConfig.Enabled {
		zLog.Info("HTTP service is disabled, skipping start")
		return
	}

	zLog.Info("Starting HTTP service...")
	if hs.server != nil {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zLog.Error("Failed to start HTTP service", zap.Error(err))
			return
		}
	}
}

// RegisterHandler 注册HTTP请求处理函数
func (hs *HTTPService) RegisterHandler(path string, handler HTTPHandlerFunc) {
	hs.routes[path] = handler
	hs.mux.HandleFunc(path, handler)
	zLog.Debug("Registered HTTP handler", zap.String("path", path))
}

// UnregisterHandler 注销HTTP请求处理函数
func (hs *HTTPService) UnregisterHandler(path string) {
	delete(hs.routes, path)
	// HTTP Mux不支持直接删除路由，需要重新创建
	hs.recreateMux()
	zLog.Debug("Unregistered HTTP handler", zap.String("path", path))
}

// recreateMux 重新创建HTTP路由Mux
func (hs *HTTPService) recreateMux() {
	newMux := http.NewServeMux()
	// 重新注册所有路由
	for path, handler := range hs.routes {
		newMux.HandleFunc(path, handler)
	}
	// 替换旧的Mux
	hs.mux = newMux
}

// registerDefaultRoutes 注册默认路由
func (hs *HTTPService) registerDefaultRoutes() {
	// 健康检查路由
	hs.RegisterHandler("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// 服务器状态路由
	hs.RegisterHandler("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server is running")
	})
}
