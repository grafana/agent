package frontendcollector

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-sourcemap/sourcemap"
	"github.com/vincent-petithory/dataurl"
)

var reSourceMap = "//[#@]\\s(source(?:Mapping)?URL)=\\s*(?P<url>\\S+)\n?$"

type sourceMap struct {
	consumer *sourcemap.Consumer
}

type SourceMapStore struct {
	sync.Mutex
	l               log.Logger
	cache           map[string]*sourceMap
	download        bool
	downloadTimeout time.Duration
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
		Timeout: store.downloadTimeout,
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

func (store *SourceMapStore) getSourceMapContent(sourceURL string) (content []byte, sourceMapURL string, err error) {
	if strings.HasPrefix(sourceURL, "http") && store.download {
		return store.downloadSourceMapContent(sourceURL)
	}
	return nil, "", nil
}

func (store *SourceMapStore) getSourceMap(sourceURL string) (*sourceMap, error) {
	store.Lock()
	defer store.Unlock()

	if smap, ok := store.cache[sourceURL]; ok {
		return smap, nil
	}
	content, sourceMapURL, err := store.downloadSourceMapContent(sourceURL)
	if err != nil || content == nil {
		store.cache[sourceURL] = nil
		return nil, err
	}
	if content != nil {
		consumer, err := sourcemap.Parse(sourceMapURL, content)
		if err != nil {
			store.cache[sourceURL] = nil
			level.Debug(store.l).Log("msg", "failed to parse sourcemap", "url", sourceMapURL, "err", err)
			return nil, err
		}
		level.Info(store.l).Log("msg", "successfully parsed sourcemap", "url", sourceMapURL)
		smap := &sourceMap{
			consumer: consumer,
		}
		store.cache[sourceURL] = smap
		return smap, nil
	}
	return nil, nil
}

func (store *SourceMapStore) resolveSourceLocation(frame sentry.Frame) (*sentry.Frame, error) {
	smap, err := store.getSourceMap(frame.Filename)
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
	return &sentry.Frame{
		Filename: file,
		Lineno:   line,
		Colno:    col,
		Function: function,
	}, nil
}
