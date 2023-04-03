// Package jaeger provides an otelcol.receiver.jaeger component.
package jaeger

import (
	"fmt"
	"time"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	"github.com/grafana/agent/pkg/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.receiver.jaeger",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := jaegerreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.jaeger component.
type Arguments struct {
	Protocols      ProtocolsArguments       `river:"protocols,block"`
	RemoteSampling *RemoteSamplingArguments `river:"remote_sampling,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ river.Unmarshaler  = (*Arguments)(nil)
	_ receiver.Arguments = Arguments{}
)

// DefaultArguments provides default settings for Arguments. All protocols are
// configured with defaults and then set to nil in UnmarshalRiver if they were
// not defined in the source config.
var DefaultArguments = Arguments{
	Protocols: ProtocolsArguments{
		GRPC: &otelcol.GRPCServerArguments{
			Endpoint:  "0.0.0.0:14250",
			Transport: "tcp",
		},
		ThriftHTTP: &otelcol.HTTPServerArguments{
			Endpoint: "0.0.0.0:14268",
		},
		ThriftBinary: &ProtocolUDP{
			Endpoint:      "0.0.0.0:6832",
			QueueSize:     1_000,
			MaxPacketSize: 65 * units.KiB,
			Workers:       10,
		},
		ThriftCompact: &ProtocolUDP{
			Endpoint:      "0.0.0.0:6831",
			QueueSize:     1_000,
			MaxPacketSize: 65 * units.KiB,
			Workers:       10,
		},
	},
}

// UnmarshalRiver implements river.Unmarshaler.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments

	// Unmarshal into a temporary struct so we can detect which protocols were
	// actually enabled by the user.
	var temp arguments
	if err := f(&temp); err != nil {
		return err
	}

	// Remove protocols from args if they weren't provided by the user.
	if temp.Protocols.GRPC == nil {
		args.Protocols.GRPC = nil
	}
	if temp.Protocols.ThriftHTTP == nil {
		args.Protocols.ThriftHTTP = nil
	}
	if temp.Protocols.ThriftBinary == nil {
		args.Protocols.ThriftBinary = nil
	}
	if temp.Protocols.ThriftCompact == nil {
		args.Protocols.ThriftCompact = nil
	}

	// Finally, unmarshal into the real struct.
	if err := f((*arguments)(args)); err != nil {
		return err
	}
	return args.Validate()
}

// Validate returns an error if args is invalid.
func (args *Arguments) Validate() error {
	if args.Protocols.GRPC == nil &&
		args.Protocols.ThriftHTTP == nil &&
		args.Protocols.ThriftBinary == nil &&
		args.Protocols.ThriftCompact == nil {

		return fmt.Errorf("at least one protocol must be enabled")
	}

	return nil
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &jaegerreceiver.Config{
		Protocols: jaegerreceiver.Protocols{
			GRPC:          args.Protocols.GRPC.Convert(),
			ThriftHTTP:    args.Protocols.ThriftHTTP.Convert(),
			ThriftBinary:  args.Protocols.ThriftBinary.Convert(),
			ThriftCompact: args.Protocols.ThriftCompact.Convert(),
		},
		RemoteSampling: args.RemoteSampling.Convert(),
	}, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	if args.RemoteSampling == nil {
		return nil
	}
	return args.RemoteSampling.Client.Extensions()
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

// ProtocolsArguments configures protocols for otelcol.receiver.jaeger to
// listen on.
type ProtocolsArguments struct {
	GRPC          *otelcol.GRPCServerArguments `river:"grpc,block,optional"`
	ThriftHTTP    *otelcol.HTTPServerArguments `river:"thrift_http,block,optional"`
	ThriftBinary  *ProtocolUDP                 `river:"thrift_binary,block,optional"`
	ThriftCompact *ProtocolUDP                 `river:"thrift_compact,block,optional"`
}

// ProtocolUDP configures a UDP server.
type ProtocolUDP struct {
	Endpoint         string           `river:"endpoint,attr,optional"`
	QueueSize        int              `river:"queue_size,attr,optional"`
	MaxPacketSize    units.Base2Bytes `river:"max_packet_size,attr,optional"`
	Workers          int              `river:"workers,attr,optional"`
	SocketBufferSize units.Base2Bytes `river:"socket_buffer_size,attr,optional"`
}

// Convert converts proto into the upstream type.
func (proto *ProtocolUDP) Convert() *jaegerreceiver.ProtocolUDP {
	if proto == nil {
		return nil
	}

	return &jaegerreceiver.ProtocolUDP{
		Endpoint: proto.Endpoint,
		ServerConfigUDP: jaegerreceiver.ServerConfigUDP{
			QueueSize:        proto.QueueSize,
			MaxPacketSize:    int(proto.MaxPacketSize),
			Workers:          proto.Workers,
			SocketBufferSize: int(proto.SocketBufferSize),
		},
	}
}

// RemoteSamplingArguments configures remote sampling settings.
type RemoteSamplingArguments struct {
	// TODO(rfratto): can we work with upstream to provide a hook to provide a
	// custom strategy file and bypass the reload interval?
	//
	// That would let users connect a local.file to otelcol.receiver.jaeger for
	// the remote sampling.

	HostEndpoint               string                      `river:"host_endpoint,attr"`
	StrategyFile               string                      `river:"strategy_file,attr"`
	StrategyFileReloadInterval time.Duration               `river:"strategy_file_reload_interval,attr"`
	Client                     otelcol.GRPCClientArguments `river:"client,block"`
}

// Convert converts args into the upstream type.
func (args *RemoteSamplingArguments) Convert() *jaegerreceiver.RemoteSamplingConfig {
	if args == nil {
		return nil
	}

	return &jaegerreceiver.RemoteSamplingConfig{
		HostEndpoint:               args.HostEndpoint,
		StrategyFile:               args.StrategyFile,
		StrategyFileReloadInterval: args.StrategyFileReloadInterval,
		GRPCClientSettings:         *args.Client.Convert(),
	}
}
