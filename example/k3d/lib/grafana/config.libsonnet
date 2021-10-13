{
  _images: {
    grafana: 'grafana/grafana:8.0.3',
  },

  _config: {
    // Optionally shard dashboards into multiple config maps.
    // Set to the number of desired config maps.  0 to disable.
    dashboard_config_maps: 8,

    provisioning_dir: '/etc/grafana/provisioning',
    grafana_root_url: 'http://grafana.default.svc.cluster.local/',

    grafana_ini: {
      sections: {
        'auth.anonymous': {
          enabled: true,
          org_role: 'Admin',
        },
        server: {
          http_port: 80,
          root_url: $._config.grafana_root_url,
        },
        analytics: {
          reporting_enabled: false,
        },
        users: {
          default_theme: 'dark',
        },
        explore+: {
          enabled: true,
        },
      },
    },
  },
}
