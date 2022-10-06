package otelcol

import (
	"github.com/alecthomas/units"
	otelconfighttp "go.opentelemetry.io/collector/config/confighttp"
)

// HTTPServerArguments holds shared settings for components which launch HTTP
// servers.
type HTTPServerArguments struct {
	Endpoint string `river:"endpoint,attr,optional"`

	TLS *TLSServerArguments `river:"tls,block,optional"`

	CORS *CORSArguments `river:"cors,block,optional"`

	// TODO(rfratto): auth
	//
	// Figuring out how to do authentication isn't very straightforward here. The
	// auth section links to an authenticator extension.
	//
	// We will need to generally figure out how we want to provide common
	// authentication extensions to all of our components.

	MaxRequestBodySize units.Base2Bytes `river:"max_request_body_size,attr,optional"`
	IncludeMetadata    bool             `river:"include_metadata,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *HTTPServerArguments) Convert() *otelconfighttp.HTTPServerSettings {
	if args == nil {
		return nil
	}

	return &otelconfighttp.HTTPServerSettings{
		Endpoint:           args.Endpoint,
		TLSSetting:         args.TLS.Convert(),
		CORS:               args.CORS.Convert(),
		MaxRequestBodySize: int64(args.MaxRequestBodySize),
		IncludeMetadata:    args.IncludeMetadata,
	}
}

// CORSArguments holds shared CORS settings for components which launch HTTP
// servers.
type CORSArguments struct {
	AllowedOrigins []string `river:"allowed_origins,attr,optional"`
	AllowedHeaders []string `river:"allowed_headers,attr,optional"`

	MaxAge int `river:"max_age,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *CORSArguments) Convert() *otelconfighttp.CORSSettings {
	if args == nil {
		return nil
	}

	return &otelconfighttp.CORSSettings{
		AllowedOrigins: args.AllowedOrigins,
		AllowedHeaders: args.AllowedHeaders,

		MaxAge: args.MaxAge,
	}
}
