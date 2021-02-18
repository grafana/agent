package instance

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

func TestSigV4_Inferred_Region(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "secret")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "token")
	os.Setenv("AWS_REGION", "us-west-2")

	sess, err := session.NewSession(&aws.Config{
		// Setting to an empty string to demostrate the default value from the yaml
		// won't override the environment's region.
		Region: aws.String(""),
	})
	require.NoError(t, err)
	_, err = sess.Config.Credentials.Get()
	require.NoError(t, err)

	require.NotNil(t, sess.Config.Region)
	require.Equal(t, "us-west-2", *sess.Config.Region)
}

func TestSigV4RoundTripper(t *testing.T) {
	var gotReq *http.Request

	rt := &sigV4RoundTripper{
		region: "us-east-2",
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
	rt.pool.New = rt.newBuf

	cli := &http.Client{Transport: rt}

	req, err := http.NewRequest(http.MethodPost, "google.com", strings.NewReader("Hello, world!"))
	require.NoError(t, err)
	_, err = cli.Do(req)
	require.NoError(t, err)

	require.NotNil(t, gotReq)
	require.NotEmpty(t, gotReq.Header.Get("Authorization"))
}
