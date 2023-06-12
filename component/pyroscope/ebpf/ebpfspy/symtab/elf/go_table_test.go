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

		var first uint64
		if goTable.Index.Entry32 != nil {
			first = uint64(goTable.Index.Entry32[0])
		} else {
			first = goTable.Index.Entry64[0]
		}
		name := goTable.Resolve(uint64(first) - 1)
		require.Empty(t, name)
		name = goTable.Resolve(goTable.Index.End)
		require.Empty(t, name)
		name = goTable.Resolve(goTable.Index.End + 1)
		require.Empty(t, name)
	}

	ts := []struct {
		f        string
		expect32 bool
	}{
		{"./testdata/elfs/go12", true},
		{"./testdata/elfs/go16", true},
		{"./testdata/elfs/go18", true},
		{"./testdata/elfs/go20", true},
		{"./testdata/elfs/go12-static", true},
		{"./testdata/elfs/go16-static", false}, // this one switches from 32 to 64 in the middle
		{"./testdata/elfs/go18-static", false}, // this one starts with 64
		{"./testdata/elfs/go20-static", true},
	}
	for _, testcase := range ts {
		t.Run(testcase.f, func(t *testing.T) {

			expectedSymbols, err := GetGoSymbols(testcase.f)

			require.NoError(t, err)

			me, err := NewMMapedElfFile(testcase.f)
			require.NoError(t, err)
			defer me.Close()

			goTable, err := me.NewGoTable()

			require.NoError(t, err)
			if testcase.expect32 {
				require.NotNil(t, goTable.Index.Entry32)
				require.Nil(t, goTable.Index.Entry64)
			} else {
				require.NotNil(t, goTable.Index.Entry64)
				require.Nil(t, goTable.Index.Entry32)
			}

			require.Greater(t, len(expectedSymbols), 1000)

			testGoSymbolTable(t, expectedSymbols, goTable)

			if testcase.expect32 {
				goTable2 := &GoTable{}
				*goTable2 = *goTable
				goTable2.Index.Entry64 = make([]uint64, len(goTable.Index.Name))
				for i := 0; i < len(goTable.Index.Name); i++ {
					goTable2.Index.Entry64[i] = uint64(goTable2.Index.Entry32[i])
				}
				goTable2.Index.Entry32 = nil
				testGoSymbolTable(t, expectedSymbols, goTable)
			}
		})
	}

}

func TestGoSymEntry64(t *testing.T) {
	t.Fail()
}
