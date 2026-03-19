package server

import (
	"strings"

	kmw "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	authnpb "github.com/Servora-Kit/servora/api/gen/go/authn/service/v1"
	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	iamv1 "github.com/Servora-Kit/servora/api/gen/go/iam/service/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"
	iammw "github.com/Servora-Kit/servora/app/iam/service/internal/server/middleware"
	"github.com/Servora-Kit/servora/app/iam/service/internal/service"
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
	ms := svrmw.NewChainBuilder(logger.With(l, logger.WithModule("grpc/server/iam-service"))).
		WithTrace(trace).
		WithMetrics(mtc).
		Build()

	publicWhitelist := svrmw.NewWhiteList(svrmw.Exact,
		authnpb.AuthnService_LoginByEmailPassword_FullMethodName,
		authnpb.AuthnService_RefreshToken_FullMethodName,
		authnpb.AuthnService_SignupByEmail_FullMethodName,
	)

	authn := iammw.Authn(iammw.WithVerifier(km.Verifier()))

	authzRules := remapAuthzRulesForGRPC(iamv1.AuthzRules)
	authzOpts := []iammw.AuthzOption{
		iammw.WithFGAClient(fga),
		iammw.WithAuthzRules(authzRules),
	}
	if rdb != nil {
		authzOpts = append(authzOpts, iammw.WithAuthzCache(rdb, openfga.DefaultCheckCacheTTL))
	}
	authz := iammw.Authz(authzOpts...)

	ms = append(ms,
		selector.Server(authn).
			Match(publicWhitelist.MatchFunc()).
			Build(),
		authz,
	)

	return ms
}

// remapAuthzRulesForGRPC converts IAM wrapper operation names to domain proto
// operation names used by gRPC service registrations.
func remapAuthzRulesForGRPC(src map[string]iamv1.AuthzRuleEntry) map[string]iamv1.AuthzRuleEntry {
	dst := make(map[string]iamv1.AuthzRuleEntry, len(src))
	for op, r := range src {
		dst[remapIAMOpToGRPC(op)] = r
	}
	return dst
}

const iamServicePrefix = "/iam.service.v1."

func remapIAMOpToGRPC(iamOp string) string {
	if !strings.HasPrefix(iamOp, iamServicePrefix) {
		return iamOp
	}
	rest := iamOp[len(iamServicePrefix):]
	slashIdx := strings.Index(rest, "/")
	if slashIdx < 0 {
		return iamOp
	}
	svcName := rest[:slashIdx]
	method := rest[slashIdx:]
	domain := strings.ToLower(strings.TrimSuffix(svcName, "Service"))
	return "/" + domain + ".service.v1." + svcName + method
}

func NewGRPCServer(
	c *conf.Server,
	mw GRPCMiddleware,
	l logger.Logger,
	authn *service.AuthnService,
	user *service.UserService,
) *kgrpc.Server {
	glog := logger.With(l, logger.WithModule("grpc/server/iam-service"))

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
