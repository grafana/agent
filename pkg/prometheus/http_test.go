package prometheus

import (
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/prometheus/instance"
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
		a.cm.ApplyConfig(instance.Config{Name: "foo"})
		a.cm.ApplyConfig(instance.Config{Name: "bar"})

		// Responses must be sorted
		rr := httptest.NewRecorder()
		a.ListInstancesHandler(rr, r)
		expect := `{"status":"success","data":["bar","foo"]}`
		require.Equal(t, expect, rr.Body.String())
	})

}
