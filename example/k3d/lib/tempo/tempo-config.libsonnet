{
  auth_enabled: false,

  server: {
    http_listen_port: 80,
    graceful_shutdown_timeout: '5s',
    http_server_idle_timeout: '120s',
    grpc_server_max_recv_msg_size: 1024 * 1024 * 64,
  },

  distributor: {
    receivers: {
      otlp: {
        protocols: {
          grpc: {
            endpoint: '0.0.0.0:55680',
          },
        },
      },
    },
  },

  ingester: {
    trace_idle_period: '10s',
    traces_per_block: 100,            
    max_block_duration: '5m',
  },

  compactor: {
    compaction: {
      compaction_window: '1h',
      max_compaction_objects: 1000000,
      block_retention: '1h',
      compacted_block_retention: '10m',
    },
  },

  storage: {
    trace: {
      backend: 'local',
      wal: {
        path: '/tmp/tempo/wal',
        bloom_filter_false_positive: 0.05, 
        index_downsample: 10,
      },
      'local': {
        path: '/tmp/tempo/data',
      },
      pool: {
        max_workers: 100,
        queue_depth: 10000,
      },
    },
  },
}
