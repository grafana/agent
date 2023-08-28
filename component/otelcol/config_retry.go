package otelcol

import (
	"time"

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

// DefaultRetryArguments holds default settings for RetryArguments.
var DefaultRetryArguments = RetryArguments{
	Enabled:         true,
	InitialInterval: 5 * time.Second,
	MaxInterval:     30 * time.Second,
	MaxElapsedTime:  5 * time.Minute,
}

// SetToDefault implements river.Defaulter.
func (args *RetryArguments) SetToDefault() {
	*args = DefaultRetryArguments
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
