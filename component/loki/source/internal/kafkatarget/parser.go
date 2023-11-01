package kafkatarget

import (
	"github.com/IBM/sarama"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
)

// KafkaTargetMessageParser implements MessageParser. It doesn't modify the content of the original `message.Value`.
type KafkaTargetMessageParser struct{}

func (p *KafkaTargetMessageParser) Parse(message *sarama.ConsumerMessage, labels model.LabelSet, relabels []*relabel.Config, useIncomingTimestamp bool) ([]loki.Entry, error) {
	return []loki.Entry{
		{
			Labels: labels,
			Entry: logproto.Entry{
				Timestamp: timestamp(useIncomingTimestamp, message.Timestamp),
				Line:      string(message.Value),
			},
		},
	}, nil
}
