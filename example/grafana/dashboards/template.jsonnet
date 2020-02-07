local agentDashboards = import 'agent-mixin/dashboards.libsonnet';

local result = agentDashboards {
  files: {
    [name]: $.grafanaDashboards[name] {
      // Use local timezone for local testing
      timezone: '',
    }
    for name in std.objectFields($.grafanaDashboards)
  },
};

result.files
