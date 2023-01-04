package snowflake_exporter

import (
	"os"
	"testing"

	"github.com/go-kit/log"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYaml(t *testing.T) {
	strConfig := `
account_name: "some_account"
username: "some_user"
password: "some_password"
warehouse: "some_warehouse"`

	var c Config

	require.NoError(t, yaml.UnmarshalStrict([]byte(strConfig), &c))

	require.Equal(t, Config{
		AccountName: "some_account",
		Username:    "some_user",
		Password:    "some_password",
		Warehouse:   "some_warehouse",
		Role:        "ACCOUNTADMIN",
	}, c)
}

func TestConfig_Identifier(t *testing.T) {
	t.Run("Identifier is in common config", func(t *testing.T) {
		c := DefaultConfig

		ik := "my-snowflake-instance-key"
		c.Common.InstanceKey = &ik

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, ik, id)
	})

	t.Run("Identifier is not in common config", func(t *testing.T) {
		c := DefaultConfig
		c.AccountName = "snowflake-acct"

		id, err := c.Identifier(integrations_v2.Globals{})
		require.NoError(t, err)
		require.Equal(t, "snowflake-acct", id)
	})
}

func TestConfig_NewIntegration(t *testing.T) {
	t.Run("integration with valid config", func(t *testing.T) {
		c := &Config{
			AccountName: "some_account",
			Username:    "some_user",
			Password:    "some_password",
			Warehouse:   "some_warehouse",
			Role:        "ACCOUNTADMIN",
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout), integrations_v2.Globals{})
		require.NoError(t, err)
		require.NotNil(t, i)
	})

	t.Run("integration with invalid config", func(t *testing.T) {
		c := &Config{
			Username:  "some_user",
			Password:  "some_password",
			Warehouse: "some_warehouse",
			Role:      "ACCOUNTADMIN",
		}

		i, err := c.NewIntegration(log.NewJSONLogger(os.Stdout), integrations_v2.Globals{})
		require.Nil(t, i)
		require.ErrorContains(t, err, "")
	})
}

func TestConfig_ApplyDefaults(t *testing.T) {
	c := DefaultConfig

	err := c.ApplyDefaults(integrations_v2.Globals{})
	require.NoError(t, err)
}
