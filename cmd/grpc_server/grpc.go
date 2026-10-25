package grpc_server

import (
	"context"
	"fmt"
	"net"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"github.com/gofreego/comms-proxy/internal/configs"
	"github.com/gofreego/comms-proxy/internal/repository"
	"github.com/gofreego/comms-proxy/internal/service"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/goutils/metrics"
	"google.golang.org/grpc"
)

// GRPCServer serves both APIs on the cluster-internal gRPC port: Notifier for other services'
// sends, and Inbox.
type GRPCServer struct {
	cfg    *configs.Configuration
	server *grpc.Server
}

func (a *GRPCServer) Name() string {
	return "GRPC_Server"
}

func (a *GRPCServer) Shutdown(ctx context.Context) {
	a.server.GracefulStop()
}

func NewGRPCServer(cfg *configs.Configuration) *GRPCServer {
	return &GRPCServer{cfg: cfg}
}

func (a *GRPCServer) Run(ctx context.Context) error {
	if a.cfg.Server.GRPCPort == 0 {
		logger.Panic(ctx, "grpc port is not provided")
	}

	svc := service.NewService(ctx, &a.cfg.Service, repository.GetInstance(ctx, &a.cfg.Repository))

	a.server = grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()))
	commsproxy_v1.RegisterNotifierServer(a.server, svc)
	commsproxy_v1.RegisterInboxServer(a.server, svc)

	logger.Info(ctx, "Starting gRPC server on port %d", a.cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.GRPCPort))
	if err != nil {
		logger.Panic(ctx, "failed to listen for grpc server: %v", err)
	}
	if err := a.server.Serve(lis); err != nil {
		logger.Panic(ctx, "failed to start grpc server: %v", err)
	}
	return nil
}
