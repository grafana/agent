package otelcol

import (
	"time"

	"github.com/alecthomas/units"
	otelconfiggrpc "go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
)

// GRPCServerArguments holds shared gRPC settings for components which launch
// gRPC servers.
type GRPCServerArguments struct {
	Endpoint  string `river:"endpoint,attr,optional"`
	Transport string `river:"transport,attr,optional"`

	TLS *TLSServerArguments `river:"tls,block,optional"`

	MaxRecvMsgSize       units.Base2Bytes `river:"max_recv_msg_size,attr,optional"`
	MaxConcurrentStreams uint32           `river:"max_concurrent_streams,attr,optional"`
	ReadBufferSize       units.Base2Bytes `river:"read_buffer_size,attr,optional"`
	WriteBufferSize      units.Base2Bytes `river:"write_buffer_size,attr,optional"`

	Keepalive *KeepaliveServerArguments `river:"keepalive,block,optional"`

	// TODO(rfratto): auth
	//
	// Figuring out how to do authentication isn't very straightforward here. The
	// auth section links to an authenticator extension.
	//
	// We will need to generally figure out how we want to provide common
	// authentication extensions to all of our components.

	IncludeMetadata bool `river:"include_metadata,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *GRPCServerArguments) Convert() *otelconfiggrpc.GRPCServerSettings {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  args.Endpoint,
			Transport: args.Transport,
		},

		TLSSetting: args.TLS.Convert(),

		MaxRecvMsgSizeMiB:    uint64(args.MaxRecvMsgSize / units.Mebibyte),
		MaxConcurrentStreams: args.MaxConcurrentStreams,
		ReadBufferSize:       int(args.ReadBufferSize),
		WriteBufferSize:      int(args.WriteBufferSize),

		Keepalive: args.Keepalive.Convert(),

		IncludeMetadata: args.IncludeMetadata,
	}
}

// KeepaliveServerArguments holds shared keepalive settings for components
// which launch servers.
type KeepaliveServerArguments struct {
	ServerParameters  *KeepaliveServerParamaters  `river:"server_parameters,block,optional"`
	EnforcementPolicy *KeepaliveEnforcementPolicy `river:"enforcement_policy,block,optional"`
}

// Convert converts args into the upstream type.
func (args *KeepaliveServerArguments) Convert() *otelconfiggrpc.KeepaliveServerConfig {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.KeepaliveServerConfig{
		ServerParameters:  args.ServerParameters.Convert(),
		EnforcementPolicy: args.EnforcementPolicy.Convert(),
	}
}

// KeepaliveServerParamaters holds shared keepalive settings for components
// which launch servers.
type KeepaliveServerParamaters struct {
	MaxConnectionIdle     time.Duration `river:"max_connection_idle,attr,optional"`
	MaxConnectionAge      time.Duration `river:"max_connection_age,attr,optional"`
	MaxConnectionAgeGrace time.Duration `river:"max_connection_age_grace,attr,optional"`
	Time                  time.Duration `river:"time,attr,optional"`
	Timeout               time.Duration `river:"timeout,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *KeepaliveServerParamaters) Convert() *otelconfiggrpc.KeepaliveServerParameters {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.KeepaliveServerParameters{
		MaxConnectionIdle:     args.MaxConnectionIdle,
		MaxConnectionAge:      args.MaxConnectionAge,
		MaxConnectionAgeGrace: args.MaxConnectionAgeGrace,
		Time:                  args.Time,
		Timeout:               args.Timeout,
	}
}

// KeepaliveEnforcementPolicy holds shared keepalive settings for components
// which launch servers.
type KeepaliveEnforcementPolicy struct {
	MinTime             time.Duration `river:"min_time,attr,optional"`
	PermitWithoutStream bool          `river:"permit_without_stream,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *KeepaliveEnforcementPolicy) Convert() *otelconfiggrpc.KeepaliveEnforcementPolicy {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.KeepaliveEnforcementPolicy{
		MinTime:             args.MinTime,
		PermitWithoutStream: args.PermitWithoutStream,
	}
}

// GRPCClientArguments holds shared gRPC settings for components which launch
// gRPC clients.
type GRPCClientArguments struct {
	Endpoint string `river:"endpoint,attr"`

	Compression CompressionType `river:"compression,attr,optional"`

	TLS       TLSClientArguments        `river:"tls,block,optional"`
	Keepalive *KeepaliveClientArguments `river:"keepalive,block,optional"`

	ReadBufferSize  units.Base2Bytes  `river:"read_buffer_size,attr,optional"`
	WriteBufferSize units.Base2Bytes  `river:"write_buffer_size,attr,optional"`
	WaitForReady    bool              `river:"wait_for_ready,attr,optional"`
	Headers         map[string]string `river:"headers,attr,optional"`
	BalancerName    string            `river:"balancer_name,attr,optional"`

	// TODO(rfratto): auth
	//
	// Figuring out how to do authentication isn't very straightforward here. The
	// auth section links to an authenticator extension.
	//
	// We will need to generally figure out how we want to provide common
	// authentication extensions to all of our components.
}

// Convert converts args into the upstream type.
func (args *GRPCClientArguments) Convert() *otelconfiggrpc.GRPCClientSettings {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.GRPCClientSettings{
		Endpoint: args.Endpoint,

		Compression: args.Compression.Convert(),

		TLSSetting: *args.TLS.Convert(),
		Keepalive:  args.Keepalive.Convert(),

		ReadBufferSize:  int(args.ReadBufferSize),
		WriteBufferSize: int(args.WriteBufferSize),
		WaitForReady:    args.WaitForReady,
		Headers:         args.Headers,
		BalancerName:    args.BalancerName,
	}
}

// KeepaliveClientArguments holds shared keepalive settings for components
// which launch clients.
type KeepaliveClientArguments struct {
	Time                time.Duration `river:"time,attr,optional"`
	Timeout             time.Duration `river:"timeout,attr,optional"`
	PermitWithoutStream bool          `river:"permit_without_stream,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *KeepaliveClientArguments) Convert() *otelconfiggrpc.KeepaliveClientConfig {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.KeepaliveClientConfig{
		Time:                args.Time,
		Timeout:             args.Timeout,
		PermitWithoutStream: args.PermitWithoutStream,
	}
}
