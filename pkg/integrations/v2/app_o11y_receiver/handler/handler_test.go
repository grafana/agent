package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/exporters"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/prometheus/client_golang/prometheus"
)

const PAYLOAD = `
{
  "traces": {
    "resourceSpans": []
  },
  "logs": [],
  "exceptions": [],
  "measurements": [],
  "meta": {}
}
`

type TestExporter struct {
	name     string
	broken   bool
	payloads []models.Payload
}

func (te *TestExporter) Name() string {
	return te.name
}

func (te *TestExporter) Export(ctx context.Context, payload models.Payload) error {
	if te.broken {
		return errors.New("this exporter is broken")
	}
	te.payloads = append(te.payloads, payload)
	return nil
}

func TestMultipleExportersAllSucceed(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	reg := prometheus.NewRegistry()

	require.NoError(t, err)

	exporter1 := TestExporter{
		name:     "exporter1",
		broken:   false,
		payloads: []models.Payload{},
	}
	exporter2 := TestExporter{
		name:     "exporter2",
		broken:   false,
		payloads: []models.Payload{},
	}

	conf := config.AppO11yReceiverConfig{}

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{&exporter1, &exporter2}, reg)
	handler := fr.HTTPHandler(log.NewNopLogger())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)

	require.Len(t, exporter1.payloads, 1)
	require.Len(t, exporter2.payloads, 1)
}

func TestMultipleExportersOneFails(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	require.NoError(t, err)

	reg := prometheus.NewRegistry()

	exporter1 := TestExporter{
		name:     "exporter1",
		broken:   true,
		payloads: []models.Payload{},
	}
	exporter2 := TestExporter{
		name:     "exporter2",
		broken:   false,
		payloads: []models.Payload{},
	}

	conf := config.AppO11yReceiverConfig{}

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{&exporter1, &exporter2}, reg)
	handler := fr.HTTPHandler(log.NewNopLogger())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	metrics, err := reg.Gather()
	require.NoError(t, err)

	metric := metrics[0]
	require.Equal(t, "app_agent_receiver_exporter_errors_total", *metric.Name)
	require.Len(t, metric.Metric, 1)
	require.Equal(t, 1.0, *metric.Metric[0].Counter.Value)
	require.Len(t, metric.Metric[0].Label, 1)
	require.Equal(t, *metric.Metric[0].Label[0].Value, "exporter1")
	require.Len(t, metrics, 1)
	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
	require.Len(t, exporter1.payloads, 0)
	require.Len(t, exporter2.payloads, 1)
}

func TestMultipleExportersAllFail(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	reg := prometheus.NewRegistry()

	require.NoError(t, err)

	exporter1 := TestExporter{
		name:     "exporter1",
		broken:   true,
		payloads: []models.Payload{},
	}
	exporter2 := TestExporter{
		name:     "exporter2",
		broken:   true,
		payloads: []models.Payload{},
	}

	conf := config.AppO11yReceiverConfig{}

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{&exporter1, &exporter2}, reg)
	handler := fr.HTTPHandler(log.NewNopLogger())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	metrics, err := reg.Gather()
	require.NoError(t, err)

	require.Len(t, metrics, 1)
	metric := metrics[0]

	require.Equal(t, "app_agent_receiver_exporter_errors_total", *metric.Name)
	require.Len(t, metric.Metric, 2)
	require.Equal(t, 1.0, *metric.Metric[0].Counter.Value)
	require.Equal(t, 1.0, *metric.Metric[1].Counter.Value)
	require.Len(t, metric.Metric[0].Label, 1)
	require.Len(t, metric.Metric[1].Label, 1)
	require.Equal(t, *metric.Metric[0].Label[0].Value, "exporter1")
	require.Equal(t, *metric.Metric[1].Label[0].Value, "exporter2")
	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
	require.Len(t, exporter1.payloads, 0)
	require.Len(t, exporter2.payloads, 0)
}

func TestNoContentLengthLimitSet(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))
	require.NoError(t, err)
	reg := prometheus.NewRegistry()

	conf := config.AppO11yReceiverConfig{}

	req.ContentLength = 89348593894

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{}, reg)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
}

func TestLargePayload(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))
	require.NoError(t, err)
	reg := prometheus.NewRegistry()

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			MaxAllowedPayloadSize: 10,
		},
	}

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{}, reg)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusRequestEntityTooLarge, rr.Result().StatusCode)
}

func TestAPIKeyRequiredButNotProvided(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			APIKey: "foo",
		},
	}

	fr := NewAppO11yHandler(conf, nil, prometheus.NewRegistry())
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestAPIKeyWrong(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))
	req.Header.Set("x-api-key", "bar")

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			APIKey: "foo",
		},
	}

	fr := NewAppO11yHandler(conf, nil, nil)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestAPIKeyCorrect(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))
	req.Header.Set("x-api-key", "foo")

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			APIKey: "foo",
		},
	}

	fr := NewAppO11yHandler(conf, nil, nil)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
}

func TestRateLimiterNoReject(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			RateLimiting: config.RateLimitingConfig{
				Burstiness: 10,
				RPS:        10,
				Enabled:    true,
			},
		},
	}

	fr := NewAppO11yHandler(conf, nil, nil)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
}

func TestRateLimiterReject(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			RateLimiting: config.RateLimitingConfig{
				Burstiness: 0,
				RPS:        0,
				Enabled:    true,
			},
		},
	}

	fr := NewAppO11yHandler(conf, nil, nil)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusTooManyRequests, rr.Result().StatusCode)
}

func TestRateLimiterDisabled(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		Server: config.ServerConfig{
			RateLimiting: config.RateLimitingConfig{
				Burstiness: 0,
				RPS:        0,
				Enabled:    false,
			},
		},
	}

	fr := NewAppO11yHandler(conf, nil, nil)
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
}
