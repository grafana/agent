package instance

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
)

// SigV4Config configures signing requests with SigV4.
type SigV4Config struct {
	Enabled bool   `yaml:"enabled"`
	Region  string `yaml:"region"`
}

type sigV4RoundTripper struct {
	cfg  SigV4Config
	next http.RoundTripper

	signer *signer.Signer
}

// NewSigV4RoundTripper returns a new http.RoundTripper that will sign requests
// using Amazon's Signature Verification V4 signing procedure. The request will
// then be handed off to the next RoundTripper provided by next. If next is nil,
// http.DefaultTransport will be used.
//
// Credentials for signing are retrieving used the default AWS credential chain.
// If credentials could not be found, an error will be returned.
func NewSigV4RoundTripper(cfg SigV4Config, next http.RoundTripper) (http.RoundTripper, error) {
	if next == nil {
		next = http.DefaultTransport
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create new AWS session: %w", err)
	}
	if _, err := sess.Config.Credentials.Get(); err != nil {
		return nil, fmt.Errorf("could not get sigv4 credentials: %w", err)
	}

	return &sigV4RoundTripper{
		cfg:  cfg,
		next: next,

		signer: signer.NewSigner(sess.Config.Credentials),
	}, nil
}

func (rt *sigV4RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// rt.signer.Sign needs a seekable body, so we replace the body with a
	// buffered reader filled with the contents of original body.
	//
	// TODO(rfratto): This could be enhanced with a buf pool for reading request
	// bodies rather than creating a new buffer each time.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, req.Body); err != nil {
		return nil, err
	}
	// Close the original body since we don't need it anymore.
	_ = req.Body.Close()

	// Ensure our seeker is back at the start of the buffer once we return.
	var seeker io.ReadSeeker = bytes.NewReader(buf.Bytes())
	defer seeker.Seek(0, io.SeekStart)
	req.Body = ioutil.NopCloser(seeker)

	_, err := rt.signer.Sign(req, seeker, "aps", rt.cfg.Region, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return rt.next.RoundTrip(req)
}
