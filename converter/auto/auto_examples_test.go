package auto

import (
	"fmt"
	"github.com/grafana/agent/component/loki/source/kafka"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKafka(t *testing.T) {
	from := kafka.Arguments{
		Brokers:  []string{"localhost:9092", "localhost:9093"},
		Topics:   []string{"topic1", "topic2"},
		GroupID:  "group1",
		Assignor: "assignor1",
		Version:  "1.0.0",
		Authentication: kafka.KafkaAuthentication{
			Type: "none",
		},
		UseIncomingTimestamp: false,
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	to := scrapeconfig.KafkaTargetConfig{}

	err := ConvertByFieldNames(&from, &to, RiverToYaml)
	require.NoError(t, err)

	fmt.Printf("%+v\n", to)
}
