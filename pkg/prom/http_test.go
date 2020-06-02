package prom

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func TestAgent_ListInstancesHandler(t *testing.T) {
	fact := newMockInstanceFactory()
	a, err := newAgent(Config{
		WALDir: "/tmp/agent",
	}, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)
	defer a.Stop()

	r := httptest.NewRequest("GET", "/agent/api/v1/instances", nil)

	t.Run("no instances", func(t *testing.T) {
		rr := httptest.NewRecorder()
		a.ListInstancesHandler(rr, r)
		expect := `{"status":"success","data":[]}`
		require.Equal(t, expect, rr.Body.String())
	})

	t.Run("non-empty", func(t *testing.T) {
		require.NoError(t, a.cm.ApplyConfig(makeInstanceConfig("foo")))
		require.NoError(t, a.cm.ApplyConfig(makeInstanceConfig("bar")))

		expect := `{"status":"success","data":["bar","foo"]}`
		test.Poll(t, time.Second, true, func() interface{} {
			rr := httptest.NewRecorder()
			a.ListInstancesHandler(rr, r)
			return expect == rr.Body.String()
		})
	})
}
