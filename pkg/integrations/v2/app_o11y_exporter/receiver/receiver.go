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

type AppReceiver struct {
	exporters   []exporters.AppReceiverExporter
	config      config.AppExporterConfig
	rateLimiter *ratelimiting.RateLimiter
}

const (
	DEFAULT_RATE_LIMITING_RPS       = 100
	DEFAULT_RATE_LIMITING_BURSTINES = 50
)

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

		for _, exporter := range ar.exporters {
			wg.Add(1)
			go func(exp exporters.AppReceiverExporter) {
				defer wg.Done()
				// Metrics exporters run asynchronously when the self scrapping
				// collects data
				if de, ok := exp.(exporters.AppMetricsExporter); ok {
					de.Process(p)
				}
				// Data exporters, export in sync with the user request
				if de, ok := exp.(exporters.AppDataExporter); ok {
					de.Export(p)
				}
			}(exporter)
		}

		wg.Wait()
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("ok"))
	})

	if len(ar.config.CORSAllowedOrigins) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins: ar.config.CORSAllowedOrigins,
		})
		handler = c.Handler(handler)
	}

	return handler
}
