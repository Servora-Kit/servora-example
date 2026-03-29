package main

import (
	"bufio"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
	confv1 "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	logger "github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/platform/bootstrap"
	bootconfig "github.com/Servora-Kit/servora/platform/bootstrap/config"
	"github.com/go-kratos/kratos/v2"
	"google.golang.org/protobuf/proto"
)

func TestMasterServiceTCPE2E(t *testing.T) {
	t.Parallel()

	cfgDir := filepath.Join("..", "..", "configs", "local")
	bc, c, err := bootconfig.LoadBootstrap(cfgDir, "master.service", false)
	if err != nil {
		t.Fatalf("load bootstrap config: %v", err)
	}
	defer func() { _ = c.Close() }()

	appLogger := logger.New(bc.GetApp())
	rt := &bootstrap.Runtime{
		Bootstrap: bc,
		Config:    c,
		Logger:    appLogger,
		Identity: bootstrap.SvcIdentity{
			Name:     "master.service",
			Version:  "test",
			ID:       "master.service-e2e",
			Metadata: map[string]string{"test": "e2e"},
		},
	}

	scannedTCP, err := bootstrap.ScanConf[tcpconf.Server](rt)
	if err != nil {
		t.Fatalf("scan tcp config: %v", err)
	}
	if got := scannedTCP.GetListen().GetAddr(); got != "0.0.0.0:8014" {
		t.Fatalf("tcp listen addr=%q want=%q", got, "0.0.0.0:8014")
	}

	serverCfg, ok := proto.Clone(bc.GetServer()).(*confv1.Server)
	if !ok {
		t.Fatalf("clone server config failed, got %T", bc.GetServer())
	}
	if serverCfg.Http == nil {
		serverCfg.Http = &confv1.Server_HTTP{}
	}
	if serverCfg.Http.Listen == nil {
		serverCfg.Http.Listen = &confv1.Server_Listen{}
	}
	if serverCfg.Grpc == nil {
		serverCfg.Grpc = &confv1.Server_GRPC{}
	}
	if serverCfg.Grpc.Listen == nil {
		serverCfg.Grpc.Listen = &confv1.Server_Listen{}
	}
	serverCfg.Http.Listen.Addr = "127.0.0.1:0"
	serverCfg.Grpc.Listen.Addr = "127.0.0.1:0"

	tcpRuntimeCfg, ok := proto.Clone(scannedTCP).(*tcpconf.Server)
	if !ok {
		t.Fatalf("clone tcp config failed, got %T", scannedTCP)
	}
	if tcpRuntimeCfg.Listen == nil {
		tcpRuntimeCfg.Listen = &confv1.Server_Listen{}
	}
	tcpRuntimeCfg.Listen.Addr = "127.0.0.1:0"
	tcpRuntimeCfg.Registry = nil

	app, cleanup, err := wireApp(
		serverCfg,
		nil, // e2e 仅验证服务启动与 tcp 链路，不依赖 discovery
		nil, // 避免测试环境下外部注册中心依赖
		bc.GetData(),
		bc.GetApp(),
		bc.GetTrace(),
		bc.GetMetrics(),
		tcpRuntimeCfg,
		rt.Identity,
		rt.Logger,
	)
	if err != nil {
		t.Fatalf("wire app: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	runErr := make(chan error, 1)
	go func() {
		runErr <- app.Run()
	}()

	tcpEndpoint := waitForTCPEndpoint(t, app, 5*time.Second)
	resp := sendTCPCommand(t, tcpEndpoint, "PING", time.Second)
	if resp != "PONG" {
		t.Fatalf("tcp response=%q want=%q", resp, "PONG")
	}

	if err := app.Stop(); err != nil {
		t.Fatalf("stop app: %v", err)
	}
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("app run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wait app stop timeout")
	}
}

func waitForTCPEndpoint(t *testing.T, app *kratos.App, timeout time.Duration) *url.URL {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, ep := range app.Endpoint() {
			if strings.HasPrefix(ep, "tcp://") || strings.HasPrefix(ep, "tcps://") {
				u, err := url.Parse(ep)
				if err == nil && u.Host != "" {
					return u
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("tcp endpoint not found within %s, endpoints=%v", timeout, app.Endpoint())
	return nil
}

func sendTCPCommand(t *testing.T, ep *url.URL, cmd string, timeout time.Duration) string {
	t.Helper()

	conn, err := net.DialTimeout("tcp", ep.Host, timeout)
	if err != nil {
		t.Fatalf("dial tcp endpoint %q: %v", ep.String(), err)
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := io.WriteString(conn, cmd+"\n"); err != nil {
		t.Fatalf("write tcp command %q: %v", cmd, err)
	}
	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("read tcp response: %v", err)
	}
	return strings.TrimSpace(resp)
}
