package file

import (
	"path/filepath"

	"github.com/bmatcuk/doublestar"
	"github.com/grafana/agent/component/discovery"
)

// watch handles a single discovery.target for file watching.
type watch struct {
	target discovery.Target
}

func (w *watch) getPaths() ([]discovery.Target, error) {
	allMatchingPaths := make([]discovery.Target, 0)

	matches, err := doublestar.Glob(w.getPath())
	if err != nil {
		return nil, err
	}
	exclude := w.getExcludePath()
	for _, m := range matches {
		if exclude != "" {
			if match, _ := doublestar.PathMatch(exclude, m); match {
				continue
			}
		}
		abs, _ := filepath.Abs(m)
		dt := discovery.Target{}
		for dk, v := range w.target {
			dt[dk] = v
		}
		dt["__path__"] = abs
		allMatchingPaths = append(allMatchingPaths, dt)
	}

	return allMatchingPaths, nil
}

func (w *watch) getPath() string {
	return w.target["__path__"]
}

func (w *watch) getExcludePath() string {
	return w.target["__path_exclude__"]
}
