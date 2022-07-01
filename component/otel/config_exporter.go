package otel

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Defaults for Exporter settings. These types don't implement gohcl.Decoder
// and so defaults must be applied by the wrapping type.
var (
	DefaultExporterQueueSettings = ExporterQueueSettings{
		Enabled:      true,
		NumConsumers: 10,
		// For 5000 queue elements at 100 requests/sec gives about 50 sec of
		// survive of destination outage. This is a pretty decent value for
		// production. Users should calculate this from the perspective of how many
		// seconds to buffer in case of a backend outage, and then multiple that by
		// the number of requests per second.
		QueueSize: 5000,
	}

	DefaultExporterRetrySettings = ExporterRetrySettings{
		Enabled:        true,
		InitalInterval: 5 * time.Second,
		MaxInterval:    30 * time.Second,
		MaxElapsedTime: 5 * time.Minute,
	}
)

// ExporterQueueSettings holds common settings for a queue_config block.
type ExporterQueueSettings struct {
	Enabled      bool `hcl:"enabled,optional"`
	NumConsumers int  `hcl:"num_consumers,optional"`
	QueueSize    int  `hcl:"queue_size,optional"`
}

// Convert transforms s into the otel QueueSettings type.
func (s *ExporterQueueSettings) Convert() exporterhelper.QueueSettings {
	return exporterhelper.QueueSettings{
		Enabled:      s.Enabled,
		NumConsumers: s.NumConsumers,
		QueueSize:    s.QueueSize,
	}
}

// Validate retursn an error if s is invalid.
func (s *ExporterQueueSettings) Validate() error {
	if !s.Enabled {
		return nil
	}

	if s.QueueSize <= 0 {
		return fmt.Errorf("queue_size must be greater than 0, got %d", s.QueueSize)
	}

	return nil
}

// ExporterRetrySettings holds common settings for a retry_on_failure block.
type ExporterRetrySettings struct {
	Enabled        bool          `hcl:"enabled,optional"`
	InitalInterval time.Duration `hcl:"initial_interval,optional"`
	MaxInterval    time.Duration `hcl:"max_interval,optional"`
	MaxElapsedTime time.Duration `hcl:"max_elapsed_time,optional"`
}

// Convert transforms s into the otel RetrySettings type.
func (s *ExporterRetrySettings) Convert() exporterhelper.RetrySettings {
	return exporterhelper.RetrySettings{
		Enabled:         s.Enabled,
		InitialInterval: s.InitalInterval,
		MaxInterval:     s.MaxInterval,
		MaxElapsedTime:  s.MaxElapsedTime,
	}
}
