package gcplogtarget

// This code is copied from Promtail. The gcplogtarget package is used to
// configure and run the targets that can read log entries from cloud resource
// logs like bucket logs, load balancer logs, and Kubernetes cluster logs
// from GCP.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/clients/pkg/promtail/targets/serverutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
)

// PushTarget defines a server for receiving messages from a GCP PubSub push
// subscription.
type PushTarget struct {
	logger         log.Logger
	jobName        string
	metrics        *Metrics
	config         *PushConfig
	entries        chan<- loki.Entry
	handler        loki.EntryHandler
	relabelConfigs []*relabel.Config
	server         *server.Server
	serverConfig   server.Config
}

// NewPushTarget constructs a PushTarget.
func NewPushTarget(metrics *Metrics, logger log.Logger, handler loki.EntryHandler, jobName string, config *PushConfig, relabel []*relabel.Config, reg prometheus.Registerer) (*PushTarget, error) {
	pt := &PushTarget{
		logger:         logger,
		jobName:        jobName,
		metrics:        metrics,
		config:         config,
		entries:        handler.Chan(),
		handler:        handler,
		relabelConfigs: relabel,
	}

	srvCfg := server.Config{
		HTTPListenPort:    config.HTTPListenPort,
		HTTPListenAddress: config.HTTPListenAddress,

		// Avoid logging entire received request on failures
		ExcludeRequestInLog: true,
	}
	mergedServerConfigs, err := serverutils.MergeWithDefaults(srvCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configs and override defaults when configuring gcp push target: %w", err)
	}

	pt.serverConfig = mergedServerConfigs
	pt.serverConfig.Registerer = reg

	err = pt.run()
	if err != nil {
		return nil, err
	}

	return pt, nil
}

func (p *PushTarget) run() error {
	level.Info(p.logger).Log("msg", "starting gcp push target", "job", p.jobName)

	// To prevent metric collisions registering in the global Prometheus registry.
	tentativeServerMetricNamespace := p.jobName + "_push_target"
	if !model.IsValidMetricName(model.LabelValue(tentativeServerMetricNamespace)) {
		return fmt.Errorf("invalid prometheus-compatible job name: %s", p.jobName)
	}
	p.serverConfig.MetricsNamespace = tentativeServerMetricNamespace

	// We don't want the /debug and /metrics endpoints running, since this is
	// not the main Flow HTTP server. We want this target to expose the least
	// surface area possible, hence disabling WeaveWorks HTTP server metrics
	// and debugging functionality.
	p.serverConfig.RegisterInstrumentation = false

	p.serverConfig.Log = logging.GoKit(p.logger)
	srv, err := server.New(p.serverConfig)
	if err != nil {
		return err
	}
	p.server = srv

	p.server.HTTP.Path("/gcp/api/v1/push").Methods("POST").Handler(http.HandlerFunc(p.push))

	go func() {
		err := srv.Run()
		if err != nil {
			level.Error(p.logger).Log("msg", "loki.source.gcplog push target shutdown with error", "err", err)
		}
	}()

	return nil
}

func (p *PushTarget) push(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Create no-op context.WithTimeout returns to simplify logic
	ctx := r.Context()
	cancel := context.CancelFunc(func() {})
	if p.config.PushTimeout != 0 {
		ctx, cancel = context.WithTimeout(r.Context(), p.config.PushTimeout)
	}
	defer cancel()

	pushMessage := PushMessage{}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		p.metrics.gcpPushErrors.WithLabelValues("read_error").Inc()
		level.Warn(p.logger).Log("msg", "failed to read incoming gcp push request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &pushMessage)
	if err != nil {
		p.metrics.gcpPushErrors.WithLabelValues("format").Inc()
		level.Warn(p.logger).Log("msg", "failed to unmarshall gcp push request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = pushMessage.Validate(); err != nil {
		p.metrics.gcpPushErrors.WithLabelValues("invalid_message").Inc()
		level.Warn(p.logger).Log("msg", "invalid gcp push request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry, err := translate(pushMessage, p.Labels(), p.config.UseIncomingTimestamp, p.relabelConfigs, r.Header.Get("X-Scope-OrgID"))
	if err != nil {
		p.metrics.gcpPushErrors.WithLabelValues("translation").Inc()
		level.Warn(p.logger).Log("msg", "failed to translate gcp push request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(p.logger).Log("msg", fmt.Sprintf("Received line: %s", entry.Line))

	if err := p.doSendEntry(ctx, entry); err != nil {
		// NOTE: timeout errors can be tracked with from the metrics exposed by
		// the spun weaveworks server.
		// loki.source.gcplog.componentid_push_target_request_duration_seconds_count{status_code="503"}
		level.Warn(p.logger).Log("msg", "error sending log entry", "err", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	p.metrics.gcpPushEntries.WithLabelValues().Inc()
	w.WriteHeader(http.StatusNoContent)
}

func (p *PushTarget) doSendEntry(ctx context.Context, entry loki.Entry) error {
	select {
	// Timeout the loki.Entry channel send operation, which is the only blocking operation in the handler
	case <-ctx.Done():
		return fmt.Errorf("timeout exceeded: %w", ctx.Err())
	case p.entries <- entry:
		return nil
	}
}

// Labels return the model.LabelSet that the target applies to log entries.
func (p *PushTarget) Labels() model.LabelSet {
	lbls := make(model.LabelSet, len(p.config.Labels))
	for k, v := range p.config.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}
	return lbls
}

// Details returns some debug information about the target.
func (p *PushTarget) Details() map[string]string {
	return map[string]string{
		"strategy":       "push",
		"labels":         p.Labels().String(),
		"server_address": p.server.HTTPListenAddr().String(),
	}
}

// Stop shuts down the push target.
func (p *PushTarget) Stop() error {
	level.Info(p.logger).Log("msg", "stopping gcp push target", "job", p.jobName)
	p.server.Stop()
	p.server.Shutdown()
	p.handler.Stop()
	return nil
}
