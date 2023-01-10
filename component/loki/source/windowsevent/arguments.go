package windowsevent

import (
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"time"
)

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
