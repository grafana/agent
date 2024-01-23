package schema

import (
	"strings"
	"testing"

	"github.com/grafana/river/parser"
	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	sch := NewSchema("t", "1")
	err := sch.AddComponent("testcomp", "", nil, nil)
	require.NoError(t, err)
	v := Verifier{schema: sch.Json}
	dag, err := parser.ParseFile("t", []byte("testcomp {}"))
	require.NoError(t, err)
	dg := v.Verify(dag)
	require.Nil(t, dg)
}

func TestFailPath(t *testing.T) {
	sch := NewSchema("t", "1")
	err := sch.AddComponent("testcomp", "", nil, nil)
	require.NoError(t, err)
	v := Verifier{schema: sch.Json}
	dag, err := parser.ParseFile("t", []byte("testcompBAD {}"))
	require.NoError(t, err)
	dg := v.Verify(dag)
	require.NotNil(t, dg)
	require.True(t, strings.Contains(dg.Message, "unknown statement"))
}
