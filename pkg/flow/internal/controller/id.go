package controller

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// ComponentID is a fully-qualified name of a component. Each element in
// ComponentID corresponds to a fragment of the period-delimited string;
// "remote.http.example" is ComponentID{"remote", "http", "example"}.
type ComponentID []string

// BlockComponentID returns the ComponentID specified by an HCL block.
func BlockComponentID(b *hcl.Block) ComponentID {
	id := make(ComponentID, 0, 1+len(b.Labels)) // add 1 for the block type
	id = append(id, b.Type)
	id = append(id, b.Labels...)
	return id
}

// String returns the string representation of a component ID.
func (id ComponentID) String() string {
	return strings.Join(id, ".")
}

// Equals returns true if id == other.
func (id ComponentID) Equals(other ComponentID) bool {
	if len(id) != len(other) {
		return false
	}
	for i := 0; i < len(id); i++ {
		if id[i] != other[i] {
			return false
		}
	}
	return true
}
