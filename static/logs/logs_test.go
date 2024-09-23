//go:build !race

package logs

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grafana/loki/pkg/loghttp/push"

	"github.com/go-kit/log"
	"github.com/grafana/agent/internal/util"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLogs_NilConfig(t *testing.T) {
	l, err := New(prometheus.NewRegistry(), nil, util.TestLogger(t), false)
	require.NoError(t, err)
	require.NoError(t, l.ApplyConfig(nil, false))

	defer l.Stop()
}

func checkConfigReloadLog(t *testing.T, logs string, expectedOccurances int) {
	logLine := `level=debug component=logs logs_config=default msg="instance config hasn't changed, not recreating Promtail"`
	actualOccurances := strings.Count(logs, logLine)
	require.Equal(t, expectedOccurances, actualOccurances)
}

func TestLogs(t *testing.T) {
	reg := prometheus.NewRegistry()

	//
	// Create a temporary file to tail
	//
	positionsDir := t.TempDir()

	tmpFile, err := os.CreateTemp(os.TempDir(), "*.log")
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
			req, err := push.ParseRequest(log.NewNopLogger(), "user_id", r, nil, nil, push.ParseLokiRequest, nil)
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
    pipeline_stages:
    - metrics:
        log_lines_total:
          type: Counter
          description: "total number of log lines"
          prefix: my_promtail_custom_
          max_idle_duration: 24h
          config:
            match_all: true
            action: inc
`, positionsDir, lis.Addr().String(), tmpFile.Name()))

	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&cfg))
	require.NoError(t, cfg.ApplyDefaults())
	logBuffer := util.SyncBuffer{}
	logger := log.NewSyncLogger(log.NewLogfmtLogger(&logBuffer))
	l, err := New(reg, &cfg, logger, false)
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

	// The config did change.
	// We expect the config reload log line to not be printed.
	checkConfigReloadLog(t, logBuffer.String(), 0)

	// Windows file paths contain `\` characters.
	// Those are not allowed in Prometheus label values:
	// https://prometheus.io/docs/instrumenting/exposition_formats/#text-format-details
	// `label_value` can be any sequence of UTF-8 characters, but the backslash (\), double-quote ("),
	// and line feed (\n) characters have to be escaped as \\, \", and \n, respectively.
	tmpFileLabelVal := strings.ReplaceAll(tmpFile.Name(), `\`, `\\`)

	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP my_promtail_custom_log_lines_total total number of log lines
# TYPE my_promtail_custom_log_lines_total counter
my_promtail_custom_log_lines_total{filename="`+tmpFileLabelVal+`",job="test",logs_config="default"} 1
`), "my_promtail_custom_log_lines_total"))

	//
	// Apply the same config and try reloading.
	// Recreate the config struct to make sure it's clean.
	//
	var sameCfg Config
	dec = yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&sameCfg))
	require.NoError(t, sameCfg.ApplyDefaults())
	require.NoError(t, l.ApplyConfig(&sameCfg, false))

	checkConfigReloadLog(t, logBuffer.String(), 1)

	// The metrics should stay the same, as the config didn't change.
	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP my_promtail_custom_log_lines_total total number of log lines
# TYPE my_promtail_custom_log_lines_total counter
my_promtail_custom_log_lines_total{filename="`+tmpFileLabelVal+`",job="test",logs_config="default"} 1
`), "my_promtail_custom_log_lines_total"))

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
    pipeline_stages:
    - metrics:
        log_lines_total2:
          type: Counter
          description: "total number of log lines"
          prefix: my_promtail_custom2_
          max_idle_duration: 24h
          config:
            match_all: true
            action: inc
`, positionsDir, lis.Addr().String(), tmpFile.Name()))

	var newCfg Config
	dec = yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&newCfg))
	require.NoError(t, newCfg.ApplyDefaults())
	require.NoError(t, l.ApplyConfig(&newCfg, false))

	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(``),
		"my_promtail_custom_log_lines_total", "my_promtail_custom2_log_lines_total2"))

	fmt.Fprintf(tmpFile, "Hello again!\n")
	select {
	case <-time.After(time.Second * 30):
		require.FailNow(t, "timed out waiting for data to be pushed")
	case req := <-pushes:
		require.Equal(t, "Hello again!", req.Streams[0].Entries[0].Line)
	}

	// The config did change this time.
	// We expect the config reload log line to not be printed again.
	checkConfigReloadLog(t, logBuffer.String(), 1)

	// The metrics changed, and the old metric is no longer visible.
	require.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(`
	# HELP my_promtail_custom2_log_lines_total2 total number of log lines
	# TYPE my_promtail_custom2_log_lines_total2 counter
	my_promtail_custom2_log_lines_total2{filename="`+tmpFileLabelVal+`",job="test-2",logs_config="default"} 1
	`), "my_promtail_custom_log_lines_total", "my_promtail_custom2_log_lines_total2"))

	t.Run("update to nil", func(t *testing.T) {
		// Applying a nil config should remove all instances.
		err := l.ApplyConfig(nil, false)
		require.NoError(t, err)
		require.Len(t, l.instances, 0)
	})

	t.Run("re-apply previous config", func(t *testing.T) {
		// Applying a nil config should remove all instances.
		l.ApplyConfig(nil, false)

		// Re-Apply the previous config and write a new line.
		var newCfg Config
		dec = yaml.NewDecoder(strings.NewReader(cfgText))
		dec.SetStrict(true)
		require.NoError(t, dec.Decode(&newCfg))
		require.NoError(t, newCfg.ApplyDefaults())
		require.NoError(t, l.ApplyConfig(&newCfg, false))

		fmt.Fprintf(tmpFile, "Hello again!\n")
		select {
		case <-time.After(time.Second * 30):
			require.FailNow(t, "timed out waiting for data to be pushed")
		case req := <-pushes:
			require.Equal(t, "Hello again!", req.Streams[0].Entries[0].Line)
		}
	})
}

func TestLogs_PositionsDirectory(t *testing.T) {
	//
	// Create a temporary file to tail
	//
	positionsDir := t.TempDir()

	//
	// Launch Loki so it starts tailing the file and writes to our server.
	//
	cfgText := util.Untab(fmt.Sprintf(`
positions_directory: %[1]s/positions
configs:
- name: instance-a
  clients:
	- url: http://127.0.0.1:80/loki/api/v1/push
- name: instance-b
  positions:
	  filename: %[1]s/other-positions/instance.yml
  clients:
	- url: http://127.0.0.1:80/loki/api/v1/push
	`, positionsDir))

	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(cfgText))
	dec.SetStrict(true)
	require.NoError(t, dec.Decode(&cfg))
	require.NoError(t, cfg.ApplyDefaults())
	logger := util.TestLogger(t)
	l, err := New(prometheus.NewRegistry(), &cfg, logger, false)
	require.NoError(t, err)
	defer l.Stop()

	_, err = os.Stat(filepath.Join(positionsDir, "positions"))
	require.NoError(t, err, "default shared positions directory did not get created")
	_, err = os.Stat(filepath.Join(positionsDir, "other-positions"))
	require.NoError(t, err, "instance-specific positions directory did not get created")
}
