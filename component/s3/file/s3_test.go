package s3

import (
	"testing"
	"time"

	"github.com/grafana/agent/component"
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
