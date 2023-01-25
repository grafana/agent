package journal

import (
	"time"

	"github.com/grafana/agent/component/common/loki"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
)

// Arguments are the arguments for the component.
type Arguments struct {
	FormatAsJson   bool                   `river:"format_as_json,attr,optional"`
	MaxAge         time.Duration          `river:"max_age,attr,optional"`
	Path           string                 `river:"path,attr,optional"`
	RelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`
	Receivers      []loki.LogsReceiver    `river:"forward_to,attr"`
}

func defaultArgs() Arguments {
	return Arguments{
		FormatAsJson: false,
		MaxAge:       7 * time.Hour,
		Path:         "",
	}
}

// UnmarshalRiver implements river.Unmarshaler.
func (r *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*r = defaultArgs()

	type arguments Arguments
	if err := f((*arguments)(r)); err != nil {
		return err
	}

	return nil
}
