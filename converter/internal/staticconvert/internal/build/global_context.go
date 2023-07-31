package build

import (
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/internal/common"
)

type GlobalContext struct {
	LabelPrefix        string
	RemoteWriteExports *remotewrite.Exports
}

func (g *GlobalContext) InitializeRemoteWriteExports() {
	if g.RemoteWriteExports == nil {
		g.RemoteWriteExports = &remotewrite.Exports{
			Receiver: common.ConvertAppendable{Expr: "prometheus.remote_write." + g.LabelPrefix + ".receiver"},
		}
	}
}
