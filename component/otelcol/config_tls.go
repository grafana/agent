package otelcol

import (
	"time"

	otelconfigtls "go.opentelemetry.io/collector/config/configtls"
)

// TLSServerArguments holds shared TLS settings for components which launch
// servers with TLS.
type TLSServerArguments struct {
	CAFile         string        `river:"ca_file,attr,optional"`
	CertFile       string        `river:"cert_file,attr,optional"`
	KeyFile        string        `river:"key_file,attr,optional"`
	MinVersion     string        `river:"min_version,attr,optional"`
	MaxVersion     string        `river:"max_version,attr,optional"`
	ReloadInterval time.Duration `river:"reload_interval,attr,optional"`
	ClientCAFile   string        `river:"client_ca_file,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *TLSServerArguments) Convert() *otelconfigtls.TLSServerSetting {
	if args == nil {
		return nil
	}

	return &otelconfigtls.TLSServerSetting{
		TLSSetting: otelconfigtls.TLSSetting{
			CAFile:         args.CAFile,
			CertFile:       args.CertFile,
			KeyFile:        args.KeyFile,
			MinVersion:     args.MinVersion,
			MaxVersion:     args.MaxVersion,
			ReloadInterval: args.ReloadInterval,
		},
		ClientCAFile: args.ClientCAFile,
	}
}

// TLSClientArguments holds shared TLS settings for components which launch
// TLS clients.
type TLSClientArguments struct {
	CAFile         string        `river:"ca_file,attr,optional"`
	CertFile       string        `river:"cert_file,attr,optional"`
	KeyFile        string        `river:"key_file,attr,optional"`
	MinVersion     string        `river:"min_version,attr,optional"`
	MaxVersion     string        `river:"max_version,attr,optional"`
	ReloadInterval time.Duration `river:"reload_interval,attr,optional"`

	Insecure           bool   `river:"insecure,attr,optional"`
	InsecureSkipVerify bool   `river:"insecure_skip_verify,attr,optional"`
	ServerName         string `river:"server_name,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *TLSClientArguments) Convert() *otelconfigtls.TLSClientSetting {
	if args == nil {
		return nil
	}

	return &otelconfigtls.TLSClientSetting{
		TLSSetting: otelconfigtls.TLSSetting{
			CAFile:         args.CAFile,
			CertFile:       args.CertFile,
			KeyFile:        args.KeyFile,
			MinVersion:     args.MinVersion,
			MaxVersion:     args.MaxVersion,
			ReloadInterval: args.ReloadInterval,
		},

		Insecure:           args.Insecure,
		InsecureSkipVerify: args.InsecureSkipVerify,
		ServerName:         args.ServerName,
	}
}
