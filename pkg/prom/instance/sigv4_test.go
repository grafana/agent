package instance

import (
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

func TestSigV4RoundTripper(t *testing.T) {
	var gotReq *http.Request

	rt := &sigV4RoundTripper{
		cfg: SigV4Config{
			Enabled: true,
			Region:  "us-east-2",
		},
		next: promhttp.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			gotReq = req
			return &http.Response{StatusCode: http.StatusOK}, nil
		}),
		signer: signer.NewSigner(credentials.NewStaticCredentials(
			"test-id",
			"secret",
			"token",
		)),
	}
	cli := &http.Client{Transport: rt}

	req, err := http.NewRequest(http.MethodPost, "google.com", strings.NewReader("Hello, world!"))
	require.NoError(t, err)
	_, err = cli.Do(req)
	require.NoError(t, err)

	require.NotNil(t, gotReq)
	require.NotEmpty(t, gotReq.Header.Get("Authorization"))
}
