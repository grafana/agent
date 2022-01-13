package receiver

import (
	"sync"

	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/exporters"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_exporter/tools/ratelimiting"
	"github.com/rs/cors"
)

// AppReceiver struct contrls the data receiver of the exporter
type AppReceiver struct {
	exporters   []exporters.AppReceiverExporter
	config      config.AppExporterConfig
	rateLimiter *ratelimiting.RateLimiter
}

// NewAppReceiver creates a new AppReceiver instance based on the given configuration
func NewAppReceiver(conf config.AppExporterConfig, exporters []exporters.AppReceiverExporter) AppReceiver {
	var rateLimiter *ratelimiting.RateLimiter
	if conf.RateLimiting.Enabled {
		var rps float64
		if conf.RateLimiting.RPS > 0 {
			rps = conf.RateLimiting.RPS
		}

		var b int
		if conf.RateLimiting.Burstiness > 0 {
			b = conf.RateLimiting.Burstiness
		}
		rateLimiter = ratelimiting.NewRateLimiter(rps, b)
	}

	return AppReceiver{
		exporters:   exporters,
		config:      conf,
		rateLimiter: rateLimiter,
	}
}

// ReceiverHandler is the http.Handler for the receiver. It will do the following
// 0. Enable CORS for the configured hosts
// 1. Check if the request should be rate limited
// 2. Verify that the payload size is within limits
// 3. Start two go routines for exporters processing and exporting data respectively
// 4. Respond with 202 once all the work is done
func (ar *AppReceiver) ReceiverHandler(logger *log.Logger) http.Handler {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check rate limiting state
		if ar.config.RateLimiting.Enabled && ar.rateLimiter.IsRateLimited() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		// Verify content length. We trust net/http to give us the correct number
		if ar.config.MaxAllowedPayloadSize > 0 && r.ContentLength > ar.config.MaxAllowedPayloadSize {
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}

		var p models.Payload
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var wg sync.WaitGroup
		wgDone := make(chan bool)
		errChan := make(chan error)

		for _, exporter := range ar.exporters {
			wg.Add(2)
			go func(exp exporters.AppReceiverExporter) {
				defer wg.Done()
				// Metrics exporters run asynchronously when the self scrapping
				// collects data
				if de, ok := exp.(exporters.AppMetricsExporter); ok {
					if err = de.Process(p); err != nil {
						errChan <- err
					}
				}
			}(exporter)

			go func(exp exporters.AppReceiverExporter) {
				defer wg.Done()
				// Data exporters, export in sync with the user request
				if de, ok := exp.(exporters.AppDataExporter); ok {
					if err = de.Export(p); err != nil {
						errChan <- err
					}
				}
			}(exporter)
		}

		go func() {
			wg.Wait()
			close(wgDone)
		}()

		select {
		case <-wgDone:
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("ok"))
		case err := <-errChan:
			close(errChan)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	if len(ar.config.CORSAllowedOrigins) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins: ar.config.CORSAllowedOrigins,
		})
		handler = c.Handler(handler)
	}

	return handler
}
