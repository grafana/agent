package lokipush

import (
	"bufio"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	frelabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/dskit/tenant"
	"github.com/grafana/loki/pkg/loghttp/push"
	"github.com/grafana/loki/pkg/logproto"
	util_log "github.com/grafana/loki/pkg/util/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	promql_parser "github.com/prometheus/prometheus/promql/parser"
)

type PushAPIServer struct {
	logger       log.Logger
	serverConfig *fnet.ServerConfig
	server       *fnet.TargetServer
	handler      loki.EntryHandler

	rwMutex       sync.RWMutex
	labels        model.LabelSet
	relabelRules  []*relabel.Config
	keepTimestamp bool
}

func NewPushAPIServer(logger log.Logger,
	serverConfig *fnet.ServerConfig,
	handler loki.EntryHandler,
	registerer prometheus.Registerer,
) (*PushAPIServer, error) {

	s := &PushAPIServer{
		logger:       logger,
		serverConfig: serverConfig,
		handler:      handler,
	}

	srv, err := fnet.NewTargetServer(logger, "loki_source_api", registerer, serverConfig)
	if err != nil {
		return nil, err
	}

	s.server = srv
	return s, nil
}

func (s *PushAPIServer) Run() error {
	level.Info(s.logger).Log("msg", "starting push API server")

	err := s.server.MountAndRun(func(router *mux.Router) {
		// This redirecting is so we can avoid breaking changes where we originally implemented it with
		// the loki prefix.
		router.Path("/api/v1/push").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = "/loki/api/v1/push"
			r.RequestURI = "/loki/api/v1/push"
			s.handleLoki(w, r)
		}))
		router.Path("/api/v1/raw").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = "/loki/api/v1/raw"
			r.RequestURI = "/loki/api/v1/raw"
			s.handlePlaintext(w, r)
		}))
		router.Path("/ready").Methods("GET").Handler(http.HandlerFunc(s.ready))
		router.Path("/loki/api/v1/push").Methods("POST").Handler(http.HandlerFunc(s.handleLoki))
		router.Path("/loki/api/v1/raw").Methods("POST").Handler(http.HandlerFunc(s.handlePlaintext))
	})
	return err
}

func (s *PushAPIServer) ServerConfig() fnet.ServerConfig {
	return *s.serverConfig
}

func (s *PushAPIServer) Shutdown() {
	level.Info(s.logger).Log("msg", "stopping push API server")
	s.server.StopAndShutdown()
}

func (s *PushAPIServer) SetLabels(labels model.LabelSet) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.labels = labels
}

func (s *PushAPIServer) getLabels() model.LabelSet {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return s.labels.Clone()
}

func (s *PushAPIServer) SetKeepTimestamp(keepTimestamp bool) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.keepTimestamp = keepTimestamp
}

func (s *PushAPIServer) getKeepTimestamp() bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return s.keepTimestamp
}

func (s *PushAPIServer) SetRelabelRules(rules frelabel.Rules) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.relabelRules = frelabel.ComponentToPromRelabelConfigs(rules)
}

func (s *PushAPIServer) getRelabelRules() []*relabel.Config {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	newRules := make([]*relabel.Config, len(s.relabelRules))
	for i, r := range s.relabelRules {
		var rCopy = *r
		newRules[i] = &rCopy
	}
	return newRules
}

// NOTE: This code is copied from Promtail (3478e180211c17bfe2f3f3305f668d5520f40481) with changes kept to the minimum.
// Only the HTTP handler functions are copied to allow for flow-specific server configuration and lifecycle management.
func (s *PushAPIServer) handleLoki(w http.ResponseWriter, r *http.Request) {
	logger := util_log.WithContext(r.Context(), util_log.Logger)
	userID, _ := tenant.TenantID(r.Context())
	req, err := push.ParseRequest(logger, userID, r, nil)
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to parse incoming push request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Take snapshot of current configs and apply consistently for the entire request.
	addLabels := s.getLabels()
	relabelRules := s.getRelabelRules()
	keepTimestamp := s.getKeepTimestamp()

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
		for k, v := range addLabels {
			lb.Set(string(k), string(v))
		}

		// Apply relabeling
		processed, keep := relabel.Process(lb.Labels(), relabelRules...)
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
			if keepTimestamp {
				e.Timestamp = entry.Timestamp
			} else {
				e.Timestamp = time.Now()
			}
			s.handler.Chan() <- e
		}
	}

	if lastErr != nil {
		level.Warn(s.logger).Log("msg", "at least one entry in the push request failed to process", "err", lastErr.Error())
		http.Error(w, lastErr.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// NOTE: This code is copied from Promtail (3478e180211c17bfe2f3f3305f668d5520f40481) with changes kept to the minimum.
// Only the HTTP handler functions are copied to allow for flow-specific server configuration and lifecycle management.
func (s *PushAPIServer) handlePlaintext(w http.ResponseWriter, r *http.Request) {
	entries := s.handler.Chan()
	defer r.Body.Close()
	body := bufio.NewReader(r.Body)
	addLabels := s.getLabels()
	for {
		line, err := body.ReadString('\n')
		if err != nil && err != io.EOF {
			level.Warn(s.logger).Log("msg", "failed to read incoming push request", "err", err.Error())
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
			Labels: addLabels,
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

// NOTE: This code is copied from Promtail (3478e180211c17bfe2f3f3305f668d5520f40481) with changes kept to the minimum.
// Only the HTTP handler functions are copied to allow for flow-specific server configuration and lifecycle management.
func (s *PushAPIServer) ready(w http.ResponseWriter, r *http.Request) {
	resp := "ready"
	if _, err := w.Write([]byte(resp)); err != nil {
		level.Error(s.logger).Log("msg", "failed to respond to ready endoint", "err", err)
	}
}
