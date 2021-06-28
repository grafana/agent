package frontendcollector

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/util/server"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type FrontendCollector struct {
	mut       sync.Mutex
	l         log.Logger
	instances map[string]*Instance
}

type Instance struct {
	cfg *InstanceConfig
	mut sync.Mutex
	l   log.Logger
	srv *server.Server
}

func New(c Config, loki *loki.Loki, l log.Logger) (*FrontendCollector, error) {
	frontendCollector := &FrontendCollector{
		instances: make(map[string]*Instance),
		l:         log.With(l, "component", "frontendcollector"),
	}

	if err := frontendCollector.ApplyConfig(c); err != nil {
		return nil, err
	}

	return frontendCollector, nil
}

func (f *FrontendCollector) ApplyConfig(c Config) error {
	f.mut.Lock()
	defer f.mut.Unlock()

	newInstances := make(map[string]*Instance, len(c.Configs))

	for _, ic := range c.Configs {
		// If an old instance existed, update it and move it to the new map.
		if old, ok := f.instances[ic.Name]; ok {
			err := old.ApplyConfig(ic)
			if err != nil {
				return err
			}

			newInstances[ic.Name] = old
			continue
		}

		inst, err := NewInstance(ic, f.l)
		if err != nil {
			return fmt.Errorf("unable to apply config for %s: %w", ic.Name, err)
		}
		newInstances[ic.Name] = inst
	}

	// Any promtail in l.instances that isn't in newInstances has been removed
	// from the config. Stop them before replacing the map.
	for key, i := range f.instances {
		if _, exist := newInstances[key]; exist {
			continue
		}
		i.Stop()
	}
	f.instances = newInstances

	return nil
}

func (f *FrontendCollector) Stop() {
	f.mut.Lock()
	defer f.mut.Unlock()

	for _, i := range f.instances {
		i.Stop()
	}
}

// NewInstance creates and starts a frontend collector instance.
func NewInstance(c *InstanceConfig, l log.Logger) (*Instance, error) {
	logger := log.With(l, "fe_collector_instance", c.Name)
	srv := server.New(prometheus.NewRegistry(), logger)
	inst := Instance{
		cfg: c,
		l:   logger,
		srv: srv,
	}
	if err := inst.ApplyConfig(c); err != nil {
		return nil, err
	}
	go func() {
		err := srv.Run()
		fmt.Println("stop", err)
		if err != nil {
			level.Error(logger).Log("msg", "Failed to start frontend collector", "err", err)
		}
	}()
	return &inst, nil
}

func (i *Instance) ApplyConfig(c *InstanceConfig) error {
	i.mut.Lock()
	defer i.mut.Unlock()
	err := i.srv.ApplyConfig(c.Server, c.wire)
	if err != nil {
		return err
	}
	i.cfg = c
	return nil
}

func (c *InstanceConfig) wire(mux *mux.Router, grpc *grpc.Server) {
	fmt.Println("wire!")
	mux.HandleFunc("/collect", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "hello world.\n")
	})
}

func (i *Instance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()
	i.srv.Close()
}
