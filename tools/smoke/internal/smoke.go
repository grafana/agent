package smoke

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Smoke is the top level object for a smoke test.
type Smoke struct {
	cfg    *Config
	logger log.Logger
	tasks  []repeatingTask

	fakeRemoteWriteHandler http.HandlerFunc
}

// Config struct to pass configuration to the Smoke constructor.
type Config struct {
	Kubeconfig        string
	Namespace         string
	PodPrefix         string
	FakeRemoteWrite   bool
	SimulateErrors    bool
	ErrorPercentage   float64
	ChaosFrequency    time.Duration
	MutationFrequency time.Duration
}

// RegisterFlags registers flags for the config to the given FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Namespace, "namespace", DefaultConfig.Namespace, "namespace smoke test should run in")
	f.StringVar(&c.Kubeconfig, "kubeconfig", DefaultConfig.Kubeconfig, "absolute path to the kubeconfig file")
	f.StringVar(&c.PodPrefix, "pod-prefix", DefaultConfig.PodPrefix, "prefix for grafana agent pods")
	f.BoolVar(&c.FakeRemoteWrite, "fake-remote-write", DefaultConfig.FakeRemoteWrite, "remote write endpoint for series which are discarded, useful for testing and not storing metrics")
	f.BoolVar(&c.SimulateErrors, "simulate-errors", DefaultConfig.SimulateErrors, "remote write endpoint will return a 500 error response randomly")
	f.Float64Var(&c.ErrorPercentage, "simulate-errors-percentage", DefaultConfig.ErrorPercentage, "percentage chance a request will result in an error")
	f.DurationVar(&c.ChaosFrequency, "chaos-frequency", DefaultConfig.ChaosFrequency, "chaos frequency duration")
	f.DurationVar(&c.MutationFrequency, "mutation-frequency", DefaultConfig.MutationFrequency, "mutation frequency duration")
}

// DefaultConfig holds defaults for Smoke settings.
var DefaultConfig = Config{
	Kubeconfig:        "",
	Namespace:         "default",
	PodPrefix:         "grafana-agent",
	FakeRemoteWrite:   false,
	SimulateErrors:    false,
	ErrorPercentage:   0.01,
	ChaosFrequency:    30 * time.Minute,
	MutationFrequency: 5 * time.Minute,
}

// New is the constructor for a Smoke object.
func New(logger log.Logger, cfg Config) (*Smoke, error) {
	s := &Smoke{
		cfg:    &cfg,
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

	if cfg.FakeRemoteWrite {
		s.fakeRemoteWriteHandler = func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		if cfg.SimulateErrors {
			rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
			s.fakeRemoteWriteHandler = func(w http.ResponseWriter, _ *http.Request) {
				if rnd := rand.Float64(); cfg.ErrorPercentage > rnd {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
		}
	}

	// add default tasks
	s.tasks = append(s.tasks,
		repeatingTask{
			Task: &deletePodTask{
				logger:    log.With(s.logger, "task", "delete_pod", "pod", "grafana-agent-0"),
				clientset: clientset,
				namespace: cfg.Namespace,
				pod:       fmt.Sprintf("%s-0", cfg.PodPrefix),
			},
			frequency: cfg.ChaosFrequency,
		},
		repeatingTask{
			Task: &deletePodBySelectorTask{
				logger:    log.With(s.logger, "task", "delete_pod_by_selector"),
				clientset: clientset,
				namespace: cfg.Namespace,
				selector:  fmt.Sprintf("name=%s-cluster", cfg.PodPrefix),
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

func (s *Smoke) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.fakeRemoteWriteHandler(w, r)
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
		g.Add(taskFn(task), func(_ error) {
			cancel()
		})
	}

	if s.cfg.FakeRemoteWrite && s.fakeRemoteWriteHandler != nil {
		level.Info(s.logger).Log("msg", "serving fake remote-write endpoint on :19090")
		g.Add(func() error {
			return http.ListenAndServe(":19090", s)
		}, func(_ error) {
			cancel()
		})
	}

	return g.Run()
}
