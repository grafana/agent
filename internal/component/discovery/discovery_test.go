package discovery

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilteredLabels(t *testing.T) {
	target := Target{
		"instance":    "instanceTest",
		"__meta_test": "metaTest",
		"job":         "jobTest",
	}
	labels := target.FilteredLabels()
	require.Equal(t, labels.Len(), 1)
	require.True(t, labels.Has("job"))
}
