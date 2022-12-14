//go:build linux
// +build linux

package s3

import (
	"bytes"
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/grafana/agent/component"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrectBucket(t *testing.T) {
	o := component.Options{
		ID:            "t1",
		OnStateChange: func(_ component.Exports) {},
		Registerer:    prometheus.NewRegistry(),
	}
	s3File, err := New(o,
		Arguments{
			Path:          "s3://bucket/file",
			PollFrequency: 30 * time.Second,
			IsSecret:      false,
		})
	require.NoError(t, err)
	require.NotNil(t, s3File)
}

func TestWatchingFile(t *testing.T) {
	var mut sync.Mutex
	_, srv := pushFilesToFakeS3(t, "test.txt", "success!")
	var output string
	s3File, err := New(component.Options{
		ID: "id1",
		OnStateChange: func(e component.Exports) {
			mut.Lock()
			defer mut.Unlock()
			output = e.(Exports).Content.Value
		},
		Registerer: prometheus.NewRegistry(),
	}, Arguments{
		Path:          "s3://mybucket/test.txt",
		PollFrequency: 1 * time.Second,
		IsSecret:      false,
		Options: ClientOptions{
			Endpoint:     srv.URL,
			DisableSSL:   true,
			UsePathStyle: true,
		},
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go s3File.Run(ctx)

	require.Eventually(t, func() bool {
		health := s3File.CurrentHealth()
		// This is due to race detector
		mut.Lock()
		defer mut.Unlock()
		t.Logf("Output: %s", output)
		t.Logf("S3 health: %s", health.Health)
		t.Logf("S3 health message: %s", health.Message)
		return output == "success!"
	}, 2*time.Second, 100*time.Millisecond)

	cancel()
}

func pushFilesToFakeS3(t *testing.T, filename string, filecontents string) (*s3mem.Backend, *httptest.Server) {
	t.Setenv("AWS_ANON", "true")

	backend := s3mem.New()
	faker := gofakes3.New(backend)
	srv := httptest.NewServer(faker.Server())
	_ = backend.CreateBucket("mybucket")
	t.Cleanup(srv.Close)
	pushFile(t, backend, filename, filecontents)
	return backend, srv
}

func pushFile(t *testing.T, backend *s3mem.Backend, filename string, filecontents string) {
	_, err := backend.PutObject(
		"mybucket",
		filename,
		map[string]string{"Content-Type": "application/yaml"},
		bytes.NewBufferString(filecontents),
		int64(len(filecontents)),
	)
	assert.NoError(t, err)
}
