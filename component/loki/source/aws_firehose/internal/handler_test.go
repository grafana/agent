package internal

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

//go:embed testdata/*
var testData embed.FS

func readTestData(t *testing.T, name string) string {
	f, err := testData.ReadFile(name)
	if err != nil {
		require.FailNow(t, fmt.Sprintf("error reading test data: %s", name))
	}
	return string(f)
}

type receiver struct {
	entries []loki.Entry
}

func (r *receiver) Send(ctx context.Context, entry loki.Entry) {
	r.entries = append(r.entries, entry)
}

type response struct {
	RequestID string `json:"requestId"`
}

func TestHandler(t *testing.T) {
	type testcase struct {
		Body   string
		Assert func(t *testing.T, res *httptest.ResponseRecorder, entries []loki.Entry)
	}

	tests := map[string]testcase{
		"direct put data": {
			Body: readTestData(t, "testdata/DirectPUT.json"),
			Assert: func(t *testing.T, res *httptest.ResponseRecorder, entries []loki.Entry) {
				r := response{}
				require.NoError(t, json.Unmarshal(res.Body.Bytes(), &r))

				require.Equal(t, 200, res.Code)
				require.Equal(t, "a1af4300-6c09-4916-ba8f-12f336176246", r.RequestID)
				require.Len(t, entries, 3)
			},
		},
		"cloudwatch logs-subscription data": {
			Body: readTestData(t, "testdata/CloudwatchLogsLambda.json"),
			Assert: func(t *testing.T, res *httptest.ResponseRecorder, entries []loki.Entry) {
				r := response{}
				require.NoError(t, json.Unmarshal(res.Body.Bytes(), &r))

				require.Equal(t, 200, res.Code)
				require.Equal(t, "86208cf6-2bcc-47e6-9010-02ca9f44a025", r.RequestID)
				require.Len(t, entries, 2)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			w := log.NewSyncWriter(os.Stderr)
			logger := log.NewLogfmtLogger(w)

			testReceiver := &receiver{entries: make([]loki.Entry, 0)}
			handler := NewHandler(testReceiver, logger, prometheus.NewRegistry())

			req, err := http.NewRequest("POST", "http://test", strings.NewReader(tc.Body))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			// delegate assertions
			tc.Assert(t, recorder, testReceiver.entries)
		})
	}
}
