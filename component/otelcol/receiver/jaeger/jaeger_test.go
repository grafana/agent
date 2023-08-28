package jaeger_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/receiver/jaeger"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

// Test ensures that otelcol.receiver.jaeger can start successfully.
func Test(t *testing.T) {
	httpAddr := getFreeAddr(t)

	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.receiver.jaeger")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		protocols {
			grpc {
				endpoint = "%s"
			}
		}

		output { /* no-op */ }
	`, httpAddr)
	var args jaeger.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second))

	// TODO(rfratto): is it worthwhile trying to make sure we can send data over
	// the client or can we trust that getting the component to run successfully
	// is enough?
	time.Sleep(100 * time.Millisecond)
}

func TestArguments_UnmarshalRiver(t *testing.T) {
	t.Run("grpc", func(t *testing.T) {
		in := `
			protocols { grpc {} }
			output {}
		`

		var args jaeger.Arguments
		require.NoError(t, river.Unmarshal([]byte(in), &args))

		defaults := &jaeger.GRPC{}
		defaults.SetToDefault()

		require.Equal(t, defaults, args.Protocols.GRPC)
		require.Nil(t, args.Protocols.ThriftHTTP)
		require.Nil(t, args.Protocols.ThriftBinary)
		require.Nil(t, args.Protocols.ThriftCompact)
	})

	t.Run("thrift_http", func(t *testing.T) {
		in := `
			protocols { thrift_http {} }
			output {} 
		`

		var args jaeger.Arguments
		require.NoError(t, river.Unmarshal([]byte(in), &args))

		defaults := &jaeger.ThriftHTTP{}
		defaults.SetToDefault()

		require.Nil(t, args.Protocols.GRPC)
		require.Equal(t, defaults, args.Protocols.ThriftHTTP)
		require.Nil(t, args.Protocols.ThriftBinary)
		require.Nil(t, args.Protocols.ThriftCompact)
	})

	t.Run("thrift_binary", func(t *testing.T) {
		in := `
			protocols { thrift_binary {} }
			output {}
		`

		var args jaeger.Arguments
		require.NoError(t, river.Unmarshal([]byte(in), &args))

		defaults := &jaeger.ThriftBinary{}
		defaults.SetToDefault()

		require.Nil(t, args.Protocols.GRPC)
		require.Nil(t, args.Protocols.ThriftHTTP)
		require.Equal(t, defaults, args.Protocols.ThriftBinary)
		require.Nil(t, args.Protocols.ThriftCompact)
	})

	t.Run("thrift_compact", func(t *testing.T) {
		in := `
			protocols { thrift_compact {} }
			output {}
		`

		var args jaeger.Arguments
		require.NoError(t, river.Unmarshal([]byte(in), &args))

		defaults := &jaeger.ThriftCompact{}
		defaults.SetToDefault()

		require.Nil(t, args.Protocols.GRPC)
		require.Nil(t, args.Protocols.ThriftHTTP)
		require.Nil(t, args.Protocols.ThriftBinary)
		require.Equal(t, defaults, args.Protocols.ThriftCompact)
	})
}

func getFreeAddr(t *testing.T) string {
	t.Helper()

	portNumber, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("localhost:%d", portNumber)
}
