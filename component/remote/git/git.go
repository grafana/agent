// Package git implements the remote.git component.
package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/remote/git/internal/vcs"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/river/rivertypes"
)

func init() {
	component.Register(component.Registration{
		Name:    "remote.git",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments configures the remote.git component.
type Arguments struct {
	Repository    string            `river:"repository,attr"`
	Revision      string            `river:"revision,attr,optional"`
	Path          string            `river:"path,attr"`
	PullFrequency time.Duration     `river:"pull_frequency,attr,optional"`
	IsSecret      bool              `river:"is_secret,attr,optional"`
	GitAuthConfig vcs.GitAuthConfig `river:",squash"`
}

// Exports holds settings exported by remote.git.
type Exports struct {
	Content rivertypes.OptionalSecret `river:"content,attr"`
}

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	Revision:      "HEAD",
	PullFrequency: time.Minute,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args.PullFrequency < 0 {
		return fmt.Errorf("poll_frequency cannot be negative %s", args.PullFrequency)
	}
	return nil
}

// Component implements the remote.git component.
type Component struct {
	opts component.Options
	log  log.Logger

	mut         sync.RWMutex
	repo        *vcs.GitRepo
	repoOpts    vcs.GitRepoOptions
	args        Arguments
	lastExports Exports

	argsChanged chan struct{}

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ component.Component       = (*Component)(nil)
	_ component.HealthComponent = (*Component)(nil)
)

// New creates a new remote.git component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,
		log:  o.Logger,

		argsChanged: make(chan struct{}, 1),

		health: component.Health{
			Health:     component.HealthTypeUnknown,
			Message:    "component started",
			UpdateTime: time.Now(),
		},
	}
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		ticker  *time.Ticker
		tickerC <-chan time.Time
	)

	for {
		select {
		case <-ctx.Done():
			if ticker != nil {
				ticker.Stop()
			}
			return nil

		case <-c.argsChanged:
			c.mut.Lock()
			pullFrequency := c.args.PullFrequency
			c.mut.Unlock()
			ticker, tickerC = c.updateTicker(pullFrequency, ticker, tickerC)

		case <-tickerC:
			level.Info(c.log).Log("msg", "updating repository")
			c.tickPollFile(ctx)
		}
	}
}

func (c *Component) updateTicker(pullFrequency time.Duration, ticker *time.Ticker, tickerC <-chan time.Time) (*time.Ticker, <-chan time.Time) {
	level.Info(c.log).Log("msg", "updating repository pull frequency, next pull attempt will be done according to the pullFrequency", "new_frequency", pullFrequency)

	if pullFrequency > 0 {
		if ticker == nil {
			ticker = time.NewTicker(pullFrequency)
			tickerC = ticker.C
		} else {
			ticker.Reset(pullFrequency)
		}
		return ticker, tickerC
	}

	if ticker != nil {
		ticker.Stop()
	}
	return nil, nil
}

func (c *Component) tickPollFile(ctx context.Context) {
	c.mut.Lock()
	err := c.pollFile(ctx, c.args)
	pullFrequency := c.args.PullFrequency
	c.mut.Unlock()

	c.updateHealth(err)

	if err != nil {
		level.Error(c.log).Log("msg", "failed to update repository", "pullFrequency", pullFrequency, "err", err)
	}
}

func (c *Component) updateHealth(err error) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	if err != nil {
		c.health = component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    err.Error(),
			UpdateTime: time.Now(),
		}
	} else {
		c.health = component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "component updated",
			UpdateTime: time.Now(),
		}
	}
}

// Update implements component.Component.
// Only acknowledge the error from Update if it's not a
// vcs.UpdateFailedError; vcs.UpdateFailedError means that the Git repo
// exists, but we were unable to update it. It makes sense to retry on the next poll and it may succeed.
func (c *Component) Update(args component.Arguments) (err error) {
	defer func() {
		c.updateHealth(err)
	}()

	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)

	// TODO(rfratto): store in a repo-specific directory so changing repositories
	// doesn't risk break the module loader if there's a SHA collision between
	// the two different repositories. (in the context of the component being used to retrieve a module)
	repoPath := filepath.Join(c.opts.DataPath, "repo")

	repoOpts := vcs.GitRepoOptions{
		Repository: newArgs.Repository,
		Revision:   newArgs.Revision,
		Auth:       newArgs.GitAuthConfig,
	}

	// Create or update the repo field.
	// Failure to update repository makes the component temporarily use cached contents on disk
	if c.repo == nil || !reflect.DeepEqual(repoOpts, c.repoOpts) {
		r, err := vcs.NewGitRepo(context.Background(), repoPath, repoOpts)
		if err != nil {
			if errors.As(err, &vcs.UpdateFailedError{}) {
				level.Error(c.log).Log("msg", "failed to update repository", "err", err)
				c.updateHealth(err)
			} else {
				return err
			}
		}
		c.repo = r
		c.repoOpts = repoOpts
	}

	if err := c.pollFile(context.Background(), newArgs); err != nil {
		if errors.As(err, &vcs.UpdateFailedError{}) {
			level.Error(c.log).Log("msg", "failed to poll file from repository", "err", err)
			c.updateHealth(err)
		} else {
			return err
		}
	}

	// Schedule an update for handling the changed arguments.
	select {
	case c.argsChanged <- struct{}{}:
	default:
	}

	c.args = newArgs
	return nil
}

// pollFile fetches the latest content from the repository and updates the
// controller. pollFile must only be called with c.mut held.
func (c *Component) pollFile(ctx context.Context, args Arguments) error {
	// Make sure our repo is up-to-date.
	if err := c.repo.Update(ctx); err != nil {
		return err
	}

	// Finally, configure our controller.
	bb, err := c.repo.ReadFile(args.Path)
	if err != nil {
		return err
	}

	newExports := Exports{
		Content: rivertypes.OptionalSecret{
			IsSecret: args.IsSecret,
			Value:    strings.TrimSpace(string(bb)),
		},
	}

	if c.lastExports != newExports {
		c.opts.OnStateChange(newExports)
	}
	c.lastExports = newExports
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (c *Component) CurrentHealth() component.Health {
	c.healthMut.RLock()
	defer c.healthMut.RUnlock()

	return c.health
}

// DebugInfo implements component.DebugComponent.
func (c *Component) DebugInfo() interface{} {
	type DebugInfo struct {
		SHA       string `river:"sha,attr"`
		RepoError string `river:"repo_error,attr,optional"`
	}

	c.mut.RLock()
	defer c.mut.RUnlock()

	rev, err := c.repo.CurrentRevision()
	if err != nil {
		return DebugInfo{RepoError: err.Error()}
	} else {
		return DebugInfo{SHA: rev}
	}
}
