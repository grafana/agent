package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/mysql"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/mysqld_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendMysqldExporter(config *mysqld_exporter.Config) discovery.Exports {
	args := toMysqldExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "mysql"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.mysql.%s.targets", compLabel))
}

func toMysqldExporter(config *mysqld_exporter.Config) *mysql.Arguments {
	return &mysql.Arguments{
		DataSourceName:    rivertypes.Secret(config.DataSourceName),
		EnableCollectors:  config.EnableCollectors,
		DisableCollectors: config.DisableCollectors,
		SetCollectors:     config.SetCollectors,
		LockWaitTimeout:   config.LockWaitTimeout,
		LogSlowFilter:     config.LogSlowFilter,
		InfoSchemaProcessList: mysql.InfoSchemaProcessList{
			MinTime:         config.InfoSchemaProcessListMinTime,
			ProcessesByUser: config.InfoSchemaProcessListProcessesByUser,
			ProcessesByHost: config.InfoSchemaProcessListProcessesByHost,
		},
		InfoSchemaTables: mysql.InfoSchemaTables{
			Databases: config.InfoSchemaTablesDatabases,
		},
		PerfSchemaEventsStatements: mysql.PerfSchemaEventsStatements{
			Limit:     config.PerfSchemaEventsStatementsLimit,
			TimeLimit: config.PerfSchemaEventsStatementsTimeLimit,
			TextLimit: config.PerfSchemaEventsStatementsTextLimit,
		},
		PerfSchemaFileInstances: mysql.PerfSchemaFileInstances{
			Filter:       config.PerfSchemaFileInstancesFilter,
			RemovePrefix: config.PerfSchemaFileInstancesRemovePrefix,
		},
		Heartbeat: mysql.Heartbeat{
			Database: config.HeartbeatDatabase,
			Table:    config.HeartbeatTable,
			UTC:      config.HeartbeatUTC,
		},
		MySQLUser: mysql.MySQLUser{
			Privileges: config.MySQLUserPrivileges,
		},
	}
}
