package smoke

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Smoke is the top level object for a smoke test.
type Smoke struct {
	logger     log.Logger
	clientset  *kubernetes.Clientset
	kubeconfig string
	namespace  string
	tasks      []Task

	chaosFrequency    time.Duration
	mutationFrequency time.Duration
}

// NewSmokeTest is the constructor for a Smoke object.
func NewSmokeTest(opts ...Option) (*Smoke, error) {
	s := &Smoke{
		namespace:         "default",
		chaosFrequency:    5 * time.Minute,
		mutationFrequency: 30 * time.Minute,
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	// add default tasks
	chaosLoop := &deletePodTask{
		namespace: s.namespace,
		pod:       "grafana-agent-0",
		duration:  s.chaosFrequency,
	}
	// TODO: need to add a deletePodTask for cluster pods,
	// TODO: currently script generates random number and appends it
	// TODO: to the pod name. Should do this with a label selector,
	// TODO: by adding a deletePodByLabelTask or such
	mutationLoop := &scaleDeploymentTask{
		namespace:   s.namespace,
		deployment:  "avalanche",
		maxReplicas: 11,
		minReplicas: 2,
		duration:    s.mutationFrequency,
	}
	s.tasks = append(s.tasks, chaosLoop, mutationLoop)

	// use the current context in kubeconfig. this falls back to in-cluster config if kubeconfig is empty
	config, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	s.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Run starts the smoke test and runs the tasks in an errgroup Group.
func (s *Smoke) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	taskFn := func(t Task) func() error {
		fn, freq := t.Task()
		return func() error {
			tick := time.NewTicker(freq)
			defer tick.Stop()
			for {
				select {
				case <-tick.C:
					if err := fn(ctx, s); err != nil {
						return err
					}
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
	for _, task := range s.tasks {
		g.Go(taskFn(task))
	}
	return g.Wait()
}

func (s *Smoke) logDebug(keyvals ...interface{}) {
	if s.logger != nil {
		level.Debug(s.logger).Log(keyvals...)
	}
}

func (s *Smoke) logError(keyvals ...interface{}) {
	if s.logger != nil {
		level.Error(s.logger).Log(keyvals...)
	}
}

// Option type for constructor option functions.
type Option func(*Smoke) error

// WithKubeConfig option function adds kubeconfig path to smoke test. If this
// not set, the smoke test will fall back to an in-cluster config.
func WithKubeConfig(k string) Option {
	return func(s *Smoke) error {
		s.kubeconfig = k
		return nil
	}
}

// WithLogger option function adds a logger to the smoke test.
func WithLogger(l log.Logger) Option {
	return func(s *Smoke) error {
		s.logger = l
		return nil
	}
}

// WithChaosFrequency option function sets the duration of the chaos task frequency.
func WithChaosFrequency(f time.Duration) Option {
	return func(s *Smoke) error {
		s.chaosFrequency = f
		return nil
	}
}

// WithMutationFrequency option function sets the duration used for mutation task frequency.
func WithMutationFrequency(f time.Duration) Option {
	return func(s *Smoke) error {
		s.mutationFrequency = f
		return nil
	}
}

// WithNamespace sets the namespace to use for the smoke test.
func WithNamespace(ns string) Option {
	return func(s *Smoke) error {
		s.namespace = ns
		return nil
	}
}

// WithTask appends a task to be executed by the smoke test.
func WithTask(t Task) Option {
	return func(s *Smoke) error {
		s.tasks = append(s.tasks, t)
		return nil
	}
}
