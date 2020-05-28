{
  auth_enabled: false,

  server: {
    graceful_shutdown_timeout: '5s',
    http_server_idle_timeout: '120s',
    grpc_server_max_recv_msg_size: 1024 * 1024 * 64,
  },

  limits_config: {
    enforce_metric_name: false,
    reject_old_samples: true,
    reject_old_samples_max_age: '24h',
  },

  ingester: {
    chunk_idle_period: '5m',
    chunk_retain_period: '30s',
    max_transfer_retries: 1,
    lifecycler: {
      address: '127.0.0.1',
      final_sleep: '0s',
      ring: {
        kvstore: { store: 'inmemory' },
        replication_factor: 1,
      },
    },
  },

  schema_config: {
    configs: [{
      from: '2020-05-25',
      store: 'boltdb',
      object_store: 'filesystem',
      schema: 'v11',
      index: {
        prefix: 'index_',
        period: '24h',
      },
    }],
  },

  storage_config: {
    boltdb: {
      directory: '/tmp/loki/index',
    },

    filesystem: {
      directory: '/tmp/loki/chunks',
    },
  },

  chunk_store_config: {
    max_look_back_period: 0,
  },

  table_manager: {
    retention_deletes_enabled: true,
    retention_period: '48h',
  },
}
