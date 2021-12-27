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

const V2_MAP_FILE = `{
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

func TestHttpMapLoader(t *testing.T) {
	loader := NewHttpMapLoader()
	assert.NotNil(t, loader)
}

type MockDoType func(req *http.Request) (*http.Response, error)

type MockHttpClient struct {
	MockDo MockDoType
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func TestHttpMapLaderLoadSuccess(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte(MAP_FILE)))
	c := MockHttpClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       rb,
			}, nil
		},
	}
	loader := HttpMapLoader{c: &c}
	scm, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.Nil(t, err)
	assert.NotNil(t, scm)
}

func TestHttpMapLaderLoadNonOKStatus(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte("Not found")))
	c := MockHttpClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       rb,
			}, nil
		},
	}
	loader := HttpMapLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), http.StatusText(http.StatusNotFound))
}

func TestHttpMapLoaderLoadRequestFail(t *testing.T) {
	c := MockHttpClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("Error while executing request")
		},
	}
	loader := HttpMapLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error while executing request")
}

func TestHttpMapLoaderLoadV2Map(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte(V2_MAP_FILE)))
	c := MockHttpClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       rb,
			}, nil
		},
	}
	loader := HttpMapLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "version=2")
}

func TestHttpMapLoaderLoadIncrrectPayload(t *testing.T) {
	rb := ioutil.NopCloser(bytes.NewReader([]byte("This is not a map file")))
	c := MockHttpClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       rb,
			}, nil
		},
	}
	loader := HttpMapLoader{c: &c}
	_, err := loader.Load(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.NotNil(t, err)
}

func TestNewMapLoaderFS(t *testing.T) {
	loader, err := NewMapLoader(config.SourceMapConfig{MapURI: "test.js.map"})
	assert.Nil(t, err)
	assert.IsType(t, &FSMapLoader{}, loader)
}

func TestNewMapLoaderHttp(t *testing.T) {
	loader, err := NewMapLoader(config.SourceMapConfig{MapURI: "grafana.com/buckets/app/test.js.map"})
	assert.Nil(t, err)
	assert.IsType(t, &FSMapLoader{}, loader)
}
