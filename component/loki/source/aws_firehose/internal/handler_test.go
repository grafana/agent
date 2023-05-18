package internal

import (
	"context"
	"encoding/json"
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

const directPutData = `{"requestId":"a1af4300-6c09-4916-ba8f-12f336176246","timestamp":1684422829730,"records":[{"data":"eyJDSEFOR0UiOi0wLjIzLCJQUklDRSI6NC44LCJUSUNLRVJfU1lNQk9MIjoiTkdDIiwiU0VDVE9SIjoiSEVBTFRIQ0FSRSJ9"},{"data":"eyJDSEFOR0UiOjYuNzYsIlBSSUNFIjo4Mi41NiwiVElDS0VSX1NZTUJPTCI6IlNMVyIsIlNFQ1RPUiI6IkVORVJHWSJ9"},{"data":"eyJDSEFOR0UiOi01LjkyLCJQUklDRSI6MTk5LjA4LCJUSUNLRVJfU1lNQk9MIjoiSEpWIiwiU0VDVE9SIjoiRU5FUkdZIn0="}]}`
const cloudwatchData = `{"requestId":"86208cf6-2bcc-47e6-9010-02ca9f44a025","timestamp":1684424042901,"records":[{"data":"H4sIAAAAAAAAADWO0QqCMBiFX2XsWsJSwrwLUW8sIYUuQmLpnxvpJttMQnz3Ztrlxzmc8424BaVIDfmnA+zjID3nlzS5n8IsO8YhtrAYOMg5aURfDUSXNBG1MkEj6liKvjPZQpmWQNoFVf9QpWSdZoJHrNEgFfZvxa8XvoHrGUfMqqWumdHQpDVj273nujvn4Nm251h/vVngmqBVD616PgoolC/Ga0SBNJoi8USVWWKczM8oYhKoULDBUzF9AScBUFTsAAAA"},{"data":"H4sIAAAAAAAAAKVW72/iRhD9VyxUqV9i2Jn9Mbv+xincKVKTqwLthx5RZPCSugKTYnPX6+n+9z5DqqIEI9LwyeyOZ997O2/G33qrWNf5Q5x8fYy9rHc5nAzvr0fj8fDDqHfRW3+p4gbL2jnHSrFW1mF5uX74sFlvH7EzyL/Ug2W+mhX5AMsPZfWQ7v/u48bNJuYrBDLeHig7ID/49MNPw8loPLkzyhbaqEKrhTdmbvJCi9OBjQ/a6IVFino7q+eb8rEp19X7ctnETd3LPvWaWDf3+3MW22re7t7jtMUuone3O3r0OVZNG/2tVxYtCbFITsoSGe/EBscGP6s0W2UBTzQrE0hIB+ySU8zBew8UTQmZmnwFxuQAlY1S2hm6+Fc+pL+6uZrcjyfD20lyu63aN5JfgRbIsqRaF/GPOiPf/+ymzbPtZHh7kyX5psqgZbYnlW3rNOZ1k3KWbfbhmSed0zwUfsEyJxd5odn7IndsbcznAk6GaK6Vy2OYL8SZ4GczcIpFUcSCp1Xv+8ULLawRb4NWhGjoYiSw18LOiVhhouCM86KVsNFWd2nhWR1q8SRD/HOL0KsiS5zmQrOoVFsTU2hu0uAipVHnYpj8HMX1n1xP9dEB2LKmgBRWA7lxCsusNCAQA7YJ2CUkdd6Sd92A+RBwW52psin5CdlM64x1HyG/TZtzkE+bq5v3H6fN73G5XCflj6skT/b3mORVkZRNnRyeTMeJOW28ZlYi0JokqGC1FbJBvMfhxrPFDSPYKSuhk5gNh8RGN5evvYe3o/PqTHS3o58/vr5Qps3ldpM3u1LR0iefrOpp865cLmORHGzt16/jar35mozLv2OWEPvk+h0W87+Sp41f6tiea3frV1XZHKQgkb4PuzRHVbF4UCoE9A2nvIhyTgWCWBAGDdM71Kdua1YcOlrXnVlRctI97GgxW8w49bldpCaIpDOXuzSKItifYyR3nnsA2LMOIcDtCq6GHmjCThR7wR2SVtyyYRgoiOtyDwD70+6xfYTAPecgf417Wqk6iHlljRKjUJvslOw6GyvXvtNWrcfVABPaBwaMdBIjf8I9Z7F5O7quifMc3Uv3nCf3QX1T39Fx9xC/xj1ut36UexB8PQQO2uvQThoO6MzWgREL5hf2Q0CjDmCPxt01YVCPfNIjgSnn3BWpnRWUYoxJitJWKaaXN8HMMBlm53kEgH3g1hvAZWDg9lsAKC1sQ2S9Y89eYTDiKrXuanUArE97RPoIgUfOQf4aj7QndxDD94HVaFQGUwg9gDHwNVoBXBWMJobj0doVo8WL6fII0psTHjmLzdvRdXnkObqXHjlP7gMj9MUct8j/dMjd938APEIoxXYLAAA="}]}`

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
			Body: directPutData,
			Assert: func(t *testing.T, res *httptest.ResponseRecorder, entries []loki.Entry) {
				r := response{}
				require.NoError(t, json.Unmarshal(res.Body.Bytes(), &r))

				require.Equal(t, 200, res.Code)
				require.Equal(t, "a1af4300-6c09-4916-ba8f-12f336176246", r.RequestID)
				require.Len(t, entries, 3)
			},
		},
		"cloudwatch logs-subscription data": {
			Body: cloudwatchData,
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

			req, err := http.NewRequest("POST", "http://test", strings.NewReader(cloudwatchData))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			// delegate assertions
			tc.Assert(t, recorder, testReceiver.entries)
		})
	}
}
