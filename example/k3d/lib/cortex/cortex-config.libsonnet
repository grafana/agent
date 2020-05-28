{
  auth_enabled: false,

  server: {
    http_listen_port: 80,
    grpc_listen_port: 9095,

    // Configure the server to allow messages up to 100MB.
    grpc_server_max_recv_msg_size: 104857600,
    grpc_server_max_send_msg_size: 104857600,
    grpc_server_max_concurrent_streams: 1000,
  },

  distributor: {
    shard_by_all_labels: true,
    pool: {
      health_check_ingesters: true,
    },
  },

  ingester_client: {
    grpc_client_config: {
      max_recv_msg_size: 104857600,
      max_send_msg_size: 104857600,
      use_gzip_compression: true,
    },
  },

  ingester: {
    lifecycler: {
      join_after: 0,
      claim_on_rollout: false,
      final_sleep: '0s',
      num_tokens: 512,

      ring: {
        kvstore: {
          store: 'inmemory',
        },
        replication_factor: 1,
      },
    },
  },

  schema: {
    configs: [{
      from: '2020-02-07',
      store: 'boltdb',
      object_store: 'filesystem',
      schema: 'v10',
      index: {
        prefix: 'index_',
        period: '168h',
      },
    }],
  },

  storage: {
    boltdb: {
      directory: '/tmp/cortex/index',
    },
    filesystem: {
      directory: '/tmp/cortex/chunks',
    },
  },

  limits: {
    ingestion_rate: 250000,
    ingestion_burst_size: 500000,
  },
}
