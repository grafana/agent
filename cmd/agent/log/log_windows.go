//go:build windows
// +build windows

package log

import (
	"github.com/grafana/agent/cmd/agent/initiate"
	"github.com/grafana/agent/pkg/server"
)

// NewLogger returns Windows Event Logger if running as a service under windows
// One non-windows platforms, this always returns a regular logger
func NewLogger(cfg *server.Config) *server.Logger {
	if initiate.IsWindowsService() {
		return server.NewWindowsEventLogger(cfg)
	} else {
		return server.NewLogger(cfg)
	}
}
