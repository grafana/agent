package file

import (
	"path/filepath"

	"github.com/bmatcuk/doublestar"
	"github.com/grafana/agent/component/discovery"
)

// watch handles a single discovery.target for file watching.
type watch struct {
	targets discovery.Target
}

func (w *watch) getPaths() ([]discovery.Target, error) {
	allMatchingPaths := make([]discovery.Target, 0)

	matches, err := doublestar.Glob(w.getPath())
	if err != nil {
		return nil, err
	}
	for _, m := range matches {
		exclude := w.getExcludePath()
		if exclude != "" {
			if match, _ := doublestar.PathMatch(m, exclude); match {
				continue
			}
		}
		abs, _ := filepath.Abs(m)
		dt := discovery.Target{}
		for dk, v := range w.targets {
			dt[dk] = v
		}
		dt["__path__"] = abs
		allMatchingPaths = append(allMatchingPaths, dt)
	}

	return allMatchingPaths, nil
}

func (w *watch) getPath() string {
	return w.targets["__path__"]
}

func (w *watch) getExcludePath() string {
	return w.targets["__path_exclude__"]
}
