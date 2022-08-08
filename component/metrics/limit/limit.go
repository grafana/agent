package limit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	component.Register(component.Registration{
		Name:    "metrics.limit",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Limit struct {
	mut      sync.Mutex
	opts     component.Options
	args     Arguments
	health   component.Health
	receiver *metrics.Receiver

	labelLengthLimit  prometheus.Counter
	numberLabelsLimit prometheus.Counter
}

type Arguments struct {
	LabelLimit  int                 `river:"label_limit,attr,optional"`
	LengthLimit int                 `river:"length_limit,attr,optional"`
	ForwardTo   []*metrics.Receiver `river:"forward_to,attr"`
}

type Exports struct {
	Receiver *metrics.Receiver `river:"receiver,attr"`
}

var (
	_ component.Component       = (*Limit)(nil)
	_ component.HealthComponent = (*Limit)(nil)
)

func New(o component.Options, args Arguments) (*Limit, error) {
	l := &Limit{
		opts:   o,
		args:   args,
		health: component.Health{},
		labelLengthLimit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_metrics_limit_label_length_exceeded_total",
			Help: "This metric is when the length of an individual label exceeds the amount specified",
		}),
		numberLabelsLimit: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_metrics_limit_number_labels_exceeded_total",
			Help: "This metrics is when the number of labels exceeds the amount specified",
		}),
	}

	l.receiver = &metrics.Receiver{Receive: l.Receive}
	o.Registerer.MustRegister(l.labelLengthLimit, l.numberLabelsLimit)
	err := l.Update(args)
	if err != nil {
		return nil, err
	}
	l.opts.OnStateChange(Exports{
		Receiver: l.receiver,
	})
	return l, nil
}

func (l *Limit) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (l *Limit) Update(args component.Arguments) error {
	l.mut.Lock()
	defer l.mut.Unlock()

	newArgs := args.(Arguments)
	if newArgs.LabelLimit <= 0 && newArgs.LengthLimit <= 0 {
		return fmt.Errorf("label limit or length limit must be greater than 0")
	}

	l.health.Health = component.HealthTypeHealthy
	l.health.Message = "limit loaded successfully"
	l.health.UpdateTime = time.Now()

	return nil
}

func (l *Limit) CurrentHealth() component.Health {
	l.mut.Lock()
	defer l.mut.Unlock()

	return l.health
}

func (l *Limit) Receive(ts int64, metricsArr []*metrics.FlowMetric) {
	l.mut.Lock()
	defer l.mut.Unlock()

	passedMetrics := make([]*metrics.FlowMetric, 0)
	for _, m := range metricsArr {
		normalLabels := 0
		keep := true
		for _, la := range m.Labels {
			// Only count non-private labels
			if !strings.HasSuffix(la.Name, "__") {
				normalLabels++
			}
			if l.args.LengthLimit != 0 {
				if len(la.Value) > l.args.LengthLimit {
					l.labelLengthLimit.Inc()
					keep = false
					break
				}
			}
		}

		if l.args.LabelLimit != 0 {
			if len(m.Labels) > normalLabels {
				l.numberLabelsLimit.Inc()
				keep = false
				break
			}
		}
		if keep {
			passedMetrics = append(passedMetrics, m)
		}
	}

	if len(passedMetrics) == 0 {
		return
	}
	for _, forward := range l.args.ForwardTo {
		forward.Receive(ts, passedMetrics)
	}
}
