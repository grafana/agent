package framework

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"

	"github.com/grafana/loki/pkg/logproto"
	loki_util "github.com/grafana/loki/pkg/util"
)

func newTestLokiServer(onWrite func(*logproto.PushRequest)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pushReq logproto.PushRequest
		err := loki_util.ParseProtoReader(context.Background(), r.Body, int(r.ContentLength), math.MaxInt32, &pushReq, loki_util.RawSnappy)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		onWrite(&pushReq)
	}))
}
