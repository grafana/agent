package windowsevent

import (
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"time"
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

func convertConfig(arg Arguments) *scrapeconfig.WindowsEventsTargetConfig {
	return &scrapeconfig.WindowsEventsTargetConfig{
		Locale:               uint32(arg.Locale),
		EventlogName:         arg.EventLogName,
		Query:                arg.XPathQuery,
		UseIncomingTimestamp: arg.UseIncomingTimestamp,
		BookmarkPath:         arg.BookmarkPath,
		PollInterval:         arg.PollInterval,
		ExcludeEventData:     arg.ExcludeEventData,
		ExcludeEventMessage:  false,
		ExcludeUserData:      arg.ExcludeUserdata,
	}
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

// UnmarshalRiver implements river.Unmarshaler.
func (r *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*r = defaultArgs()

	type arguments Arguments
	if err := f((*arguments)(r)); err != nil {
		return err
	}

	return nil
}
