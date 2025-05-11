package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/dataseap/pkg/common/constants"
	"github.com/turtacn/dataseap/pkg/config"
	"github.com/turtacn/dataseap/pkg/logger"
	// Import domain service interfaces for handlers
	"github.com/turtacn/dataseap/pkg/domain/ingestion"
	"github.com/turtacn/dataseap/pkg/domain/query"
	"github.com/turtacn/dataseap/pkg/domain/management/lifecycle"
	"github.com/turtacn/dataseap/pkg/domain/management/metadata"
	"github.com/turtacn/dataseap/pkg/domain/management/workload"
)

// Server holds the HTTP server instance (Gin engine) and its configuration.
// Server 保存HTTP服务器实例 (Gin引擎) 及其配置。
type Server struct {
	engine *gin.Engine
	server *http.Server
	cfg    config.ServerConfig
	// services (passed to router setup)
}

// ServiceRegistry holds all domain services required by HTTP handlers.
// ServiceRegistry 保存HTTP处理器所需的所有领域服务。
type ServiceRegistry struct {
	IngestionSvc ingestion.Service
	QuerySvc     query.Service
	WorkloadSvc  workload.Service
	MetadataSvc  metadata.Service
	LifecycleSvc lifecycle.Service
}


// NewServer creates a new HTTP server instance using Gin.
// NewServer 使用Gin创建一个新的HTTP服务器实例。
func NewServer(cfg config.ServerConfig, services ServiceRegistry) (*Server, error) {
	l := logger.L().With("component", "HTTPServer")

	if strings.ToLower(cfg.Mode) == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	// TODO: Add standard middleware: Logger, Recovery, CORS, Metrics, Tracing, RequestID
	engine.Use(GinLogger(l)) // Custom logger middleware
	engine.Use(gin.Recovery())
	// engine.Use(cors.Default()) // Example CORS

	// Setup routes
	// Pass domain services to the router setup function
	SetupRouter(engine, services)
	l.Info("HTTP routes configured")

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if cfg.Port == 0 {
		address = fmt.Sprintf("%s:%d", cfg.Host, constants.DefaultAPIPort)
	}

	httpServer := &http.Server{
		Addr:           address,
		Handler:        engine,
		ReadTimeout:    time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	return &Server{
		engine: engine,
		server: httpServer,
		cfg:    cfg,
	}, nil
}

// ListenAndServe starts the HTTP server and blocks until the server stops.
// ListenAndServe 启动HTTP服务器并阻塞直到服务器停止。
func (s *Server) ListenAndServe() error {
	l := logger.L().With("component", "HTTPServer", "address", s.server.Addr)
	l.Info("Starting HTTP server...")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Errorw("HTTP server failed to listen and serve", "error", err)
		return fmt.Errorf("HTTP server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP server.
// Shutdown 优雅地关闭HTTP服务器。
func (s *Server) Shutdown(ctx context.Context) error {
	l := logger.L().With("component", "HTTPServer")
	l.Info("Attempting graceful shutdown of HTTP server...")
	if err := s.server.Shutdown(ctx); err != nil {
		l.Errorw("HTTP server graceful shutdown failed", "error", err)
		return fmt.Errorf("HTTP server shutdown failed: %w", err)
	}
	l.Info("HTTP server shutdown gracefully.")
	return nil
}

// Address returns the address the server is listening on.
// Address 返回服务器正在监听的地址。
func (s *Server) Address() string {
	if s.server != nil {
		return s.server.Addr
	}
	return ""
}

// GinLogger is a custom Gin middleware for logging using zap.
// GinLogger 是一个使用zap进行日志记录的自定义Gin中间件。
func GinLogger(l *logger.zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next() // Process request

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String() // Get private errors

		logFields := []interface{}{
			"status_code", statusCode,
			"latency_ms", latency.Milliseconds(),
			"client_ip", clientIP,
			"method", method,
			"path", path,
		}
		if query != "" {
			logFields = append(logFields, "query", query)
		}
		if errorMessage != "" {
			logFields = append(logFields, "error", errorMessage)
		}

		if statusCode >= http.StatusInternalServerError {
			l.Errorw("HTTP request error", logFields...)
		} else if statusCode >= http.StatusBadRequest {
			l.Warnw("HTTP request warning/client error", logFields...)
		} else {
			l.Infow("HTTP request", logFields...)
		}
	}
}