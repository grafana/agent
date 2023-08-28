package stages

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ValidateConfigs(t *testing.T) {
	source := "ip"
	tests := []struct {
		config    GeoIPConfig
		wantError error
	}{
		{
			GeoIPConfig{
				DB:     "test",
				Source: &source,
				DBType: "city",
			},
			nil,
		},
		{
			GeoIPConfig{
				DB:     "test",
				Source: &source,
				CustomLookups: map[string]string{
					"field": "lookup",
				},
			},
			nil,
		},
		{
			GeoIPConfig{
				DB:     "test",
				Source: &source,
			},
			ErrEmptyDBTypeAndValuesGeoIPStageConfig,
		},
		{
			GeoIPConfig{
				Source: &source,
				DBType: "city",
			},
			ErrEmptyDBPathGeoIPStageConfig,
		},
		{
			GeoIPConfig{
				DB:     "test",
				DBType: "city",
			},
			ErrEmptySourceGeoIPStageConfig,
		},
		{
			GeoIPConfig{
				DB:     "test",
				DBType: "fake",
				Source: &source,
			},
			ErrEmptyDBTypeGeoIPStageConfig,
		},
		{
			GeoIPConfig{
				DB:     "test",
				Source: &source,
				CustomLookups: map[string]string{
					"field": ".-badlookup",
				},
			},
			errors.New(ErrCouldNotCompileJMES),
		},
	}
	for _, tt := range tests {
		_, err := validateGeoIPConfig(tt.config)
		if err != nil {
			require.Equal(t, tt.wantError.Error(), err.Error())
		}
		if tt.wantError == nil {
			require.Nil(t, err)
		}
	}
}
