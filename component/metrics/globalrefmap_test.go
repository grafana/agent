package metrics

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestAddingMarker(t *testing.T) {
	mapping := GlobalRefMapping
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})
	globalID := mapping.CreateGlobalRefID(l)
	shouldBeSameGlobalID := mapping.CreateGlobalRefID(l)
	require.True(t, globalID == shouldBeSameGlobalID)
	require.Len(t, mapping.labelsHashToGlobal, 1)
}

func TestAddingDifferentMarkers(t *testing.T) {
	mapping := GlobalRefMapping
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})
	l2 := labels.Labels{}
	l2 = append(l, labels.Label{
		Name:  "__name__",
		Value: "roar",
	})
	globalID := mapping.CreateGlobalRefID(l)
	shouldBeDifferentID := mapping.CreateGlobalRefID(l2)
	require.True(t, globalID != shouldBeDifferentID)
	require.Len(t, mapping.labelsHashToGlobal, 2)
}

func TestAddingLocalMapping(t *testing.T) {
	mapping := GlobalRefMapping
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	globalID := mapping.CreateGlobalRefID(l)
	shouldBeSameGlobalID := mapping.GetOrAddLink("1", 1, l)
	require.True(t, globalID == shouldBeSameGlobalID)
	require.Len(t, mapping.labelsHashToGlobal, 1)
	require.Len(t, mapping.mappings, 1)
	require.True(t, mapping.mappings["1"].RemoteWriteID == "1")
	require.True(t, mapping.mappings["1"].globalToLocal[shouldBeSameGlobalID] == 1)
	require.True(t, mapping.mappings["1"].localToGlobal[1] == shouldBeSameGlobalID)
}

func TestAddingLocalMappings(t *testing.T) {
	mapping := GlobalRefMapping
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	globalID := mapping.CreateGlobalRefID(l)
	shouldBeSameGlobalID := mapping.GetOrAddLink("1", 1, l)
	shouldBeSameGlobalID2 := mapping.GetOrAddLink("2", 1, l)
	require.True(t, globalID == shouldBeSameGlobalID)
	require.True(t, globalID == shouldBeSameGlobalID2)
	require.Len(t, mapping.labelsHashToGlobal, 1)
	require.Len(t, mapping.mappings, 2)

	require.True(t, mapping.mappings["1"].RemoteWriteID == "1")
	require.True(t, mapping.mappings["1"].globalToLocal[shouldBeSameGlobalID] == 1)
	require.True(t, mapping.mappings["1"].localToGlobal[1] == shouldBeSameGlobalID)

	require.True(t, mapping.mappings["2"].RemoteWriteID == "2")
	require.True(t, mapping.mappings["2"].globalToLocal[shouldBeSameGlobalID2] == 1)
	require.True(t, mapping.mappings["2"].localToGlobal[1] == shouldBeSameGlobalID2)
}

func TestAddingLocalMappingsWithoutCreatingGlobalUpfront(t *testing.T) {
	mapping := GlobalRefMapping
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	shouldBeSameGlobalID := mapping.GetOrAddLink("1", 1, l)
	shouldBeSameGlobalID2 := mapping.GetOrAddLink("2", 1, l)
	require.True(t, shouldBeSameGlobalID2 == shouldBeSameGlobalID)
	require.Len(t, mapping.labelsHashToGlobal, 1)
	require.Len(t, mapping.mappings, 2)

	require.True(t, mapping.mappings["1"].RemoteWriteID == "1")
	require.True(t, mapping.mappings["1"].globalToLocal[shouldBeSameGlobalID] == 1)
	require.True(t, mapping.mappings["1"].localToGlobal[1] == shouldBeSameGlobalID)

	require.True(t, mapping.mappings["2"].RemoteWriteID == "2")
	require.True(t, mapping.mappings["2"].globalToLocal[shouldBeSameGlobalID2] == 1)
	require.True(t, mapping.mappings["2"].localToGlobal[1] == shouldBeSameGlobalID2)
}
