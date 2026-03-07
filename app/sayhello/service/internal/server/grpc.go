package server

import (
	"crypto/tls"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	sayhellov1 "github.com/Servora-Kit/servora/api/gen/go/sayhello/service/v1"
	"github.com/Servora-Kit/servora/app/sayhello/service/internal/service"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/logger"
	srvmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

func NewGRPCServer(c *conf.Server, trace *conf.Trace, mtc *telemetry.Metrics, l logger.Logger, sayhello *service.SayHelloService) *grpc.Server {
	helper := logger.NewHelper(l)
	grpcLogger := logger.With(l, logger.WithModule("grpc/server/sayhello-service"))

	mds := srvmw.NewChainBuilder(grpcLogger).
		WithTrace(trace).
		WithMetrics(mtc).
		WithoutRateLimit().
		Build()

	var opts = []grpc.ServerOption{
		grpc.Middleware(mds...),
	}

	if c != nil && c.Grpc != nil {
		if c.Grpc.Network != "" {
			opts = append(opts, grpc.Network(c.Grpc.Network))
		}
		if c.Grpc.Addr != "" {
			opts = append(opts, grpc.Address(c.Grpc.Addr))
		}
		if c.Grpc.Timeout != nil {
			opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
		}
	}

	if c != nil && c.Grpc != nil && c.Grpc.Tls != nil && c.Grpc.Tls.Enable {
		cert, err := tls.LoadX509KeyPair(c.Grpc.Tls.CertPath, c.Grpc.Tls.KeyPath)
		if err != nil {
			helper.Fatalf("gRPC Server TLS: Failed to load key pair: %v", err)
		}
		creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
		opts = append(opts, grpc.Options(gogrpc.Creds(creds)))
	}

	srv := grpc.NewServer(opts...)
	sayhellov1.RegisterSayHelloServiceServer(srv, sayhello)
	return srv
}
