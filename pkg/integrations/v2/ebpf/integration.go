package ebpf

import (
	"context"
	"fmt"
	"net/http"
	"path"

	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`
}

type ebpfHandler struct{}

func init() {
	integrations.Register(&config{}, integrations.TypeSingleton)
}

var defaultConfig = config{
	Programs: []ebpf_config.Program{},
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = defaultConfig
	type plain config

	return unmarshal((*plain)(c))
}

func (c *config) ApplyDefaults(globals integrations.Globals) error {
	return nil
}

func (c *config) Identifier(globals integrations.Globals) (string, error) {
	return "ebpf", nil
}

func (c *config) Name() string { return "ebpf" }

func (c *config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	ebpf := &ebpfHandler{}

	return ebpf, nil
}

// RunIntegration implements the Integration interface and is
// the entrypoint for our integration
func (e *ebpfHandler) RunIntegration(ctx context.Context) error {
	fmt.Println("Running epbf handler!")
	<-ctx.Done()
	fmt.Println("Exiting from ebpf handler...")
	return nil
}

// Handler implements the HTTPIntegration interface.
func (e *ebpfHandler) Handler(prefix string) (http.Handler, error) {
	r := mux.NewRouter()
	r.Handle(path.Join(prefix, "metrics"), createHandler())
	return r, nil
}

func createHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ln := []byte(`an_important_metric_total{method="GET",code="200"}  3`)
		w.Write(ln)
		return
	}
}

// Targets implements the MetricsIntegration interface.
func (e *ebpfHandler) Targets(ep integrations.Endpoint) []*targetgroup.Group {
	return []*targetgroup.Group{{}}
}

// ScrapeConfigs implements the MetricsIntegration interface.
func (e *ebpfHandler) ScrapeConfigs(sd discovery.Configs) []*autoscrape.ScrapeConfig {
	return nil
}

// ServeHTTP kicks off the integration's HTTP handler.
func (e *ebpfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	createHandler()
}
