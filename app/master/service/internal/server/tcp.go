package server

import (
	"log/slog"

	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	tcpsrv "github.com/Servora-Kit/servora-transport/server/tcp"
	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
)

func NewTCPServer(c *tcpconf.Server, l *slog.Logger, svc *service.TCPCommandService) *tcpsrv.Server {
	tcpLogger := l.With("scope", "tcp/server/master")
	return tcpsrv.NewServer(
		tcpsrv.WithConfig(c),
		tcpsrv.WithLogger(tcpLogger),
		tcpsrv.WithConnectionHandler(newTCPConnectionHandler(svc, tcpLogger)),
	)
}
