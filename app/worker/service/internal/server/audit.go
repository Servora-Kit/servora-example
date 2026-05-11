package server

import (
	"github.com/Servora-Kit/servora/obs/audit"
	"github.com/Servora-Kit/servora/obs/audit/stdout"
)

// ProvideAuditor wires a stdout-based Auditor so audit events surface as
// JSON on stdout. Zero external deps — fits the demo branch.
func ProvideAuditor() audit.Auditor {
	return stdout.NewAuditor()
}
