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

// Arguments holds values which are used to configure the discovery.file
// component.
type Arguments struct {
	Paths         []discovery.Target `river:"paths,attr"`
	ExcludedPaths []discovery.Target `river:"excluded_paths,attr,optional"`
	UpdatePeriod  time.Duration      `river:"update_period,attr,optional"`
}

func (a *Arguments) getPaths() []string {
	paths := make([]string, 0)
	index := 0
	for _, v := range a.Paths {
		val, found := v["__path__"]
		if !found {
			continue
		}
		paths = append(paths, val)
		index++
	}
	return paths
}

func (a *Arguments) getExcluded() []string {
	paths := make([]string, 0)
	index := 0
	for _, v := range a.Paths {
		val, found := v["__path_exclude__"]
		if !found {
			continue
		}
		paths = append(paths, val)
		index++
	}
	return paths
}

// Exports exposes targets.
type Exports struct {
	Targets []discovery.Target `river:"targets,attr"`
}

var _ component.Component = (*Component)(nil)

// Component implements the discovery.file component.
type Component struct {
	opts component.Options

	mut            sync.RWMutex
	args           Arguments
	watcher        *fsnotify.Watcher
	watchesUpdated bool
	watchedFiles   map[string]struct{}
}

// New creates a new discovery.file component.
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

func getDefault() Arguments {
	return Arguments{UpdatePeriod: 10 * time.Second}
}

// UnmarshalRiver implements river.Unmarshaler.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = getDefault()
	type arguments Arguments
	return f((*arguments)(a))
}

// Update satisfies the component interface.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.args = args.(Arguments)
	c.reconcileWatchesWithWatcher()
	return nil
}

// Run satisfies the component interface.
func (c *Component) Run(ctx context.Context) error {
	watchDog := time.NewTicker(c.args.UpdatePeriod)
	timerDuration := c.args.UpdatePeriod
	update := func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		// See if there is anything new we need to check.
		c.reconcileWatchesWithWatcher()
		// Update the exports with the targets. Should only be called if changes occurred.
		c.checkOnStateChanged()
		// Check to see if our ticker timer needs to be reset.
		if timerDuration != c.args.UpdatePeriod {
			watchDog.Reset(c.args.UpdatePeriod)
			timerDuration = c.args.UpdatePeriod
		}
	}
	// Trigger initial check
	update()
	defer watchDog.Stop()
	for {
		select {
		case fe := <-c.watcher.Events:
			c.fsnotifyTrigger(fe)
		case err := <-c.watcher.Errors:
			level.Error(c.opts.Logger).Log("msg", "error with fsnotify", "err", err)
		case <-watchDog.C:
			// This triggers a check for any new paths, along with pushing new targets.
			update()
		case <-ctx.Done():
			c.mut.Lock()
			err := c.watcher.Close()
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error closing watcher", "err", err)
			}
			c.mut.Unlock()
			return nil
		}
	}
}

// reconcileWatchesWithWatcher checks for any new directories that have been added along with verifying
// that the args and watchers are in sync.
func (c *Component) reconcileWatchesWithWatcher() {
	includedPaths := c.args.getPaths()
	excludedPaths := c.args.getExcluded()
	expandedPaths, err := getPaths(includedPaths)
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
			for _, excluded := range excludedPaths {
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
	filesToRemove := make([]string, 0)
	for p := range c.watchedFiles {
		found := false
		for _, np := range expandedPaths {
			if p == np {
				found = true
				break
			}
		}

		if !found {
			filesToRemove = append(filesToRemove, p)
		}
		// Scan to see if we need to exclude any new files.
		for _, exclude := range excludedPaths {
			matched, _ := doublestar.PathMatch(exclude, p)
			if matched {
				filesToRemove = append(filesToRemove, p)
			}
		}
	}

	if len(filesToRemove) > 0 {
		c.watchesUpdated = true
	}
	for _, p := range filesToRemove {
		cleaned := filepath.Dir(p)
		// Check to see if there are paths we no longer need to watch.
		for _, exclude := range excludedPaths {
			excludeDir, _ := doublestar.PathMatch(exclude, cleaned)
			if excludeDir {
				_ = c.watcher.Remove(cleaned)
			}
		}
		delete(c.watchedFiles, p)
	}
}

// checkOnStateChanged will see if onStateChanged needs to be called.
func (c *Component) checkOnStateChanged() {
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

	if fe.Has(fsnotify.Create) {
		fi, _ := os.Stat(fe.Name)
		if fi.IsDir() {
			err := c.watcher.Add(fe.Name)
			if err != nil {
				level.Error(c.opts.Logger).Log("msg", "error adding to watcher", "folder", fe.Name, "err", err)
			}
		} else {
			for _, p := range c.args.getPaths() {
				match, err := doublestar.Match(p, fe.Name)
				if err != nil {
					level.Error(c.opts.Logger).Log("msg", "error matching pattern", "err", err)
				}
				if match {
					c.addToWatchedFiles(p)
					break
				}
			}
		}
	} else if fe.Has(fsnotify.Remove) {
		for k := range c.watchedFiles {
			if k == fe.Name {
				c.removeFromWatched(k)
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
	c.watchesUpdated = true
}

func (c *Component) removeFromWatched(fp string) {
	abs, _ := filepath.Abs(fp)
	delete(c.watchedFiles, abs)
	c.watchesUpdated = true
}

func (c *Component) getWatchedFiles() []discovery.Target {
	c.mut.Lock()
	defer c.mut.Unlock()

	foundFiles := make([]discovery.Target, 0)
	for k := range c.watchedFiles {
		foundFiles = append(foundFiles, discovery.Target{"__path__": k})
	}
	return foundFiles
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
