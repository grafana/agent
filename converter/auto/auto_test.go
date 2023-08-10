package auto

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type riverExample struct {
	TestRiver int    `river:"test,attr"`
	Str       string `river:"str,attr"`
}

type yamlExample struct {
	TestYAML int    `yaml:"test"`
	String   string `yaml:"str"`
}

func TestFoo(t *testing.T) {
	from := &riverExample{
		TestRiver: 42,
		Str:       "foo",
	}
	to := &yamlExample{}
	err := Convert(from, to, ConversionCfg{
		FromTags:             "river",
		ToTags:               "yaml",
		FromTagNameExtractor: FistInCSV,
		ToTagNameExtractor:   FistInCSV,
	})
	require.NoError(t, err)
	require.Equal(t, 42, to.TestYAML)
	require.Equal(t, "foo", to.String)
}
