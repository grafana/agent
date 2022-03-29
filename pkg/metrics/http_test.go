package metrics

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/util/test"
	"github.com/go-kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

func TestAgent_ListInstancesHandler(t *testing.T) {
	fact := newFakeInstanceFactory()
	a, err := newAgent(prometheus.NewRegistry(), Config{
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
		require.NoError(t, a.mm.ApplyConfig(makeInstanceConfig("foo")))
		require.NoError(t, a.mm.ApplyConfig(makeInstanceConfig("bar")))

		expect := `{"status":"success","data":["bar","foo"]}`
		test.Poll(t, time.Second, true, func() interface{} {
			rr := httptest.NewRecorder()
			a.ListInstancesHandler(rr, r)
			return expect == rr.Body.String()
		})
	})
}

func TestAgent_ListTargetsHandler(t *testing.T) {
	fact := newFakeInstanceFactory()
	a, err := newAgent(prometheus.NewRegistry(), Config{
		WALDir: "/tmp/agent",
	}, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)

	mockManager := &instance.MockManager{
		ListInstancesFunc: func() map[string]instance.ManagedInstance { return nil },
		ListConfigsFunc:   func() map[string]instance.Config { return nil },
		ApplyConfigFunc:   func(_ instance.Config) error { return nil },
		DeleteConfigFunc:  func(name string) error { return nil },
		StopFunc:          func() {},
	}
	a.mm, err = instance.NewModalManager(prometheus.NewRegistry(), a.logger, mockManager, instance.ModeDistinct)
	require.NoError(t, err)

	r := httptest.NewRequest("GET", "/agent/api/v1/targets", nil)

	t.Run("scrape manager not ready", func(t *testing.T) {
		mockManager.ListInstancesFunc = func() map[string]instance.ManagedInstance {
			return map[string]instance.ManagedInstance{
				"test_instance": &mockInstanceScrape{},
			}
		}

		rr := httptest.NewRecorder()
		a.ListTargetsHandler(rr, r)
		expect := `{"status": "success", "data": []}`
		require.JSONEq(t, expect, rr.Body.String())
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
	})

	t.Run("scrape manager targets", func(t *testing.T) {
		tgt := scrape.NewTarget(labels.FromMap(map[string]string{
			model.JobLabel:         "job",
			model.InstanceLabel:    "instance",
			"foo":                  "bar",
			model.SchemeLabel:      "http",
			model.AddressLabel:     "localhost:12345",
			model.MetricsPathLabel: "/metrics",
		}), labels.FromMap(map[string]string{
			"__discovered__": "yes",
		}), nil)

		startTime := time.Date(1994, time.January, 12, 0, 0, 0, 0, time.UTC)
		tgt.Report(startTime, time.Minute, fmt.Errorf("something went wrong"))

		mockManager.ListInstancesFunc = func() map[string]instance.ManagedInstance {
			return map[string]instance.ManagedInstance{
				"test_instance": &mockInstanceScrape{
					tgts: map[string][]*scrape.Target{
						"group_a": {tgt},
					},
				},
			}
		}

		rr := httptest.NewRecorder()
		a.ListTargetsHandler(rr, r)
		expect := `{
			"status": "success",
			"data": [{
				"instance": "test_instance",
				"target_group": "group_a",
				"endpoint": "http://localhost:12345/metrics",
				"state": "down",
				"labels": {
					"foo": "bar",
					"instance": "instance",
					"job": "job"
				},
				"discovered_labels": {
					"__discovered__": "yes"
				},
				"last_scrape": "1994-01-12T00:00:00Z",
				"scrape_duration_ms": 60000,
				"scrape_error":"something went wrong"
			}]
		}`
		require.JSONEq(t, expect, rr.Body.String())
		require.Equal(t, http.StatusOK, rr.Result().StatusCode)
	})
}

type mockInstanceScrape struct {
	instance.NoOpInstance
	tgts map[string][]*scrape.Target
}

func (i *mockInstanceScrape) TargetsActive() map[string][]*scrape.Target {
	return i.tgts
}

func TestRemoteWriteHandler(t *testing.T) {
	// Create a mock instance.
	fact := newFakeInstanceFactory()
	a, err := newAgent(prometheus.NewRegistry(), Config{
		WALDir: "/tmp/agent",
		Configs: []instance.Config{
			makeInstanceConfig("instanceOne"),
		},
	}, log.NewNopLogger(), fact.factory)
	require.NoError(t, err)
	defer a.Stop()

	// Build the router that will be used to execute all requests.
	router := mux.NewRouter()
	router.HandleFunc("/agent/api/v1/metrics/instance/{instance}/write", a.PushMetricsHandler)

	// Build a 'valid' push request.
	rr := httptest.NewRecorder()
	buf, _, err := buildWriteRequest(writeRequestFixture.Timeseries, nil, nil, nil)
	require.NoError(t, err)
	r := httptest.NewRequest("POST", "http://localhost:12345/agent/api/v1/metrics/instance/instanceOne/write", bytes.NewReader(buf))

	// Assign a mockAppendable to that instance, so we can verify
	// that data was pushed correctly.
	mockAppendable := &mockAppendable{}
	var managedInstance instance.ManagedInstance
	gotInstance := func() bool {
		managedInstance, err = a.InstanceManager().GetInstance("instanceOne")
		if managedInstance == nil || err != nil {
			return false
		}
		return true
	}
	// The instance may not have started yet, so let's be patient here.
	require.Eventually(
		t,
		gotInstance,
		5*time.Second,
		1*time.Second,
	)
	managedInstance.(*fakeInstance).appender = mockAppendable

	// Execute the request and make sure that the status code and recorded data were correct.
	router.ServeHTTP(rr, r)
	require.Equal(t, http.StatusNoContent, rr.Code)

	i := 0
	j := 0
	for _, ts := range writeRequestFixture.Timeseries {
		labels := labelProtosToLabels(ts.Labels)
		for _, s := range ts.Samples {
			require.Equal(t, mockSample{labels, s.Timestamp, s.Value}, mockAppendable.samples[i])
			i++
		}

		for _, e := range ts.Exemplars {
			exemplarLabels := labelProtosToLabels(e.Labels)
			require.Equal(t, mockExemplar{labels, exemplarLabels, e.Timestamp, e.Value}, mockAppendable.exemplars[j])
			j++
		}
	}

	// Set a commit error and validate it's returned correctly.
	mockAppendable.commitErr = fmt.Errorf("commit failed")
	rr = httptest.NewRecorder()
	buf, _, err = buildWriteRequest(writeRequestFixture.Timeseries, nil, nil, nil)
	require.NoError(t, err)
	r = httptest.NewRequest("POST", "http://localhost:12345/agent/api/v1/metrics/instance/instanceOne/write", bytes.NewReader(buf))

	router.ServeHTTP(rr, r)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Equal(t, "commit failed\n", string(body))

	// Set the `latestSample`, build a request that contains an out-of-order
	// datapoint, and validate that it's rejected.
	rr = httptest.NewRecorder()
	mockAppendable.latestSample = 100
	buf, _, err = buildWriteRequest([]prompb.TimeSeries{{
		Labels:  []prompb.Label{{Name: "__name__", Value: "test_metric"}},
		Samples: []prompb.Sample{{Value: 1, Timestamp: 0}},
	}}, nil, nil, nil)
	require.NoError(t, err)
	r = httptest.NewRequest("POST", "http://localhost:12345/agent/api/v1/metrics/instance/instanceOne/write", bytes.NewReader(buf))

	router.ServeHTTP(rr, r)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

var writeRequestFixture = &prompb.WriteRequest{
	Timeseries: []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "test_metric1"},
				{Name: "b", Value: "c"},
				{Name: "baz", Value: "qux"},
				{Name: "d", Value: "e"},
				{Name: "foo", Value: "bar"},
			},
			Samples:   []prompb.Sample{{Value: 1, Timestamp: timestamp.FromTime(time.Now())}},
			Exemplars: []prompb.Exemplar{{Labels: []prompb.Label{{Name: "f", Value: "g"}}, Value: 1, Timestamp: 0}},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "test_metric1"},
				{Name: "b", Value: "c"},
				{Name: "baz", Value: "qux"},
				{Name: "d", Value: "e"},
				{Name: "foo", Value: "bar"},
			},
			Samples:   []prompb.Sample{{Value: 2, Timestamp: timestamp.FromTime(time.Now())}},
			Exemplars: []prompb.Exemplar{{Labels: []prompb.Label{{Name: "h", Value: "i"}}, Value: 2, Timestamp: 1}},
		},
	},
}

func buildWriteRequest(samples []prompb.TimeSeries, metadata []prompb.MetricMetadata, pBuf *proto.Buffer, buf []byte) ([]byte, int64, error) {
	var highest int64
	for _, ts := range samples {
		// At the moment we only ever append a TimeSeries with a single sample or exemplar in it.
		if len(ts.Samples) > 0 && ts.Samples[0].Timestamp > highest {
			highest = ts.Samples[0].Timestamp
		}
		if len(ts.Exemplars) > 0 && ts.Exemplars[0].Timestamp > highest {
			highest = ts.Exemplars[0].Timestamp
		}
	}

	req := &prompb.WriteRequest{
		Timeseries: samples,
		Metadata:   metadata,
	}

	if pBuf == nil {
		pBuf = proto.NewBuffer(nil) // For convenience in tests. Not efficient.
	} else {
		pBuf.Reset()
	}
	err := pBuf.Marshal(req)
	if err != nil {
		return nil, highest, err
	}

	// snappy uses len() to see if it needs to allocate a new slice. Make the
	// buffer as long as possible.
	if buf != nil {
		buf = buf[0:cap(buf)]
	}
	compressed := snappy.Encode(buf, pBuf.Bytes())
	return compressed, highest, nil
}

func labelProtosToLabels(labelPairs []prompb.Label) labels.Labels {
	result := make(labels.Labels, 0, len(labelPairs))
	for _, l := range labelPairs {
		result = append(result, labels.Label{
			Name:  l.Name,
			Value: l.Value,
		})
	}
	sort.Sort(result)
	return result
}

type mockAppendable struct {
	latestSample   int64
	samples        []mockSample
	latestExemplar int64
	exemplars      []mockExemplar
	commitErr      error
}

type mockSample struct {
	l labels.Labels
	t int64
	v float64
}

type mockExemplar struct {
	l  labels.Labels
	el labels.Labels
	t  int64
	v  float64
}

func (m *mockAppendable) Appender(_ context.Context) storage.Appender {
	return m
}

func (m *mockAppendable) Append(_ storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	if t < m.latestSample {
		return 0, storage.ErrOutOfOrderSample
	}

	m.latestSample = t
	m.samples = append(m.samples, mockSample{l, t, v})
	return 0, nil
}

func (m *mockAppendable) Commit() error {
	return m.commitErr
}

func (*mockAppendable) Rollback() error {
	return fmt.Errorf("not implemented")
}

func (m *mockAppendable) AppendExemplar(_ storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	if e.Ts < m.latestExemplar {
		return 0, storage.ErrOutOfOrderExemplar
	}

	m.latestExemplar = e.Ts
	m.exemplars = append(m.exemplars, mockExemplar{l, e.Labels, e.Ts, e.Value})
	return 0, nil
}
