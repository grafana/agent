local cortex = import 'cortex/main.libsonnet';
local datasource = import 'grafana/datasource.libsonnet';
local grafana = import 'grafana/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';
local loki = import 'loki/main.libsonnet';
local metrics = import 'metrics-server/main.libsonnet';
local promtail = import 'promtail/promtail.libsonnet';

local mixins = import './mixins.libsonnet';

{
  new(namespace=''):: {
    ns: k.core.v1.namespace.new(namespace),

    grafana:
      grafana.new(namespace=namespace) +
      grafana.withDashboards(mixins.grafanaDashboards) +
      grafana.withDataSources([
        datasource.new('Cortex', 'http://cortex.default.svc.cluster.local/api/prom'),
        datasource.new('Loki', 'http://loki.default.svc.cluster.local', type='loki'),
      ]),

    loki: loki.new(namespace),

    promtail: promtail {
      _config+:: {
        namespace: namespace,
        promtail_config+: {
          clients: [{
            scheme:: 'http',
            hostname:: 'loki.default.svc.cluster.local',
            external_labels: {},
          }],

          pipeline_stages: [
            // k3d uses cri for logging
            { cri: {} },

            // Bad words metrics, used in Agent dashboard
            {
              regex: {
                expression: '(?i)(?P<bad_words>panic:|core_dumped|failure|error|attack| bad |illegal |denied|refused|unauthorized|fatal|failed|Segmentation Fault|Corrupted)',
              },
            },
            {
              metrics: {
                panic_total: {
                  type: 'Counter',
                  description: 'total count of panic: found in log lines',
                  source: 'panic',
                  config: {
                    action: 'inc',
                  },
                },
                bad_words_total: {
                  type: 'Counter',
                  description: 'total count of bad words found in log lines',
                  source: 'bad_words',
                  config: {
                    action: 'inc',
                  },
                },
              },
            },
          ],
        },
      },
    },

    cortex: cortex.new(namespace),
  },
}
