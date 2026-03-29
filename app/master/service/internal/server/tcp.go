package server

import (
	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	tcpsrv "github.com/Servora-Kit/servora-transport/server/tcp"
	tcpconf "github.com/Servora-Kit/servora-transport/server/tcp/gen/conf"
	logger "github.com/Servora-Kit/servora/obs/logging"
)

func NewTCPServer(c *tcpconf.Server, l logger.Logger, svc *service.TCPCommandService) *tcpsrv.Server {
	tcpLogger := logger.With(l, "tcp/server/master")
	return tcpsrv.NewServer(
		tcpsrv.WithConfig(c),
		tcpsrv.WithLogger(tcpLogger),
		tcpsrv.WithConnectionHandler(newTCPConnectionHandler(svc, tcpLogger)),
	)
}
