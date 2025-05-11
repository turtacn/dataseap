package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/turtacn/dataseap/pkg/common/constants"
	"github.com/turtacn/dataseap/pkg/config"
	"github.com/turtacn/dataseap/pkg/logger"

	// Import domain service interfaces
	"github.com/turtacn/dataseap/pkg/domain/ingestion"
	"github.com/turtacn/dataseap/pkg/domain/management/lifecycle"
	"github.com/turtacn/dataseap/pkg/domain/management/metadata"
	"github.com/turtacn/dataseap/pkg/domain/management/workload"
	"github.com/turtacn/dataseap/pkg/domain/query"

	// Import generated gRPC server code
	apiv1 "github.com/turtacn/dataseap/api/v1" // Alias for generated proto code
)

// Server holds the gRPC server instance and its configuration.
// Server 保存gRPC服务器实例及其配置。
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	cfg        config.ServerConfig
	// Add other dependencies if needed, like a global error handler/interceptor config
}

// ServiceRegistry holds all domain services required by gRPC handlers.
// ServiceRegistry 保存gRPC处理器所需的所有领域服务。
type ServiceRegistry struct {
	IngestionSvc ingestion.Service
	QuerySvc     query.Service
	WorkloadSvc  workload.Service
	MetadataSvc  metadata.Service
	LifecycleSvc lifecycle.Service
}

// NewServer creates a new gRPC server instance.
// NewServer 创建一个新的gRPC服务器实例。
func NewServer(cfg config.ServerConfig, services ServiceRegistry, grpcOpts ...grpc.ServerOption) (*Server, error) {
	l := logger.L().With("component", "gRPCServer")

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPCPort)
	if cfg.GRPCPort == 0 {
		address = fmt.Sprintf("%s:%d", cfg.Host, constants.DefaultGRPCPort)
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		l.Errorw("Failed to listen for gRPC", "address", address, "error", err)
		return nil, fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	l.Infow("gRPC server listening", "address", address)

	// Default gRPC server options
	// TODO: Add interceptors for logging, metrics, auth, recovery
	opts := []grpc.ServerOption{
		// grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		// 	// logging.UnaryServerInterceptor(l),
		// 	// metrics.UnaryServerInterceptor(),
		// 	// recovery.UnaryServerInterceptor(),
		// )),
		// grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		// 	// Stream interceptors here
		// )),
	}
	opts = append(opts, grpcOpts...)
	s := grpc.NewServer(opts...)

	// Register services
	if services.IngestionSvc != nil {
		ingestionHandler := NewIngestionHandler(services.IngestionSvc)
		apiv1.RegisterIngestionServiceServer(s, ingestionHandler)
		l.Info("Registered IngestionService")
	}
	if services.QuerySvc != nil {
		queryHandler := NewQueryHandler(services.QuerySvc)
		apiv1.RegisterQueryServiceServer(s, queryHandler)
		l.Info("Registered QueryService")
	}
	if services.WorkloadSvc != nil || services.MetadataSvc != nil || services.LifecycleSvc != nil {
		managementHandler := NewManagementHandler(services.WorkloadSvc, services.MetadataSvc, services.LifecycleSvc)
		apiv1.RegisterManagementServiceServer(s, managementHandler)
		l.Info("Registered ManagementService")
	}

	// Register health check service
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	l.Info("Registered HealthCheckService")

	// Register reflection service (for tools like grpcurl)
	reflection.Register(s)
	l.Info("Registered gRPC ReflectionService")

	return &Server{
		grpcServer: s,
		listener:   lis,
		cfg:        cfg,
	}, nil
}

// ListenAndServe starts the gRPC server and blocks until the server stops.
// ListenAndServe 启动gRPC服务器并阻塞直到服务器停止。
func (s *Server) ListenAndServe() error {
	l := logger.L().With("component", "gRPCServer", "address", s.listener.Addr().String())
	l.Info("Starting gRPC server...")
	if err := s.grpcServer.Serve(s.listener); err != nil {
		l.Errorw("gRPC server failed to serve", "error", err)
		return fmt.Errorf("gRPC server failed: %w", err)
	}
	return nil
}

// GracefulStop attempts to gracefully stop the gRPC server.
// GracefulStop 尝试优雅地停止gRPC服务器。
func (s *Server) GracefulStop(ctx context.Context) error {
	l := logger.L().With("component", "gRPCServer")
	l.Info("Attempting graceful shutdown of gRPC server...")

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		l.Warn("gRPC server graceful shutdown timed out, forcing stop.")
		s.grpcServer.Stop() // Force stop if context deadline is exceeded
		return ctx.Err()
	case <-stopped:
		l.Info("gRPC server shutdown gracefully.")
		return nil
	}
}

// Stop stops the gRPC server immediately.
// Stop 立即停止gRPC服务器。
func (s *Server) Stop() {
	logger.L().Info("Forcing gRPC server to stop.")
	s.grpcServer.Stop()
}

// Address returns the address the server is listening on.
// Address 返回服务器正在监听的地址。
func (s *Server) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}
