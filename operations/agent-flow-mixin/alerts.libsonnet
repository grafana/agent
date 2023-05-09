{
  prometheusAlerts+: {
    groups+: [
      (import './alerts/clustering.libsonnet'),
    ],
  },
}
