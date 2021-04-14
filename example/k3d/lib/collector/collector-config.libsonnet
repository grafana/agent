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
      loglevel: "debug",
    },
  },

  service: {
    pipelines: {
      traces: {
        receivers: [
          "otlp",
        ],
        exporters: [
          "logging",
        ],
      },
    },
  },
}
