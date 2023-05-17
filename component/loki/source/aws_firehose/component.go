package aws_firehose

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/relabel"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.awsfirehose",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Arguments struct {
	Server       *fnet.ServerConfig  `river:",squash"`
	ForwardTo    []loki.LogsReceiver `river:"forward_to,attr"`
	RelabelRules flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

func (a *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*a = Arguments{}
	type arguments Arguments
	err := f((*arguments)(a))
	if err != nil {
		return err
	}

	return nil
}

type Component struct {
	// mut controls concurrent access to fanout
	mut    sync.RWMutex
	fanout []loki.LogsReceiver

	// handler is the main destination where the TargetServer writes received log entries to
	handler loki.LogsReceiver
	rbs     []*relabel.Config

	server *fnet.TargetServer

	opts component.Options
	args Arguments

	// utils
	serverMetrics *util.UncheckedCollector
	metrics       *metrics
	logger        log.Logger
}

func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts:          o,
		handler:       make(loki.LogsReceiver),
		fanout:        args.ForwardTo,
		serverMetrics: util.NewUncheckedCollector(nil),

		// todo(pablo): should use unchecked collector here?
		metrics: newMetrics(o.Registerer),
		logger:  log.With(o.Logger, "component", "aws_firehose_logs"),
	}

	o.Registerer.MustRegister(c.serverMetrics)

	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Component) Run(ctx context.Context) error {
	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()
		c.shutdownServer()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler:
			c.mut.RLock()
			for _, receiver := range c.fanout {
				receiver <- entry
			}
			c.mut.RUnlock()
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	var err error
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)
	c.fanout = newArgs.ForwardTo

	// todo(pablo): is it a good practice to keep a reference to the arguments in the
	// component struct, used for comparing here rather than destructuring them?
	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules) > 0 {
		c.rbs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules)
	}

	serverNeedsUpdate := !reflect.DeepEqual(c.args.Server, newArgs.Server)
	if !serverNeedsUpdate {
		c.args = newArgs
		return nil
	}

	c.shutdownServer()

	jobName := strings.Replace(c.opts.ID, ".", "_", -1)

	registry := prometheus.NewRegistry()
	c.serverMetrics.SetCollector(registry)

	wlog := log.With(c.logger, "component", "aws_firehose_logs")
	c.server, err = fnet.NewTargetServer(wlog, jobName, registry, newArgs.Server)
	if err != nil {
		return err
	}

	if err = c.server.MountAndRun(func(router *mux.Router) {
		router.Path("/api/v1/aws-firehose").Methods("POST").Handler(http.HandlerFunc(c.handle))
	}); err != nil {
		return err
	}

	c.args = newArgs
	return nil
}

// shutdownServer will shut down the currently used server.
// It is not goroutine-safe and mut write lock must be held when it's called.
func (c *Component) shutdownServer() {
	if c.server != nil {
		c.server.StopAndShutdown()
		c.server = nil
	}
}

type FirehoseRequest struct {
	RequestID string           `json:"requestId"`
	Timestamp int64            `json:"timestamp"`
	Records   []FirehoseRecord `json:"records"`
}

type FirehoseRecord struct {
	Data string `json:"data"`
}

func (c *Component) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	firehoseReq := FirehoseRequest{}

	bs, err := io.ReadAll(r.Body)
	if err != nil {
		c.metrics.errors.WithLabelValues("read_error").Inc()
		level.Warn(c.logger).Log("msg", "failed to read incoming request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &firehoseReq)
	if err != nil {
		c.metrics.errors.WithLabelValues("format").Inc()
		level.Warn(c.logger).Log("msg", "failed to unmarshall request", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, rec := range firehoseReq.Records {
		decodedRec, err := base64.StdEncoding.DecodeString(rec.Data)
		if err != nil {
			// handle
		}
		gzipReader, err := gzip.NewReader(bytes.NewReader(decodedRec))
		if err != nil {
			// handle
		}
		var sb strings.Builder
		if _, err := io.Copy(&sb, gzipReader); err != nil {
			// handle
		}

		//
		if err := gzipReader.Close(); err != nil {
			level.Error(c.logger).Log("msg", "failed to close gzip reader")
		}
	}
	//if err = pushMessage.Validate(); err != nil {
	//	p.metrics.gcpPushErrors.WithLabelValues("invalid_message").Inc()
	//	level.Warn(p.logger).Log("msg", "invalid gcp push request", "err", err.Error())
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}

	//entry, err := translate(pushMessage, p.Labels(), p.config.UseIncomingTimestamp, p.relabelConfigs, r.Header.Get("X-Scope-OrgID"))
	//if err != nil {
	//	p.metrics.gcpPushErrors.WithLabelValues("translation").Inc()
	//	level.Warn(p.logger).Log("msg", "failed to translate gcp push request", "err", err.Error())
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	//level.Debug(p.logger).Log("msg", fmt.Sprintf("Received line: %s", entry.Line))
	//
	//if err := p.doSendEntry(ctx, entry); err != nil {
	//	// NOTE: timeout errors can be tracked with from the metrics exposed by
	//	// the spun weaveworks server.
	//	// loki.source.gcplog.componentid_push_target_request_duration_seconds_count{status_code="503"}
	//	level.Warn(p.logger).Log("msg", "error sending log entry", "err", err.Error())
	//	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	//	return
	//}
	//
	c.metrics.entriesReceived.WithLabelValues().Inc()
	w.WriteHeader(http.StatusNoContent)
}
