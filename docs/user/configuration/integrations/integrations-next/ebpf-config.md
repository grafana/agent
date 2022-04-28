+++
title = "ebpf_config"
+++

# ebpf_config

The `ebpf_config` block configures the Agent's eBPF integration.
It is an embedded version of
[`ebpf_exporter`](https://github.com/cloudflare/ebpf_exporter)
that allows the Agent to attach eBPF programs to the host kernel
and export defined metrics in a Prometheus-compatible format.

As such, this integration comes with the relevant caveats of
running eBPF programs on your host, like being on a kernel 
version >4.1, specific kernel flags being enabled, plus 
superuser access is most usually required.

Currently, the exporter only supports `kprobes`, that is
kernel-space probes.

Configuration reference:

```yaml
  ## ebpf runs the provided 'programs' on the host's kernel
  ## and reports back on the metrics attached to them.
  programs: 
     [- <program_config> ... ]
```

Each provided [`<program_config>`](https://pkg.go.dev/github.com/cloudflare/ebpf_exporter@v1.2.5/config#Program) block defines a single eBPF program that the integration should run, along with what metrics should be attached to it.

Here's an [example](https://github.com/cloudflare/ebpf_exporter/blob/master/examples/cachestat.yaml) of a valid configuration that includes a program to measure hits and misses to the file system page cache.

```yaml
programs:
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
    account_page_dirtied: do_count
    mark_buffer_dirty: do_count
  code: |
    #include <uapi/linux/ptrace.h>
    BPF_HASH(counts, u64);
    int do_count(struct pt_regs *ctx) {
        counts.increment(PT_REGS_IP(ctx) - 1);
        return 0;
    }
```
