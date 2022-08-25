//go:build !windows
// +build !windows

package log

import (
	"github.com/grafana/agent/pkg/server"
)

// NewLogger returns Windows Event Logger if running as a service under windows
// One non-windows platforms, this always returns a regular logger
func NewLogger(cfg *server.Config) *server.Logger {
	return server.NewLogger(cfg)
}
