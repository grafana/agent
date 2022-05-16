package flow

import (
	"context"
	"sync"

	"github.com/go-kit/log"
)

// The scheduler manages running components.
type scheduler struct {
	log log.Logger

	ctx     context.Context
	cancel  context.CancelFunc
	running sync.WaitGroup

	tasksMut sync.Mutex
	tasks    map[string]*task
}

// newScheduler creates a new scheduler
func newScheduler(l log.Logger) *scheduler {
	if l == nil {
		l = log.NewNopLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())

	sched := &scheduler{
		log: l,

		ctx:    ctx,
		cancel: cancel,

		tasks: make(map[string]*task),
	}
	return sched
}

// Synchronize synchronizes the current tasks to those defined by rr.
//
// New runnables will be launched as tasks. Runnables already managed by the
// scheduler will be kept running, while runnables that are no longer present
// in rr will be removed.
func (s *scheduler) Synchronize(rr []runnable) {
	s.tasksMut.Lock()
	defer s.tasksMut.Unlock()

	newRunnables := make(map[string]runnable, len(rr))
	for _, r := range rr {
		newRunnables[r.NodeID()] = r
	}

	// Stop tasks that are not defined in rr.
	var stopping sync.WaitGroup
	for id, t := range s.tasks {
		if _, keep := newRunnables[id]; keep {
			continue
		}

		stopping.Add(1)
		go func(t *task) {
			defer stopping.Done()
			t.Stop()
		}(t)
	}

	// Launch new runnables that have appeared.
	for id, r := range newRunnables {
		if _, exist := s.tasks[id]; exist {
			continue
		}

		var (
			nodeID      = id
			newRunnable = r
		)

		opts := taskOptions{
			Context:  s.ctx,
			Logger:   log.With(s.log, "component", id),
			Runnable: newRunnable,
			OnDone: func() {
				defer s.running.Done()

				s.tasksMut.Lock()
				defer s.tasksMut.Unlock()
				delete(s.tasks, nodeID)
			},
		}

		s.running.Add(1)
		s.tasks[nodeID] = newTask(opts)
	}

	// Wait for all stopping runnables to exit.
	stopping.Wait()
}

func (s *scheduler) Close() error {
	s.cancel()
	s.running.Wait()
	return nil
}

// task is a scheduled runnable.
type task struct {
	ctx    context.Context
	cancel context.CancelFunc
	exited chan struct{}
}

type taskOptions struct {
	Context  context.Context
	Logger   log.Logger
	Runnable runnable
	OnDone   func()
}

func newTask(opts taskOptions) *task {
	ctx, cancel := context.WithCancel(opts.Context)

	t := &task{
		ctx:    ctx,
		cancel: cancel,
		exited: make(chan struct{}),
	}

	go func() {
		defer opts.OnDone()
		defer close(t.exited)
		_ = opts.Runnable.Run(t.ctx)
	}()
	return t
}

func (t *task) Stop() {
	t.cancel()
	<-t.exited
}

type runnable interface {
	NodeID() string
	Run(ctx context.Context) error
}
