// Package httpclient provides an httpgrpc client to nodes in the cluster.
package httpclient

import (
	"bytes"
	"io"
	"net/http"

	"github.com/weaveworks/common/httpgrpc"
	httpgrpc_server "github.com/weaveworks/common/httpgrpc/server"
)

// RoundTripper performs an HTTP request against a fixed client.
type RoundTripper struct {
	// Client is the httpgrpc client where the HTTP request will be sent.
	Client httpgrpc.HTTPClient
}

func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	finalR := r.Clone(r.Context())
	if finalR.Body == nil {
		finalR.Body = noOpReader{}
	}
	if finalR.RequestURI == "" {
		finalR.RequestURI = r.URL.RequestURI()
	}

	grpcReq, err := httpgrpc_server.HTTPRequest(finalR)
	if err != nil {
		return nil, err
	}

	grpcResp, err := rt.Client.Handle(finalR.Context(), grpcReq)
	if err != nil {
		var ok bool
		grpcResp, ok = httpgrpc.HTTPResponseFromError(err)
		if !ok {
			return nil, err
		}
	}

	resp := &http.Response{
		Status:     http.StatusText(int(grpcResp.Code)),
		StatusCode: int(grpcResp.Code),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,

		Header:        make(http.Header, len(grpcReq.Headers)),
		Body:          io.NopCloser(bytes.NewReader(grpcResp.Body)),
		ContentLength: int64(len(grpcResp.Body)),
		Request:       r,
	}
	for _, header := range grpcReq.Headers {
		for _, value := range header.Values {
			resp.Header.Add(header.Key, value)
		}
	}
	return resp, nil
}

type noOpReader struct{}

func (noOpReader) Read(b []byte) (n int, err error) { return 0, io.EOF }
func (noOpReader) Close() error                     { return nil }
