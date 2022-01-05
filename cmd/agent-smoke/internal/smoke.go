package smoke

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/oklog/run"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Smoke is the top level object for a smoke test.
type Smoke struct {
	logger     log.Logger
	kubeconfig string
	namespace  string
	tasks      []repeatingTask

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

	if s.logger == nil {
		s.logger = log.NewNopLogger()
	}

	// use the current context in kubeconfig. this falls back to in-cluster config if kubeconfig is empty
	config, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// add default tasks
	// TODO: need to add a deletePodTask for cluster pods,
	// TODO: currently script generates random number and appends it
	// TODO: to the pod name. Should do this with a label selector,
	// TODO: by adding a deletePodByLabelTask or such
	s.tasks = append(s.tasks,
		repeatingTask{
			Task: &deletePodTask{
				logger:    s.logger,
				clientset: clientset,
				namespace: s.namespace,
				pod:       "grafana-agent-0",
			},
			frequency: s.chaosFrequency,
		},
		repeatingTask{
			Task: &scaleDeploymentTask{
				logger:      s.logger,
				clientset:   clientset,
				namespace:   s.namespace,
				deployment:  "avalanche",
				maxReplicas: 11,
				minReplicas: 2,
			},
			frequency: s.mutationFrequency,
		})

	return s, nil
}

// Run starts the smoke test and runs the tasks concurrently.
func (s *Smoke) Run(ctx context.Context) error {
	var g run.Group
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	taskFn := func(t repeatingTask) func() error {
		return func() error {
			tick := time.NewTicker(t.frequency)
			defer tick.Stop()
			for {
				select {
				case <-tick.C:
					if err := t.Run(ctx); err != nil {
						return err
					}
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
	for _, task := range s.tasks {
		g.Add(taskFn(task), func(err error) {
			cancel()
		})
	}
	return g.Run()
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
