package generator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
)

func init() {
	component.Register(component.Registration{
		Name:    "metrics.generator",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

type Generator struct {
	mut           sync.Mutex
	opts          component.Options
	args          Arguments
	health        component.Health
	refreshTicker *time.Ticker

	generatedMetricTotal prometheus.Counter
}

type Arguments struct {
	// This will be the optional labels, the total amount of labels will be 2 more
	// __name__ and series
	LabelCount     int           `river:"label_count,attr,optional"`
	LabelLength    int           `river:"label_length,attr,optional"`
	SeriesCount    int           `river:"series_count,attr,optional"`
	MetricCount    int           `river:"metric_count,attr,optional"`
	ScrapeInterval time.Duration `river:"scraper_interval,attr,optional"`

	ForwardTo []*metrics.Receiver `river:"forward_to,attr"`
}

func (a *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*a = Arguments{
		ScrapeInterval: 1 * time.Minute,
		SeriesCount:    1,
		LabelLength:    1,
		LabelCount:     1,
		MetricCount:    1,
	}
	type arguments Arguments
	return f((*arguments)(a))
}

type Exports struct {
}

var (
	_ component.Component       = (*Generator)(nil)
	_ component.HealthComponent = (*Generator)(nil)
)

func New(o component.Options, args Arguments) (*Generator, error) {
	l := &Generator{
		opts:   o,
		args:   args,
		health: component.Health{},
		generatedMetricTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_metrics_generator_created_metrics_total",
			Help: "The total number of metrics generater",
		}),
		refreshTicker: time.NewTicker(args.ScrapeInterval),
	}

	o.Registerer.MustRegister(l.generatedMetricTotal)
	err := l.Update(args)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Generator) Run(ctx context.Context) error {
	l.mut.Lock()
	refreshTicker := time.NewTicker(l.args.ScrapeInterval)
	l.mut.Unlock()

	for {
		select {
		case <-refreshTicker.C:
			metricsArr := l.generate()
			for _, r := range l.args.ForwardTo {
				r.Receive(time.Now().Unix(), metricsArr)
			}
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func (l *Generator) Update(args component.Arguments) error {
	l.mut.Lock()
	defer l.mut.Unlock()

	newArgs := args.(Arguments)
	l.args = newArgs
	l.refreshTicker.Reset(newArgs.ScrapeInterval)
	l.health.Health = component.HealthTypeHealthy
	l.health.Message = "generator updated successfully"
	l.health.UpdateTime = time.Now()

	return nil
}

func (l *Generator) CurrentHealth() component.Health {
	l.mut.Lock()
	defer l.mut.Unlock()

	return l.health
}

func (l *Generator) generate() []*metrics.FlowMetric {
	l.mut.Lock()
	defer l.mut.Unlock()

	return l.generateDynamic()
}

func (l *Generator) generateDynamic() []*metrics.FlowMetric {
	metricArr := make([]*metrics.FlowMetric, l.args.MetricCount)

	for i := 0; i < l.args.MetricCount; i++ {
		m := &metrics.FlowMetric{
			Labels: make([]labels.Label, 0),
			Value:  rand.Float64(),
		}
		m.Labels = append(m.Labels, labels.Label{
			Name:  "__name__",
			Value: fmt.Sprintf("metric_%d", i),
		})
		metricArr[i] = m
		for seriesIndex := 0; seriesIndex < l.args.SeriesCount; seriesIndex++ {
			for labelIndex := 0; labelIndex < l.args.LabelCount; labelIndex++ {
				m.Labels = append(m.Labels, labels.Label{
					Name:  fmt.Sprintf("series_%d_label_%d", seriesIndex, labelIndex),
					Value: RandomString(l.args.LabelLength),
				})
			}
		}

	}
	return metricArr

}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
