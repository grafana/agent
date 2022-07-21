package value

import (
	"reflect"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
)

// tagsCache caches the river tags for a struct type. This is never cleared,
// but since most structs will be statically created throughout the lifetime
// of the process, this will consume a negligible amount of memory.
var tagsCache = make(map[reflect.Type]objectFields)

func getCachedTags(t reflect.Type) objectFields {
	if t.Kind() != reflect.Struct {
		panic("getCachedTags called with non-struct type")
	}

	if entry, ok := tagsCache[t]; ok {
		return entry
	}

	ff := rivertags.Get(t)

	ofs := objectFields{
		lookup: make(map[string]rivertags.Field, len(ff)),
		keys:   make([]string, len(ff)),
	}
	for i, f := range ff {
		ofs.keys[i] = f.Name
		ofs.lookup[f.Name] = f
	}

	tagsCache[t] = ofs
	return ofs
}

type objectFields struct {
	lookup map[string]rivertags.Field
	keys   []string
}

func (ff objectFields) Get(name string) (rivertags.Field, bool) {
	f, ok := ff.lookup[name]
	return f, ok
}

func (ff objectFields) Len() int { return len(ff.lookup) }

// Index gets the field by index i. Panics if i < 0 or i >= ff.Len().
func (ff objectFields) Index(i int) rivertags.Field {
	return ff.lookup[ff.keys[i]]
}

func (ff objectFields) Keys() []string { return ff.keys }
