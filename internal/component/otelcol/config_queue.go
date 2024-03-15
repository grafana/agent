package otelcol

import (
	"fmt"

	otelexporterhelper "go.opentelemetry.io/collector/exporter/exporterhelper"
)

// QueueArguments holds shared settings for components which can queue
// requests.
type QueueArguments struct {
	Enabled      bool `river:"enabled,attr,optional"`
	NumConsumers int  `river:"num_consumers,attr,optional"`
	QueueSize    int  `river:"queue_size,attr,optional"`

	// TODO(rfratto): queues can send to persistent storage through an extension.
}

// DefaultQueueArguments holds default settings for QueueArguments.
var DefaultQueueArguments = QueueArguments{
	Enabled:      true,
	NumConsumers: 10,

	// Copied from [upstream](https://github.com/open-telemetry/opentelemetry-collector/blob/241334609fc47927b4a8533dfca28e0f65dad9fe/exporter/exporterhelper/queue_sender.go#L50-L53)
	//
	// By default, batches are 8192 spans, for a total of up to 8 million spans in the queue
	// This can be estimated at 1-4 GB worth of maximum memory usage
	// This default is probably still too high, and may be adjusted further down in a future release
	QueueSize: 1000,
}

// SetToDefault implements river.Defaulter.
func (args *QueueArguments) SetToDefault() {
	*args = DefaultQueueArguments
}

// Convert converts args into the upstream type.
func (args *QueueArguments) Convert() *otelexporterhelper.QueueSettings {
	if args == nil {
		return nil
	}

	return &otelexporterhelper.QueueSettings{
		Enabled:      args.Enabled,
		NumConsumers: args.NumConsumers,
		QueueSize:    args.QueueSize,
	}
}

// Validate returns an error if args is invalid.
func (args *QueueArguments) Validate() error {
	if args == nil || !args.Enabled {
		return nil
	}

	if args.QueueSize <= 0 {
		return fmt.Errorf("queue_size must be greater than zero")
	}

	return nil
}
