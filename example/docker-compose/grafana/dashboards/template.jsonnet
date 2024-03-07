local agentDashboards = import 'agent-static-mixin/dashboards.libsonnet';
local agentDebugging = import 'agent-static-mixin/debugging.libsonnet';

local result = agentDashboards + agentDebugging {
  files: {
    [name]: $.grafanaDashboards[name] {
      // Use local timezone for local testing
      timezone: '',
    }
    for name in std.objectFields($.grafanaDashboards)
  },
};

result.files
