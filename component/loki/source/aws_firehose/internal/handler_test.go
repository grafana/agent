package internal

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki"
	"github.com/klauspost/compress/gzip"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"io"
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
		for _, gzipContentEncoding := range []bool{true, false} {
			suffix := ""
			if gzipContentEncoding {
				suffix = " - with gzip content encoding"
			}
			t.Run(fmt.Sprintf("%s%s", name, suffix), func(t *testing.T) {
				w := log.NewSyncWriter(os.Stderr)
				logger := log.NewLogfmtLogger(w)

				testReceiver := &receiver{entries: make([]loki.Entry, 0)}
				handler := NewHandler(testReceiver, logger, prometheus.NewRegistry())

				bs := bytes.NewBuffer(nil)
				var bodyReader io.Reader = strings.NewReader(tc.Body)

				// if testing gzip content encoding, use the following read/writer chain
				// to compress the body: string reader -> gzip writer -> bytes buffer
				// after that use the same bytes buffer as reader
				if gzipContentEncoding {
					gzipWriter := gzip.NewWriter(bs)
					_, err := io.Copy(gzipWriter, bodyReader)
					require.NoError(t, err)
					require.NoError(t, gzipWriter.Close())
					bodyReader = bs
				}

				req, err := http.NewRequest("POST", "http://test", bodyReader)
				require.NoError(t, err)

				// Also content-encoding header needs to be set
				if gzipContentEncoding {
					req.Header.Set("Content-Encoding", "gzip")
				}

				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, req)
				// delegate assertions
				tc.Assert(t, recorder, testReceiver.entries)
			})
		}
	}
}
