package cloudwatch_exporter

import (
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"testing"
)

const configString = `
stsRegion: us-east-2
discovery:
  jobs:
    - type: AWS/EC2
      regions:
        - us-east-2
      roles:
        - roleArn: arn:aws:iam::878167871295:role/yace_testing
      period: 5m
      customTags:
        - key: alias
          value: tesis
      metrics:
        - name: CPUUtilization
          statistics:
            - Maximum
            - Average
`

var roundingPeriod5Minutes = int64(300)
var truePtr = true
var falsePtr = false

var expectedConfig = yaceConf.ScrapeConf{
	StsRegion: "us-east-2",
	Discovery: yaceConf.Discovery{
		Jobs: []*yaceConf.Job{
			{
				Type:    "AWS/EC2",
				Regions: []string{"us-east-2"},
				Roles: []yaceConf.Role{
					{
						RoleArn: "arn:aws:iam::878167871295:role/yace_testing",
					},
				},
				CustomTags: []yaceModel.Tag{
					{
						Key:   "alias",
						Value: "tesis",
					},
				},
				Metrics: []*yaceConf.Metric{
					{
						Name:       "CPUUtilization",
						Statistics: []string{"Maximum", "Average"},
						// Defaults get configured from general settings
						Period: 300,
						Length: 300,
						// Default YACE delay applied
						Delay:                  300,
						NilToZero:              &nilToZero,
						AddCloudwatchTimestamp: &addCloudwatchTimestamp,
					},
				},
				Period:                 300,
				Length:                 300,
				Delay:                  0,
				RoundingPeriod:         &roundingPeriod5Minutes,
				AddCloudwatchTimestamp: &falsePtr,
				NilToZero:              &nilToZero,
			},
		},
	},
	Static:          []*yaceConf.Static{},
	CustomNamespace: []*yaceConf.CustomNamespace{},
}

func TestTranslateConfigToYACEConfig(t *testing.T) {
	c := Config{}
	err := yaml.Unmarshal([]byte(configString), &c)
	require.NoError(t, err, "failed to unmarshall config")

	yaceConf, err := ToYACEConfig(&c)
	require.NoError(t, err, "failed to translate to YACE configuration")

	require.EqualValues(t, expectedConfig, yaceConf)
}
