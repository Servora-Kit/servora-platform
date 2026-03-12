package server

import (
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/Servora-Kit/servora/pkg/jwks"
	"github.com/Servora-Kit/servora/pkg/openfga"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(registry.NewRegistrar, telemetry.NewMetrics, jwks.NewKeyManagerFromConfig, openfga.NewClientOptional, NewGRPCMiddleware, NewGRPCServer, NewHTTPMiddleware, NewHealthHandler, NewHTTPServer)
