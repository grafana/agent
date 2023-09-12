package http

import (
	"crypto/tls"
	"github.com/grafana/agent/pkg/server"
)

// tlsConfig generates a tls.Config from args.
func (args *TLSArguments) winTlsConfig(win *server.WinCertStoreHandler) (*tls.Config, error) {
	config := &tls.Config{
		MinVersion:            uint16(args.MinVersion),
		MaxVersion:            uint16(args.MaxVersion),
		ClientAuth:            tls.ClientAuthType(args.ClientAuth),
		VerifyPeerCertificate: win.VerifyPeer,
		GetCertificate:        win.CertificateHandler,
	}

	for _, c := range args.CipherSuites {
		config.CipherSuites = append(config.CipherSuites, uint16(c))
	}
	for _, c := range args.CurvePreferences {
		config.CurvePreferences = append(config.CurvePreferences, tls.CurveID(c))
	}
	return config, nil
}
