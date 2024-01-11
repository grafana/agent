package importsource

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/logging/level"
	vcs "github.com/grafana/agent/pkg/util/git"
	"github.com/grafana/river/vm"
)

// The difference between this import source and the others is that there is no git component.
// The git logic in the internal package is a copy of the one used in the old module.
type ImportGit struct {
	opts            component.Options
	log             log.Logger
	eval            *vm.Evaluator
	mut             sync.RWMutex
	repo            *vcs.GitRepo
	repoOpts        vcs.GitRepoOptions
	args            Arguments
	onContentChange func(string)

	argsChanged chan struct{}

	healthMut sync.RWMutex
	health    component.Health
}

var (
	_ ImportSource              = (*ImportGit)(nil)
	_ component.Component       = (*ImportGit)(nil)
	_ component.HealthComponent = (*ImportGit)(nil)
)

type Arguments struct {
	Repository    string            `river:"repository,attr"`
	Revision      string            `river:"revision,attr,optional"`
	Path          string            `river:"path,attr"`
	PullFrequency time.Duration     `river:"pull_frequency,attr,optional"`
	GitAuthConfig vcs.GitAuthConfig `river:",squash"`
}

var DefaultArguments = Arguments{
	Revision:      "HEAD",
	PullFrequency: time.Minute,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

func NewImportGit(managedOpts component.Options, eval *vm.Evaluator, onContentChange func(string)) *ImportGit {
	return &ImportGit{
		opts:            managedOpts,
		log:             managedOpts.Logger,
		eval:            eval,
		argsChanged:     make(chan struct{}, 1),
		onContentChange: onContentChange,
	}
}

func (im *ImportGit) Evaluate(scope *vm.Scope) error {
	var arguments Arguments
	if err := im.eval.Evaluate(scope, &arguments); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	if reflect.DeepEqual(im.args, arguments) {
		return nil
	}

	if err := im.Update(arguments); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}
	return nil
}

func (im *ImportGit) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		ticker  *time.Ticker
		tickerC <-chan time.Time
	)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-im.argsChanged:
			im.mut.Lock()
			{
				level.Info(im.log).Log("msg", "updating repository pull frequency", "new_frequency", im.args.PullFrequency)

				if im.args.PullFrequency > 0 {
					if ticker == nil {
						ticker = time.NewTicker(im.args.PullFrequency)
						tickerC = ticker.C
					} else {
						ticker.Reset(im.args.PullFrequency)
					}
				} else {
					if ticker != nil {
						ticker.Stop()
					}
					ticker = nil
					tickerC = nil
				}
			}
			im.mut.Unlock()

		case <-tickerC:
			level.Info(im.log).Log("msg", "updating repository", "new_frequency", im.args.PullFrequency)
			im.tickPollFile(ctx)
		}
	}
}

func (im *ImportGit) tickPollFile(ctx context.Context) {
	im.mut.Lock()
	err := im.pollFile(ctx, im.args)
	im.mut.Unlock()

	im.updateHealth(err)
}

func (im *ImportGit) updateHealth(err error) {
	im.healthMut.Lock()
	defer im.healthMut.Unlock()

	if err != nil {
		im.health = component.Health{
			Health:     component.HealthTypeUnhealthy,
			Message:    err.Error(),
			UpdateTime: time.Now(),
		}
	} else {
		im.health = component.Health{
			Health:     component.HealthTypeHealthy,
			Message:    "module updated",
			UpdateTime: time.Now(),
		}
	}
}

// Update implements component.Component.
func (im *ImportGit) Update(args component.Arguments) (err error) {
	defer func() {
		im.updateHealth(err)
	}()

	im.mut.Lock()
	defer im.mut.Unlock()

	newArgs := args.(Arguments)

	// TODO(rfratto): store in a repo-specific directory so changing repositories
	// doesn't risk break the module loader if there's a SHA collision between
	// the two different repositories.
	repoPath := filepath.Join(im.opts.DataPath, "repo")

	repoOpts := vcs.GitRepoOptions{
		Repository: newArgs.Repository,
		Revision:   newArgs.Revision,
		Auth:       newArgs.GitAuthConfig,
	}

	// Create or update the repo field.
	// Failure to update repository makes the module loader temporarily use cached contents on disk
	if im.repo == nil || !reflect.DeepEqual(repoOpts, im.repoOpts) {
		r, err := vcs.NewGitRepo(context.Background(), repoPath, repoOpts)
		if err != nil {
			if errors.As(err, &vcs.UpdateFailedError{}) {
				level.Error(im.log).Log("msg", "failed to update repository", "err", err)
				im.updateHealth(err)
			} else {
				return err
			}
		}
		im.repo = r
		im.repoOpts = repoOpts
	}

	if err := im.pollFile(context.Background(), newArgs); err != nil {
		return err
	}

	// Schedule an update for handling the changed arguments.
	select {
	case im.argsChanged <- struct{}{}:
	default:
	}

	im.args = newArgs
	return nil
}

// pollFile fetches the latest content from the repository and updates the
// controller. pollFile must only be called with im.mut held.
func (im *ImportGit) pollFile(ctx context.Context, args Arguments) error {
	// Make sure our repo is up-to-date.
	if err := im.repo.Update(ctx); err != nil {
		return err
	}

	// Finally, configure our controller.
	bb, err := im.repo.ReadFile(args.Path)
	if err != nil {
		return err
	}
	im.onContentChange(string(bb))
	return nil
}

// CurrentHealth implements component.HealthComponent.
func (im *ImportGit) CurrentHealth() component.Health {
	im.healthMut.RLock()
	defer im.healthMut.RUnlock()
	return im.health
}

// DebugInfo implements component.DebugComponent.
func (im *ImportGit) DebugInfo() interface{} {
	type DebugInfo struct {
		SHA       string `river:"sha,attr"`
		RepoError string `river:"repo_error,attr,optional"`
	}

	im.mut.RLock()
	defer im.mut.RUnlock()

	rev, err := im.repo.CurrentRevision()
	if err != nil {
		return DebugInfo{RepoError: err.Error()}
	} else {
		return DebugInfo{SHA: rev}
	}
}

func (im *ImportGit) Arguments() component.Arguments {
	return im.args
}

func (im *ImportGit) Component() component.Component {
	return im
}
