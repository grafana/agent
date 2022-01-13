package sourcemaps

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
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

const V2MapFile = `{
  "version": 2,
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

func TestHTTPMapLoader(t *testing.T) {
	loader := NewHTTPMapLoader()
	assert.NotNil(t, loader)
}

type MockDoType func(req *http.Request) (*http.Response, error)

type MockHTTPClient struct {
	MockDo MockDoType
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func TestHTTPMapLaderLoadSuccess(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte(MapFile)))
	c := MockHTTPClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       rb,
			}, nil
		},
	}
	loader := HTTPMappLoader{c: &c}
	scm, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.Nil(t, err)
	assert.NotNil(t, scm)
}

func TestHTTPMapLaderLoadNonOKStatus(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte("Not found")))
	c := MockHTTPClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       rb,
			}, nil
		},
	}
	loader := HTTPMappLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), http.StatusText(http.StatusNotFound))
}

func TestHTTPMapLoaderLoadRequestFail(t *testing.T) {
	c := MockHTTPClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("Error while executing request")
		},
	}
	loader := HTTPMappLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error while executing request")
}

func TestHTTPMapLoaderLoadV2Map(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte(V2MapFile)))
	c := MockHTTPClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       rb,
			}, nil
		},
	}
	loader := HTTPMappLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "version=2")
}

func TestHTTPMapLoaderLoadIncrrectPayload(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte("This is not a map file")))
	c := MockHTTPClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       rb,
			}, nil
		},
	}
	loader := HTTPMappLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
}

func TestNewMapLoaderFS(t *testing.T) {
	loader, err := NewMapLoader(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.Nil(t, err)
	assert.IsType(t, &FSMapLoader{}, loader)
}

func TestNewMapLoaderHTTP(t *testing.T) {
	loader, err := NewMapLoader(config.SourceMapConfig{MapURI: "grafana.com/buckets/app/test.js.map"})
	assert.Nil(t, err)
	assert.IsType(t, &FSMapLoader{}, loader)
}
