package otelcol

import (
	"time"

	"github.com/grafana/agent/pkg/river"
	otelexporterhelper "go.opentelemetry.io/collector/exporter/exporterhelper"
)

// RetryArguments holds shared settings for components which can retry
// requests.
type RetryArguments struct {
	Enabled         bool          `river:"enabled,attr,optional"`
	InitialInterval time.Duration `river:"initial_interval,attr,optional"`
	MaxInterval     time.Duration `river:"max_interval,attr,optional"`
	MaxElapsedTime  time.Duration `river:"max_elapsed_time,attr,optional"`
}

var _ river.Unmarshaler = (*RetryArguments)(nil)

// DefaultRetryArguments holds default settings for RetryArguments.
var DefaultRetryArguments = RetryArguments{
	Enabled:         true,
	InitialInterval: 5 * time.Second,
	MaxInterval:     30 * time.Second,
	MaxElapsedTime:  5 * time.Minute,
}

// UnmarshalRiver implements river.Unmarshaler.
func (args *RetryArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultRetryArguments
	type arguments RetryArguments
	return f((*arguments)(args))
}

// Convert converts args into the upstream type.
func (args *RetryArguments) Convert() *otelexporterhelper.RetrySettings {
	if args == nil {
		return nil
	}

	return &otelexporterhelper.RetrySettings{
		Enabled:         args.Enabled,
		InitialInterval: args.InitialInterval,
		MaxInterval:     args.MaxInterval,
		MaxElapsedTime:  args.MaxElapsedTime,
	}
}
