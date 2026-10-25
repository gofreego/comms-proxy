package http_server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"github.com/gofreego/comms-proxy/internal/configs"
	"github.com/gofreego/comms-proxy/internal/repository"
	"github.com/gofreego/comms-proxy/internal/service"

	"github.com/gofreego/goutils/api"
	"github.com/gofreego/goutils/api/debug"
	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/goutils/metrics"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// HTTPServer serves the Inbox API (what the gateway routes to) under /comms-proxy/. Notifier
// is deliberately not registered here: sends only come in over gRPC, inside the cluster.
type HTTPServer struct {
	cfg    *configs.Configuration
	server *http.Server
}

func (a *HTTPServer) Name() string {
	return "HTTP_Server"
}

func (a *HTTPServer) Shutdown(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Panic(ctx, "failed to shutdown %s : %v", a.Name(), err)
	}
}

func NewHTTPServer(cfg *configs.Configuration) *HTTPServer {
	return &HTTPServer{cfg: cfg}
}

func (a *HTTPServer) Run(ctx context.Context) error {
	if a.cfg.Server.HTTPPort == 0 {
		logger.Panic(ctx, "http port is not provided")
	}

	svc := service.NewService(ctx, &a.cfg.Service, repository.GetInstance(ctx, &a.cfg.Repository))

	// Forward the gateway's identity headers into gRPC metadata so the inbox knows whose it is.
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "X-Profile-Id", "X-User-Id", "X-Profile-Ids", "X-User-Perms":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	api.RegisterSwaggerHandler(ctx, mux, "/comms-proxy/v1/swagger", "./api/docs/proto", "/commsproxy_v1/commsproxy.swagger.json")
	if err := commsproxy_v1.RegisterInboxHandlerServer(ctx, mux, svc); err != nil {
		logger.Panic(ctx, "failed to register inbox handlers : %v", err)
	}
	if a.cfg.Debug.Enabled {
		debug.RegisterDebugHandlersWithGateway(ctx, &a.cfg.Debug, mux, a.cfg.Logger.AppName, string(a.cfg.Logger.Build), "/comms-proxy/v1")
	}

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.Server.HTTPPort),
		Handler: metrics.WithHTTPMetrics(logger.WithRequestMiddleware(logger.WithRequestTimeMiddleware(api.CorsMiddleware(mux, a.cfg.Server.GetCORSConfig)))),
	}

	logger.Info(ctx, "Starting HTTP server on port %d", a.cfg.Server.HTTPPort)
	logger.Info(ctx, "Swagger UI is available at `http://localhost:%d/comms-proxy/v1/swagger`", a.cfg.Server.HTTPPort)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Panic(ctx, "failed to start http server : %v", err)
	}
	return nil
}
