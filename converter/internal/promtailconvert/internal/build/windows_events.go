package build

import (
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/loki/source/windowsevent"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
)

func (s *ScrapeConfigBuilder) AppendWindowsEventsConfig() {
	if s.cfg.WindowsConfig == nil {
		return
	}
	winCfg := s.cfg.WindowsConfig
	if len(winCfg.Labels) != 0 {
		// TODO: Add support for labels - see https://github.com/grafana/agent/issues/4634 for more details
		s.diags.Add(diag.SeverityLevelError, "windows_events.labels are currently not supported")
	}
	args := windowsevent.Arguments{
		Locale:               int(winCfg.Locale),
		EventLogName:         winCfg.EventlogName,
		XPathQuery:           winCfg.Query,
		BookmarkPath:         winCfg.BookmarkPath,
		PollInterval:         winCfg.PollInterval,
		ExcludeEventData:     winCfg.ExcludeEventData,
		ExcludeUserdata:      winCfg.ExcludeUserData,
		UseIncomingTimestamp: winCfg.UseIncomingTimestamp,
		ForwardTo:            make([]loki.LogsReceiver, 0),
	}

	override := func(val interface{}) interface{} {
		switch val.(type) {
		case []loki.LogsReceiver:
			return common.CustomTokenizer{Expr: s.getOrNewLokiRelabel()}
		default:
			return val
		}
	}
	compLabel := common.LabelForParts(s.globalCtx.LabelPrefix, s.cfg.JobName)
	s.f.Body().AppendBlock(common.NewBlockWithOverrideFn(
		[]string{"loki", "source", "windowsevent"},
		compLabel,
		args,
		override,
	))
}
