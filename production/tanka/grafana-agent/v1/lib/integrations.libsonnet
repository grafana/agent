{
  // withIntegrations controls the integrations component of the Agent.
  //
  // For the full list of options, refer to the configuration reference:
  // https://github.com/grafana/agent/blob/master/docs/configuration-reference.md#integrations_config
  withIntegrations(integrations):: {
    assert std.objectHasAll(self, '_mode') : |||
      withLokiConfig must be merged with the result of calling new.
    |||,
    _integrations:: integrations
  },
}
