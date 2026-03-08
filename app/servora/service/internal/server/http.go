package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/Servora-Kit/servora/api/gen/go/conf/v1"
	servorav1 "github.com/Servora-Kit/servora/api/gen/go/servora/service/v1"
	"github.com/Servora-Kit/servora/app/servora/service/internal/consts"
	innermw "github.com/Servora-Kit/servora/app/servora/service/internal/server/middleware"
	"github.com/Servora-Kit/servora/app/servora/service/internal/service"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/transport/server/http"
	svrmw "github.com/Servora-Kit/servora/pkg/transport/server/middleware"
)

// HTTPMiddleware 用于 Wire 注入的中间件切片包装类型
type HTTPMiddleware []middleware.Middleware

func NewHTTPMiddleware(
	trace *conf.Trace,
	m *telemetry.Metrics,
	l logger.Logger,
	authJWT innermw.AuthJWT,
) HTTPMiddleware {
	ms := svrmw.NewChainBuilder(logger.With(l, logger.WithModule("http/server/servora-service"))).
		WithTrace(trace).
		WithMetrics(m).
		Build()

	publicWhitelist := svrmw.NewWhiteList(svrmw.Exact,
		servorav1.OperationAuthServiceLoginByEmailPassword,
		servorav1.OperationAuthServiceRefreshToken,
		servorav1.OperationAuthServiceSignupByEmail,
		servorav1.OperationTestServiceTest,
		servorav1.OperationTestServiceHello,
	)

	userWhitelist := svrmw.NewWhiteList(svrmw.Exact,
		servorav1.OperationUserServiceCurrentUserInfo,
		servorav1.OperationUserServiceUpdateUser,
		servorav1.OperationAuthServiceLogout,
		servorav1.OperationTestServicePrivateTest,
	)

	// Admin 权限排除白名单 = 公开接口 ∪ User 级接口
	adminExcludeWhitelist := publicWhitelist.Merge(userWhitelist)

	ms = append(ms,
		selector.Server(authJWT(consts.User)).
			Match(publicWhitelist.MatchFunc()).
			Build(),
		selector.Server(authJWT(consts.Admin)).
			Match(adminExcludeWhitelist.MatchFunc()).
			Build(),
	)

	return ms
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	mw HTTPMiddleware,
	mtc *telemetry.Metrics,
	l logger.Logger,
	auth *service.AuthService,
	user *service.UserService,
	test *service.TestService,
) *khttp.Server {
	hlog := logger.With(l, logger.WithModule("http/server/servora-service"))

	opts := []http.ServerOption{
		http.WithLogger(hlog),
		http.WithMiddleware(mw...),
		http.WithMetrics(mtc),
		http.WithServices(
			func(s *khttp.Server) { servorav1.RegisterAuthServiceHTTPServer(s, auth) },
			func(s *khttp.Server) { servorav1.RegisterUserServiceHTTPServer(s, user) },
			func(s *khttp.Server) { servorav1.RegisterTestServiceHTTPServer(s, test) },
		),
	}
	if c != nil && c.Http != nil {
		opts = append(opts, http.WithConfig(c.Http))
		opts = append(opts, http.WithCORS(c.Http.Cors))
	}

	return http.NewServer(opts...)
}
