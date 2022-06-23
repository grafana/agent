{
  _images+:: {
    tempo: 'grafana/tempo:latest',
    tempo_query: 'grafana/tempo-query:latest',
    tempo_vulture: 'grafana/tempo-vulture:latest',
  },

  _config+:: {
    tempo: {
      port: 3200,
      replicas: 1,
      headless_service_name: 'tempo-members',
    },
    pvc_size: '30Gi',
    pvc_storage_class: 'local-path',
    receivers: {
      jaeger: {
        protocols: {
          thrift_http: null,
        },
      },
      otlp: {
        protocols: {
          grpc: {
            endpoint: "0.0.0.0:4317"
          },
        },
      },
    },
    ballast_size_mbs: '1024',
    jaeger_ui: {
      base_path: '/',
    },
    search_enabled: false,
  },
}