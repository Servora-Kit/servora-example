package main

import (
	"flag"

	"github.com/Servora-Kit/servora/core/bootstrap"
	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/registry"
	"github.com/go-kratos/kratos/v3/transport/grpc"

	_ "go.uber.org/automaxprocs"
)

var (
	Name     = "worker.service"
	Version  = "dev"
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func newApp(rt *bootstrap.Runtime, reg registry.Registrar, gs *grpc.Server) *kratos.App {
	return rt.NewApp(
		kratos.Server(gs),
		kratos.Registrar(reg),
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
	return rt.Run(func() (*kratos.App, func(), error) {
		return wireApp(rt)
	})
}
