package otelcol

import (
	"time"

	otelconfiggrpc "go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
)

// GRPCServerArguments holds shared gRPC settings for components which launch
// gRPC servers.
type GRPCServerArguments struct {
	Endpoint  string `river:"endpoint,attr,optional"`
	Transport string `river:"transport,attr,optional"`

	TLS *TLSServerArguments `river:"tls,block,optional"`

	MaxRecvMsgSizeMiB    uint64 `river:"max_recv_msg_size_mib,attr,optional"`
	MaxConcurrentStreams uint32 `river:"max_concurrent_streams,attr,optional"`
	ReadBufferSize       int    `river:"read_buffer_size,attr,optional"`
	WriteBufferSize      int    `river:"write_buffer_size,attr,optional"`

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

		MaxRecvMsgSizeMiB:    args.MaxRecvMsgSizeMiB,
		MaxConcurrentStreams: args.MaxConcurrentStreams,
		ReadBufferSize:       args.ReadBufferSize,
		WriteBufferSize:      args.WriteBufferSize,

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
