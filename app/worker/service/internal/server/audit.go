package server

import (
	"log/slog"

	"github.com/Servora-Kit/servora/obs/audit"
	auditlog "github.com/Servora-Kit/servora/obs/audit/log"
)

func ProvideAuditor(l *slog.Logger) audit.Auditor {
	return auditlog.NewAuditor(l.With("scope", "audit"))
}
