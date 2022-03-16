package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
)

// TLSConfig holds dynamic configuration options for TLS.
type TLSConfig struct {
	TLSCertPath              string      `yaml:"cert_file"`
	TLSKeyPath               string      `yaml:"key_file"`
	ClientAuth               string      `yaml:"client_auth_type"`
	ClientCAs                string      `yaml:"client_ca_file"`
	CipherSuites             []TLSCipher `yaml:"cipher_suites"`
	CurvePreferences         []TLSCurve  `yaml:"curve_preferences"`
	MinVersion               TLSVersion  `yaml:"min_version"`
	MaxVersion               TLSVersion  `yaml:"max_version"`
	PreferServerCipherSuites bool        `yaml:"prefer_server_cipher_suites"`
}

// TLSCipher holds the ID of a tls.CipherSuite.
type TLSCipher uint16

// UnmarshalYAML unmarshals the name of a cipher suite to its ID.
func (c *TLSCipher) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	for _, cs := range tls.CipherSuites() {
		if cs.Name == s {
			*c = (TLSCipher)(cs.ID)
			return nil
		}
	}
	return errors.New("unknown cipher: " + s)
}

// MarshalYAML marshals the name of the cipher suite.
func (c TLSCipher) MarshalYAML() (interface{}, error) {
	return tls.CipherSuiteName((uint16)(c)), nil
}

// TLSCurve holds the ID of a TLS elliptic curve.
type TLSCurve tls.CurveID

var curves = map[string]TLSCurve{
	"CurveP256": (TLSCurve)(tls.CurveP256),
	"CurveP384": (TLSCurve)(tls.CurveP384),
	"CurveP521": (TLSCurve)(tls.CurveP521),
	"X25519":    (TLSCurve)(tls.X25519),
}

// UnmarshalYAML unmarshals the name of a TLS elliptic curve into its ID.
func (c *TLSCurve) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	if curveid, ok := curves[s]; ok {
		*c = curveid
		return nil
	}
	return errors.New("unknown curve: " + s)
}

// MarshalYAML marshals the ID of a TLS elliptic curve into its name.
func (c *TLSCurve) MarshalYAML() (interface{}, error) {
	for s, curveid := range curves {
		if *c == curveid {
			return s, nil
		}
	}
	return fmt.Sprintf("%v", c), nil
}

// TLSVersion holds a TLS version ID.
type TLSVersion uint16

var tlsVersions = map[string]TLSVersion{
	"TLS13": (TLSVersion)(tls.VersionTLS13),
	"TLS12": (TLSVersion)(tls.VersionTLS12),
	"TLS11": (TLSVersion)(tls.VersionTLS11),
	"TLS10": (TLSVersion)(tls.VersionTLS10),
}

// UnmarshalYAML unmarshals the name of a TLS version into its ID.
func (tv *TLSVersion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	if v, ok := tlsVersions[s]; ok {
		*tv = v
		return nil
	}
	return errors.New("unknown TLS version: " + s)
}

// MarshalYAML marshals the ID of a TLS version into its name.
func (tv *TLSVersion) MarshalYAML() (interface{}, error) {
	for s, v := range tlsVersions {
		if *tv == v {
			return s, nil
		}
	}
	return fmt.Sprintf("%v", tv), nil
}

// tlsListener is a net.Listener for establishing TLS connections. tlsListener
// supports dynamically updating the TLS settings used to establish
// connections.
type tlsListener struct {
	mut       sync.RWMutex
	cfg       TLSConfig
	tlsConfig *tls.Config

	innerListener net.Listener
}

// newTLSListener creates and configures a new tlsListener.
func newTLSListener(inner net.Listener, c TLSConfig) (*tlsListener, error) {
	tl := &tlsListener{
		innerListener: inner,
	}
	return tl, tl.ApplyConfig(c)
}

// Accept implements net.Listener and returns the next connection. Connections
func (l *tlsListener) Accept() (net.Conn, error) {
	nc, err := l.innerListener.Accept()
	if err != nil {
		return nc, err
	}

	l.mut.RLock()
	defer l.mut.RUnlock()
	return tls.Server(nc, l.tlsConfig), nil
}

// Close implements net.Listener and closes the tlsListener, preventing any new
// connections from being formed. Existing connections will be kept alive.
func (l *tlsListener) Close() error {
	return l.innerListener.Close()
}

// Addr implements net.Listener and returns the listener's network address.
func (l *tlsListener) Addr() net.Addr {
	return l.innerListener.Addr()
}

// ApplyConfig updates the tlsListener with new settings for creating TLS
// connections.
//
// Existing TLS connections will be kept alive after updating the TLS settings.
// New connections cannot be established while ApplyConfig is running.
func (l *tlsListener) ApplyConfig(c TLSConfig) error {
	l.mut.Lock()
	defer l.mut.Unlock()

	// Convert our TLSConfig into a new *tls.Config.
	//
	// While *tls.Config supports callbacks and doesn't need to be fully
	// replaced, some of our dynamic settings from TLSConfig can't be dynamically
	// updated (e.g., ciphers, min/max version, etc.).
	//
	// To make life easier on ourselves we just replace the whole thing with a new TLS listener.

	// Make sure that the certificates exist
	if c.TLSCertPath == "" {
		return fmt.Errorf("missing certificate file")
	}
	if c.TLSKeyPath == "" {
		return fmt.Errorf("missing key file")
	}
	_, err := tls.LoadX509KeyPair(c.TLSCertPath, c.TLSKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load key pair: %w", err)
	}

	newConfig := &tls.Config{
		MinVersion:               (uint16)(c.MinVersion),
		MaxVersion:               (uint16)(c.MaxVersion),
		PreferServerCipherSuites: c.PreferServerCipherSuites,

		GetCertificate: l.getCertificate,
	}

	var cf []uint16
	for _, c := range c.CipherSuites {
		cf = append(cf, (uint16)(c))
	}
	if len(cf) > 0 {
		newConfig.CipherSuites = cf
	}

	var cp []tls.CurveID
	for _, c := range c.CurvePreferences {
		cp = append(cp, (tls.CurveID)(c))
	}
	if len(cp) > 0 {
		newConfig.CurvePreferences = cp
	}

	if c.ClientCAs != "" {
		clientCAPool := x509.NewCertPool()
		clientCAFile, err := ioutil.ReadFile(c.ClientCAs)
		if err != nil {
			return err
		}
		clientCAPool.AppendCertsFromPEM(clientCAFile)
		newConfig.ClientCAs = clientCAPool
	}

	switch c.ClientAuth {
	case "RequestClientCert":
		newConfig.ClientAuth = tls.RequestClientCert
	case "RequireAnyClientCert", "RequireClientCert": // Preserved for backwards compatibility.
		newConfig.ClientAuth = tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		newConfig.ClientAuth = tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		newConfig.ClientAuth = tls.RequireAndVerifyClientCert
	case "", "NoClientCert":
		newConfig.ClientAuth = tls.NoClientCert
	default:
		return fmt.Errorf("Invalid ClientAuth %q", c.ClientAuth)
	}
	if c.ClientCAs != "" && newConfig.ClientAuth == tls.NoClientCert {
		return fmt.Errorf("Client CAs have been configured without a ClientAuth policy")
	}

	l.tlsConfig = newConfig
	l.cfg = c
	return nil
}

func (l *tlsListener) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	l.mut.RLock()
	defer l.mut.RUnlock()

	cert, err := tls.LoadX509KeyPair(l.cfg.TLSCertPath, l.cfg.TLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}
	return &cert, nil
}
