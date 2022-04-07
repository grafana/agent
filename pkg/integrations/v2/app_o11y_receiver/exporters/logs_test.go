package exporters

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	prommodel "github.com/prometheus/common/model"

	"github.com/stretchr/testify/assert"
)

func loadTestData(t *testing.T, file string) models.Payload {
	t.Helper()
	// Safe to disable, this is a test.
	// nolint:gosec
	content, err := ioutil.ReadFile(filepath.Join("../models/testdata", file))
	assert.NoError(t, err, "expected to be able to read file")
	assert.True(t, len(content) > 0)
	var payload models.Payload
	err = json.Unmarshal(content, &payload)
	assert.NoError(t, err)
	return payload
}

type testLogsInstance struct {
	Entries []api.Entry
}

func (i *testLogsInstance) SendEntry(entry api.Entry, dur time.Duration) bool {
	i.Entries = append(i.Entries, entry)
	return true
}

type MockSourceMapStore struct{}

func (store *MockSourceMapStore) TransformException(ex *models.Exception, release string) *models.Exception {
	if ex.Stacktrace == nil {
		return ex
	}
	frames := []models.Frame{}
	for _, frame := range ex.Stacktrace.Frames {
		frame.Filename = strings.Replace(frame.Filename, ".js", ".ts", 1)
		frames = append(frames, frame)
	}
	transformed := *ex
	transformed.Stacktrace.Frames = frames
	return &transformed
}

func (store *MockSourceMapStore) ResolveSourceLocation(frame *models.Frame, release string) (*models.Frame, error) {
	return frame, nil
}

func TestExportLogs(t *testing.T) {
	inst := testLogsInstance{
		Entries: []api.Entry{},
	}

	logger := kitlog.NewNopLogger()

	logsExporter := NewLogsExporter(
		logger,
		LogsExporterConfig{
			LogsInstance: &inst,
			Labels: map[string]string{
				"app":  "frontend",
				"kind": "",
			},
			SendEntryTimeout: 100,
		},
		&MockSourceMapStore{},
	)

	payload := loadTestData(t, "payload.json")

	err := logsExporter.Export(payload)
	assert.NoError(t, err)

	assert.Len(t, inst.Entries, 4)

	// log1
	assert.Equal(t, prommodel.LabelSet{
		prommodel.LabelName("app"):  prommodel.LabelValue("frontend"),
		prommodel.LabelName("kind"): prommodel.LabelValue("log"),
	}, inst.Entries[0].Labels)
	expectedLine := "timestamp=\"2021-09-30 10:46:17.68 +0000 UTC\" kind=log message=\"opened pricing page\" level=info context_component=AppRoot context_page=Pricing sdk_name=grafana-frontend-agent sdk_version=1.0.0 app_name=testapp app_release=0.8.2 app_version=abcdefg app_environment=production user_attr_foo=bar session_id=abcd session_attr_time_elapsed=100s page_url=https://example.com/page browser_name=chrome browser_version=88.12.1 browser_os=linux browser_mobile=false"
	assert.Equal(t, expectedLine, inst.Entries[0].Line)

	// log2
	assert.Equal(t, prommodel.LabelSet{
		prommodel.LabelName("app"):  prommodel.LabelValue("frontend"),
		prommodel.LabelName("kind"): prommodel.LabelValue("log"),
	}, inst.Entries[1].Labels)
	expectedLine = "timestamp=\"2021-09-30 10:46:17.68 +0000 UTC\" kind=log message=\"loading price list\" level=trace context_component=AppRoot context_page=Pricing sdk_name=grafana-frontend-agent sdk_version=1.0.0 app_name=testapp app_release=0.8.2 app_version=abcdefg app_environment=production user_attr_foo=bar session_id=abcd session_attr_time_elapsed=100s page_url=https://example.com/page browser_name=chrome browser_version=88.12.1 browser_os=linux browser_mobile=false"
	assert.Equal(t, expectedLine, inst.Entries[1].Line)

	// exception
	assert.Equal(t, prommodel.LabelSet{
		prommodel.LabelName("app"):  prommodel.LabelValue("frontend"),
		prommodel.LabelName("kind"): prommodel.LabelValue("exception"),
	}, inst.Entries[2].Labels)
	expectedLine = "timestamp=\"2021-09-30 10:46:17.68 +0000 UTC\" kind=exception type=Error value=\"Cannot read property 'find' of undefined\" stacktrace=\"Error: Cannot read property 'find' of undefined\\n  at ? (http://fe:3002/static/js/vendors~main.chunk.ts:8639:42)\\n  at dispatchAction (http://fe:3002/static/js/vendors~main.chunk.ts:268095:9)\\n  at scheduleUpdateOnFiber (http://fe:3002/static/js/vendors~main.chunk.ts:273726:13)\\n  at flushSyncCallbackQueue (http://fe:3002/static/js/vendors~main.chunk.ts:263362:7)\\n  at flushSyncCallbackQueueImpl (http://fe:3002/static/js/vendors~main.chunk.ts:263374:13)\\n  at runWithPriority$1 (http://fe:3002/static/js/vendors~main.chunk.ts:263325:14)\\n  at unstable_runWithPriority (http://fe:3002/static/js/vendors~main.chunk.ts:291265:16)\\n  at ? (http://fe:3002/static/js/vendors~main.chunk.ts:263379:30)\\n  at performSyncWorkOnRoot (http://fe:3002/static/js/vendors~main.chunk.ts:274126:22)\\n  at renderRootSync (http://fe:3002/static/js/vendors~main.chunk.ts:274509:11)\\n  at workLoopSync (http://fe:3002/static/js/vendors~main.chunk.ts:274543:9)\\n  at performUnitOfWork (http://fe:3002/static/js/vendors~main.chunk.ts:274606:16)\\n  at beginWork$1 (http://fe:3002/static/js/vendors~main.chunk.ts:275746:18)\\n  at beginWork (http://fe:3002/static/js/vendors~main.chunk.ts:270944:20)\\n  at updateFunctionComponent (http://fe:3002/static/js/vendors~main.chunk.ts:269291:24)\\n  at renderWithHooks (http://fe:3002/static/js/vendors~main.chunk.ts:266969:22)\\n  at ? (http://fe:3002/static/js/main.chunk.ts:2600:74)\\n  at useGetBooksQuery (http://fe:3002/static/js/main.chunk.ts:1299:65)\\n  at Module.useQuery (http://fe:3002/static/js/vendors~main.chunk.ts:8495:85)\\n  at useBaseQuery (http://fe:3002/static/js/vendors~main.chunk.ts:8656:83)\\n  at useDeepMemo (http://fe:3002/static/js/vendors~main.chunk.ts:8696:14)\\n  at ? (http://fe:3002/static/js/vendors~main.chunk.ts:8657:55)\\n  at QueryData.execute (http://fe:3002/static/js/vendors~main.chunk.ts:7883:47)\\n  at QueryData.getExecuteResult (http://fe:3002/static/js/vendors~main.chunk.ts:7944:23)\\n  at QueryData._this.getQueryResult (http://fe:3002/static/js/vendors~main.chunk.ts:7790:19)\\n  at new ApolloError (http://fe:3002/static/js/vendors~main.chunk.ts:5164:24)\" sdk_name=grafana-frontend-agent sdk_version=1.0.0 app_name=testapp app_release=0.8.2 app_version=abcdefg app_environment=production user_attr_foo=bar session_id=abcd session_attr_time_elapsed=100s page_url=https://example.com/page browser_name=chrome browser_version=88.12.1 browser_os=linux browser_mobile=false"
	assert.Equal(t, expectedLine, inst.Entries[2].Line)

	// measurement
	assert.Equal(t, prommodel.LabelSet{
		prommodel.LabelName("app"):  prommodel.LabelValue("frontend"),
		prommodel.LabelName("kind"): prommodel.LabelValue("measurement"),
	}, inst.Entries[3].Labels)
	expectedLine = "timestamp=\"2021-09-30 10:46:17.68 +0000 UTC\" kind=measurement ttfb=14.000000 ttfcp=22.120000 ttfp=20.120000 sdk_name=grafana-frontend-agent sdk_version=1.0.0 app_name=testapp app_release=0.8.2 app_version=abcdefg app_environment=production user_attr_foo=bar session_id=abcd session_attr_time_elapsed=100s page_url=https://example.com/page browser_name=chrome browser_version=88.12.1 browser_os=linux browser_mobile=false"
	assert.Equal(t, expectedLine, inst.Entries[3].Line)
}
