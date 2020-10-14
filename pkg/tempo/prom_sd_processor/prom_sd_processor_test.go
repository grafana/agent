package prom_sd_processor

import (
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/relabel"
	"gotest.tools/assert"
)

func TestSyncGroups(t *testing.T) {

	/*	cfgs := map[string][]*relabel.Config{
				"test": {
					{
						Action:       "replace",
						SourceLabels: model.LabelNames([]model.LabelName{"blerg"}),
						TargetLabel:  "blarg",
						Regex:        relabel.MustNewRegexp(""),
					},
				},
			}

			p := &promServiceDiscoProcessor{
				logger:         log.NewNopLogger(),
				relabelConfigs: cfgs,
			}

			groups := []*targetgroup.Group{
				{
					Targets: []model.LabelSet{
						{
							"__address__": "127.0.0.1",
							"blerg":       "?",
						},
					},
				},
			}

			hostLabels := make(map[string]model.LabelSet

		p.syncGroups("test", groups, hostLabels)

		fmt.Println(hostLabels) */

	tests := []struct {
		name        string
		jobToSync   string
		relabelCfgs map[string][]*relabel.Config
		targets     []model.LabelSet
		expected    map[string]model.LabelSet
	}{
		{
			name:        "empty",
			jobToSync:   "",
			relabelCfgs: map[string][]*relabel.Config{},
			targets:     []model.LabelSet{},
			expected:    map[string]model.LabelSet{},
		},
		{
			name:      "no relabeling",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {},
			},
		},
		{
			name:      "strip port",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1:8888",
					"label":       "val",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {
					"label": "val",
				},
			},
		},
		{
			name:      "passthrough",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1",
					"label":       "val",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {
					"label": "val",
				},
			},
		},
		{
			name:      "ignore metadata",
			jobToSync: "job",
			relabelCfgs: map[string][]*relabel.Config{
				"job": {},
			},
			targets: []model.LabelSet{
				{
					"__address__": "127.0.0.1",
					"__ignore":    "ignore",
				},
			},
			expected: map[string]model.LabelSet{
				"127.0.0.1": {},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups := []*targetgroup.Group{
				{
					Targets: tc.targets,
				},
			}

			p := &promServiceDiscoProcessor{
				logger:         log.NewNopLogger(),
				relabelConfigs: tc.relabelCfgs,
			}

			hostLabels := make(map[string]model.LabelSet)
			p.syncGroups(tc.jobToSync, groups, hostLabels)

			assert.DeepEqual(t, tc.expected, hostLabels)
		})
	}
}
