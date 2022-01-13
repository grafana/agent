package models

import (
	"encoding/json"
	"testing"

	"github.com/go-sourcemap/sourcemap"
	"github.com/stretchr/testify/assert"
)

const MapFile = `{
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
	stacktracePayload := `
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
	err := json.Unmarshal([]byte(stacktracePayload), &stacktrace)
	assert.NoError(t, err)
	assert.Len(t, stacktrace.Frames, 2)

	scm, err := sourcemap.Parse("index.bundle.js", []byte(MapFile))
	assert.NoError(t, err)

	newTrace := stacktrace.MapFrames(scm)
	assert.Equal(t, "boom", newTrace.Frames[1].Function)
	assert.Equal(t, 5, newTrace.Frames[1].Lineno)
	assert.Equal(t, 0, newTrace.Frames[1].Colno)
}

func TestStackTraceMappingFallback(t *testing.T) {
	var stacktrace Stacktrace
	stacktracePayload := `
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
	err := json.Unmarshal([]byte(stacktracePayload), &stacktrace)
	assert.NoError(t, err)

	scm, err := sourcemap.Parse("index.bundle.js", []byte(MapFile))
	assert.NoError(t, err)

	newTrace := stacktrace.MapFrames(scm)
	assert.Equal(t, "fallback_name", newTrace.Frames[0].Function)
	assert.Equal(t, 5, newTrace.Frames[0].Lineno)
	assert.Equal(t, 30, newTrace.Frames[0].Colno)
}
