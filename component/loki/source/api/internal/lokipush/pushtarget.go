package lokipush

// This code is copied from Promtail (52655706b8fc9393983d8709c20aab5a851c222d) with changes kept to the minimum.
// The lokipush package is used to configure and run the HTTP server that can receive loki push API requests and
// forward them to other loki components.

import (
	"bufio"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	promql_parser "github.com/prometheus/prometheus/promql/parser"
	"github.com/weaveworks/common/server"

	"github.com/grafana/dskit/tenant"

	"github.com/grafana/agent/component/common/loki"

	"github.com/grafana/loki/pkg/loghttp/push"
	"github.com/grafana/loki/pkg/logproto"
	util_log "github.com/grafana/loki/pkg/util/log"
)

type PushTargetConfig struct {
	// Server is the weaveworks server config for listening connections
	Server server.Config
	// Labels optionally holds labels to associate with each record received on the push api.
	Labels model.LabelSet
	// If promtail should maintain the incoming log timestamp or replace it with the current time.
	KeepTimestamp bool
	// RelabelConfig used to relabel incoming entries.
	RelabelConfig []*relabel.Config
}

type PushTarget struct {
	logger  log.Logger
	handler loki.EntryHandler
	config  *PushTargetConfig
	jobName string
	server  *server.Server
}

func NewPushTarget(logger log.Logger,
	handler loki.EntryHandler,
	jobName string,
	config *PushTargetConfig,
) (*PushTarget, error) {

	pt := &PushTarget{
		logger:  logger,
		handler: handler,
		jobName: jobName,
		config:  config,
	}

	if err := pt.run(); err != nil {
		return nil, err
	}

	return pt, nil
}

func (t *PushTarget) run() error {
	level.Info(t.logger).Log("msg", "starting push server", "job", t.jobName)

	srv, err := server.New(t.config.Server)
	if err != nil {
		return err
	}

	t.server = srv
	t.server.HTTP.Path("/api/v1/push").Methods("POST").Handler(http.HandlerFunc(t.handleLoki))
	t.server.HTTP.Path("/api/v1/raw").Methods("POST").Handler(http.HandlerFunc(t.handlePlaintext))
	t.server.HTTP.Path("/ready").Methods("GET").Handler(http.HandlerFunc(t.ready))

	go func() {
		err := srv.Run()
		if err != nil {
			level.Error(t.logger).Log("msg", "loki.source.api server shutdown with error", "err", err)
		}
	}()

	return nil
}

func (t *PushTarget) handleLoki(w http.ResponseWriter, r *http.Request) {
	logger := util_log.WithContext(r.Context(), util_log.Logger)
	userID, _ := tenant.TenantID(r.Context())
	req, err := push.ParseRequest(logger, userID, r, nil)
	if err != nil {
		level.Warn(t.logger).Log("msg", "failed to parse incoming push request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var lastErr error
	for _, stream := range req.Streams {
		ls, err := promql_parser.ParseMetric(stream.Labels)
		if err != nil {
			lastErr = err
			continue
		}
		sort.Sort(ls)

		lb := labels.NewBuilder(ls)

		// Add configured labels
		for k, v := range t.config.Labels {
			lb.Set(string(k), string(v))
		}

		// Apply relabeling
		processed, keep := relabel.Process(lb.Labels(nil), t.config.RelabelConfig...)
		if !keep || len(processed) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Convert to model.LabelSet
		filtered := model.LabelSet{}
		for i := range processed {
			if strings.HasPrefix(processed[i].Name, "__") {
				continue
			}
			filtered[model.LabelName(processed[i].Name)] = model.LabelValue(processed[i].Value)
		}

		for _, entry := range stream.Entries {
			e := loki.Entry{
				Labels: filtered.Clone(),
				Entry: logproto.Entry{
					Line: entry.Line,
				},
			}
			if t.config.KeepTimestamp {
				e.Timestamp = entry.Timestamp
			} else {
				e.Timestamp = time.Now()
			}
			t.handler.Chan() <- e
		}
	}

	if lastErr != nil {
		level.Warn(t.logger).Log("msg", "at least one entry in the push request failed to process", "err", lastErr.Error())
		http.Error(w, lastErr.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handlePlaintext handles newline delimited input such as plaintext or NDJSON.
func (t *PushTarget) handlePlaintext(w http.ResponseWriter, r *http.Request) {
	entries := t.handler.Chan()
	defer r.Body.Close()
	body := bufio.NewReader(r.Body)
	for {
		line, err := body.ReadString('\n')
		if err != nil && err != io.EOF {
			level.Warn(t.logger).Log("msg", "failed to read incoming push request", "err", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if err == io.EOF {
				break
			}
			continue
		}
		entries <- loki.Entry{
			Labels: t.Labels().Clone(),
			Entry: logproto.Entry{
				Timestamp: time.Now(),
				Line:      line,
			},
		}
		if err == io.EOF {
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Labels returns the set of labels that statically apply to all log entries
// produced by the PushTarget.
func (t *PushTarget) Labels() model.LabelSet {
	return t.config.Labels
}

// Stop shuts down the PushTarget.
func (t *PushTarget) Stop() error {
	level.Info(t.logger).Log("msg", "stopping push server", "job", t.jobName)
	t.server.Shutdown()
	t.server.Stop() // Required to stop signal handler.
	t.handler.Stop()
	return nil
}

func (t *PushTarget) CurrentConfig() PushTargetConfig {
	return *t.config
}

// ready function serves the ready endpoint
func (t *PushTarget) ready(w http.ResponseWriter, r *http.Request) {
	resp := "ready"
	if _, err := w.Write([]byte(resp)); err != nil {
		level.Error(t.logger).Log("msg", "failed to respond to ready endoint", "err", err)
	}
}
