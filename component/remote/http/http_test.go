package http_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/log/level"
	http_component "github.com/grafana/agent/component/remote/http"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/dskit/backoff"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	ctx := componenttest.TestContext(t)

	var handler lazyHandler
	srv := httptest.NewServer(&handler)
	defer srv.Close()

	handler.SetHandler(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "remote.http")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		url = "%s"

		poll_frequency = "50ms" 
		poll_timeout   = "25ms" 
	`, srv.URL)
	var args http_component.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

	requireExports := func(expect http_component.Exports) {
		eventually(t, 10*time.Millisecond, 100*time.Millisecond, 5, func() error {
			actual := ctrl.Exports().(http_component.Exports)
			if expect != actual {
				return fmt.Errorf("expected %#v, got %#v", expect, actual)
			}
			return nil
		})
	}

	requireExports(http_component.Exports{
		Content: rivertypes.OptionalSecret{
			IsSecret: false,
			Value:    "Hello, world!",
		},
	})

	// Change the content to ensure new exports get produced.
	handler.SetHandler(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Testing!")
	})
	require.NoError(t, ctrl.WaitExports(time.Second), "component didn't update exports")
	requireExports(http_component.Exports{
		Content: rivertypes.OptionalSecret{
			IsSecret: false,
			Value:    "Testing!",
		},
	})
}

func eventually(t *testing.T, min, max time.Duration, retries int, f func() error) {
	t.Helper()

	l := util.TestLogger(t)

	bo := backoff.New(context.Background(), backoff.Config{
		MinBackoff: min,
		MaxBackoff: max,
		MaxRetries: retries,
	})
	for bo.Ongoing() {
		err := f()
		if err == nil {
			return
		}

		level.Error(l).Log("msg", "condition failed", "err", err)
		bo.Wait()
		continue
	}

	require.NoError(t, bo.Err(), "condition failed")
}

type lazyHandler struct {
	mut   sync.Mutex
	inner http.HandlerFunc
}

func (lh *lazyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lh.mut.Lock()
	defer lh.mut.Unlock()

	if lh.inner == nil {
		http.NotFound(w, r)
		return
	}
	lh.inner.ServeHTTP(w, r)
}

func (lh *lazyHandler) SetHandler(h http.HandlerFunc) {
	lh.mut.Lock()
	defer lh.mut.Unlock()

	lh.inner = h
}
