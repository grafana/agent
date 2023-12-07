//go:build !nodocker

package vault

import (
	"fmt"
	stdlog "log"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/docker/go-connections/nat"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/grafana/river/rivertypes"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_GetSecrets(t *testing.T) {
	var (
		ctx = componenttest.TestContext(t)
		l   = util.TestLogger(t)
	)

	cli := getTestVaultServer(t)

	// Store a secret in value to use from the component.
	_, err := cli.KVv2("secret").Put(ctx, "test", map[string]any{
		"key": "value",
	})
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		server = "%s"
		path   = "secret/test"

		reread_frequency = "0s"

		auth.token {
			token = "%s"
		}
	`, cli.Address(), cli.Token())

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	ctrl, err := componenttest.NewControllerFromID(l, "remote.vault")
	require.NoError(t, err)

	go func() {
		require.NoError(t, ctrl.Run(ctx, args))
	}()

	require.NoError(t, ctrl.WaitRunning(time.Minute))
	require.NoError(t, ctrl.WaitExports(time.Minute))

	var (
		expectExports = Exports{
			Data: map[string]rivertypes.Secret{
				"key": rivertypes.Secret("value"),
			},
		}
		actualExports = ctrl.Exports().(Exports)
	)
	require.Equal(t, expectExports, actualExports)
}

func Test_PollSecrets(t *testing.T) {
	var (
		ctx = componenttest.TestContext(t)
		l   = util.TestLogger(t)
	)

	cli := getTestVaultServer(t)

	// Store a secret in value to use from the component.
	_, err := cli.KVv2("secret").Put(ctx, "test", map[string]any{
		"key": "value",
	})
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		server = "%s"
		path   = "secret/test"

		reread_frequency = "100ms"

		auth.token {
			token = "%s"
		}
	`, cli.Address(), cli.Token())

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	ctrl, err := componenttest.NewControllerFromID(l, "remote.vault")
	require.NoError(t, err)

	go func() {
		require.NoError(t, ctrl.Run(ctx, args))
	}()
	require.NoError(t, ctrl.WaitRunning(time.Minute))

	// Get the initial secret.
	{
		require.NoError(t, ctrl.WaitExports(time.Minute))

		var (
			expectExports = Exports{
				Data: map[string]rivertypes.Secret{
					"key": rivertypes.Secret("value"),
				},
			}
			actualExports = ctrl.Exports().(Exports)
		)
		require.Equal(t, expectExports, actualExports)
	}

	// Get an updated secret.
	{
		_, err := cli.KVv2("secret").Put(ctx, "test", map[string]any{
			"key": "newvalue",
		})
		require.NoError(t, err)

		require.NoError(t, ctrl.WaitExports(time.Minute))

		var (
			expectExports = Exports{
				Data: map[string]rivertypes.Secret{
					"key": rivertypes.Secret("newvalue"),
				},
			}
			actualExports = ctrl.Exports().(Exports)
		)
		require.Equal(t, expectExports, actualExports)
	}
}

func getTestVaultServer(t *testing.T) *vaultapi.Client {
	// TODO: this is broken with go 1.20.6
	// waiting on https://github.com/testcontainers/testcontainers-go/issues/1359
	t.Skip()
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "hashicorp/vault:1.13.2",
			ExposedPorts: []string{"80/tcp"},
			Env: map[string]string{
				"VAULT_DEV_ROOT_TOKEN_ID":  "secretkey",
				"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:80",
			},
			WaitingFor: wait.ForHTTP("/v1/sys/health"),
		},
		Started: true,
		Logger:  stdlog.New(log.NewStdlibAdapter(l), "", 0),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	ep, err := container.PortEndpoint(ctx, nat.Port("80/tcp"), "http")
	require.NoError(t, err)

	cli, err := vaultapi.NewClient(&vaultapi.Config{Address: ep})
	require.NoError(t, err)

	cli.SetToken("secretkey")
	return cli
}
