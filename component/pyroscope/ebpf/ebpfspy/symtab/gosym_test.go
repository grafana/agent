package symtab

import (
	"testing"

	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/elf"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) { //todo remove or commit elf
	elfPath := "/Users/korniltsev/Desktop/go_elf"
	expectedSymbols, err := newGoSymbols(elfPath)
	require.NoError(t, err)

	me, err := elf.NewMMapedElfFile(elfPath)
	require.NoError(t, err)
	newTable, err := me.ReadGoSymbols()
	require.NoError(t, err)

	for _, symbol := range expectedSymbols.Symbols {
		name := newTable.Resolve(symbol.Start)
		require.Equal(t, symbol.Name, name)
	}

}
