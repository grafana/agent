package automaticloggingprocessor

import (
	"context"
	"fmt"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/promtail/api"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

type promServiceDiscoProcessor struct {
	nextConsumer consumer.TracesConsumer
	cfg          *AutomaticLoggingConfig
	lokiChan     chan<- api.Entry

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.TracesConsumer, cfg *AutomaticLoggingConfig) (component.TracesProcessor, error) {
	logger := log.With(util.Logger, "component", "tempo automatic logging")

	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}
	return &promServiceDiscoProcessor{
		nextConsumer: nextConsumer,
		cfg:          cfg,
		logger:       logger,
	}, nil
}

func (p *promServiceDiscoProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	p.lokiChan <- api.Entry{ // do something real
		Labels: model.LabelSet{
			"test": "test",
		},
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      "ooga booga",
		},
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *promServiceDiscoProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{}
}

// Start is invoked during service startup.
func (p *promServiceDiscoProcessor) Start(ctx context.Context, _ component.Host) error {
	loki := ctx.Value(contextkeys.Loki).(*loki.Loki)
	if loki == nil {
		return fmt.Errorf("key %s does not contain a Loki instance", contextkeys.Loki)
	}
	lokiInstance := loki.Instance(p.cfg.LokiName)
	if lokiInstance == nil {
		return fmt.Errorf("loki instance %s not found", p.cfg.LokiName)
	}
	p.lokiChan = lokiInstance.Promtail().Client().Chan()
	if p.lokiChan == nil {
		return fmt.Errorf("loki chan is unexpectedly nil")
	}
	return nil
}

// Shutdown is invoked during service shutdown.
func (p *promServiceDiscoProcessor) Shutdown(context.Context) error {
	return nil
}
