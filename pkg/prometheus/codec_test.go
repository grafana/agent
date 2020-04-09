package prometheus

import (
	"testing"
	"time"

	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
)

func TestCodec(t *testing.T) {
	// Not a full scrape config but fills out enough to make
	// sure the Codec is working properly.
	in := InstanceConfig{
		Name:                "test",
		HostFilter:          true,
		RemoteFlushDeadline: 10 * time.Minute,
		RemoteWrite: []*config.RemoteWriteConfig{{
			Name: "remote",
		}},
		ScrapeConfigs: []*config.ScrapeConfig{{
			JobName: "job",
			Scheme:  "http",
		}},
	}

	c := &jsonCodec{}
	bb, err := c.Encode(in)
	require.NoError(t, err)

	out, err := c.Decode(bb)
	require.NoError(t, err)
	require.Equal(t, in, out)
}
