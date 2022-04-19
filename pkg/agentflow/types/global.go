package types

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type Global struct {
	Log            log.Logger
	RootContext    *actor.RootContext
	MetricRegistry prometheus.Registerer
	Mux            *mux.Router
}
