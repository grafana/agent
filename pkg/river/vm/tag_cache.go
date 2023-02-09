package vm

import (
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
)

// tagsCache caches the river tags for a struct type. This is never cleared,
// but since most structs will be statically created throughout the lifetime
// of the process, this will consume a negligible amount of memory.
var tagsCache = make(map[reflect.Type]*tagInfo)

func getCachedTagInfo(t reflect.Type) *tagInfo {
	if t.Kind() != reflect.Struct {
		panic("getCachedTagInfo called with non-struct type")
	}

	if entry, ok := tagsCache[t]; ok {
		return entry
	}

	tfs := rivertags.Get(t)
	ti := &tagInfo{
		Tags:      tfs,
		TagLookup: make(map[string]rivertags.Field, len(tfs)),
	}

	for _, tf := range tfs {
		fullName := strings.Join(tf.Name, ".")
		ti.TagLookup[fullName] = tf
	}

	tagsCache[t] = ti
	return ti
}

type tagInfo struct {
	Tags      []rivertags.Field
	TagLookup map[string]rivertags.Field
}
