package sourcemaps

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-sourcemap/sourcemap"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
	"github.com/vincent-petithory/dataurl"
)

var reSourceMap = "//[#@]\\s(source(?:Mapping)?URL)=\\s*(?P<url>\\S+)\n?$"

type sourceMap struct {
	consumer *sourcemap.Consumer
}

type SourceMapStore struct {
	sync.Mutex
	l      log.Logger
	config config.SourceMapConfig
	cache  map[string]*sourceMap
}

func NewSourceMapStore(l log.Logger, config config.SourceMapConfig) *SourceMapStore {
	return &SourceMapStore{
		l:      l,
		config: config,
		cache:  make(map[string]*sourceMap),
	}
}

func downloadFileContents(client http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (store *SourceMapStore) downloadSourceMapContent(sourceURL string) (content []byte, resolvedSourceMapURL string, err error) {
	level.Debug(store.l).Log("msg", "attempting to download source file", "url", sourceURL)
	client := http.Client{
		Timeout: store.config.DownloadTimeout,
	}
	result, err := downloadFileContents(client, sourceURL)
	if err != nil {
		level.Debug(store.l).Log("failed to download source file", "url", sourceURL, "err", err)
		return nil, "", err
	}
	r := regexp.MustCompile(reSourceMap)
	match := r.FindAllStringSubmatch(string(result), -1)
	if match == nil || len(match) == 0 {
		level.Debug(store.l).Log("msg", "no sourcemap url found in source", "url", sourceURL)
		return nil, "", nil
	}
	sourceMapURL := match[len(match)-1][2]

	// inline sourcemap
	if strings.HasPrefix(sourceMapURL, "data:") {
		dataURL, err := dataurl.DecodeString(sourceMapURL)
		if err != nil {
			level.Debug(store.l).Log("msg", "failed to parse inline sourcemap data url", "url", sourceURL, "err", err)
			return nil, "", err
		}

		level.Info(store.l).Log("msg", "successfully parsed inline sourcemap data url", "url", sourceURL)
		return dataURL.Data, sourceURL + ".map", nil
	}
	// remote sourcemap
	resolvedSourceMapURL = sourceMapURL

	// if url is relative, attempt to resolve absolute
	if !strings.HasPrefix(resolvedSourceMapURL, "http") {
		base, err := url.Parse(sourceURL)
		if err != nil {
			level.Debug(store.l).Log("msg", "failed to parse source url", "url", sourceURL, "err", err)
			return nil, "", err
		}
		relative, err := url.Parse(sourceMapURL)
		if err != nil {
			level.Debug(store.l).Log("msg", "failed to parse source map url", "url", sourceURL, "sourceMapURL", sourceMapURL, "err", err)
			return nil, "", err
		}
		resolvedSourceMapURL = base.ResolveReference(relative).String()
		level.Debug(store.l).Log("msg", "resolved absolute soure map url", "url", sourceURL, "sourceMapURL", resolvedSourceMapURL)
	}
	level.Debug(store.l).Log("msg", "attempting to download sourcemap file", "url", resolvedSourceMapURL)
	result, err = downloadFileContents(client, resolvedSourceMapURL)
	if err != nil {
		level.Debug(store.l).Log("failed to download source map file", "url", resolvedSourceMapURL, "err", err)
		return nil, "", err
	}
	return result, resolvedSourceMapURL, nil
}

func (store *SourceMapStore) getSourceMapFromFileSystem(sourceURL string, release string, fileconf config.SourceMapFileLocation) (content []byte, sourceMapURL string, err error) {
	if len(sourceURL) == 0 || !strings.HasPrefix(sourceURL, fileconf.MinifiedPathPrefix) || strings.HasSuffix(sourceURL, "/") {
		return nil, "", nil
	}
	pathParts := []string{strings.Replace(fileconf.Path, "{RELEASE}", cleanForFilePath(release), 1)}
	for _, part := range strings.Split(strings.TrimPrefix(sourceURL, fileconf.MinifiedPathPrefix), "/") {
		if len(part) > 0 && part != "." && part != ".." {
			pathParts = append(pathParts, part)
		}
	}
	mapFilePath := filepath.Join(pathParts...) + ".map"

	if _, err := os.Stat(mapFilePath); err != nil {
		level.Debug(store.l).Log("msg", "sourcemap not found on filesystem", "url", sourceURL, "file_path", mapFilePath)
		return nil, "", nil
	}
	level.Debug(store.l).Log("msg", "sourcemap found on filesystem", "url", mapFilePath, "file_path", mapFilePath)

	content, err = os.ReadFile(mapFilePath)
	return content, sourceURL, err
}

func (store *SourceMapStore) getSourceMapContent(sourceURL string, release string) (content []byte, sourceMapURL string, err error) {

	//attempt to find in fs
	for _, fileconf := range store.config.FileSystem {
		content, sourceMapURL, err = store.getSourceMapFromFileSystem(sourceURL, release, fileconf)
		if content != nil || err != nil {
			return content, sourceMapURL, err
		}
	}

	//attempt to download
	if strings.HasPrefix(sourceURL, "http") && utils.URLMatchesOrigins(sourceMapURL, store.config.DownloadFromOrigins) {
		return store.downloadSourceMapContent(sourceURL)
	}
	return nil, "", nil
}

func (store *SourceMapStore) getSourceMap(sourceURL string, release string) (*sourceMap, error) {
	store.Lock()
	defer store.Unlock()

	cacheKey := fmt.Sprintf("%s__%s", sourceURL, release)

	if smap, ok := store.cache[cacheKey]; ok {
		return smap, nil
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
			level.Debug(store.l).Log("msg", "failed to parse sourcemap", "url", sourceMapURL, "release", release, "err", err)
			return nil, err
		}
		level.Info(store.l).Log("msg", "successfully parsed sourcemap", "url", sourceMapURL, "release", release)
		smap := &sourceMap{
			consumer: consumer,
		}
		store.cache[cacheKey] = smap
		return smap, nil
	}
	return nil, nil
}

func (store *SourceMapStore) ResolveSourceLocation(frame models.Frame, release string) (*models.Frame, error) {
	smap, err := store.getSourceMap(frame.Filename, release)
	if err != nil {
		return nil, err
	}
	if smap == nil {
		return nil, nil
	}

	file, function, line, col, ok := smap.consumer.Source(frame.Lineno, frame.Colno)
	if !ok {
		return nil, nil
	}
	// unfortunately in many cases go-sourcemap fails to determine the original function name.
	// not a big issue as long as file, line and column are correct
	if len(function) == 0 {
		function = "?"
	}
	return &models.Frame{
		Filename: file,
		Lineno:   line,
		Colno:    col,
		Function: function,
	}, nil
}

func (store *SourceMapStore) TransformException(ex *models.Exception, release string) *models.Exception {
	if ex.Stacktrace == nil {
		return ex
	}
	frames := []models.Frame{}

	for _, frame := range ex.Stacktrace.Frames {
		mappedFrame, err := store.ResolveSourceLocation(frame, release)
		if err != nil {
			level.Error(store.l).Log("msg", "Error resolving stack trace frame source location", "err", err)
			frames = append(frames, frame)
		} else {
			if mappedFrame != nil {
				frames = append(frames, *mappedFrame)
			} else {
				frames = append(frames, frame)
			}
		}
	}

	return &models.Exception{
		Type:       ex.Type,
		Value:      ex.Value,
		Stacktrace: &models.Stacktrace{Frames: frames},
		Timestamp:  ex.Timestamp,
	}
}

func cleanForFilePath(x string) string {
	return strings.TrimLeft(strings.ReplaceAll(strings.ReplaceAll(x, "\\", ""), "/", ""), ".")
}
