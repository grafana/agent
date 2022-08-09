package scrape

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/metrics"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestForwardingToAppendable(t *testing.T) {
	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	opts := component.Options{
		Logger:     l,
		Registerer: prometheus.NewRegistry(),
	}

	nilReceivers := []*metrics.Receiver{nil, nil}

	args := Arguments{
		Targets:      []Target{},
		ForwardTo:    nilReceivers,
		ScrapeConfig: DefaultConfig,
	}

	s, err := New(opts, args)
	require.NoError(t, err)

	// List the Appendable's receivers; they are nil.
	require.Equal(t, nilReceivers, s.appendable.ListReceivers())

	// Forwarding samples to the nil receivers shouldn't fail.
	appender := s.appendable.Appender(context.Background())
	_, err = appender.Append(0, labels.FromStrings("foo", "bar"), 0, 0)
	require.NoError(t, err)

	err = appender.Commit()
	require.NoError(t, err)

	// Update the component with a mock receiver; it should be passed along to the Appendable.
	var receivedTs int64
	var receivedSamples []*metrics.FlowMetric
	mockReceiver := []*metrics.Receiver{
		{
			Receive: func(t int64, m []*metrics.FlowMetric) {
				receivedTs = t
				receivedSamples = m
			},
		},
	}

	args.ForwardTo = mockReceiver
	err = s.Update(args)
	require.NoError(t, err)

	require.Equal(t, mockReceiver, s.appendable.ListReceivers())

	// Forwarding a sample to the mock receiver should succeed.
	appender = s.appendable.Appender(context.Background())
	sample := metrics.NewFlowMetric(1, labels.FromStrings("foo", "bar"), 42.0)
	timestamp := time.Now().Unix()
	_, err = appender.Append(0, sample.LabelsCopy(), timestamp, sample.Value())
	require.NoError(t, err)

	err = appender.Commit()
	require.NoError(t, err)

	require.Equal(t, receivedTs, timestamp)
	require.Len(t, receivedSamples, 1)
	require.Equal(t, receivedSamples[0], sample)
}
