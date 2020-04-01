package prometheus

import (
	"testing"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/stretchr/testify/require"
)

func makeGroup(labels []model.LabelSet) *targetgroup.Group {
	return &targetgroup.Group{
		Targets: labels,
		Labels:  model.LabelSet{},
	}
}

func TestFilterGroups(t *testing.T) {
	tt := []struct {
		name         string
		labelHost    string
		inputHost    string
		shouldRemove bool
	}{
		{
			name:         "complete match",
			labelHost:    "myhost",
			inputHost:    "myhost",
			shouldRemove: false,
		},
		{
			name:         "mismatch",
			labelHost:    "notmyhost",
			inputHost:    "myhost",
			shouldRemove: true,
		},
		{
			name:         "match with port",
			labelHost:    "myhost:12345",
			inputHost:    "myhost",
			shouldRemove: false,
		},
		{
			name:         "mismatch with port",
			labelHost:    "notmyhost:12345",
			inputHost:    "myhost",
			shouldRemove: true,
		},
	}

	// Sets of labels we want to test against.
	labels := []model.LabelName{
		model.AddressLabel,
		model.LabelName(kubernetesNodeNameLabel),
		model.LabelName(kubernetesPodNodeNameLabel),
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for _, label := range labels {
				t.Run(string(label), func(t *testing.T) {
					lset := model.LabelSet{
						label: model.LabelValue(tc.labelHost),
					}

					// Special case: if label is not model.AddressLabel, we need to give
					// it a fake value. model.AddressLabel is always expected to be present and
					// is considered an error if it isn't.
					if label != model.AddressLabel {
						lset[model.AddressLabel] = "fake"
					}

					group := makeGroup([]model.LabelSet{lset})

					groups := DiscoveredGroups{"test": []*targetgroup.Group{group}}
					result := FilterGroups(groups, tc.inputHost)

					require.NotNil(t, result["test"])
					if tc.shouldRemove {
						require.NotEqual(t, len(result["test"][0].Targets), len(groups["test"][0].Targets))
					} else {
						require.Equal(t, len(result["test"][0].Targets), len(groups["test"][0].Targets))
					}
				})
			}
		})
	}
}
