package aggregator

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

// Aggregator implement storage.Aggregator and runs a set of PromQL queries
// on the in-memory batch before committing to the underlying appendable.
type Aggregator struct {
	logger     log.Logger
	appendable storage.Appendable
	engine     *promql.Engine
	rules      []Rule
}

// Rule for scrape-time aggregation.
type Rule struct {
	Name   string            `yaml:"name"`
	Expr   string            `yaml:"expr"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

// New makes a fresh Aggregator.
func New(logger log.Logger, appendable storage.Appendable, rules []Rule) *Aggregator {
	engine := promql.NewEngine(promql.EngineOpts{
		Logger:        logger,
		MaxSamples:    1000000,
		Timeout:       time.Minute,
		LookbackDelta: 15 * time.Minute,
	})

	return &Aggregator{
		logger:     logger,
		appendable: appendable,
		engine:     engine,
		rules:      rules,
	}
}

// Appender implements storage.Appendable.
func (a *Aggregator) Appender(ctx context.Context) storage.Appender {
	return &batch{
		ctx:        ctx,
		aggregator: a,
		appender:   a.appendable.Appender(ctx),
	}
}

type batch struct {
	ctx        context.Context
	aggregator *Aggregator
	appender   storage.Appender
	samples    []sample
	ts         int64
}

func (b *batch) Add(l labels.Labels, t int64, v float64) (uint64, error) {
	b.samples = append(b.samples, sample{l, t, v})
	return b.appender.Add(l, t, v)
}

func (b *batch) AddFast(ref uint64, t int64, v float64) error {
	return b.appender.AddFast(ref, t, v)
}

func (b *batch) Commit() error {
	for _, r := range b.aggregator.rules {
		if err := b.execute(r); err != nil {
			return err
		}
	}

	return b.appender.Commit()
}

func (b *batch) Rollback() error {
	return b.appender.Commit()
}

func (b *batch) execute(rule Rule) error {
	q, err := b.aggregator.engine.NewInstantQuery(b, rule.Expr, time.Unix(0, b.ts*int64(time.Millisecond)))
	if err != nil {
		return err
	}
	defer q.Close()

	r := q.Exec(b.ctx)
	if r.Err != nil {
		return r.Err
	}

	v, err := r.Vector()
	if err != nil {
		return err
	}

	for i := range v {
		ls := v[i].Metric.Copy()
		ls = ls.WithoutLabels(labels.MetricName)
		ls = append(ls, labels.Label{Name: labels.MetricName, Value: rule.Name})
		for k, v := range rule.Labels {
			ls = append(ls, labels.Label{Name: k, Value: v})
		}

		_, err := b.Add(ls, v[i].T, v[i].V)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *batch) Querier(ctx context.Context, mint, maxt int64) (storage.Querier, error) {
	return &querier{
		ctx:  ctx,
		b:    b,
		mint: mint,
		maxt: maxt,
	}, nil
}

type querier struct {
	storage.LabelQuerier

	ctx        context.Context
	b          *batch
	mint, maxt int64
}

func (q *querier) Select(sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	return &seriesSet{
		q:        q,
		matchers: matchers,
		i:        -1,
	}
}

func (*querier) Close() error {
	return nil
}

type seriesSet struct {
	q        *querier
	matchers []*labels.Matcher
	i        int
}

func (s *seriesSet) Next() bool {
	s.i++
	for s.i < len(s.q.b.samples) {
		if matchLabels(s.q.b.samples[s.i].l, s.matchers) {
			return true
		}
		s.i++
	}
	return false
}

func matchLabels(lset labels.Labels, matchers []*labels.Matcher) bool {
	for _, m := range matchers {
		if !m.Matches(lset.Get(m.Name)) {
			return false
		}
	}
	return true
}

func (s *seriesSet) At() storage.Series {
	return &s.q.b.samples[s.i]
}

func (s *seriesSet) Err() error {
	return nil
}

func (s *seriesSet) Warnings() storage.Warnings {
	return nil
}

type sample struct {
	l labels.Labels
	t int64
	v float64
}

func (s *sample) Labels() labels.Labels {
	return s.l
}

func (s *sample) Iterator() chunkenc.Iterator {
	return &iter{s: s}
}

type iter struct {
	s    *sample
	next bool
}

func (i *iter) Next() bool {
	if i.next {
		return false
	}
	i.next = true
	return true
}

func (*iter) Seek(t int64) bool {
	return true
}

func (i *iter) At() (int64, float64) {
	return i.s.t, i.s.v
}

func (*iter) Err() error {
	return nil
}
