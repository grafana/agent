package docker

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/positions"
	dt "github.com/grafana/agent/component/loki/source/docker/internal/dockertarget"
	"github.com/grafana/agent/pkg/runner"
	"github.com/prometheus/common/model"
)

// A Manager manages a set of running tailers.
type Manager struct {
	log log.Logger

	mut   sync.Mutex
	opts  *Options
	tasks []*tailerTask

	runner *runner.Runner[*tailerTask]
}

// NewManager returns a new Manager which manages a set of running tailers.
// Options must not be modified after passing it to a Manager.
//
// If NewManager is called with a nil set of options, no targets will be
// scheduled for running until UpdateOptions is called.
func NewManager(l log.Logger, opts *Options) *Manager {
	return &Manager{
		log: l,
		runner: runner.New(func(t *tailerTask) runner.Worker {
			return newTailer(l, t)
		}),
	}
}

// Options passed to all tailers.
type Options struct {
	// Client to use to request logs from Docker.
	Client client.APIClient

	// Handler to send discovered logs to.
	Handler loki.EntryHandler

	// Positions interface so tailers can save/restore offsets in log files.
	Positions positions.Positions
}

// tailerTask is the payload used to create tailers. It implements runner.Task.
type tailerTask struct {
	Options *Options
	Target  *dt.Target
}

var _ runner.Task = (*tailerTask)(nil)

func (tt *tailerTask) Hash() uint64 { return tt.Target.Hash() }

func (tt *tailerTask) Equals(other runner.Task) bool {
	otherTask := other.(*tailerTask)

	// Quick path: pointers are exactly the same.
	if tt == otherTask {
		return true
	}

	// Slow path: check individual fields which are part of the task.
	return tt.Options == otherTask.Options &&
		tt.Target.Labels().String() == otherTask.Target.Labels().String()
}

// A tailer tails the logs of a docker container. It is created by a [Manager].
type tailer struct {
	log    log.Logger
	opts   *Options
	target *dt.Target

	lset model.LabelSet
}

// newTailer returns a new tailer which tails logs from the target specified by
// the task.
func newTailer(l log.Logger, task *tailerTask) *tailer {
	return &tailer{
		log:    log.WithPrefix(l, "target", task.Target.Name()),
		opts:   task.Options,
		target: task.Target,

		lset: task.Target.Labels(),
	}
}

func (t *tailer) Run(ctx context.Context) {
	ch, chErr := t.opts.Client.ContainerWait(ctx, t.target.Name(), container.WaitConditionNextExit)

	t.target.StartIfNotRunning()

	select {
	case err := <-chErr:
		// Error setting up the Wait request from the client; either failed to
		// read from /containers/{containerID}/wait, or couldn't parse the
		// response. Stop the target and exit the task after logging; if it was
		// a transient error, the target will be retried on the next discovery
		// refresh.
		level.Error(t.log).Log("msg", "could not set up a wait request to the Docker client", "error", err)
		t.target.Stop()
		return
	case <-ch:
		t.target.Stop()
		return
	}
}

// SyncTargets synchronizes the set of running tailers to the set specified by
// targets.
func (m *Manager) SyncTargets(ctx context.Context, targets []*dt.Target) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	// Convert targets into tasks to give to the runner.
	tasks := make([]*tailerTask, 0, len(targets))
	for _, target := range targets {
		tasks = append(tasks, &tailerTask{
			Options: m.opts,
			Target:  target,
		})
	}

	// Sync our tasks to the runner. If the Manager doesn't have any options,
	// the runner will be cleared of tasks until UpdateOptions is called with a
	// non-nil set of options.
	switch m.opts {
	default:
		if err := m.runner.ApplyTasks(ctx, tasks); err != nil {
			return err
		}
	case nil:
		if err := m.runner.ApplyTasks(ctx, nil); err != nil {
			return err
		}
	}

	// Delete positions for targets which have gone away.
	newEntries := make(map[positions.Entry]struct{}, len(targets))
	for _, target := range targets {
		newEntries[entryForTarget(target)] = struct{}{}
	}

	for _, task := range m.tasks {
		ent := entryForTarget(task.Target)

		// The task from the last call to SyncTargets is no longer in newEntries;
		// remove it from the positions file. We do this _after_ calling ApplyTasks
		// to ensure that the old tailers have shut down, otherwise the tailer
		// might write its position again during shutdown after we removed it.
		if _, found := newEntries[ent]; !found {
			level.Info(m.log).Log("msg", "removing entry from positions file", "path", ent.Path, "labels", ent.Labels)
			m.opts.Positions.Remove(ent.Path, ent.Labels)
		}
	}

	m.tasks = tasks
	return nil
}

func entryForTarget(t *dt.Target) positions.Entry {
	// The positions entry is keyed by container_id; the path is fed into
	// positions.CursorKey to treat it as a "cursor"; otherwise
	// positions.Positions will try to read the path as a file and delete the
	// entry when it can't find it.
	return positions.Entry{
		Path:   positions.CursorKey(t.Name()),
		Labels: t.Labels().String(),
	}
}

// UpdateOptions updates the Options shared with all Tailers. All Tailers will
// be updated with the new set of Options. Options should not be modified after
// passing to UpdateOptions.
//
// If newOptions is nil, all tasks will be cleared until UpdateOptions is
// called again with a non-nil set of options.
func (m *Manager) UpdateOptions(ctx context.Context, newOptions *Options) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	// Iterate through the previous set of tasks and create a new task with the
	// new set of options.
	tasks := make([]*tailerTask, 0, len(m.tasks))
	for _, oldTask := range m.tasks {
		tasks = append(tasks, &tailerTask{
			Options: newOptions,
			Target:  oldTask.Target,
		})
	}

	switch newOptions {
	case nil:
		if err := m.runner.ApplyTasks(ctx, nil); err != nil {
			return err
		}
	default:
		if err := m.runner.ApplyTasks(ctx, tasks); err != nil {
			return err
		}
	}

	m.opts = newOptions
	m.tasks = tasks
	return nil
}

// Targets returns the set of targets which are actively being tailed. Targets
// for tailers which have terminated are not included. The returned set of
// targets are deduplicated.
func (m *Manager) Targets() []*dt.Target {
	tasks := m.runner.Tasks()

	targets := make([]*dt.Target, 0, len(tasks))
	for _, task := range tasks {
		targets = append(targets, task.Target)
	}
	return targets
}

// Stop stops the manager and all running Tailers. It blocks until all Tailers
// have exited.
func (m *Manager) Stop() {
	m.runner.Stop()
}
