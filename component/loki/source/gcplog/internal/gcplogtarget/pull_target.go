package gcplogtarget

// This code is copied from Promtail. The gcplogtarget package is used to
// configure and run the targets that can read log entries from cloud resource
// logs like bucket logs, load balancer logs, and Kubernetes cluster logs
// from GCP.

import (
	"context"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"google.golang.org/api/option"
)

// PullTarget represents a target that scrapes logs from a GCP project id and
// subscription and converts them to Loki log entries.
type PullTarget struct {
	// why was this here above?? nolint:revive
	metrics       *Metrics
	logger        log.Logger
	handler       loki.EntryHandler
	config        *PullConfig
	relabelConfig []*relabel.Config
	jobName       string

	// lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// pubsub
	ps   *pubsub.Client
	msgs chan *pubsub.Message
}

// NewPullTarget returns the new instance of PullTarget.
func NewPullTarget(metrics *Metrics, logger log.Logger, handler loki.EntryHandler, jobName string, config *PullConfig, relabel []*relabel.Config, clientOptions ...option.ClientOption) (*PullTarget, error) {
	// why was this here above?? nolint:revive,govet
	ctx, cancel := context.WithCancel(context.Background())

	ps, err := pubsub.NewClient(ctx, config.ProjectID, clientOptions...)
	if err != nil {
		return nil, err
	}

	target := &PullTarget{
		metrics:       metrics,
		logger:        logger,
		handler:       handler,
		relabelConfig: relabel,
		config:        config,
		jobName:       jobName,
		ctx:           ctx,
		cancel:        cancel,
		ps:            ps,
		msgs:          make(chan *pubsub.Message),
	}

	go func() {
		err := target.run()
		if err != nil {
			level.Error(logger).Log("msg", "loki.source.gcplog pull target shutdown with error", "err", err)
		}
	}()

	return target, nil
}

func (t *PullTarget) run() error {
	t.wg.Add(1)
	defer t.wg.Done()

	send := t.handler.Chan()

	sub := t.ps.SubscriptionInProject(t.config.Subscription, t.config.ProjectID)
	go func() {
		// NOTE(kavi): `cancel` the context as exiting from this goroutine should stop main `run` loop
		// It makesense as no more messages will be received.
		defer t.cancel()

		err := sub.Receive(t.ctx, func(ctx context.Context, m *pubsub.Message) {
			t.msgs <- m
		})
		if err != nil {
			level.Error(t.logger).Log("msg", "failed to receive pubsub messages", "error", err)
			t.metrics.gcplogErrors.WithLabelValues(t.config.ProjectID).Inc()
			t.metrics.gcplogTargetLastSuccessScrape.WithLabelValues(t.config.ProjectID, t.config.Subscription).SetToCurrentTime()
		}
	}()

	lbls := make(model.LabelSet, len(t.config.Labels))
	for k, v := range t.config.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}

	for {
		select {
		case <-t.ctx.Done():
			return t.ctx.Err()
		case m := <-t.msgs:
			entry, err := parseGCPLogsEntry(m.Data, lbls, nil, t.config.UseIncomingTimestamp, t.relabelConfig)
			if err != nil {
				level.Error(t.logger).Log("event", "error formating log entry", "cause", err)
				m.Ack()
				break
			}
			send <- entry
			m.Ack() // Ack only after log is sent.
			t.metrics.gcplogEntries.WithLabelValues(t.config.ProjectID).Inc()
		}
	}
}

// Labels return the model.LabelSet that the target applies to log entries.
func (t *PullTarget) Labels() model.LabelSet {
	lbls := make(model.LabelSet, len(t.config.Labels))
	for k, v := range t.config.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}
	return lbls
}

// Details returns some debug information about the target.
func (t *PullTarget) Details() map[string]string {
	return map[string]string{
		"strategy": "pull",
		"labels":   t.Labels().String(),
	}
}

// Stop shuts the target down.
func (t *PullTarget) Stop() error {
	t.cancel()
	t.wg.Wait()
	t.handler.Stop()
	t.ps.Close()
	return nil
}

// Used to add a mock pubsub clients for testing.
func (t *PullTarget) setPubSubClientAndMessageChan(cl *pubsub.Client, ch chan *pubsub.Message) {
	t.ps = cl
	t.msgs = ch
}
