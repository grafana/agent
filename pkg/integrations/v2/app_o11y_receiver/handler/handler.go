package handler

import (
	"sync"

	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/config"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/exporters"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/tools/ratelimiting"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
)

// AppO11yHandler struct controls the data ingestion http handler of the receiver
type AppO11yHandler struct {
	exporters               []exporters.AppO11yReceiverExporter
	config                  config.AppO11yReceiverConfig
	rateLimiter             *ratelimiting.RateLimiter
	exporterErrorsCollector *prometheus.CounterVec
}

// NewAppO11yHandler creates a new AppReceiver instance based on the given configuration
func NewAppO11yHandler(conf config.AppO11yReceiverConfig, exporters []exporters.AppO11yReceiverExporter, reg *prometheus.Registry) AppO11yHandler {
	var rateLimiter *ratelimiting.RateLimiter
	if conf.Server.RateLimiting.Enabled {
		var rps float64
		if conf.Server.RateLimiting.RPS > 0 {
			rps = conf.Server.RateLimiting.RPS
		}

		var b int
		if conf.Server.RateLimiting.Burstiness > 0 {
			b = conf.Server.RateLimiting.Burstiness
		}
		rateLimiter = ratelimiting.NewRateLimiter(rps, b)
	}

	exporterErrorsCollector := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: utils.MetricsNamespace,
		Subsystem: "exporter",
		Name:      "errors",
		Help:      "Total number of errors produced by a receiver exporter",
	}, []string{"exporter"})

	if reg != nil {
		reg.MustRegister(exporterErrorsCollector)
	}

	return AppO11yHandler{
		exporters:               exporters,
		config:                  conf,
		rateLimiter:             rateLimiter,
		exporterErrorsCollector: exporterErrorsCollector,
	}
}

// HTTPHandler is the http.Handler for the receiver. It will do the following
// 0. Enable CORS for the configured hosts
// 1. Check if the request should be rate limited
// 2. Verify that the payload size is within limits
// 3. Start two go routines for exporters processing and exporting data respectively
// 4. Respond with 202 once all the work is done
func (ar *AppO11yHandler) HTTPHandler(logger log.Logger) http.Handler {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check rate limiting state
		if ar.config.Server.RateLimiting.Enabled && ar.rateLimiter.IsRateLimited() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		// check API key if one is provided
		if len(ar.config.Server.APIKey) > 0 && subtle.ConstantTimeCompare([]byte(r.Header.Get("x-api-key")), []byte(ar.config.Server.APIKey)) == 0 {
			http.Error(w, "api key not provided or incorrect", http.StatusUnauthorized)
			return
		}

		// Verify content length. We trust net/http to give us the correct number
		if ar.config.Server.MaxAllowedPayloadSize > 0 && r.ContentLength > ar.config.Server.MaxAllowedPayloadSize {
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
			go func(exp exporters.AppO11yReceiverExporter) {
				defer wg.Done()
				if err := exp.Export(r.Context(), p); err != nil {
					level.Error(logger).Log("msg", "exporter error", "exporter", exp.Name(), "error", err)
					ar.exporterErrorsCollector.WithLabelValues(exp.Name()).Inc()
				}
			}(exporter)
		}

		wg.Wait()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	})

	if len(ar.config.Server.CORSAllowedOrigins) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins: ar.config.Server.CORSAllowedOrigins,
			AllowedHeaders: []string{"x-api-key", "content-type"},
		})
		handler = c.Handler(handler)
	}

	return handler
}
