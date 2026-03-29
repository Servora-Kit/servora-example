package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Servora-Kit/servora-example/app/master/service/internal/service"
	"github.com/go-kratos/kratos/v2/log"
)

func newTCPConnectionHandler(svc *service.TCPCommandService, l log.Logger) func(context.Context, net.Conn) {
	return func(ctx context.Context, conn net.Conn) {
		if conn == nil {
			return
		}
		if conn.RemoteAddr() != nil {
			_ = l.Log(log.LevelDebug, "msg", "tcp connection accepted", "remote", conn.RemoteAddr().String())
		}
		defer func() { _ = conn.Close() }()

		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				_ = l.Log(log.LevelWarn, "msg", "tcp read failed", "err", err)
				return
			}

			cmdLine := strings.TrimSpace(line)
			if cmdLine == "" {
				continue
			}
			cmd, arg := parseTCPCommand(cmdLine)
			if svc == nil {
				if err := writeTCPLine(conn, "ERR tcp command service not configured"); err != nil {
					return
				}
				continue
			}
			resp, err := svc.Handle(ctx, cmd, arg)
			if err != nil {
				_ = l.Log(log.LevelWarn, "msg", "tcp command failed", "cmd", cmd, "err", err)
				if writeErr := writeTCPLine(conn, "ERR "+err.Error()); writeErr != nil {
					return
				}
				continue
			}
			if err := writeTCPLine(conn, resp); err != nil {
				_ = l.Log(log.LevelWarn, "msg", "tcp write failed", "err", err)
				return
			}
		}
	}
}

func parseTCPCommand(line string) (cmd string, arg string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", ""
	}
	cmd = strings.ToUpper(fields[0])
	if len(fields) > 1 {
		arg = strings.Join(fields[1:], " ")
	}
	return cmd, arg
}

func writeTCPLine(conn net.Conn, msg string) error {
	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := io.WriteString(conn, msg+"\n")
	return err
}
