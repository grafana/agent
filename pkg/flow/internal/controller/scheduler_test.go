package controller_test

import (
	"context"
	"sync"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/stretchr/testify/require"
)

func TestScheduler_Synchronize(t *testing.T) {
	t.Run("Can start new jobs", func(t *testing.T) {
		var started, finished sync.WaitGroup
		started.Add(3)
		finished.Add(3)

		runFunc := func(ctx context.Context) error {
			defer finished.Done()
			started.Done()

			<-ctx.Done()
			return nil
		}

		sched := controller.NewScheduler()
		sched.Synchronize([]controller.RunnableNode{
			fakeRunnable{ID: "component-a", Component: mockComponent{RunFunc: runFunc}},
			fakeRunnable{ID: "component-b", Component: mockComponent{RunFunc: runFunc}},
			fakeRunnable{ID: "component-c", Component: mockComponent{RunFunc: runFunc}},
		})

		started.Wait()
		require.NoError(t, sched.Close())
		finished.Wait()
	})

	t.Run("Ignores existing jobs", func(t *testing.T) {
		var started sync.WaitGroup
		started.Add(1)

		runFunc := func(ctx context.Context) error {
			started.Done()
			<-ctx.Done()
			return nil
		}

		sched := controller.NewScheduler()

		for i := 0; i < 10; i++ {
			// If a new runnable is created, runFunc will panic since the WaitGroup
			// only supports 1 goroutine.
			sched.Synchronize([]controller.RunnableNode{
				fakeRunnable{ID: "component-a", Component: mockComponent{RunFunc: runFunc}},
			})
		}

		started.Wait()
		require.NoError(t, sched.Close())
	})

	t.Run("Removes stale jobs", func(t *testing.T) {
		var started, finished sync.WaitGroup
		started.Add(1)
		finished.Add(1)

		runFunc := func(ctx context.Context) error {
			defer finished.Done()
			started.Done()
			<-ctx.Done()
			return nil
		}

		sched := controller.NewScheduler()

		sched.Synchronize([]controller.RunnableNode{
			fakeRunnable{ID: "component-a", Component: mockComponent{RunFunc: runFunc}},
		})
		started.Wait()

		sched.Synchronize([]controller.RunnableNode{})

		finished.Wait()
		require.NoError(t, sched.Close())
	})
}

type fakeRunnable struct {
	ID        string
	Component component.Component
}

var _ controller.RunnableNode = fakeRunnable{}

func (fr fakeRunnable) NodeID() string                { return fr.ID }
func (fr fakeRunnable) Run(ctx context.Context) error { return fr.Component.Run(ctx) }

type mockComponent struct {
	RunFunc    func(ctx context.Context) error
	UpdateFunc func(newConfig component.Arguments) error
}

var _ component.Component = (*mockComponent)(nil)

func (mc mockComponent) Run(ctx context.Context) error              { return mc.RunFunc(ctx) }
func (mc mockComponent) Update(newConfig component.Arguments) error { return mc.UpdateFunc(newConfig) }
