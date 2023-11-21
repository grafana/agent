package mssql

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.mssql",
		Args:    Arguments{},
		Exports: exporter.Exports{},

		Build: exporter.New(createExporter, "mssql"),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

// DefaultArguments holds the default settings for the mssql exporter
var DefaultArguments = Arguments{
	MaxIdleConnections: 3,
	MaxOpenConnections: 3,
	Timeout:            10 * time.Second,
}

// Arguments controls the mssql exporter.
type Arguments struct {
	ConnectionString   rivertypes.Secret         `river:"connection_string,attr"`
	MaxIdleConnections int                       `river:"max_idle_connections,attr,optional"`
	MaxOpenConnections int                       `river:"max_open_connections,attr,optional"`
	Timeout            time.Duration             `river:"timeout,attr,optional"`
	QueryConfigFile    string                    `river:"query_config_file,attr,optional"`
	QueryConfig        rivertypes.OptionalSecret `river:"query_config,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	if a.MaxOpenConnections < 1 {
		return errors.New("max_open_connections must be at least 1")
	}

	if a.MaxIdleConnections < 1 {
		return errors.New("max_idle_connections must be at least 1")
	}

	if a.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	if a.QueryConfigFile != "" {
		_, err := os.Stat(a.QueryConfigFile)

		if err == nil {
			return nil
		}

		if errors.Is(err, os.ErrNotExist) {
			return errors.New("query_config_file must be a valid path of a YAML config file")
		} else {
			return fmt.Errorf("query_config_file file has issues: %w", err)
		}
	}

	return nil
}

func (a *Arguments) Convert() *mssql.Config {
	return &mssql.Config{
		ConnectionString:   config_util.Secret(a.ConnectionString),
		MaxIdleConnections: a.MaxIdleConnections,
		MaxOpenConnections: a.MaxOpenConnections,
		Timeout:            a.Timeout,
		QueryConfigFile:    a.QueryConfigFile,
		QueryConfig:        util.RawYAML(a.QueryConfig.Value),
	}
}
