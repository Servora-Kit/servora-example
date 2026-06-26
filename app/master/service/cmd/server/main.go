package main

import (
	"context"
	"flag"

	tcpk "github.com/Servora-Kit/servora-transport/server/tcp"
	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"

	"github.com/Servora-Kit/servora/core/bootstrap"
	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/registry"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	_ "go.uber.org/automaxprocs"
)

var (
	Name     = "master.service"
	Version  = "dev"
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func newApp(rt *bootstrap.Runtime, reg registry.Registrar, gs *grpc.Server, hs *http.Server, ts *tcpk.Server) *kratos.App {
	return rt.NewApp(
		kratos.Server(gs, hs), // TCP server is managed manually to avoid Consul registration
		kratos.Registrar(reg),
		// 不将 TCP server 注册到 Consul，手动管理生命周期。
		kratos.AfterStart(func(ctx context.Context) error {
			return ts.Start(ctx)
		}),
		kratos.BeforeStop(func(ctx context.Context) error {
			return ts.Stop(ctx)
		}),
	)
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() (err error) {
	rt, err := bootstrap.NewRuntime(flagconf, bootstrap.Name(Name), bootstrap.Version(Version))
	if err != nil {
		return err
	}
	tcpCfg := &tcpconf.Server{}
	if err := bootstrap.Scan(rt, tcpCfg); err != nil {
		return err
	}

	return rt.Run(func() (*kratos.App, func(), error) {
		return wireApp(rt, tcpCfg)
	})
}
