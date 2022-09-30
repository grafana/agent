package relabel_test

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/relabel"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRelabelConfigApplication(t *testing.T) {
	riverArguments := `
targets = [ 
	{ "__meta_foo" = "foo", "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "one",   "app" = "backend",  "__tmp_a" = "tmp" },
	{ "__meta_foo" = "foo", "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "two",   "app" = "db",       "__tmp_b" = "tmp" },
	{ "__meta_baz" = "baz", "__meta_qux" = "qux", "__address__" = "localhost", "instance" = "three", "app" = "frontend", "__tmp_c" = "tmp" },
]

rule {
	source_labels = ["__address__", "instance"]
	separator     = "/"
	target_label  = "destination"
	action        = "replace"
} 

rule {
	source_labels = ["app"]
	action        = "drop"
	regex         = "frontend"
}

rule {
	source_labels = ["app"]
	action        = "keep"
	regex         = "backend"
}

rule {
	source_labels = ["instance"]
	target_label  = "name"
}

rule {
	action      = "labelmap"
	regex       = "__meta_(.*)"
	replacement = "meta_$1"
}

rule {
	action = "labeldrop"
	regex  = "__meta(.*)|__tmp(.*)|instance"
}
`
	expectedExports := relabel.Exports{
		Output: []discovery.Target{
			map[string]string{"__address__": "localhost", "app": "backend", "destination": "localhost/one", "meta_bar": "bar", "meta_foo": "foo", "name": "one"},
		},
	}

	var args relabel.Arguments
	require.NoError(t, river.Unmarshal([]byte(riverArguments), &args))

	tc, err := componenttest.NewControllerFromID(nil, "discovery.relabel")
	require.NoError(t, err)
	go func() {
		err = tc.Run(componenttest.TestContext(t), args)
		require.NoError(t, err)
	}()

	require.NoError(t, tc.WaitExports(time.Second))
	require.Equal(t, expectedExports, tc.Exports())
}
