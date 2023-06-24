package cluster

import (
	"encoding/base64"
	"net/http"
)

type Transport struct {
	auth Authentication
	T    http.RoundTripper
}

func NewTransport(T http.RoundTripper, auth Authentication) *Transport {
	if T == nil {
		T = http.DefaultTransport
	}
	return &Transport{auth, T}
}

func (adt *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if adt.auth.Enabled {
		basicAuth := base64.StdEncoding.EncodeToString([]byte(adt.auth.Username + ":" + adt.auth.Password))
		req.Header.Set("Authorization", "Basic "+basicAuth)
	}

	return adt.T.RoundTrip(req)
}
