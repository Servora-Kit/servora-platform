package main

import (
	"flag"

	iamconf "github.com/Servora-Kit/servora/api/gen/go/servora/iam/conf/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data"
	"github.com/Servora-Kit/servora/pkg/bootstrap"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// Injected at build time via -ldflags:
//
//	-X main.Name=<service-name>.service
//	-X main.Version=<git-tag>
var (
	Name     = "iam.service"
	Version  = "dev"
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func newApp(identity bootstrap.SvcIdentity, l log.Logger, reg registry.Registrar, gs *grpc.Server, hs *http.Server, seeder *data.Seeder) *kratos.App {
	return kratos.New(
		kratos.ID(identity.ID),
		kratos.Name(identity.Name),
		kratos.Version(identity.Version),
		kratos.Metadata(identity.Metadata),
		kratos.Logger(l),
		kratos.Server(gs, hs),
		kratos.Registrar(reg),
		kratos.BeforeStart(seeder.Run),
	)
}

func main() {
	flag.Parse()

	err := bootstrap.BootstrapAndRun(flagconf, Name, Version, func(runtime *bootstrap.Runtime) (*kratos.App, func(), error) {
		bc := runtime.Bootstrap
		bizConf, err := bootstrap.ScanBiz[iamconf.Biz](runtime)
		if err != nil {
			return nil, nil, err
		}
		return wireApp(bc.Server, bc.Discovery, bc.Registry, bc.Data, bc.App, bc.Trace, bc.Metrics, bc.Mail, bizConf, runtime.Identity, runtime.Logger)
	})
	if err != nil {
		panic(err)
	}
}
