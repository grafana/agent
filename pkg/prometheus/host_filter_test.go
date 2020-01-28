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
		group        *targetgroup.Group
		inputHost    string
		shouldRemove bool
	}{
		{
			name:         "__address__ match",
			group:        makeGroup([]model.LabelSet{{model.AddressLabel: "myhost"}}),
			inputHost:    "myhost",
			shouldRemove: false,
		},
		{
			name:         "__address__ mismatch",
			group:        makeGroup([]model.LabelSet{{model.AddressLabel: "notmyhost"}}),
			inputHost:    "myhost",
			shouldRemove: true,
		},
		{
			name:         "__address__ match with port",
			group:        makeGroup([]model.LabelSet{{model.AddressLabel: "myhost:12345"}}),
			inputHost:    "myhost",
			shouldRemove: false,
		},
		{
			name:         "__address__ mismatch with port",
			group:        makeGroup([]model.LabelSet{{model.AddressLabel: "notmyhost:12345"}}),
			inputHost:    "myhost",
			shouldRemove: true,
		},
		{
			name: "__host__ match",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "mycontainer",
				model.LabelName(hostAddressLabel): "myhost",
			}}),
			inputHost:    "myhost",
			shouldRemove: false,
		},
		{
			name: "__host__ mismatch",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "mycontainer",
				model.LabelName(hostAddressLabel): "notmyhost",
			}}),
			inputHost:    "myhost",
			shouldRemove: true,
		},
		{
			name: "__host__ match with port",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "mycontainer:12345",
				model.LabelName(hostAddressLabel): "myhost:12345",
			}}),
			inputHost:    "myhost",
			shouldRemove: false,
		},
		{
			name: "__host__ mismatch with port",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "mycontainer:12345",
				model.LabelName(hostAddressLabel): "notmyhost:12345",
			}}),
			inputHost:    "myhost",
			shouldRemove: true,
		},
		{
			name: "__host__ mismatch, __address__ match",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "mycontainer",
				model.LabelName(hostAddressLabel): "notmyhost",
			}}),
			inputHost:    "mycontainer",
			shouldRemove: false,
		},
		{
			name: "__host__ mismatch, __address__ match with port",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "mycontainer:12345",
				model.LabelName(hostAddressLabel): "notmyhost:12345",
			}}),
			inputHost:    "mycontainer",
			shouldRemove: false,
		},
		{
			name: "always allow localhost",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "localhost:12345",
				model.LabelName(hostAddressLabel): "notmyhost:12345",
			}}),
			inputHost:    "mycontainer",
			shouldRemove: false,
		},
		{
			name: "always allow 127.0.0.1",
			group: makeGroup([]model.LabelSet{{
				model.AddressLabel:                "127.0.0.1:12345",
				model.LabelName(hostAddressLabel): "notmyhost:12345",
			}}),
			inputHost:    "mycontainer",
			shouldRemove: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			groups := DiscoveredGroups{"test": []*targetgroup.Group{tc.group}}
			result := FilterGroups(groups, tc.inputHost)

			require.NotNil(t, result["test"])
			if tc.shouldRemove {
				require.NotEqual(t, len(result["test"][0].Targets), len(groups["test"][0].Targets))
			} else {
				require.Equal(t, len(result["test"][0].Targets), len(groups["test"][0].Targets))
			}
		})
	}
}
