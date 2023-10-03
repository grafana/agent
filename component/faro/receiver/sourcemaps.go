package receiver

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-sourcemap/sourcemap"
	"github.com/grafana/agent/component/faro/receiver/internal/payload"
	"github.com/minio/pkg/wildcard"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vincent-petithory/dataurl"
)

// sourceMapsStore is an interface for a sourcemap service capable of
// transforming minified source locations to the original source location.
type sourceMapsStore interface {
	GetSourceMap(sourceURL string, release string) (*sourcemap.Consumer, error)
}

// Stub interfaces for easier mocking.
type (
	httpClient interface {
		Get(url string) (*http.Response, error)
	}

	fileService interface {
		Stat(name string) (fs.FileInfo, error)
		ReadFile(name string) ([]byte, error)
	}
)

type osFileService struct{}

func (fs osFileService) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (fs osFileService) ReadFile(name string) ([]byte, error)  { return os.ReadFile(name) }

type sourceMapMetrics struct {
	cacheSize *prometheus.CounterVec
	downloads *prometheus.CounterVec
	fileReads *prometheus.CounterVec
}

func newSourceMapMetrics(reg prometheus.Registerer) *sourceMapMetrics {
	m := &sourceMapMetrics{
		cacheSize: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "faro_receiver_sourcemap_cache_size",
			Help: "number of items in source map cache, per origin",
		}, []string{"origin"}),
		downloads: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "faro_receiver_sourcemap_downloads_total",
			Help: "downloads by the source map service",
		}, []string{"origin", "http_status"}),
		fileReads: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "faro_receiver_sourcemap_file_reads_total",
			Help: "source map file reads from file system, by origin and status",
		}, []string{"origin", "status"}),
	}

	reg.MustRegister(m.cacheSize, m.downloads, m.fileReads)
	return m
}

type sourcemapFileLocation struct {
	LocationArguments
	pathTemplate *template.Template
}

type sourceMapsStoreImpl struct {
	log     log.Logger
	cli     httpClient
	fs      fileService
	args    SourceMapsArguments
	metrics *sourceMapMetrics
	locs    []*sourcemapFileLocation

	cacheMut sync.Mutex
	cache    map[string]*sourcemap.Consumer
}

// newSourceMapStore creates an implementation of sourceMapsStore. The returned
// implementation is not dynamically updatable; create a new sourceMapsStore
// implementation if arguments change.
func newSourceMapsStore(log log.Logger, args SourceMapsArguments, metrics *sourceMapMetrics, cli httpClient, fs fileService) *sourceMapsStoreImpl {
	// TODO(rfratto): it would be nice for this to be dynamically updatable, but
	// that will require swapping out the http client (when the timeout changes)
	// or to find a way to inject a download timeout without modifying the http
	// client.

	if cli == nil {
		cli = &http.Client{Timeout: args.DownloadTimeout}
	}
	if fs == nil {
		fs = osFileService{}
	}

	locs := []*sourcemapFileLocation{}
	for _, loc := range args.Locations {
		tpl, err := template.New(loc.Path).Parse(loc.Path)
		if err != nil {
			panic(err) // TODO(rfratto): why is this set to panic?
		}

		locs = append(locs, &sourcemapFileLocation{
			LocationArguments: loc,
			pathTemplate:      tpl,
		})
	}

	return &sourceMapsStoreImpl{
		log:     log,
		cli:     cli,
		fs:      fs,
		args:    args,
		cache:   make(map[string]*sourcemap.Consumer),
		metrics: metrics,
		locs:    locs,
	}
}

func (store *sourceMapsStoreImpl) GetSourceMap(sourceURL string, release string) (*sourcemap.Consumer, error) {
	// TODO(rfratto): GetSourceMap is weak to transient errors, since it always
	// caches the result, even when there's an error. This means that transient
	// errors will be cached forever, preventing source maps from being retrieved.

	store.cacheMut.Lock()
	defer store.cacheMut.Unlock()

	cacheKey := fmt.Sprintf("%s__%s", sourceURL, release)
	if sm, ok := store.cache[cacheKey]; ok {
		return sm, nil
	}

	content, sourceMapURL, err := store.getSourceMapContent(sourceURL, release)
	if err != nil || content == nil {
		store.cache[cacheKey] = nil
		return nil, err
	}
	if content != nil {
		consumer, err := sourcemap.Parse(sourceMapURL, content)
		if err != nil {
			store.cache[cacheKey] = nil
			level.Debug(store.log).Log("msg", "failed to parse source map", "url", sourceMapURL, "release", release, "err", err)
			return nil, err
		}
		level.Info(store.log).Log("msg", "successfully parsed source map", "url", sourceMapURL, "release", release)
		store.cache[cacheKey] = consumer
		store.metrics.cacheSize.WithLabelValues(getOrigin(sourceURL)).Inc()
		return consumer, nil
	}

	return nil, nil
}

func (store *sourceMapsStoreImpl) getSourceMapContent(sourceURL string, release string) (content []byte, sourceMapURL string, err error) {
	// Attempt to find the source map in the filesystem first.
	for _, loc := range store.locs {
		content, sourceMapURL, err = store.getSourceMapFromFileSystem(sourceURL, release, loc)
		if content != nil || err != nil {
			return content, sourceMapURL, err
		}
	}

	// Attempt to download the sourcemap.
	//
	// TODO(rfratto): check if downloading is enabled.
	if strings.HasPrefix(sourceURL, "http") && urlMatchesOrigins(sourceURL, store.args.DownloadFromOrigins) {
		return store.downloadSourceMapContent(sourceURL)
	}
	return nil, "", nil
}

func (store *sourceMapsStoreImpl) getSourceMapFromFileSystem(sourceURL string, release string, loc *sourcemapFileLocation) (content []byte, sourceMapURL string, err error) {
	if len(sourceURL) == 0 || !strings.HasPrefix(sourceURL, loc.MinifiedPathPrefix) || strings.HasSuffix(sourceURL, "/") {
		return nil, "", nil
	}

	var rootPath bytes.Buffer

	err = loc.pathTemplate.Execute(&rootPath, struct{ Release string }{Release: cleanFilePathPart(release)})
	if err != nil {
		return nil, "", err
	}

	pathParts := []string{rootPath.String()}
	for _, part := range strings.Split(strings.TrimPrefix(strings.Split(sourceURL, "?")[0], loc.MinifiedPathPrefix), "/") {
		if len(part) > 0 && part != "." && part != ".." {
			pathParts = append(pathParts, part)
		}
	}
	mapFilePath := filepath.Join(pathParts...) + ".map"

	if _, err := store.fs.Stat(mapFilePath); err != nil {
		store.metrics.fileReads.WithLabelValues(getOrigin(sourceURL), "not_found").Inc()
		level.Debug(store.log).Log("msg", "source map not found on filesystem", "url", sourceURL, "file_path", mapFilePath)
		return nil, "", nil
	}
	level.Debug(store.log).Log("msg", "source map found on filesystem", "url", mapFilePath, "file_path", mapFilePath)

	content, err = store.fs.ReadFile(mapFilePath)
	if err != nil {
		store.metrics.fileReads.WithLabelValues(getOrigin(sourceURL), "error").Inc()
	} else {
		store.metrics.fileReads.WithLabelValues(getOrigin(sourceURL), "ok").Inc()
	}

	return content, sourceURL, err
}

func (store *sourceMapsStoreImpl) downloadSourceMapContent(sourceURL string) (content []byte, resolvedSourceMapURL string, err error) {
	level.Debug(store.log).Log("msg", "attempting to download source file", "url", sourceURL)

	result, err := store.downloadFileContents(sourceURL)
	if err != nil {
		level.Debug(store.log).Log("msg", "failed to download source file", "url", sourceURL, "err", err)
		return nil, "", err
	}

	match := reSourceMap.FindAllStringSubmatch(string(result), -1)
	if len(match) == 0 {
		level.Debug(store.log).Log("msg", "no source map url found in source", "url", sourceURL)
		return nil, "", nil
	}
	sourceMapURL := match[len(match)-1][2]

	// Inline sourcemap
	if strings.HasPrefix(sourceMapURL, "data:") {
		dataURL, err := dataurl.DecodeString(sourceMapURL)
		if err != nil {
			level.Debug(store.log).Log("msg", "failed to parse inline source map data url", "url", sourceURL, "err", err)
			return nil, "", err
		}

		level.Info(store.log).Log("msg", "successfully parsed inline source map data url", "url", sourceURL)
		return dataURL.Data, sourceURL + ".map", nil
	}
	// Remote sourcemap
	resolvedSourceMapURL = sourceMapURL

	// If the URL is relative, we need to attempt to resolve the absolute URL.
	if !strings.HasPrefix(resolvedSourceMapURL, "http") {
		base, err := url.Parse(sourceURL)
		if err != nil {
			level.Debug(store.log).Log("msg", "failed to parse source URL", "url", sourceURL, "err", err)
			return nil, "", err
		}
		relative, err := url.Parse(sourceMapURL)
		if err != nil {
			level.Debug(store.log).Log("msg", "failed to parse source map URL", "url", sourceURL, "sourceMapURL", sourceMapURL, "err", err)
			return nil, "", err
		}

		resolvedSourceMapURL = base.ResolveReference(relative).String()
		level.Debug(store.log).Log("msg", "resolved absolute source map URL", "url", sourceURL, "sourceMapURL", sourceMapURL)
	}

	level.Debug(store.log).Log("msg", "attempting to download source map file", "url", resolvedSourceMapURL)
	result, err = store.downloadFileContents(resolvedSourceMapURL)
	if err != nil {
		level.Debug(store.log).Log("msg", "failed to download source map file", "url", resolvedSourceMapURL, "err", err)
		return nil, "", err
	}

	return result, resolvedSourceMapURL, nil
}

func (store *sourceMapsStoreImpl) downloadFileContents(url string) ([]byte, error) {
	resp, err := store.cli.Get(url)
	if err != nil {
		store.metrics.downloads.WithLabelValues(getOrigin(url), "?").Inc()
		return nil, err
	}
	defer resp.Body.Close()

	store.metrics.downloads.WithLabelValues(getOrigin(url), fmt.Sprint(resp.StatusCode)).Inc()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

var reSourceMap = regexp.MustCompile("//[#@]\\s(source(?:Mapping)?URL)=\\s*(?P<url>\\S+)\r?\n?$")

func getOrigin(URL string) string {
	// TODO(rfratto): why are we parsing this every time? Let's parse it once.

	parsed, err := url.Parse(URL)
	if err != nil {
		return "?" // TODO(rfratto): should invalid URLs be permitted?
	}
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}

// urlMatchesOrigins returns true if URL matches at least one of origin prefix. Wildcard '*' and '?' supported
func urlMatchesOrigins(URL string, origins []string) bool {
	for _, origin := range origins {
		if origin == "*" || wildcard.Match(origin+"*", URL) {
			return true
		}
	}
	return false
}

func cleanFilePathPart(x string) string {
	return strings.TrimLeft(strings.ReplaceAll(strings.ReplaceAll(x, "\\", ""), "/", ""), ".")
}

func transformException(log log.Logger, store sourceMapsStore, ex *payload.Exception, release string) *payload.Exception {
	if ex.Stacktrace == nil {
		return ex
	}

	var frames []payload.Frame
	for _, frame := range ex.Stacktrace.Frames {
		mappedFrame, err := resolveSourceLocation(store, &frame, release)
		if err != nil {
			level.Error(log).Log("msg", "Error resolving stack trace frame source location", "err", err)
			frames = append(frames, frame)
		} else if mappedFrame != nil {
			frames = append(frames, *mappedFrame)
		} else {
			frames = append(frames, frame)
		}
	}

	return &payload.Exception{
		Type:       ex.Type,
		Value:      ex.Value,
		Stacktrace: &payload.Stacktrace{Frames: frames},
		Timestamp:  ex.Timestamp,
	}
}

func resolveSourceLocation(store sourceMapsStore, frame *payload.Frame, release string) (*payload.Frame, error) {
	smap, err := store.GetSourceMap(frame.Filename, release)
	if err != nil {
		return nil, err
	}
	if smap == nil {
		return nil, nil
	}

	file, function, line, col, ok := smap.Source(frame.Lineno, frame.Colno)
	if !ok {
		return nil, nil
	}
	// unfortunately in many cases go-sourcemap fails to determine the original function name.
	// not a big issue as long as file, line and column are correct
	if len(function) == 0 {
		function = "?"
	}
	return &payload.Frame{
		Filename: file,
		Lineno:   line,
		Colno:    col,
		Function: function,
	}, nil
}
