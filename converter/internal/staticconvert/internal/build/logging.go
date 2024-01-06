package build

import (
	"reflect"

	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/pkg/server"
)

func (b *IntegrationsConfigBuilder) appendLogging(config *server.Config) {
	args := toLogging(config)
	if !reflect.DeepEqual(*args, logging.DefaultOptions) {
		b.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"logging"},
			"",
			args,
		))
	}
}

func toLogging(config *server.Config) *logging.Options {
	return &logging.Options{
		Level:  logging.Level(config.LogLevel.String()),
		Format: logging.Format(config.LogFormat),
	}
}
