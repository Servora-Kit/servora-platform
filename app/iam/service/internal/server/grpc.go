package server

import (
	kmw "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	authnpb "github.com/Servora-Kit/servora/api/gen/go/servora/authn/service/v1"
	"github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/servora/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/service"
	"github.com/Servora-Kit/servora/pkg/authn"
	authjwt "github.com/Servora-Kit/servora/pkg/authn/jwt"
	"github.com/Servora-Kit/servora/pkg/authz"
	authzopenfga "github.com/Servora-Kit/servora/pkg/authz/openfga"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/jwks"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/openfga"
	"github.com/Servora-Kit/servora/pkg/redis"
	"github.com/Servora-Kit/servora/pkg/transport/server/grpc"
	svrmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

type GRPCMiddleware []kmw.Middleware

func NewGRPCMiddleware(
	trace *conf.Trace,
	mtc *telemetry.Metrics,
	l logger.Logger,
	km *jwks.KeyManager,
	fga *openfga.Client,
	rdb *redis.Client,
) GRPCMiddleware {
	ms := svrmw.NewChainBuilder(logger.With(l, "grpc/server/iam")).
		WithTrace(trace).
		WithMetrics(mtc).
		Build()

	publicWhitelist := svrmw.NewWhiteList(svrmw.Exact,
		authnpb.AuthnService_LoginByEmailPassword_FullMethodName,
		authnpb.AuthnService_RefreshToken_FullMethodName,
		authnpb.AuthnService_SignupByEmail_FullMethodName,
	)

	authnMw := authn.Server(authjwt.NewAuthenticator(authjwt.WithVerifier(km.Verifier())))

	fgaAuthorizerOpts := []authzopenfga.Option{}
	if rdb != nil {
		fgaAuthorizerOpts = append(fgaAuthorizerOpts, authzopenfga.WithRedisCache(rdb, openfga.DefaultCheckCacheTTL))
	}
	authzMw := authz.Server(
		authzopenfga.NewAuthorizer(fga, fgaAuthorizerOpts...),
		authz.WithRulesFuncs(userpb.AuthzRules, authnpb.AuthzRules),
	)

	ms = append(ms,
		selector.Server(authnMw).
			Match(publicWhitelist.MatchFunc()).
			Build(),
		authzMw,
	)

	return ms
}

func NewGRPCServer(
	c *conf.Server,
	mw GRPCMiddleware,
	l logger.Logger,
	authn *service.AuthnService,
	user *service.UserService,
) *kgrpc.Server {
	glog := logger.With(l, "grpc/server/iam")

	opts := []grpc.ServerOption{
		grpc.WithLogger(glog),
		grpc.WithMiddleware(mw...),
	}
	if c != nil && c.Grpc != nil {
		opts = append(opts, grpc.WithConfig(c.Grpc))
	}

	srv := grpc.NewServer(opts...)

	authnpb.RegisterAuthnServiceServer(srv, authn)
	userpb.RegisterUserServiceServer(srv, user)

	return srv
}
