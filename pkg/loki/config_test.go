package loki

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConfigVersioning(t *testing.T) {
	expectStr := `
  version: v1
  configs:
  - name: default
    positions:
      filename: /tmp/positions.yml
    clients:
      - url: http://loki:3100/loki/api/v1/push
    scrape_configs:
    - job_name: varlog
      static_configs:
        - targets: [localhost]
          labels:
            job: varlog
            __path__: /var/log/*log
  `
	var expect LatestConfig
	err := yaml.UnmarshalStrict([]byte(expectStr), &expect)
	require.NoError(t, err)

	input := map[string]string{
		"unversioned": `
      positions:
        filename: /tmp/positions.yml
      clients:
        - url: http://loki:3100/loki/api/v1/push
      scrape_configs:
      - job_name: varlog
        static_configs:
          - targets: [localhost]
            labels:
              job: varlog
              __path__: /var/log/*log
    `,
		"v0": `
     version: v0
     positions:
       filename: /tmp/positions.yml
     clients:
       - url: http://loki:3100/loki/api/v1/push
     scrape_configs:
      - job_name: varlog
        static_configs:
          - targets: [localhost]
            labels:
              job: varlog
              __path__: /var/log/*log
    `,
		"v1": `
      version: v1
      configs:
      - name: default
        positions:
          filename: /tmp/positions.yml
        clients:
          - url: http://loki:3100/loki/api/v1/push
        scrape_configs:
        - job_name: varlog
          static_configs:
            - targets: [localhost]
              labels:
                job: varlog
                __path__: /var/log/*log
    `,
	}

	for name, ii := range input {
		t.Run(name, func(t *testing.T) {
			var cfg Config
			err := yaml.UnmarshalStrict([]byte(ii), &cfg)
			require.NoError(t, err)
			require.Equal(t, expect, cfg.Config)
		})
	}
}
