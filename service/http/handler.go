//go:build !windows

package http

import (
	"crypto/tls"
	"github.com/grafana/agent/pkg/server"
)

// tlsConfig generates a tls.Config from args.
func (args *TLSArguments) winTlsConfig(_ *server.WinCertStoreHandler) (*tls.Config, error) {
	panic("Windows Certificate filter is only available on Windows platforms.")
}
