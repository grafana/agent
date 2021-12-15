package mysqld_exporter //nolint:golint

import (
	"testing"

	"github.com/grafana/agent/pkg/config"
)

func TestConfig_SecretMysqlD(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  mysqld_exporter:
    enabled: true
    data_source_name: root:secret_password@myserver:3306`
	config.CheckSecret(t, stringCfg, "secret_password")
}
