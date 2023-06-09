package elf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveOutOfBounds(t *testing.T) {
	elfPath := "/Users/korniltsev/Desktop/go_elf"
	me, err := NewMMapedElfFile(elfPath)
	require.NoError(t, err)
	newTable, err := me.ReadGoSymbols()
	require.NoError(t, err)

	name := newTable.Resolve(newTable.Symbols[0].Entry - 1)
	require.Empty(t, name)
	name = newTable.Resolve(newTable.Symbols[len(newTable.Symbols)-1].End)
	require.Empty(t, name)
}
