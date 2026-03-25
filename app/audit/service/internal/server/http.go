package server

import (
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	auditsvcpb "github.com/Servora-Kit/servora-platform/api/gen/go/servora/audit/service/v1"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora-platform/app/audit/service/internal/service"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/transport/server/middleware"
	svrhttp "github.com/Servora-Kit/servora/pkg/transport/server/http"
)

// NewHTTPServer creates the HTTP server for the audit service.
func NewHTTPServer(c *conf.Server, trace *conf.Trace, l logger.Logger, svc *service.AuditService) *khttp.Server {
	hlog := logger.With(l, "audit/server/http")

	ms := middleware.NewChainBuilder(hlog).
		WithTrace(trace).
		Build()

	opts := []svrhttp.ServerOption{
		svrhttp.WithLogger(hlog),
		svrhttp.WithMiddleware(ms...),
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
