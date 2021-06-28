package frontendcollector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/pkg/util/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

type Instance struct {
	cfg  *InstanceConfig
	mut  sync.Mutex
	l    log.Logger
	srv  *server.Server
	smap SourceMapStore
}

// NewInstance creates and starts a frontend collector instance.
func NewInstance(c *InstanceConfig, l log.Logger) (*Instance, error) {
	logger := log.With(l, "collector", c.Name)
	srv := server.New(prometheus.NewRegistry(), logger)
	inst := Instance{
		cfg: c,
		l:   logger,
		srv: srv,
	}
	if err := inst.ApplyConfig(c); err != nil {
		return nil, err
	}
	go func() {
		err := srv.Run()
		fmt.Println("stop", err)
		if err != nil {
			level.Error(logger).Log("msg", "Failed to start frontend collector", "err", err)
		}
	}()
	return &inst, nil
}

func (i *Instance) ApplyConfig(c *InstanceConfig) error {
	i.mut.Lock()
	defer i.mut.Unlock()
	c.Server.Log = util.GoKitLogger(i.l)
	err := i.srv.ApplyConfig(c.Server, i.wire)
	if err != nil {
		return err
	}
	i.cfg = c
	return nil
}

func (i *Instance) wire(mux *mux.Router, grpc *grpc.Server) {

	c := cors.New(cors.Options{
		AllowedOrigins: i.cfg.AllowedOrigins,
	})

	mux.Handle("/collect", c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		logctx := evt.ToLogContext(&i.smap, i.l)
		keyvals := LogContextToKeyVals(logctx)
		if i.cfg.LogToStdout {
			level.Info(i.l).Log(keyvals...)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})))
}

func (i *Instance) Stop() {
	i.mut.Lock()
	defer i.mut.Unlock()
	i.srv.Close()
}
