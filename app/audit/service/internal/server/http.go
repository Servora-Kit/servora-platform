package server

import (
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	auditsvcpb "github.com/Servora-Kit/servora-platform/api/gen/go/servora/audit/service/v1"
	"github.com/Servora-Kit/servora-platform/app/audit/service/internal/service"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/obs/telemetry"
	svrhttp "github.com/Servora-Kit/servora/transport/server/http"
	"github.com/Servora-Kit/servora/transport/server/middleware"
)

// NewHTTPServer creates the HTTP server for the audit service.
func NewHTTPServer(c *conf.Server, trace *conf.Trace, m *telemetry.Metrics, l logger.Logger, svc *service.AuditService) *khttp.Server {
	hlog := logger.With(l, "audit/server/http")

	ms := middleware.NewChainBuilder(hlog).
		WithTrace(trace).
		WithMetrics(m).
		Build()

	opts := []svrhttp.ServerOption{
		svrhttp.WithLogger(hlog),
		svrhttp.WithMiddleware(ms...),
		svrhttp.WithMetrics(m),
		svrhttp.WithServices(func(s *khttp.Server) {
			auditsvcpb.RegisterAuditHTTPServiceHTTPServer(s, svc)
		}),
	}
	if c != nil && c.Http != nil {
		opts = append(opts, svrhttp.WithConfig(c.Http))
		opts = append(opts, svrhttp.WithCORS(c.Http.Cors))
	}

	return svrhttp.NewServer(opts...)
}
