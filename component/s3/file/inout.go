package s3

import (
	"time"

	"github.com/grafana/agent/pkg/flow/rivertypes"
)

// Arguments implements the input for the s3 component
type Arguments struct {
	Path string `river:"path,attr"`
	// PollFrequency determines the frequency to check for changes
	// defaults to 5m
	PollFrequency time.Duration `river:"poll_frequency,attr,optional"`
	// IsSecret determines if the content should be displayed to the user
	IsSecret bool `river:"is_secret,attr,optional"`
	// Options allows you to override default settings
	Options AWSOptions `river:"options,block,optional"`
}

// AWSOptions implements specific AWS configuration options
type AWSOptions struct {
	AccessKey    string            `river:"key,attr,optional"`
	Secret       rivertypes.Secret `river:"secret,attr,optional"`
	Endpoint     string            `river:"endpoint,attr,optional"`
	DisableSSL   bool              `river:"disable_ssl,attr,optional"`
	UsePathStyle bool              `river:"use_path_style,attr,optional"`
	Region       string            `river:"region,attr,optional"`
}

// DefaultArguments sets the poll frequency
var DefaultArguments = Arguments{
	PollFrequency: 10 * time.Minute,
}

// UnmarshalRiver implements the unmarshaller
func (a *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*a = DefaultArguments
	type arguments Arguments
	return f((*arguments)(a))
}

// Exports implements the file content
type Exports struct {
	Content rivertypes.OptionalSecret `river:"content,attr"`
}
