package labelstore

import (
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestAddingMarker(t *testing.T) {
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})
	globalID := mapping.GetOrAddGlobalRefID(l)
	shouldBeSameGlobalID := mapping.GetOrAddGlobalRefID(l)
	require.True(t, globalID == shouldBeSameGlobalID)
	require.Len(t, mapping.labelsHashToGlobal, 1)
}

func TestAddingDifferentMarkers(t *testing.T) {
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})
	l2 := labels.Labels{}
	l2 = append(l2, labels.Label{
		Name:  "__name__",
		Value: "roar",
	})
	globalID := mapping.GetOrAddGlobalRefID(l)
	shouldBeDifferentID := mapping.GetOrAddGlobalRefID(l2)
	require.True(t, globalID != shouldBeDifferentID)
	require.Len(t, mapping.labelsHashToGlobal, 2)
}

func TestAddingLocalMapping(t *testing.T) {
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	globalID := mapping.GetOrAddGlobalRefID(l)
	shouldBeSameGlobalID := mapping.GetOrAddLink("1", 1, l)
	require.True(t, globalID == shouldBeSameGlobalID)
	require.Len(t, mapping.labelsHashToGlobal, 1)
	require.Len(t, mapping.mappings, 1)
	require.True(t, mapping.mappings["1"].RemoteWriteID == "1")
	require.True(t, mapping.mappings["1"].globalToLocal[shouldBeSameGlobalID] == 1)
	require.True(t, mapping.mappings["1"].localToGlobal[1] == shouldBeSameGlobalID)
}

func TestAddingLocalMappings(t *testing.T) {
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	globalID := mapping.GetOrAddGlobalRefID(l)
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
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
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

func TestStaleness(t *testing.T) {
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})
	l2 := labels.Labels{}
	l2 = append(l2, labels.Label{
		Name:  "__name__",
		Value: "test2",
	})

	global1 := mapping.GetOrAddLink("1", 1, l)
	_ = mapping.GetOrAddLink("2", 1, l2)
	mapping.AddStaleMarker(global1, l)
	require.Len(t, mapping.staleGlobals, 1)
	require.Len(t, mapping.labelsHashToGlobal, 2)
	staleDuration = 1 * time.Millisecond
	time.Sleep(10 * time.Millisecond)
	mapping.CheckAndRemoveStaleMarkers()
	require.Len(t, mapping.staleGlobals, 0)
	require.Len(t, mapping.labelsHashToGlobal, 1)
}

func TestRemovingStaleness(t *testing.T) {
	mapping := New(log.NewNopLogger(), prometheus.DefaultRegisterer)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	global1 := mapping.GetOrAddLink("1", 1, l)
	mapping.AddStaleMarker(global1, l)
	require.Len(t, mapping.staleGlobals, 1)
	mapping.RemoveStaleMarker(global1)
	require.Len(t, mapping.staleGlobals, 0)
}
