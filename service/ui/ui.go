// Package ui implements the UI service.
package ui

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/service"
	"github.com/grafana/agent/service/cluster"
	http_service "github.com/grafana/agent/service/http"
	"github.com/grafana/agent/web/api"
	"github.com/grafana/agent/web/ui"
	"github.com/tg123/go-htpasswd"
)

// ServiceName defines the name used for the UI service.
const ServiceName = "ui"

// Options are used to configure the UI service. Options are constant for the
// lifetime of the UI service.
type Options struct {
	BasicAuthUserFilepath string          // Path of Basic Auth user and hashed password.
	Cluster               cluster.Cluster // Cluster
	Logger                log.Logger      // Logger
	UIPrefix              string          // Path prefix to host the UI at.
}

// Service implements the UI service.
type Service struct {
	opts Options
}

// New returns a new, unstarted UI service.
func New(opts Options) *Service {
	return &Service{
		opts: opts,
	}
}

var (
	_ service.Service             = (*Service)(nil)
	_ http_service.ServiceHandler = (*Service)(nil)
)

// Definition returns the definition of the HTTP service.
func (s *Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil, // ui does not accept configuration
		DependsOn:  []string{http_service.ServiceName},
	}
}

// Run starts the UI service. It will run until the provided context is
// canceled or there is a fatal error.
func (s *Service) Run(ctx context.Context, host service.Host) error {
	<-ctx.Done()
	return nil
}

// Update implements [service.Service]. It is a no-op since the UI service
// does not support runtime configuration.
func (s *Service) Update(newConfig any) error {
	return fmt.Errorf("UI service does not support configuration")
}

// Data implements [service.Service]. It returns nil, as the UI service does
// not have any runtime data.
func (s *Service) Data() any {
	return nil
}

// ServiceHandler implements [http_service.ServiceHandler]. It returns the HTTP
// endpoints to host the UI.
func (s *Service) ServiceHandler(host service.Host) (base string, handler http.Handler) {
	r := mux.NewRouter()

	// If toggled, setup Basic Auth for all UI and API routes
	if s.opts.BasicAuthUserFilepath != "" {
		_, err := os.Stat(s.opts.BasicAuthUserFilepath)
		if errors.Is(err, fs.ErrNotExist) {
			level.Info(s.opts.Logger).Log("msg", "Basic Auth user file doesn't exist")
		} else {
			r.Use(func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

					htpassFile, err := htpasswd.New(s.opts.BasicAuthUserFilepath, htpasswd.DefaultSystems, nil)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					username, password, authOk := r.BasicAuth()
					if !authOk {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
					if ok := htpassFile.Match(username, password); ok {
						h.ServeHTTP(w, r)
					}
				})
			})
		}
	}

	// TODO(rfratto): allow service.Host to return services so we don't have to
	// pass the clustering service in Options.
	fa := api.NewFlowAPI(host, s.opts.Cluster)
	fa.RegisterRoutes(path.Join(s.opts.UIPrefix, "/api/v0/web"), r)
	ui.RegisterRoutes(s.opts.UIPrefix, r)

	return s.opts.UIPrefix, r
}
