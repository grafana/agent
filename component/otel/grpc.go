package otel

import (
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
)

// GRPCClientSettings defines common settings for a gRPC client configuration.
type GRPCClientSettings struct {
	Endpoint    string                            `hcl:"endpoint,attr"`
	Compression configcompression.CompressionType `hcl:"compression,optional"`
	// TODO(rfratto): TLSSetting
	// TODO(rfratto): keepalive
	ReadBufferSize  int               `hcl:"read_buffer_size,optional"`
	WriteBufferSize int               `hcl:"write_buffer_size,optional"`
	WaitForReady    bool              `hcl:"wait_for_ready,optional"`
	Headers         map[string]string `hcl:"headers,optional"`
	BalancerName    string            `hcl:"balancer_name,optional"`
	// TODO(rfratto): Auth
}

// Convert converts s into otel's GRPCClientSettings type.
func (s *GRPCClientSettings) Convert() configgrpc.GRPCClientSettings {
	return configgrpc.GRPCClientSettings{
		Endpoint:    s.Endpoint,
		Compression: s.Compression,
		// TODO(rfratto): TLSSetting
		TLSSetting: configtls.TLSClientSetting{
			Insecure: true,
		},
		// TODO(rfratto): keepalive
		ReadBufferSize:  s.ReadBufferSize,
		WriteBufferSize: s.WriteBufferSize,
		WaitForReady:    s.WaitForReady,
		Headers:         s.Headers,
		BalancerName:    s.BalancerName,
		// TODO(rfratto): auth
	}
}
