package file

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/grafana/agent/component/discovery"

	"github.com/bmatcuk/doublestar"
	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.file",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the discovery.loki.path
// component.
type Arguments struct {
	Paths         []string      `river:"paths,attr"`
	ExcludedPaths []string      `river:"excluded_paths,attr,optional"`
	UpdatePeriod  time.Duration `river:"update_period,attr,optional"`
}

type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

var _ component.Component = (*Component)(nil)

// Component implements the discovery.loki.path component.
type Component struct {
	opts component.Options

	mut            sync.RWMutex
	args           Arguments
	watcher        *fsnotify.Watcher
	watchesUpdated bool
	watchedFiles   map[string]struct{}
}

// New creates a new loki.source.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:         o,
		mut:          sync.RWMutex{},
		args:         args,
		watchedFiles: make(map[string]struct{}),
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	c.watcher = watcher
	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = args.(Arguments)
	c.reconcileWatchesWithWatcher()
	return nil
}

func (c *Component) Run(ctx context.Context) error {
	watchDog := time.NewTicker(c.args.UpdatePeriod)
	reconcileLoop := time.NewTicker(c.args.UpdatePeriod * 2)
	timerDuration := c.args.UpdatePeriod
	keepGoing := true

	defer watchDog.Stop()
	defer reconcileLoop.Stop()
	for keepGoing {
		select {
		case fe := <-c.watcher.Events:
			c.fsnotifyTrigger(fe)
		case err := <-c.watcher.Errors:
			level.Error(c.opts.Logger).Log("msg", "error with fsnotify", "err", err)
		case <-watchDog.C:
			c.checkOnStateChanged()
			c.mut.Lock()
			// Check to see if our ticker timer needs to be reset.
			if timerDuration != c.args.UpdatePeriod {
				watchDog.Reset(c.args.UpdatePeriod)
				timerDuration = c.args.UpdatePeriod
			}
			c.mut.Unlock()
		case <-reconcileLoop.C:
			// There is a window between fsnotify watches being added and files being added.
			// This means every so often the system will manually true up the files and dir.
			func() {
				c.mut.Lock()
				defer c.mut.Unlock()
				c.reconcileWatchesWithWatcher()
			}()
			// Check to see if our ticker timer needs to be reset.
			if timerDuration != c.args.UpdatePeriod {
				reconcileLoop.Reset(c.args.UpdatePeriod * 2)
				timerDuration = c.args.UpdatePeriod
			}
		case <-ctx.Done():
			c.mut.Lock()
			err := c.watcher.Close()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error closing watcher", "err", err)
			}
			keepGoing = false
			c.mut.Unlock()
			break
		}
	}
	return nil
}

// reconcileWatchesWithWatcher checks for any new directories that have been added along with verifying
// that the args and watchers are in sync.
func (c *Component) reconcileWatchesWithWatcher() {
	c.watchesUpdated = true

	expandedPaths, err := getPaths(c.args.Paths)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "error expanding paths", "err", err)
		return
	}
	alreadyWatching := c.watcher.WatchList()
	alreadyWatchingDir := make(map[string]struct{})
	for _, p := range alreadyWatching {
		alreadyWatchingDir[p] = struct{}{}
	}
	// Ensure all the paths are added.
	for _, n := range expandedPaths {

		fi, fileErr := os.Stat(n)
		if fileErr != nil {
			level.Error(c.opts.Logger).Log("msg", "error getting os stats", "err", err)
			continue
		}
		if fi.IsDir() {
			// Check to see if we are already watching.
			if _, found := alreadyWatchingDir[n]; found {
				continue
			}
			err = c.watcher.Add(n)
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error adding path to watcher", "err", err)
			}
		} else {
			exclude := false
			for _, excluded := range c.args.ExcludedPaths {
				if match, _ := doublestar.Match(excluded, n); match {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
			c.addToWatchedFiles(n)
			dir := filepath.Dir(n)
			err = c.watcher.Add(dir)
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error adding path to watcher", "err", err)
			}
		}
	}
	// Find all the removed paths.
	pathsToRemove := make([]string, 0)
	for p := range c.watchedFiles {
		found := false
		for _, np := range expandedPaths {
			if p == np {
				found = true
				break
			}
		}
		if !found {
			pathsToRemove = append(pathsToRemove, p)
		}
	}
	for _, p := range pathsToRemove {
		cleaned := filepath.Dir(p)
		_ = c.watcher.Remove(cleaned)
		delete(c.watchedFiles, p)
	}
}

// checkOnStateChanged will see if onStateChanged needs to be called.
func (c *Component) checkOnStateChanged() {
	c.mut.Lock()
	defer c.mut.Unlock()

	if !c.watchesUpdated {
		return
	}
	c.watchesUpdated = false
	output := make([]discovery.Target, len(c.watchedFiles))
	i := 0
	for k := range c.watchedFiles {
		output[i] = discovery.Target{"__path__": k}
		i++
	}
	c.opts.OnStateChange(Exports{Targets: output})
}

func (c *Component) fsnotifyTrigger(fe fsnotify.Event) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.watchesUpdated = true
	if fe.Has(fsnotify.Create) {
		fi, _ := os.Stat(fe.Name)
		if fi.IsDir() {
			err := c.watcher.Add(fe.Name)
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error adding to watcher", "folder", fe.Name, "err", err)
			}
		} else {
			for _, p := range c.args.Paths {
				match, err := doublestar.Match(p, fe.Name)
				if err != nil {
					level.Error(c.opts.Logger).Log("msg", "error matching pattern", "err", err)
				}
				if match {
					c.watchedFiles[fe.Name] = struct{}{}
					break
				}
			}
		}
	} else if fe.Has(fsnotify.Remove) {
		for k := range c.watchedFiles {
			if k == fe.Name {
				delete(c.watchedFiles, k)
				break
			}
		}
	}
}

func (c *Component) addToWatchedFiles(fp string) {
	absFp, err := filepath.Abs(fp)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "error adding to watched files", "err", err)
	}
	if _, found := c.watchedFiles[absFp]; found {
		return
	}
	c.watchedFiles[absFp] = struct{}{}
}

func getPaths(paths []string) ([]string, error) {
	allMatchingPaths := make([]string, 0)
	for _, p := range paths {
		matches, err := doublestar.Glob(p)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			abs, _ := filepath.Abs(m)
			allMatchingPaths = append(allMatchingPaths, abs)
		}

	}

	return allMatchingPaths, nil
}
