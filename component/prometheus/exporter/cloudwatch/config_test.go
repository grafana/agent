package cloudwatch

import (
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test(t *testing.T) {
	config := `
sts_region = "us-east-2"
debug = true
discovery {
	type = "sqs"
	regions = ["us-east-2"]
	search_tags = {
		"scrape" = "true",
	}
	metric {
		name = "NumberOfMessagesSent"
		statistics = ["Sum", "Average"]
		period = "1m"
	}
	metric {
		name = "NumberOfMessagesReceived"
		statistics = ["Sum", "Average"]
		period = "1m"
	}
}
static "super_ec2_instance_id" {
	regions = ["us-east-2"]
	namespace = "AWS/EC2"
	dimensions = {
		"InstanceID" = "i01u29u12ue1u2c",
	}
	metric {
		name = "CPUUsage"
		statistics = ["Sum", "Average"]
		period = "1m"
	}
}
`
	args := Arguments{}
	require.NoError(t, river.Unmarshal([]byte(config), &args))
}
