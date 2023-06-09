//go:build linux

package symtab

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	gosym2 "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab/gosym"
	"github.com/stretchr/testify/require"
)

func TestGoSymSelfTest(t *testing.T) {
	var ptr = reflect.ValueOf(TestGoSymSelfTest).Pointer()
	mod := "/proc/self/exe"
	symtab, err := newGoSymbols(mod)
	if err != nil {
		t.Fatalf("failed to create symtab %v", err)
	}
	sym := symtab.Resolve(uint64(ptr))
	expectedSym := "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab.TestGoSymSelfTest"
	require.NotNil(t, sym)
	require.Equal(t, expectedSym, sym.Name)
	require.Equal(t, uint64(ptr), sym.Start)
}

func TestPclntab18(t *testing.T) {
	s := "f0 ff ff ff 00 00 01 08 9a 05 00 00 00 00 00 00 " +
		" bb 00 00 00 00 00 00 00 a0 23 40 00 00 00 00 00" +
		" 60 00 00 00 00 00 00 00 c0 bb 00 00 00 00 00 00" +
		" c0 c3 00 00 00 00 00 00 c0 df 00 00 00 00 00 00"
	bs, _ := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	textStart := gosym2.ParseRuntimeTextFromPclntab18(bs)
	expected := uint64(0x4023a0)
	require.Equal(t, expected, textStart)
}

func BenchmarkGoSym(b *testing.B) {
	gosym, _ := newGoSymbols("/proc/self/exe")
	if gosym.Length() < 1000 {
		b.FailNow()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, symbol := range gosym.Symbols {
			gosym.Resolve(symbol.Start)
		}
	}
}
