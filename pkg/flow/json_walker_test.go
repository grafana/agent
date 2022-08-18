package flow

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/pkg/flow/rivertypes"

	_ "github.com/grafana/agent/component/discovery/kubernetes" // Import discovery.k8s
	_ "github.com/grafana/agent/component/local/file"           // Import local.file
	_ "github.com/grafana/agent/component/metrics/mutate"       // Import metrics.mutate
	_ "github.com/grafana/agent/component/metrics/remotewrite"  // Import metrics.remotewrite
	_ "github.com/grafana/agent/component/metrics/scrape"       // Import metrics.scrape
	_ "github.com/grafana/agent/component/remote/s3"            // Import s3.file
	_ "github.com/grafana/agent/component/targets/mutate"       // Import targets.mutate
	"github.com/stretchr/testify/require"
)

func TestSimpleWalking(t *testing.T) {
	type simple struct {
		Int    int64       `river:"int_test,attr"`
		String string      `river:"string_test,attr"`
		Bool   bool        `river:"bool_test,attr"`
		Float  float64     `river:"float_test,attr"`
		Nil    interface{} `river:"nil_test,attr"`
	}
	test := simple{
		Int:    1,
		String: "cool",
		Bool:   true,
		Float:  3.14,
	}
	fields := ConvertToField(test, "simple")
	require.True(t, fields.Key == "simple")
	sub := fields.Value.([]*Field)
	require.Len(t, sub, 5)

	require.True(t, sub[0].Key == "int_test")
	require.True(t, sub[0].Value.(*Field).Value.(int64) == 1)

	require.True(t, sub[1].Key == "string_test")
	require.True(t, sub[1].Value.(*Field).Value.(string) == "cool")

	require.True(t, sub[2].Key == "bool_test")
	require.True(t, sub[2].Value.(*Field).Value.(bool) == true)

	require.True(t, sub[3].Key == "float_test")
	require.True(t, sub[3].Value.(*Field).Value.(float64) == test.Float)

	_, err := json.Marshal(fields)
	require.NoError(t, err)

}

func TestArrayWalking(t *testing.T) {
	type simple struct {
		Array []int `river:"array_test,attr"`
	}
	test := simple{
		Array: []int{1, 2, 3},
	}
	fields := ConvertToField(test, "simple")
	sub := fields.Value.([]*Field)
	require.Len(t, sub, 1)

	require.True(t, sub[0].Key == "array_test")
	arr := sub[0].Value.([]*Field)
	require.Len(t, arr, 3)

	str, err := json.Marshal(fields)
	require.NoError(t, err)
	println(str)
}

func TestMapWalking(t *testing.T) {
	type simple struct {
		Map map[string]string `river:"map_test,attr"`
	}
	test := simple{
		Map: map[string]string{
			"p1": "bob",
			"p2": "sam",
		},
	}
	fields := ConvertToField(test, "simple")
	sub := fields.Value.([]*Field)
	require.Len(t, sub, 1)

	require.True(t, sub[0].Key == "map_test")
	arr := sub[0].Value.([]*Field)
	require.Len(t, arr, 2)
}

func TestSecretsWalking(t *testing.T) {
	type simple struct {
		Secret       rivertypes.Secret         `river:"secret_test,attr"`
		AlwaysSecret rivertypes.OptionalSecret `river:"opt_yes,attr"`
		NotSecret    rivertypes.OptionalSecret `river:"opt_no,attr"`
	}
	test := simple{
		Secret:       "password",
		AlwaysSecret: rivertypes.OptionalSecret{IsSecret: true, Value: "password"},
		NotSecret:    rivertypes.OptionalSecret{IsSecret: false, Value: "password"},
	}
	fields := ConvertToField(test, "simple")
	sub := fields.Value.([]*Field)
	require.Len(t, sub, 3)

	require.True(t, sub[0].Key == "secret_test")
	require.True(t, sub[0].Value.(*Field).Value == "(secret)")

	require.True(t, sub[1].Key == "opt_yes")
	require.True(t, sub[1].Value.(*Field).Value == "(secret)")

	require.True(t, sub[2].Key == "opt_no")
	require.True(t, sub[2].Value.(*Field).Value == "password")

}
