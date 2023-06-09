package symtab

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) { //todo remove or commit elf
	symtab, err := newGoSymbols("/Users/korniltsev/Desktop/go_elf")
	require.NoError(t, err)
	_ = symtab
}
