package server

import (
	"github.com/Servora-Kit/servora/pkg/governance/registry"
	"github.com/Servora-Kit/servora/pkg/governance/telemetry"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(registry.NewRegistrar, telemetry.NewMetrics, NewGRPCServer, NewHTTPServer)
