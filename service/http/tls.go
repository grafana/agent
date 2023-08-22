package http

import (
	"crypto/tls"
	"crypto/x509"
	"encoding"
	"fmt"
	"os"

	"github.com/grafana/river"
	"github.com/grafana/river/rivertypes"
)

// TLSArguments configures TLS settings for the HTTP service.
type TLSArguments struct {
	Cert                     string            `river:"cert_pem,attr,optional"`
	CertFile                 string            `river:"cert_file,attr,optional"`
	Key                      rivertypes.Secret `river:"key_pem,attr,optional"`
	KeyFile                  string            `river:"key_file,attr,optional"`
	ClientCA                 string            `river:"client_ca_pem,attr,optional"`
	ClientCAFile             string            `river:"client_ca_file,attr,optional"`
	ClientAuth               ClientAuth        `river:"client_auth_type,attr,optional"`
	CipherSuites             []TLSCipher       `river:"cipher_suites,attr,optional"`
	CurvePreferences         []TLSCurve        `river:"curve_preferences,attr,optional"`
	MinVersion               TLSVersion        `river:"min_version,attr,optional"`
	MaxVersion               TLSVersion        `river:"max_version,attr,optional"`
	PreferServerCipherSuites bool              `river:"prefer_server_cipher_suites,attr,optional"`

	// TODO(rfratto): windows certificate filter.
}

var _ river.Validator = (*TLSArguments)(nil)

// Validate returns whether args is valid. It checks that mutually exclusive
// fields are not both set, and that required fields are set.
func (args *TLSArguments) Validate() error {
	if len(args.ClientCA) > 0 && len(args.ClientCAFile) > 0 {
		return fmt.Errorf("cannot specify both client_ca_pem and client_ca_file")
	}
	if len(args.Cert) > 0 && len(args.CertFile) > 0 {
		return fmt.Errorf("cannot specify both cert_pem and cert_file")
	}
	if len(args.Key) > 0 && len(args.KeyFile) > 0 {
		return fmt.Errorf("cannot specify both key_pem and key_file")
	}

	var (
		usingCert     = len(args.Cert) > 0 || len(args.CertFile) > 0
		usingKey      = len(args.Key) > 0 || len(args.KeyFile) > 0
		usingClientCA = len(args.ClientCA) > 0 || len(args.ClientCAFile) > 0
	)
	if !usingCert {
		return fmt.Errorf("must specify either cert_pem or cert_file")
	}
	if !usingKey {
		return fmt.Errorf("must specify either key_pem or key_file")
	}
	if usingClientCA && args.ClientAuth == ClientAuth(tls.NoClientCert) {
		return fmt.Errorf("cannot specify client_ca_pem or client_ca_file when client_auth_type is NoClientCert")
	}

	return nil
}

// tlsConfig generates a tls.Config from args.
func (args *TLSArguments) tlsConfig() (*tls.Config, error) {
	config := &tls.Config{
		MinVersion:               uint16(args.MinVersion),
		MaxVersion:               uint16(args.MaxVersion),
		PreferServerCipherSuites: args.PreferServerCipherSuites,
		ClientAuth:               tls.ClientAuthType(args.ClientAuth),

		GetCertificate: func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return args.tlsCertificate()
		},
	}

	for _, c := range args.CipherSuites {
		config.CipherSuites = append(config.CipherSuites, uint16(c))
	}
	for _, c := range args.CurvePreferences {
		config.CurvePreferences = append(config.CurvePreferences, tls.CurveID(c))
	}

	caPool, err := args.clientCAPool()
	if err != nil {
		return nil, err
	}
	config.ClientCAs = caPool

	return config, nil
}

// tlsCertificate generates a TLS certificate from the arguments.
func (args *TLSArguments) tlsCertificate() (*tls.Certificate, error) {
	var (
		certPEM []byte
		keyPEM  []byte
	)

	if len(args.Cert) > 0 {
		certPEM = []byte(args.Cert)
	} else {
		var err error
		certPEM, err = os.ReadFile(args.CertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read cert_file: %w", err)
		}
	}

	if len(args.Key) > 0 {
		keyPEM = []byte(args.Key)
	} else {
		var err error
		keyPEM, err = os.ReadFile(args.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key_file: %w", err)
		}
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// clientCAPool generates a client CA pool from the arguments. If client CA
// isn't configured, clientCAPool returns nil.
func (args *TLSArguments) clientCAPool() (*x509.CertPool, error) {
	var caPEM []byte

	if len(args.ClientCA) > 0 {
		caPEM = []byte(args.ClientCA)
	} else if len(args.ClientCAFile) > 0 {
		var err error
		caPEM, err = os.ReadFile(args.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read client_ca_file: %w", err)
		}
	}

	if len(caPEM) == 0 {
		return nil, nil
	}

	clientCAPool := x509.NewCertPool()
	clientCAPool.AppendCertsFromPEM(caPEM)
	return clientCAPool, nil
}

// ClientAuth configures the type of TLS client authentication to use.
type ClientAuth tls.ClientAuthType

var (
	_ encoding.TextUnmarshaler = (*ClientAuth)(nil)
	_ encoding.TextMarshaler   = (ClientAuth)(0)
)

var clientAuths = map[string]ClientAuth{
	"NoClientCert":               (ClientAuth)(tls.NoClientCert),
	"RequestClientCert":          (ClientAuth)(tls.RequestClientCert),
	"RequireAnyClientCert":       (ClientAuth)(tls.RequireAnyClientCert),
	"VerifyClientCertIfGiven":    (ClientAuth)(tls.VerifyClientCertIfGiven),
	"RequireAndVerifyClientCert": (ClientAuth)(tls.RequireAndVerifyClientCert),
}

// UnmarshalText unmarshals the name of a client auth type to its ID.
func (c *ClientAuth) UnmarshalText(text []byte) error {
	str := string(text)

	auth, ok := clientAuths[str]
	if !ok {
		return fmt.Errorf("unknown client auth type %q", str)
	}

	*c = auth
	return nil
}

// MarshalText marshals the ID of a client auth type to its name.
func (c ClientAuth) MarshalText() ([]byte, error) {
	for name, auth := range clientAuths {
		if auth == c {
			return []byte(name), nil
		}
	}

	return nil, fmt.Errorf("unknown client auth type %d", c)
}

// TLSCipher holds the ID of a TLS cipher suite.
type TLSCipher uint16

var (
	_ encoding.TextUnmarshaler = (*TLSCipher)(nil)
	_ encoding.TextMarshaler   = (TLSCipher)(0)
)

// UnmarshalText unmarshals the name of a cipher suite to its ID.
func (c *TLSCipher) UnmarshalText(text []byte) error {
	str := string(text)

	for _, cs := range tls.CipherSuites() {
		if cs.Name == str {
			*c = TLSCipher(cs.ID)
			return nil
		}
	}

	return fmt.Errorf("unknown cipher %q", str)
}

// MarshalText marshals the ID of a cipher suite to its name.
func (c TLSCipher) MarshalText() ([]byte, error) {
	return []byte(tls.CipherSuiteName(uint16(c))), nil
}

// TLSCurve holds the ID of a [tls.CurveID].
type TLSCurve tls.CurveID

var (
	_ encoding.TextUnmarshaler = (*TLSCurve)(nil)
	_ encoding.TextMarshaler   = (TLSCurve)(0)
)

var curves = map[string]TLSCurve{
	"CurveP256": (TLSCurve)(tls.CurveP256),
	"CurveP384": (TLSCurve)(tls.CurveP384),
	"CurveP521": (TLSCurve)(tls.CurveP521),
	"X25519":    (TLSCurve)(tls.X25519),
}

// UnmarshalText unmarshals the name of a curve to its ID.
func (c *TLSCurve) UnmarshalText(text []byte) error {
	str := string(text)

	curve, ok := curves[str]
	if !ok {
		return fmt.Errorf("unknown curve %q", str)
	}

	*c = curve
	return nil
}

// MarshalText marshals the ID of a curve to its name.
func (c TLSCurve) MarshalText() ([]byte, error) {
	for name, curve := range curves {
		if curve == c {
			return []byte(name), nil
		}
	}

	return nil, fmt.Errorf("unknown curve %d", c)
}

// TLSVersion holds the ID of a TLS version.
type TLSVersion uint16

var (
	_ encoding.TextUnmarshaler = (*TLSVersion)(nil)
	_ encoding.TextMarshaler   = (TLSVersion)(0)
)

var tlsVersions = map[string]TLSVersion{
	"TLS13": (TLSVersion)(tls.VersionTLS13),
	"TLS12": (TLSVersion)(tls.VersionTLS12),
	"TLS11": (TLSVersion)(tls.VersionTLS11),
	"TLS10": (TLSVersion)(tls.VersionTLS10),
}

// UnmarshalText unmarshals the name of a TLS version to its ID.
func (v *TLSVersion) UnmarshalText(text []byte) error {
	str := string(text)

	version, ok := tlsVersions[str]
	if !ok {
		return fmt.Errorf("unknown TLS version %q", str)
	}

	*v = version
	return nil
}

// MarshalText marshals the ID of a TLS version to its name.
func (v TLSVersion) MarshalText() ([]byte, error) {
	for name, version := range tlsVersions {
		if version == v {
			return []byte(name), nil
		}
	}

	return nil, fmt.Errorf("unknown TLS version %d", v)
}
