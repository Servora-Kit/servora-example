package service

import (
	"context"
	"fmt"
	"strings"

	workerpb "github.com/Servora-Kit/servora-example/api/gen/go/servora/worker/service/v1"
)

// TCPCommandService holds transport-independent command semantics for TCP adapter.
type TCPCommandService struct {
	master *MasterService
}

func NewTCPCommandService(master *MasterService) *TCPCommandService {
	return &TCPCommandService{master: master}
}

func (s *TCPCommandService) Handle(ctx context.Context, cmd string, arg string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(cmd)) {
	case "PING":
		return "PONG", nil
	case "HELLO":
		if s.master == nil {
			return "", fmt.Errorf("master service not configured")
		}
		if arg == "" {
			arg = "tcp-client"
		}
		resp, err := s.master.Hello(ctx, &workerpb.HelloRequest{Greeting: arg})
		if err != nil {
			return "", err
		}
		return "OK " + resp.GetReply(), nil
	default:
		return "", fmt.Errorf("unsupported command")
	}
}
