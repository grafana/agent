package sourcemaps

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	responses []struct {
		*http.Response
		error
	}
	requests []string
}

type mockFileService struct {
	files map[string][]byte
	stats []string
	reads []string
}

func (s *mockFileService) Stat(name string) (fs.FileInfo, error) {
	s.stats = append(s.stats, name)
	_, ok := s.files[name]
	if !ok {
		return nil, errors.New("file not found")
	}
	return nil, nil
}

func (s *mockFileService) ReadFile(name string) ([]byte, error) {
	s.reads = append(s.reads, name)
	content, ok := s.files[name]
	if ok {
		return content, nil
	}
	return nil, errors.New("file not found")
}

func (cl *mockHTTPClient) Get(url string) (resp *http.Response, err error) {
	r := cl.responses[len(cl.requests)]
	cl.requests = append(cl.requests, url)
	return r.Response, r.error
}

func loadTestData(t *testing.T, file string) []byte {
	t.Helper()
	// Safe to disable, this is a test.
	// nolint:gosec
	content, err := ioutil.ReadFile(filepath.Join("testdata", file))
	require.NoError(t, err, "expected to be able to read file")
	require.True(t, len(content) > 0)
	return content
}

func newResponseFromTestData(t *testing.T, file string) *http.Response {
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader(loadTestData(t, file))),
		StatusCode: 200,
	}
}

func mockException() *models.Exception {
	return &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    6,
					Filename: "http://localhost:1234/foo.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    5,
					Filename: "http://localhost:1234/foo.js",
					Function: "callUndefined",
					Lineno:   6,
				},
			},
		},
	}
}

func Test_RealSourceMapStore_DownloadSuccess(t *testing.T) {
	conf := config.SourceMapConfig{
		Download:            true,
		DownloadFromOrigins: []string{"*"},
	}

	httpClient := &mockHTTPClient{
		responses: []struct {
			*http.Response
			error
		}{
			{newResponseFromTestData(t, "foo.js"), nil},
			{newResponseFromTestData(t, "foo.js.map"), nil},
		},
	}

	sourceMapStore := NewSourceMapStore(testLogger(t), conf, prometheus.NewRegistry(), httpClient, &mockFileService{})

	exception := mockException()

	transformed := sourceMapStore.TransformException(exception, "123")

	require.Equal(t, []string{"http://localhost:1234/foo.js", "http://localhost:1234/foo.js.map"}, httpClient.requests)

	expected := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    37,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   6,
				},
				{
					Colno:    2,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   7,
				},
			},
		},
	}

	require.Equal(t, *expected, *transformed)
}

func Test_RealSourceMapStore_DownloadError(t *testing.T) {
	conf := config.SourceMapConfig{
		Download:            true,
		DownloadFromOrigins: []string{"*"},
	}

	resp := &http.Response{
		StatusCode: 500,
	}

	httpClient := &mockHTTPClient{
		responses: []struct {
			*http.Response
			error
		}{
			{resp, nil},
		},
	}

	sourceMapStore := NewSourceMapStore(testLogger(t), conf, prometheus.NewRegistry(), httpClient, &mockFileService{})

	exception := mockException()

	transformed := sourceMapStore.TransformException(exception, "123")

	require.Equal(t, []string{"http://localhost:1234/foo.js"}, httpClient.requests)
	require.Equal(t, exception, transformed)
}

func Test_RealSourceMapStore_DownloadHTTPOriginFiltering(t *testing.T) {
	conf := config.SourceMapConfig{
		Download:            true,
		DownloadFromOrigins: []string{"http://bar.com/"},
	}

	httpClient := &mockHTTPClient{
		responses: []struct {
			*http.Response
			error
		}{
			{newResponseFromTestData(t, "foo.js"), nil},
			{newResponseFromTestData(t, "foo.js.map"), nil},
		},
	}

	sourceMapStore := NewSourceMapStore(testLogger(t), conf, prometheus.NewRegistry(), httpClient, &mockFileService{})

	exception := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    6,
					Filename: "http://foo.com/foo.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    5,
					Filename: "http://bar.com/foo.js",
					Function: "callUndefined",
					Lineno:   6,
				},
			},
		},
	}

	transformed := sourceMapStore.TransformException(exception, "123")

	require.Equal(t, []string{"http://bar.com/foo.js", "http://bar.com/foo.js.map"}, httpClient.requests)

	expected := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    6,
					Filename: "http://foo.com/foo.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    2,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   7,
				},
			},
		},
	}

	require.Equal(t, *expected, *transformed)
}

func Test_RealSourceMapStore_ReadFromFileSystem(t *testing.T) {
	conf := config.SourceMapConfig{
		Download: false,
		FileSystem: []config.SourceMapFileLocation{
			{
				MinifiedPathPrefix: "http://foo.com/",
				Path:               filepath.FromSlash("/var/build/latest/"),
			},
			{
				MinifiedPathPrefix: "http://bar.com/",
				Path:               filepath.FromSlash("/var/build/{RELEASE}/"),
			},
		},
	}

	mapFile := loadTestData(t, "foo.js.map")

	fileService := &mockFileService{
		files: map[string][]byte{
			filepath.FromSlash("/var/build/latest/foo.js.map"): mapFile,
			filepath.FromSlash("/var/build/123/foo.js.map"):    mapFile,
		},
	}

	sourceMapStore := NewSourceMapStore(log.NewNopLogger(), conf, prometheus.NewRegistry(), &mockHTTPClient{}, fileService)

	exception := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    6,
					Filename: "http://foo.com/foo.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    6,
					Filename: "http://foo.com/bar.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    5,
					Filename: "http://bar.com/foo.js",
					Function: "callUndefined",
					Lineno:   6,
				},
				{
					Colno:    5,
					Filename: "http://baz.com/foo.js",
					Function: "callUndefined",
					Lineno:   6,
				},
			},
		},
	}

	transformed := sourceMapStore.TransformException(exception, "123")

	require.Equal(t, []string{
		filepath.FromSlash("/var/build/latest/foo.js.map"),
		filepath.FromSlash("/var/build/latest/bar.js.map"),
		filepath.FromSlash("/var/build/123/foo.js.map"),
	}, fileService.stats)
	require.Equal(t, []string{
		filepath.FromSlash("/var/build/latest/foo.js.map"),
		filepath.FromSlash("/var/build/123/foo.js.map"),
	}, fileService.reads)

	expected := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    37,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   6,
				},
				{
					Colno:    6,
					Filename: "http://foo.com/bar.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    2,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   7,
				},
				{
					Colno:    5,
					Filename: "http://baz.com/foo.js",
					Function: "callUndefined",
					Lineno:   6,
				},
			},
		},
	}

	require.Equal(t, *expected, *transformed)
}

func Test_RealSourceMapStore_ReadFromFileSystemAndDownload(t *testing.T) {
	conf := config.SourceMapConfig{
		Download:            true,
		DownloadFromOrigins: []string{"*"},
		FileSystem: []config.SourceMapFileLocation{
			{
				MinifiedPathPrefix: "http://foo.com/",
				Path:               filepath.FromSlash("/var/build/latest/"),
			},
		},
	}

	mapFile := loadTestData(t, "foo.js.map")

	fileService := &mockFileService{
		files: map[string][]byte{
			filepath.FromSlash("/var/build/latest/foo.js.map"): mapFile,
		},
	}

	httpClient := &mockHTTPClient{
		responses: []struct {
			*http.Response
			error
		}{
			{newResponseFromTestData(t, "foo.js"), nil},
			{newResponseFromTestData(t, "foo.js.map"), nil},
		},
	}

	sourceMapStore := NewSourceMapStore(testLogger(t), conf, prometheus.NewRegistry(), httpClient, fileService)

	exception := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    6,
					Filename: "http://foo.com/foo.js",
					Function: "eval",
					Lineno:   5,
				},
				{
					Colno:    5,
					Filename: "http://bar.com/foo.js",
					Function: "callUndefined",
					Lineno:   6,
				},
			},
		},
	}

	transformed := sourceMapStore.TransformException(exception, "123")

	require.Equal(t, []string{filepath.FromSlash("/var/build/latest/foo.js.map")}, fileService.stats)
	require.Equal(t, []string{filepath.FromSlash("/var/builsd/latest/foo.js.map")}, fileService.reads)
	require.Equal(t, []string{"http://bar.com/foo.js", "http://bar.com/foo.js.map"}, httpClient.requests)

	expected := &models.Exception{
		Stacktrace: &models.Stacktrace{
			Frames: []models.Frame{
				{
					Colno:    37,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   6,
				},
				{
					Colno:    2,
					Filename: "/__parcel_source_root/demo/src/actions.ts",
					Function: "?",
					Lineno:   7,
				},
			},
		},
	}

	require.Equal(t, *expected, *transformed)
}

type testLogWriter struct {
	t *testing.T
}

func (w *testLogWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func testLogger(t *testing.T) log.Logger {
	return log.NewSyncLogger(log.NewLogfmtLogger(&testLogWriter{t}))
}
