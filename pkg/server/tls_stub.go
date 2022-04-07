//go:build !windows
// +build !windows

package server

import "fmt"

func (l *tlsListener) applyWindowsCertificateStore(_ TLSConfig) error {
	return fmt.Errorf("cannot use windows certificate store on non windows platforms")
}

type winCertStoreHandler struct {
}
