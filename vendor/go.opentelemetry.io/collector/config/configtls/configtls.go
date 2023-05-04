// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configtls // import "go.opentelemetry.io/collector/config/configtls"

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// We should avoid that users unknowingly use a vulnerable TLS version.
// The defaults should be a safe configuration
const defaultMinTLSVersion = tls.VersionTLS12

// Uses the default MaxVersion from "crypto/tls" which is the maximum supported version
const defaultMaxTLSVersion = 0

// TLSSetting exposes the common client and server TLS configurations.
// Note: Since there isn't anything specific to a server connection. Components
// with server connections should use TLSSetting.
type TLSSetting struct {
	// Path to the CA cert. For a client this verifies the server certificate.
	// For a server this verifies client certificates. If empty uses system root CA.
	// (optional)
	CAFile string `mapstructure:"ca_file"`

	// In memory PEM encoded cert. (optional)
	CAPem []byte `mapstructure:"ca_pem"`

	// Path to the TLS cert to use for TLS required connections. (optional)
	CertFile string `mapstructure:"cert_file"`

	// In memory PEM encoded TLS cert to use for TLS required connections. (optional)
	CertPem []byte `mapstructure:"cert_pem"`

	// Path to the TLS key to use for TLS required connections. (optional)
	KeyFile string `mapstructure:"key_file"`

	// In memory PEM encoded TLS key to use for TLS required connections. (optional)
	KeyPem []byte `mapstructure:"key_pem"`

	// MinVersion sets the minimum TLS version that is acceptable.
	// If not set, TLS 1.2 will be used. (optional)
	MinVersion string `mapstructure:"min_version"`

	// MaxVersion sets the maximum TLS version that is acceptable.
	// If not set, refer to crypto/tls for defaults. (optional)
	MaxVersion string `mapstructure:"max_version"`

	// ReloadInterval specifies the duration after which the certificate will be reloaded
	// If not set, it will never be reloaded (optional)
	ReloadInterval time.Duration `mapstructure:"reload_interval"`
}

// TLSClientSetting contains TLS configurations that are specific to client
// connections in addition to the common configurations. This should be used by
// components configuring TLS client connections.
type TLSClientSetting struct {
	// squash ensures fields are correctly decoded in embedded struct.
	TLSSetting `mapstructure:",squash"`

	// These are config options specific to client connections.

	// In gRPC when set to true, this is used to disable the client transport security.
	// See https://godoc.org/google.golang.org/grpc#WithInsecure.
	// In HTTP, this disables verifying the server's certificate chain and host name
	// (InsecureSkipVerify in the tls Config). Please refer to
	// https://godoc.org/crypto/tls#Config for more information.
	// (optional, default false)
	Insecure bool `mapstructure:"insecure"`
	// InsecureSkipVerify will enable TLS but not verify the certificate.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
	// ServerName requested by client for virtual hosting.
	// This sets the ServerName in the TLSConfig. Please refer to
	// https://godoc.org/crypto/tls#Config for more information. (optional)
	ServerName string `mapstructure:"server_name_override"`
}

// TLSServerSetting contains TLS configurations that are specific to server
// connections in addition to the common configurations. This should be used by
// components configuring TLS server connections.
type TLSServerSetting struct {
	// squash ensures fields are correctly decoded in embedded struct.
	TLSSetting `mapstructure:",squash"`

	// These are config options specific to server connections.

	// Path to the TLS cert to use by the server to verify a client certificate. (optional)
	// This sets the ClientCAs and ClientAuth to RequireAndVerifyClientCert in the TLSConfig. Please refer to
	// https://godoc.org/crypto/tls#Config for more information. (optional)
	ClientCAFile string `mapstructure:"client_ca_file"`
}

// certReloader is a wrapper object for certificate reloading
// Its GetCertificate method will either return the current certificate or reload from disk
// if the last reload happened more than ReloadInterval ago
type certReloader struct {
	nextReload time.Time
	cert       *tls.Certificate
	lock       sync.RWMutex
	tls        TLSSetting
}

func (c TLSSetting) newCertReloader() (*certReloader, error) {
	cert, err := c.loadCertificate()
	if err != nil {
		return nil, err
	}
	return &certReloader{
		tls:        c,
		nextReload: time.Now().Add(c.ReloadInterval),
		cert:       &cert,
	}, nil
}

func (r *certReloader) GetCertificate() (*tls.Certificate, error) {
	now := time.Now()
	// Read locking here before we do the time comparison
	// If a reload is in progress this will block and we will skip reloading in the current
	// call once we can continue
	r.lock.RLock()
	if r.tls.ReloadInterval != 0 && r.nextReload.Before(now) && (r.tls.hasCertFile() || r.tls.hasKeyFile()) {
		// Need to release the read lock, otherwise we deadlock
		r.lock.RUnlock()
		r.lock.Lock()
		defer r.lock.Unlock()
		cert, err := r.tls.loadCertificate()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS cert and key: %w", err)
		}
		r.cert = &cert
		r.nextReload = now.Add(r.tls.ReloadInterval)
		return r.cert, nil
	}
	defer r.lock.RUnlock()
	return r.cert, nil
}

// loadTLSConfig loads TLS certificates and returns a tls.Config.
// This will set the RootCAs and Certificates of a tls.Config.
func (c TLSSetting) loadTLSConfig() (*tls.Config, error) {
	certPool, err := c.loadCACertPool()
	if err != nil {
		return nil, err
	}

	var getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	var getClientCertificate func(*tls.CertificateRequestInfo) (*tls.Certificate, error)

	if c.hasCert() || c.hasKey() {
		var certReloader *certReloader
		certReloader, err = c.newCertReloader()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS cert and key: %w", err)
		}

		getCertificate = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) { return certReloader.GetCertificate() }
		getClientCertificate = func(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) { return certReloader.GetCertificate() }
	}

	minTLS, err := convertVersion(c.MinVersion, defaultMinTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS min_version: %w", err)
	}
	maxTLS, err := convertVersion(c.MaxVersion, defaultMaxTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS max_version: %w", err)
	}

	return &tls.Config{
		RootCAs:              certPool,
		GetCertificate:       getCertificate,
		GetClientCertificate: getClientCertificate,
		MinVersion:           minTLS,
		MaxVersion:           maxTLS,
	}, nil
}

func (c TLSSetting) loadCACertPool() (*x509.CertPool, error) {
	// There is no need to load the System Certs for RootCAs because
	// if the value is nil, it will default to checking against th System Certs.
	var err error
	var certPool *x509.CertPool

	switch {
	case c.hasCAFile() && c.hasCAPem():
		return nil, fmt.Errorf("failed to load CA CertPool: CA File and PEM cannot both be provided")
	case c.hasCAFile():
		// Set up user specified truststore from file
		certPool, err = c.loadCertFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA CertPool File: %w", err)
		}
	case c.hasCAPem():
		// Set up user specified truststore from PEM
		certPool, err = c.loadCertPem(c.CAPem)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA CertPool PEM: %w", err)
		}
	}

	return certPool, nil
}

func (c TLSSetting) loadCertFile(certPath string) (*x509.CertPool, error) {
	certPem, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load cert %s: %w", certPath, err)
	}

	return c.loadCertPem(certPem)
}

func (c TLSSetting) loadCertPem(certPem []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPem) {
		return nil, fmt.Errorf("failed to parse cert")
	}
	return certPool, nil
}

func (c TLSSetting) loadCertificate() (tls.Certificate, error) {
	switch {
	case c.hasCert() != c.hasKey():
		return tls.Certificate{}, fmt.Errorf("for auth via TLS, either both certificate and key must be supplied, or neither")
	case !c.hasCert() && !c.hasKey():
		return tls.Certificate{}, nil
	case c.hasCertFile() && c.hasCertPem():
		return tls.Certificate{}, fmt.Errorf("for auth via TLS, certificate file and PEM cannot both be provided")
	case c.hasKeyFile() && c.hasKeyPem():
		return tls.Certificate{}, fmt.Errorf("for auth via TLS, key file and PEM cannot both be provided")
	}

	var certPem, keyPem []byte
	var err error
	if c.hasCertFile() {
		certPem, err = os.ReadFile(c.CertFile)
		if err != nil {
			return tls.Certificate{}, err
		}
	} else {
		certPem = c.CertPem
	}

	if c.hasKeyFile() {
		keyPem, err = os.ReadFile(c.KeyFile)
		if err != nil {
			return tls.Certificate{}, err
		}
	} else {
		keyPem = c.KeyPem
	}

	certificate, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load TLS cert and key PEMs: %w", err)
	}

	return certificate, err
}

func (c TLSSetting) hasCA() bool   { return c.hasCAFile() || c.hasCAPem() }
func (c TLSSetting) hasCert() bool { return c.hasCertFile() || c.hasCertPem() }
func (c TLSSetting) hasKey() bool  { return c.hasKeyFile() || c.hasKeyPem() }

func (c TLSSetting) hasCAFile() bool { return c.CAFile != "" }
func (c TLSSetting) hasCAPem() bool  { return len(c.CAPem) != 0 }

func (c TLSSetting) hasCertFile() bool { return c.CertFile != "" }
func (c TLSSetting) hasCertPem() bool  { return len(c.CertPem) != 0 }

func (c TLSSetting) hasKeyFile() bool { return c.KeyFile != "" }
func (c TLSSetting) hasKeyPem() bool  { return len(c.KeyPem) != 0 }

// LoadTLSConfig loads the TLS configuration.
func (c TLSClientSetting) LoadTLSConfig() (*tls.Config, error) {
	if c.Insecure && !c.hasCA() {
		return nil, nil
	}

	tlsCfg, err := c.TLSSetting.loadTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}
	tlsCfg.ServerName = c.ServerName
	tlsCfg.InsecureSkipVerify = c.InsecureSkipVerify
	return tlsCfg, nil
}

// LoadTLSConfig loads the TLS configuration.
func (c TLSServerSetting) LoadTLSConfig() (*tls.Config, error) {
	tlsCfg, err := c.loadTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}
	if c.ClientCAFile != "" {
		certPool, err := c.loadCertFile(c.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: failed to load client CA CertPool: %w", err)
		}
		tlsCfg.ClientCAs = certPool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsCfg, nil
}

func convertVersion(v string, defaultVersion uint16) (uint16, error) {
	// Use a default that is explicitly defined
	if v == "" {
		return defaultVersion, nil
	}
	val, ok := tlsVersions[v]
	if !ok {
		return 0, fmt.Errorf("unsupported TLS version: %q", v)
	}
	return val, nil
}

var tlsVersions = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}
