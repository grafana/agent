package mutate_test

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/targets/mutate"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rfratto/gohcl"
	"github.com/stretchr/testify/require"
)

func TestRelabelConfigApplication(t *testing.T) {
	hclArguments := `
targets = [ 
    { "__meta_foo" = "foo", "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "one", "app" = "backend", __tmp_a = "tmp" },
    { "__meta_foo" = "foo", "__meta_bar" = "bar", "__address__" = "localhost", "instance" = "two", "app" = "db", "__tmp_b" = "tmp" },
    { "__meta_baz" = "baz", "__meta_qux" = "qux", "__address__" = "localhost", "instance" = "three", "app" = "frontend", "__tmp_c" = "tmp" }
]

relabel_config {
	source_labels = ["__address__", "instance"]
	separator     = "/"
	target_label  = "destination"
    action        = "replace"
} 

relabel_config {
    source_labels = ["app"]
	action = "drop"
	regex  = "frontend"
}

relabel_config {
    source_labels = ["app"]
	action = "keep"
	regex  = "backend"
}

relabel_config {
	source_labels = ["instance"]
	target_label  = "name"
}

relabel_config {
	action = "labelmap"
	regex  = "__meta_(.*)"
	replacement = "meta_$1"
}

relabel_config {
    action = "labeldrop"
	regex  = "__meta(.*)|__tmp(.*)|instance"
}

`
	expectedExports := mutate.Exports{
		Output: []mutate.Target{
			map[string]string{"__address__": "localhost", "app": "backend", "destination": "localhost/one", "meta_bar": "bar", "meta_foo": "foo", "name": "one"},
		},
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(hclArguments), "agent-config.flow")
	require.False(t, diags.HasErrors())

	var args mutate.Arguments
	diags = gohcl.DecodeBody(file.Body, nil, &args)
	require.False(t, diags.HasErrors())

	tc, err := componenttest.NewControllerFromID(nil, "targets.mutate")
	require.NoError(t, err)
	go func() {
		err = tc.Run(componenttest.TestContext(t), args)
		require.NoError(t, err)
	}()

	require.NoError(t, tc.WaitExports(time.Second))
	require.Equal(t, expectedExports, tc.Exports())
}
