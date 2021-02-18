//+build !race

package loki

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/loki/pkg/distributor"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLoki(t *testing.T) {
	//
	// Create a temporary file to tail
	//
	positionsDir, err := ioutil.TempDir(os.TempDir(), "positions-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(positionsDir)
	})

	tmpFile, err := ioutil.TempFile(os.TempDir(), "*.log")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpFile.Name())
	})

	//
	// Listen for push requests and pass them through to a channel
	//
	pushes := make(chan *logproto.PushRequest)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, lis.Close())
	})
	go func() {
		_ = http.Serve(lis, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			req, err := distributor.ParseRequest(r)
			require.NoError(t, err)

			pushes <- req
			_, _ = rw.Write(nil)
		}))
	}()

	//
	// Launch Loki so it starts tailing the file and writes to our server.
	//
	cfgText := util.Untab(fmt.Sprintf(`
positions_directory: %s
configs:
- name: default
  clients:
  - url: http://%s/loki/api/v1/push
		batchwait: 50ms
		batchsize: 1
  scrape_configs:
  - job_name: system
    static_configs:
    - targets: [localhost]
      labels:
        job: test
        __path__: %s
	`, positionsDir, lis.Addr().String(), tmpFile.Name()))

	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&cfg))

	logger := log.NewSyncLogger(log.NewNopLogger())
	l, err := New(prometheus.NewRegistry(), cfg, logger)
	require.NoError(t, err)
	defer l.Stop()

	//
	// Write a log line and wait for it to come through.
	//
	fmt.Fprintf(tmpFile, "Hello, world!\n")
	select {
	case <-time.After(time.Second * 30):
		require.FailNow(t, "timed out waiting for data to be pushed")
	case req := <-pushes:
		require.Equal(t, "Hello, world!", req.Streams[0].Entries[0].Line)
	}

	//
	// Apply a new config and write a new line.
	//
	cfgText = util.Untab(fmt.Sprintf(`
positions_directory: %s
configs:
- name: default
  clients:
  - url: http://%s/loki/api/v1/push
		batchwait: 50ms
		batchsize: 5
  scrape_configs:
  - job_name: system
    static_configs:
    - targets: [localhost]
      labels:
        job: test-2
        __path__: %s
	`, positionsDir, lis.Addr().String(), tmpFile.Name()))

	var newCfg Config
	dec = yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&newCfg))

	require.NoError(t, l.ApplyConfig(newCfg))

	fmt.Fprintf(tmpFile, "Hello again!\n")
	select {
	case <-time.After(time.Second * 30):
		require.FailNow(t, "timed out waiting for data to be pushed")
	case req := <-pushes:
		require.Equal(t, "Hello again!", req.Streams[0].Entries[0].Line)
	}
}
