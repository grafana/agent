package elf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelfGoSymbolComparison(t *testing.T) {
	testGoSymbolTable := func(t *testing.T, expectedSymbols []TestSym, goTable *GoTable) {
		for _, symbol := range expectedSymbols {
			name := goTable.Resolve(symbol.Start)
			require.Equal(t, symbol.Name, name)
		}

		name := goTable.Resolve(uint64(goTable.Index.Entry32[0]) - 1)
		require.Empty(t, name)
		name = goTable.Resolve(goTable.Index.End)
		require.Empty(t, name)
		name = goTable.Resolve(goTable.Index.End + 1)
		require.Empty(t, name)
	}

	fs := []string{
		"./testdata/elfs/go12",
		"./testdata/elfs/go16",
		"./testdata/elfs/go18",
		"./testdata/elfs/go20",
		"./testdata/elfs/go12-static",
		"./testdata/elfs/go16-static",
		"./testdata/elfs/go18-static",
		"./testdata/elfs/go20-static",
	}
	for _, f := range fs {
		t.Run(f, func(t *testing.T) {

			expectedSymbols, err := GetGoSymbols(f)

			require.NoError(t, err)

			me, err := NewMMapedElfFile(f)
			require.NoError(t, err)
			defer me.Close()

			goTable, err := me.NewGoTable()
			require.NotNil(t, goTable.Index.Entry32)
			require.Nil(t, goTable.Index.Entry64)
			require.NoError(t, err)

			require.Greater(t, len(expectedSymbols), 1000)

			testGoSymbolTable(t, expectedSymbols, goTable)

			goTable2 := &GoTable{}
			*goTable2 = *goTable
			goTable2.Index.Entry64 = make([]uint64, len(goTable.Index.Name))
			for i := 0; i < len(goTable.Index.Name); i++ {
				goTable2.Index.Entry64[i] = uint64(goTable2.Index.Entry32[i])
			}
			goTable2.Index.Entry32 = nil
			testGoSymbolTable(t, expectedSymbols, goTable)
		})
	}

}

func TestGoSymEntry64(t *testing.T) {
	t.Fail()
}
