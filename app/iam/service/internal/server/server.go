package server

import (
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/jwks"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/openfga"
	"github.com/google/wire"
)

// iamComputedRelations defines computed relations for cache invalidation
// specific to the IAM service's OpenFGA model.
var iamComputedRelations = map[string][]string{
	"tenant":       {"can_view", "can_manage"},
	"organization": {"can_view", "can_manage", "can_manage_members"},
	"project":      {"can_view", "can_edit", "can_admin", "can_manage_members"},
}

// NewOpenFGAClient wraps openfga.NewClientOptional with IAM-specific options.
func NewOpenFGAClient(cfg *conf.App, l logger.Logger) *openfga.Client {
	return openfga.NewClientOptional(cfg, l,
		openfga.WithComputedRelations(iamComputedRelations),
	)
}

var ProviderSet = wire.NewSet(registry.NewRegistrar, telemetry.NewMetrics, jwks.NewKeyManagerFromConfig, NewOpenFGAClient, NewGRPCMiddleware, NewGRPCServer, NewHTTPMiddleware, NewHealthHandler, NewHTTPServer)
