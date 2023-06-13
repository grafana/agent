package stages

import (
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
				Source: &source,
			},
			ErrEmptyDBTypeGeoIPStageConfig,
		},
	}
	for _, tt := range tests {
		err := validateGeoIPConfig(tt.config)
		if err != nil {
			require.Equal(t, tt.wantError.Error(), err.Error())
		}
		if tt.wantError == nil {
			require.Nil(t, err)
		}
	}
}
