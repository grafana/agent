package metrics

import (
	"github.com/grafana/agent/component/common/metrics/instance"
)

type Config struct {
	global instance.globalConfig `river:"global,block,optional"`
}
