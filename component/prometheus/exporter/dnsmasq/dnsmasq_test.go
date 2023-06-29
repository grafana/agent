package dnsmasq

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/dnsmasq_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalRiver(t *testing.T) {
	rawCfg := `
  address       = "localhost:9999"
  leases_file   = "/etc/dnsmasq.leases"
  expose_leases = true
`
	var args Arguments
	err := river.Unmarshal([]byte(rawCfg), &args)
	assert.NoError(t, err)

	expected := Arguments{
		Address:      "localhost:9999",
		LeasesFile:   "/etc/dnsmasq.leases",
		ExposeLeases: true,
	}
	assert.Equal(t, expected, args)
}

func TestUnmarshalRiverDefaults(t *testing.T) {
	rawCfg := ``
	var args Arguments
	err := river.Unmarshal([]byte(rawCfg), &args)
	assert.NoError(t, err)

	expected := DefaultArguments
	assert.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverArguments := Arguments{
		Address:      "localhost:9999",
		LeasesFile:   "/etc/dnsmasq.leases",
		ExposeLeases: true,
	}

	expected := &dnsmasq_exporter.Config{
		DnsmasqAddress: "localhost:9999",
		LeasesPath:     "/etc/dnsmasq.leases",
		ExposeLeases:   true,
	}

	assert.Equal(t, expected, riverArguments.Convert())
}

func TestCustomizeTarget(t *testing.T) {
	args := Arguments{
		Address: "localhost:53",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	assert.Equal(t, 1, len(newTargets))
	assert.Equal(t, "localhost:53", newTargets[0]["instance"])
}
