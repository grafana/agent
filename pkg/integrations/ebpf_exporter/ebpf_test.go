package ebpf

import (
	"testing"

	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEBPFConfig(t *testing.T) {
	yamlCfg := `
programs:
# Count timers fired in the kernel
- name: cachestat
  metrics:
  counters:
  - name: page_cache_ops_total
    help: Page cache operation counters by type
    table: counts
    labels:
    - name: op
      size: 8
      decoders:
      - name: ksym
  kprobes:
    add_to_page_cache_lru: do_count
    mark_page_accessed: do_count
  code: |
    #include <uapi/linux/ptrace.h>
    struct key_t {
      u64 ip;
      char command[128];
    };
    BPF_HASH(counts, struct key_t);
    int do_count(struct pt_regs *ctx) {
      struct key_t key = { .ip = PT_REGS_IP(ctx) - 1 };
      bpf_get_current_comm(&key.command, sizeof(key.command));
      counts.increment(key);
      return 0;
    }
`
	var cfg Config
	err := yaml.Unmarshal([]byte(yamlCfg), &cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Programs, 1)
	require.Equal(t, cfg.Programs[0].Name, "cachestat")

	_, err = cfg.NewIntegration(util.TestLogger(t))
	require.NoError(t, err)
}
