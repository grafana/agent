package scrape

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestManager(t *testing.T) {
	reloadInterval = time.Millisecond

	m := NewManager(pyroscope.AppendableFunc(func(ctx context.Context, labels labels.Labels, samples []*pyroscope.RawSample) error {
		return nil
	}), util.TestLogger(t))

	defer m.Stop()
	targetSetsChan := make(chan map[string][]*targetgroup.Group)
	require.NoError(t, m.ApplyConfig(NewDefaultArguments()))
	go m.Run(targetSetsChan)

	targetSetsChan <- map[string][]*targetgroup.Group{
		"group1": {
			{
				Targets: []model.LabelSet{
					{model.AddressLabel: "localhost:9090", serviceNameLabel: "s"},
					{model.AddressLabel: "localhost:8080", serviceNameK8SLabel: "k"},
				},
				Labels: model.LabelSet{"foo": "bar"},
			},
		},
	}
	require.Eventually(t, func() bool {
		return len(m.TargetsActive()["group1"]) == 10
	}, time.Second, 10*time.Millisecond)

	new := NewDefaultArguments()
	new.ScrapeInterval = 1 * time.Second

	// Trigger a config reload
	require.NoError(t, m.ApplyConfig(new))

	targetSetsChan <- map[string][]*targetgroup.Group{
		"group2": {
			{
				Targets: []model.LabelSet{
					{model.AddressLabel: "localhost:9090", serviceNameLabel: "s"},
					{model.AddressLabel: "localhost:8080", serviceNameK8SLabel: "k"},
				},
				Labels: model.LabelSet{"foo": "bar"},
			},
		},
	}

	require.Eventually(t, func() bool {
		return len(m.TargetsActive()["group2"]) == 10
	}, time.Second, 10*time.Millisecond)

	for _, ts := range m.targetsGroups {
		require.Equal(t, 1*time.Second, ts.config.ScrapeInterval)
	}

	targetSetsChan <- map[string][]*targetgroup.Group{"group1": {}, "group2": {}}

	require.Eventually(t, func() bool {
		return len(m.TargetsAll()["group2"]) == 0 && len(m.TargetsAll()["group1"]) == 0
	}, time.Second, 10*time.Millisecond)
}
