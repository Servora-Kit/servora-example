package server

import (
	"github.com/Servora-Kit/servora/obs/audit"
	logger "github.com/Servora-Kit/servora/obs/logging"
	"github.com/Servora-Kit/servora/platform/bootstrap"
)

// ProvideAuditEmitter wires a stdout/log-based Emitter so audit events surface
// in the service log stream. Zero external deps — fits the demo/audit branch.
func ProvideAuditEmitter(l logger.Logger) audit.Emitter {
	return audit.NewLogEmitter(l)
}

// ProvideAuditRecorder builds the per-service Recorder. serviceName is taken
// from bootstrap identity so events carry the same name kratos registers.
func ProvideAuditRecorder(emitter audit.Emitter, identity bootstrap.SvcIdentity) *audit.Recorder {
	return audit.NewRecorder(emitter, identity.Name)
}
