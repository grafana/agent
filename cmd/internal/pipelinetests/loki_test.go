package pipelinetests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/cmd/internal/pipelinetests/internal/framework"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/client"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/phayes/freeport"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline_Loki_SelfLogsWrite(topT *testing.T) {
	framework.PipelineTest{
		ConfigFile: "testdata/self_logs_write.river",
		EventuallyAssert: func(t *assert.CollectT, context *framework.RuntimeContext) {
			// Reload the config in order to generate some logs
			_, _ = http.Get(fmt.Sprintf("http://localhost:%d/-/reload", context.AgentPort))
			line, labels := context.DataSentToLoki.FindLineContaining("config reloaded")
			assert.NotNil(t, line)
			assert.Equal(t, "{component=\"agent\"}", labels)
		},
	}.RunTest(topT)
}

func TestPipeline_Loki_APILogsWrite(topT *testing.T) {
	apiServerPort, err := freeport.GetFreePort()
	assert.NoError(topT, err)

	lokiClient := newTestLokiClient(topT, fmt.Sprintf("http://127.0.0.1:%d/loki/api/v1/push", apiServerPort))
	defer lokiClient.Stop()

	testLogEntry := loki.Entry{
		Labels: map[model.LabelName]model.LabelValue{"source": "test"},
		Entry:  logproto.Entry{Timestamp: time.Now(), Line: "hello world!"},
	}

	logLineSent := false

	framework.PipelineTest{
		ConfigFile: "testdata/loki_source_api_write.river",
		Environment: map[string]string{
			"API_SERVER_PORT": fmt.Sprintf("%d", apiServerPort),
		},
		EventuallyAssert: func(t *assert.CollectT, context *framework.RuntimeContext) {
			// Send the line if not yet sent
			if !logLineSent {
				lokiClient.Chan() <- testLogEntry
				logLineSent = true
			}
			// Verify we have received the line at the other end of the pipeline
			line, labels := context.DataSentToLoki.FindLineContaining("hello world!")
			assert.NotNil(t, line)
			assert.Equal(t, "{forwarded=\"true\", source=\"test\"}", labels)
		},
	}.RunTest(topT)
}

func newTestLokiClient(t *testing.T, url string) client.Client {
	fUrl := flagext.URLValue{}
	err := fUrl.Set(url)
	require.NoError(t, err)

	lokiClient, err := client.New(
		client.NewMetrics(nil),
		client.Config{
			URL:     fUrl,
			Timeout: 5 * time.Second,
		},
		0,
		0,
		false,
		log.NewNopLogger(),
	)
	require.NoError(t, err)
	return lokiClient
}
