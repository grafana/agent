package windowsevent

import (
	"time"

	"github.com/grafana/agent/component/common/loki"
)

// Arguments holds values which are used to configure the loki.source.windowsevent
// component.
type Arguments struct {
	Locale               int                 `river:"locale,attr,optional"`
	EventLogName         string              `river:"eventlog_name,attr,optional"`
	XPathQuery           string              `river:"xpath_query,attr,optional"`
	BookmarkPath         string              `river:"bookmark_path,attr,optional"`
	PollInterval         time.Duration       `river:"poll_interval,attr,optional"`
	ExcludeEventData     bool                `river:"exclude_event_data,attr,optional"`
	ExcludeUserdata      bool                `river:"exclude_user_data,attr,optional"`
	UseIncomingTimestamp bool                `river:"use_incoming_timestamp,attr,optional"`
	ForwardTo            []loki.LogsReceiver `river:"forward_to,attr"`
}

func defaultArgs() Arguments {
	return Arguments{
		Locale:               0,
		EventLogName:         "",
		XPathQuery:           "*",
		BookmarkPath:         "",
		PollInterval:         3 * time.Second,
		ExcludeEventData:     false,
		ExcludeUserdata:      false,
		UseIncomingTimestamp: false,
	}
}

// SetToDefault implements river.Defaulter.
func (r *Arguments) SetToDefault() {
	*r = defaultArgs()
}
