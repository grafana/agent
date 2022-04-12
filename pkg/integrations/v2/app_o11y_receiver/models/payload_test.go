package models

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func loadTestData(t *testing.T, file string) []byte {
	t.Helper()
	// Safe to disable, this is a test.
	// nolint:gosec
	content, err := ioutil.ReadFile(filepath.Join("testdata", file))
	assert.NoError(t, err, "expected to be able to read file")
	assert.True(t, len(content) > 0)
	return content
}

func TestUnmarshalPayloadJSON(t *testing.T) {
	content := loadTestData(t, "payload.json")
	var payload Payload
	err := json.Unmarshal(content, &payload)
	assert.NoError(t, err)

	now, err := time.Parse("2006-01-02T15:04:05Z0700", "2021-09-30T10:46:17.680Z")
	assert.NoError(t, err)

	assert.Equal(t, Meta{
		SDK: SDK{
			Name:    "grafana-frontend-agent",
			Version: "1.0.0",
		},
		App: App{
			Name:        "testapp",
			Release:     "0.8.2",
			Version:     "abcdefg",
			Environment: "production",
		},
		User: User{
			Username:   "domasx2",
			Attributes: map[string]string{"foo": "bar"},
		},
		Session: Session{
			ID:         "abcd",
			Attributes: map[string]string{"time_elapsed": "100s"},
		},
		Page: Page{
			URL: "https://example.com/page",
		},
		Browser: Browser{
			Name:    "chrome",
			Version: "88.12.1",
			OS:      "linux",
			Mobile:  false,
		},
	}, payload.Meta)

	assert.Len(t, payload.Exceptions, 1)
	assert.Len(t, payload.Exceptions[0].Stacktrace.Frames, 26)
	assert.Equal(t, "Error", payload.Exceptions[0].Type)
	assert.Equal(t, "Cannot read property 'find' of undefined", payload.Exceptions[0].Value)

	assert.Equal(t, []Log{
		{
			Message:  "opened pricing page",
			LogLevel: LogLevelInfo,
			Context: map[string]string{
				"component": "AppRoot",
				"page":      "Pricing",
			},
			Timestamp: now,
			Trace: TraceContext{
				TraceID: "abcd",
				SpanID:  "def",
			},
		},
		{
			Message:  "loading price list",
			LogLevel: LogLevelTrace,
			Context: map[string]string{
				"component": "AppRoot",
				"page":      "Pricing",
			},
			Timestamp: now,
			Trace: TraceContext{
				TraceID: "abcd",
				SpanID:  "ghj",
			},
		},
	}, payload.Logs)
}
