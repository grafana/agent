package smoke

import (
	"context"
	"flag"
	"time"

	"github.com/go-kit/log"
	"github.com/oklog/run"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Smoke is the top level object for a smoke test.
type Smoke struct {
	logger log.Logger
	tasks  []repeatingTask
}

// Config struct to pass configuration to the Smoke constructor.
type Config struct {
	Kubeconfig        string
	Namespace         string
	ChaosFrequency    time.Duration
	MutationFrequency time.Duration
}

// RegisterFlags registers flags for the config to the given FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Namespace, "namespace", DefaultConfig.Namespace, "namespace smoke test should run in")
	f.StringVar(&c.Kubeconfig, "kubeconfig", DefaultConfig.Kubeconfig, "absolute path to the kubeconfig file")
	f.DurationVar(&c.ChaosFrequency, "chaos-frequency", DefaultConfig.ChaosFrequency, "chaos frequency duration")
	f.DurationVar(&c.MutationFrequency, "mutation-frequency", DefaultConfig.MutationFrequency, "mutation frequency duration")
}

// DefaultConfig holds defaults for Smoke settings.
var DefaultConfig = Config{
	Kubeconfig:        "",
	Namespace:         "default",
	ChaosFrequency:    30 * time.Minute,
	MutationFrequency: 5 * time.Minute,
}

// New is the constructor for a Smoke object.
func New(logger log.Logger, cfg Config) (*Smoke, error) {
	s := &Smoke{
		logger: logger,
	}
	if s.logger == nil {
		s.logger = log.NewNopLogger()
	}

	// use the current context in kubeconfig. this falls back to in-cluster config if kubeconfig is empty
	config, err := clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// add default tasks
	s.tasks = append(s.tasks,
		repeatingTask{
			Task: &deletePodTask{
				logger:    log.With(s.logger, "task", "delete_pod", "pod", "grafana-agent-0"),
				clientset: clientset,
				namespace: cfg.Namespace,
				pod:       "grafana-agent-0",
			},
			frequency: cfg.ChaosFrequency,
		},
		repeatingTask{
			Task: &deletePodBySelectorTask{
				logger:    log.With(s.logger, "task", "delete_pod_by_selector"),
				clientset: clientset,
				namespace: cfg.Namespace,
				selector:  "name=grafana-agent-cluster",
			},
			frequency: cfg.ChaosFrequency,
		},
		repeatingTask{
			Task: &scaleDeploymentTask{
				logger:      log.With(s.logger, "task", "scale_deployment", "deployment", "avalanche"),
				clientset:   clientset,
				namespace:   cfg.Namespace,
				deployment:  "avalanche",
				maxReplicas: 11,
				minReplicas: 2,
			},
			frequency: cfg.MutationFrequency,
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
