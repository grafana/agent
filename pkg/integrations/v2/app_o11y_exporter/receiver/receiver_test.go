package receiver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/exporters"
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

func TestNoLimitSet(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.ReceiverConfig{}

	req.ContentLength = 89348593894

	fr := NewAppReceiver(conf, []exporters.AppReceiverExporter{})
	handler := fr.ReceiverHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Result().StatusCode, http.StatusAccepted)
}

func TestLargePayload(t *testing.T) {
	req, err := http.NewRequest("POST", "/collect", bytes.NewBuffer([]byte(PAYLOAD)))

	if err != nil {
		t.Fatal(err)
	}

	conf := config.ReceiverConfig{
		MaxAllowedPayloadSize: 10,
	}

	fr := NewAppReceiver(conf, []exporters.AppReceiverExporter{})
	handler := fr.ReceiverHandler(nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Result().StatusCode, http.StatusRequestEntityTooLarge)
}
