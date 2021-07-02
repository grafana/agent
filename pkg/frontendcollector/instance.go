package frontendcollector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-logfmt/logfmt"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/pkg/util/server"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

type Instance struct {
	cfg          *InstanceConfig
	mut          sync.Mutex
	l            log.Logger
	srv          *server.Server
	smap         SourceMapStore
	lokiInstance *loki.Instance
}

// NewInstance creates and starts a frontend collector instance.
func NewInstance(loki *loki.Loki, c *InstanceConfig, l log.Logger) (*Instance, error) {
	logger := log.With(l, "collector", c.Name)
	srv := server.New(prometheus.NewRegistry(), logger)
	inst := Instance{
		cfg: c,
		l:   logger,
		srv: srv,
		smap: SourceMapStore{
			l:               logger,
			cache:           make(map[string]*sourceMap),
			download:        c.DownloadSourcemaps,
			downloadTimeout: c.DownloadSourcemapTimeout,
		},
	}
	if err := inst.ApplyConfig(loki, c); err != nil {
		return nil, err
	}
	go func() {
		err := srv.Run()
		if err != nil {
			level.Error(logger).Log("msg", "Failed to start frontend collector", "err", err)
		}
	}()
	return &inst, nil
}

func (i *Instance) ApplyConfig(loki *loki.Loki, c *InstanceConfig) error {
	i.mut.Lock()
	defer i.mut.Unlock()
	c.Server.Log = util.GoKitLogger(i.l)
	err := i.srv.ApplyConfig(c.Server, i.wire)
	if err != nil {
		return err
	}
	i.cfg = c
	if len(c.LokiName) > 0 {
		i.lokiInstance = loki.Instance(c.LokiName)
		if i.lokiInstance == nil {
			return fmt.Errorf("loki instance %s not found", c.LokiName)
		}
	} else {
		i.lokiInstance = nil
	}
	return nil
}

func (i *Instance) sendEventToStdout(event FrontendSentryEvent) error {
	logctx := event.ToLogContext(&i.smap, i.l)
	keyvals := LogContextToKeyVals(logctx)
	for labelName, labelValue := range i.cfg.StaticLabels {
		keyvals = append([]interface{}{labelName, labelValue}, keyvals...)
	}
	switch event.Level {
	case sentry.LevelError:
		return level.Error(i.l).Log(keyvals...)
	case sentry.LevelWarning:
		return level.Warn(i.l).Log(keyvals...)
	case sentry.LevelDebug:
		return level.Debug(i.l).Log(keyvals...)
	default:
		return level.Info(i.l).Log(keyvals...)
	}
}

func (i *Instance) sendEventToLoki(event FrontendSentryEvent) error {
	logctx := event.ToLogContext(&i.smap, i.l)
	logctx["level"] = event.Level
	keyvals := LogContextToKeyVals(logctx)
	line, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		return err
	}
	labels := model.LabelSet{
		model.LabelName("collector"): model.LabelValue(i.cfg.Name),
	}

	for labelName, labelValue := range i.cfg.StaticLabels {
		labels[model.LabelName(labelName)] = model.LabelValue(labelValue)
	}

	sent := i.lokiInstance.SendEntry(api.Entry{
		Labels: labels,
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}, i.cfg.LokiTimeout)
	if !sent {
		return fmt.Errorf("Failed to send to loki")
	}
	return nil
}

func (i *Instance) HandleEvent(event FrontendSentryEvent) error {
	if len(event.GetKind()) == 0 {
		level.Debug(i.l).Log("msg", "skipping frontend event, unknown kind", "event_id", event.EventID)
		return nil
	}
	var err error = nil
	if i.cfg.LogToStdout {
		err = i.sendEventToStdout(event)
		if err != nil {
			level.Error(i.l).Log("msg", "error logging event", "err", err)
		}
	}
	if i.lokiInstance != nil {
		err = i.sendEventToLoki(event)
		if err != nil {
			level.Error(i.l).Log("msg", "error sending event to loki", "err", err)
		}
	}
	return err
}

func (i *Instance) handleHTTPEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}
		var evt FrontendSentryEvent
		err := json.NewDecoder(r.Body).Decode(&evt)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err.Error()), 400)
			return
		}

		i.HandleEvent(evt)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, "")
}

func (i *Instance) wire(mux *mux.Router, grpc *grpc.Server) {

	c := cors.New(cors.Options{
		AllowedOrigins:     i.cfg.AllowedOrigins,
		OptionsPassthrough: false,
		AllowedHeaders:     []string{"*"},
	})

	mux.Handle("/collect", c.Handler(rateLimit(i.cfg.RateLimitRPS, i.cfg.RateLimitBurst, time.Now, http.HandlerFunc(i.handleHTTPEvent))))
}

func (i *Instance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()
	i.srv.Close()
}
