package sourcemaps

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/go-sourcemap/sourcemap"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
)

// Basic Interface to load source maps from either a file or an HTTP URL
// this can be extended in the future to load from a source file using the
// inline source map.
type MapLoader interface {
	Load(config.SourceMapConfig) (*sourcemap.Consumer, error)
}

// Used for mocking
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpMapLoader struct {
	c HTTPClient
}

func NewHttpMapLoader() *HttpMapLoader {
	c := &http.Client{}
	return &HttpMapLoader{c: c}
}

func loadFromReader(ior io.Reader, url string) (scm *sourcemap.Consumer, err error) {
	mapData, err := ioutil.ReadAll(ior)
	if err != nil {
		return nil, err
	}

	scm, err = sourcemap.Parse(url, mapData)
	if err != nil {
		return nil, err
	}

	fmt.Println("Sourcemap Consumer created")
	return scm, nil
}

func (hl *HttpMapLoader) Load(conf config.SourceMapConfig) (scm *sourcemap.Consumer, err error) {
	req, err := http.NewRequest(http.MethodGet, conf.MapURI, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hl.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(http.StatusText(resp.StatusCode))
	}

	return loadFromReader(resp.Body, conf.MapURI)
}

type FSMapLoader struct{}

func (fl *FSMapLoader) Load(conf config.SourceMapConfig) (scm *sourcemap.Consumer, err error) {
	f, err := os.Open(conf.MapURI)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	ior := bufio.NewReader(f)

	return loadFromReader(ior, conf.MapURI)
}

func NewMapLoader(conf config.SourceMapConfig) (MapLoader, error) {
	u, err := url.Parse(conf.MapURI)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" || u.Host == "" {
		fmt.Println("Loading source map file from file system")
		return &FSMapLoader{}, nil
	} else {
		fmt.Println("Loading source map external source")
		return NewHttpMapLoader(), nil
	}
}
