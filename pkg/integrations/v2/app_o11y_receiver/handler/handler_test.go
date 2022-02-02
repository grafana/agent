package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/exporters"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/stretchr/testify/assert"
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

func (te *TestExporter) Export(payload models.Payload) error {
	if te.broken {
		return errors.New("this exporter is broken")
	}
	te.payloads = append(te.payloads, payload)
	return nil
}

func TestMultipleExportersAllSucceed(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

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

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{&exporter1, &exporter2})
	handler := fr.HTTPHandler(log.NewNopLogger())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Result().StatusCode)

	assert.Len(t, exporter1.payloads, 1)
	assert.Len(t, exporter2.payloads, 1)
}

func TestMultipleExportersOneFails(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	require.NoError(t, err)

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

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{&exporter1, &exporter2})
	handler := fr.HTTPHandler(log.NewNopLogger())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Result().StatusCode)

	assert.Len(t, exporter1.payloads, 0)
	assert.Len(t, exporter2.payloads, 1)
}

func TestMultipleExportersAllFail(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

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

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{&exporter1, &exporter2})
	handler := fr.HTTPHandler(log.NewNopLogger())

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Result().StatusCode)

	assert.Len(t, exporter1.payloads, 0)
	assert.Len(t, exporter2.payloads, 0)
}

func TestNoContentLengthLimitSet(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	require.NoError(t, err)

	conf := config.AppO11yReceiverConfig{}

	req.ContentLength = 89348593894

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{})
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Result().StatusCode)
}

func TestLargePayload(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.AppO11yReceiverConfig{
		MaxAllowedPayloadSize: 10,
	}

	fr := NewAppO11yHandler(conf, []exporters.AppO11yReceiverExporter{})
	handler := fr.HTTPHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rr.Result().StatusCode)
}
