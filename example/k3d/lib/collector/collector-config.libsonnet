{
  receivers: {
    otlp: {
      protocols: {
        grpc: null,
      },
    },
  },

  exporters: {
    logging: {
      loglevel: 'info',
    },
  },

  service: {
    pipelines: {
      traces: {
        receivers: [
          'otlp',
        ],
        exporters: [
          'logging',
        ],
      },
    },
  },
}
