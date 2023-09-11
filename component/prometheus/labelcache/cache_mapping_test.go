package labelcache

import (
	"github.com/go-kit/log"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddingMarker(t *testing.T) {
	mapping := newCacheTest(t)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})
	globalID := mapping.GetOrAddGlobalRefID(l)
	shouldBeSameGlobalID := mapping.GetOrAddGlobalRefID(l)
	require.True(t, globalID == shouldBeSameGlobalID)
}

func TestAddingDifferentMarkers(t *testing.T) {
	mapping := newCacheTest(t)
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
}

func TestAddingLocalMapping(t *testing.T) {
	mapping := newCacheTest(t)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	globalID := mapping.GetOrAddGlobalRefID(l)
	shouldBeSameGlobalID := mapping.GetOrAddLink("1", 1, l)
	require.True(t, globalID == shouldBeSameGlobalID)

}

func TestAddingLocalMappings(t *testing.T) {
	mapping := newCacheTest(t)
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

}

func TestAddingLocalMappingsWithoutCreatingGlobalUpfront(t *testing.T) {
	mapping := newCacheTest(t)
	l := labels.Labels{}
	l = append(l, labels.Label{
		Name:  "__name__",
		Value: "test",
	})

	shouldBeSameGlobalID := mapping.GetOrAddLink("1", 1, l)
	shouldBeSameGlobalID2 := mapping.GetOrAddLink("2", 1, l)
	require.True(t, shouldBeSameGlobalID2 == shouldBeSameGlobalID)

}

func newCacheTest(t *testing.T) *Cache {
	l := log.NewNopLogger()
	return NewCache(t.TempDir(), l)
}
