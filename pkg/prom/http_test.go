package prom

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/scrape"
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

func TestAgent_ListTargetsHandler(t *testing.T) {
	fact := newMockInstanceFactory()
	a, err := newAgent(Config{
		WALDir: "/tmp/agent",
	}, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)

	r := httptest.NewRequest("GET", "/agent/api/v1/targets", nil)

	t.Run("scrape manager not ready", func(t *testing.T) {
		a.instances = map[string]inst{
			"test_instance": &mockInstanceScrape{},
		}

		rr := httptest.NewRecorder()
		a.ListTargetsHandler(rr, r)
		expect := `{"status":"success","data":[]}`
		require.Equal(t, expect, rr.Body.String())
	})

	t.Run("scrape manager targets", func(t *testing.T) {
		tgt := scrape.NewTarget(labels.FromMap(map[string]string{
			model.JobLabel:         "job",
			model.InstanceLabel:    "instance",
			model.SchemeLabel:      "http",
			model.AddressLabel:     "localhost:12345",
			model.MetricsPathLabel: "/metrics",
			"foo":                  "bar",
		}), nil, nil)

		startTime := time.Date(1994, time.January, 12, 0, 0, 0, 0, time.UTC)
		tgt.Report(startTime, time.Minute, fmt.Errorf("something went wrong"))

		a.instances = map[string]inst{
			"test_instance": &mockInstanceScrape{
				tgts: map[string][]*scrape.Target{
					"group_a": {tgt},
				},
			},
		}

		rr := httptest.NewRecorder()
		a.ListTargetsHandler(rr, r)
		expect := `{"status":"success","data":[{"instance":"test_instance","target_group":"group_a","endpoint":"http://localhost:12345/metrics","state":"down","labels":{"foo":"bar","instance":"instance","job":"job"},"last_scrape":"1994-01-12T00:00:00Z","scrape_duration_ms":60000,"scrape_error":"something went wrong"}]}`
		require.Equal(t, expect, rr.Body.String())
	})
}

type mockInstanceScrape struct {
	tgts map[string][]*scrape.Target
}

func (i *mockInstanceScrape) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (i *mockInstanceScrape) TargetsActive() map[string][]*scrape.Target {
	return i.tgts
}
