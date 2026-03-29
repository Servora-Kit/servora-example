package server

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
	confv1 "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	logger "github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/platform/bootstrap"
	bootconfig "github.com/Servora-Kit/servora/platform/bootstrap/config"
	"google.golang.org/protobuf/proto"
)

func TestNewTCPServerLoadsConfigViaScanConf(t *testing.T) {
	t.Parallel()

	cfgDir := filepath.Join("..", "..", "configs", "local")

	bc, c, err := bootconfig.LoadBootstrap(cfgDir, "master.service", false)
	if err != nil {
		t.Fatalf("load bootstrap config: %v", err)
	}
	defer func() { _ = c.Close() }()

	rt := &bootstrap.Runtime{
		Bootstrap: bc,
		Config:    c,
		Logger:    logger.New(bc.GetApp()),
	}

	tcpCfg, err := bootstrap.ScanConf[tcpconf.Server](rt)
	if err != nil {
		t.Fatalf("scan tcp config: %v", err)
	}
	if got := tcpCfg.GetListen().GetAddr(); got != "0.0.0.0:8014" {
		t.Fatalf("tcp listen addr=%q want=%q", got, "0.0.0.0:8014")
	}

	// 使用项目真实配置完成 ScanConf，再将监听地址覆写为随机端口，避免测试端口冲突。
	runtimeCfg, ok := proto.Clone(tcpCfg).(*tcpconf.Server)
	if !ok {
		t.Fatalf("clone tcp config failed, got %T", tcpCfg)
	}
	if runtimeCfg.Listen == nil {
		runtimeCfg.Listen = &confv1.Server_Listen{}
	}
	runtimeCfg.Listen.Addr = "127.0.0.1:0"

	srv := NewTCPServer(runtimeCfg, rt.Logger, service.NewTCPCommandService(nil))
	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("resolve tcp endpoint: %v", err)
	}
	if ep == nil {
		t.Fatal("tcp endpoint is nil")
	}

	if err := srv.Start(context.Background()); err != nil {
		t.Fatalf("start tcp server: %v", err)
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Stop(stopCtx); err != nil {
		t.Fatalf("stop tcp server: %v", err)
	}
}
