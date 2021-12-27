package models

import (
	"encoding/json"
	"testing"

	"github.com/go-sourcemap/sourcemap"
	"github.com/stretchr/testify/assert"
)

const MAP_FILE = `{
  "version": 3,
  "file": "index.bundle.js",
  "mappings": "CAAA,WACE,MAAM,IAAIA,MAAM,SAGlBC",
  "sources": [
    "webpack://jsmap_test/./index.js"
  ],
  "sourcesContent": [
    "function boom() {\n  throw new Error('Error');\n}\n\nboom();\n"
  ],
  "names": [
    "Error",
    "boom"
  ],
  "sourceRoot": ""
}`

func TestStackTraceMapping(t *testing.T) {
	var stacktrace Stacktrace
	stacktrace_payload := `
{
	"frames": [
		{
			"function": "",
			"filename": "index.bundle.js",
			"colno": 18,
			"lineno": 1,
			"in_app": true
		},
		{
			"function": "Object.<anonymous>",
			"filename": "index.min.js",
			"colno": 37,
			"lineno": 1,
			"in_app": true
		}
	]
}
`
	err := json.Unmarshal([]byte(stacktrace_payload), &stacktrace)
	assert.NoError(t, err)
	assert.Len(t, stacktrace.Frames, 2)

	scm, err := sourcemap.Parse("index.bundle.js", []byte(MAP_FILE))
	assert.NoError(t, err)

	new_trace := stacktrace.MapFrames(scm)
	assert.Equal(t, "boom", new_trace.Frames[1].Function)
	assert.Equal(t, 5, new_trace.Frames[1].Lineno)
	assert.Equal(t, 0, new_trace.Frames[1].Colno)
}

func TestStackTraceMappingFallback(t *testing.T) {
	var stacktrace Stacktrace
	stacktrace_payload := `
{
	"frames": [
		{
			"function": "fallback_name",
			"filename": "index.bundle.js",
			"colno": 30,
			"lineno": 5,
			"in_app": true
		}
	]
}
`
	json.Unmarshal([]byte(stacktrace_payload), &stacktrace)

	scm, err := sourcemap.Parse("index.bundle.js", []byte(MAP_FILE))
	assert.NoError(t, err)

	new_trace := stacktrace.MapFrames(scm)
	assert.Equal(t, "fallback_name", new_trace.Frames[0].Function)
	assert.Equal(t, 5, new_trace.Frames[0].Lineno)
	assert.Equal(t, 30, new_trace.Frames[0].Colno)
}
