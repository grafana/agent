package s3

import (
	"bytes"
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrectBucket(t *testing.T) {
	o := component.Options{
		ID: "t1",
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
	u := pushFilesToFakeS3(t, "test.txt", "success!")

	s3File, err := New(component.Options{
		ID: "id1",
	}, Arguments{
		Path:          "s3://mybucket?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&endpoint=" + u.Host +  "/test.txt",
		PollFrequency: 1 * time.Second,
		IsSecret:      false,
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	go s3File.Run(ctx)
	time.Sleep(5 * time.Second)
	require.True(t, s3File.content == "success!")
	cancel()
}

func pushFilesToFakeS3(t *testing.T, filename string, filecontents string) *url.URL {
	t.Setenv("AWS_ANON", "true")

	backend := s3mem.New()
	faker := gofakes3.New(backend)

	srv := httptest.NewServer(faker.Server())
	_ = backend.CreateBucket("mybucket")
	t.Cleanup(srv.Close)
	_, err := backend.PutObject(
		"mybucket",
		filename,
		map[string]string{"Content-Type": "application/yaml"},
		bytes.NewBufferString(filecontents),
		int64(len(filecontents)),
	)
	assert.NoError(t, err)
	u, err := url.Parse(srv.URL)
	assert.NoError(t, err)
	return u
}
