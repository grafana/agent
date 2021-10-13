local config = import 'config.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local configMap = k.core.v1.configMap;
local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local service = k.core.v1.service;

{
  new(dashboards={}, datasources=[], namespace='default'):: {
    _images:: config._images,
    _config:: config._config { namespace: namespace },
    _dashboards:: {},
    _datasources:: datasources,

    local _images = self._images,
    local _config = self._config,
    local _dashboards = self._dashboards,
    local _datasources = self._datasources,

    grafana_cm:
      configMap.new('grafana-config') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'grafana.ini': std.manifestIni(_config.grafana_ini),
      }),

    grafana_dashboard_cm:
      if _config.dashboard_config_maps > 0
      then {}
      else
        configMap.new('dashboards') +
        configMap.mixin.metadata.withNamespace(namespace) +
        configMap.withDataMixin({
          [name]: std.toString(
            $.dashboards[name]
            { uid: std.substr(std.md5(std.toString($.dashboards[name])), 0, 9) }
          )
          for name in std.objectFields($.dashboards)
        }),

    grafana_dashboard_cms: {
      ['dashboard-%d' % shard]:
        configMap.new('dashboards-%d' % shard) +
        configMap.mixin.metadata.withNamespace(namespace) +
        configMap.withDataMixin({
          [name]: std.toString(
            _dashboards[name]
            { uid: std.substr(std.md5(std.toString(_dashboards[name])), 0, 9) }
          )
          for name in std.objectFields(_dashboards)
          if std.codepoint(std.md5(name)[1]) % _config.dashboard_config_maps == shard
        })
      for shard in std.range(0, _config.dashboard_config_maps - 1)
    },

    grafana_datasource_cm:
      configMap.new('grafana-datasources') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withDataMixin(std.foldl(function(acc, obj) acc {
        ['%s.yml' % obj.datasources[0].name]: k.util.manifestYaml(obj),
      }, self._datasources, {})),

    grafana_dashboard_provisioning_cm:
      configMap.new('grafana-dashboard-provisioning') +
      configMap.mixin.metadata.withNamespace(namespace) +
      configMap.withData({
        'dashboards.yml': k.util.manifestYaml({
          apiVersion: 1,
          providers: [{
            name: 'dashboards',
            orgId: 1,
            folder: '',
            folderUid: '',
            type: 'file',
            disableDeletion: true,
            editable: false,
            updateIntervalSeconds: 10,
            allowUiUpdates: false,
            options: {
              path: '/grafana/dashboards',
            },
          }],
        }),
      }),

    container::
      container.new('grafana', _images.grafana) +
      container.withPorts(containerPort.new('grafana', 80)) +
      container.withCommand([
        '/usr/share/grafana/bin/grafana-server',
        '--homepath=/usr/share/grafana',
        '--config=/etc/grafana-config/grafana.ini',
      ]) +
      k.util.resourcesRequests('10m', '40Mi'),

    deployment:
      deployment.new('grafana', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(0) +
      k.util.configMapVolumeMount(self.grafana_cm, '/etc/grafana-config') +
      k.util.configMapVolumeMount(self.grafana_datasource_cm, '%(provisioning_dir)s/datasources' % _config) +
      k.util.configMapVolumeMount(self.grafana_dashboard_provisioning_cm, '%(provisioning_dir)s/dashboards' % _config) +
      (
        if self._config.dashboard_config_maps == 0
        then k.util.configMapVolumeMount(self.grafana_dashboard_config_map, '/grafana/dashboards')
        else
          std.foldr(
            function(m, acc) m + acc,
            [
              k.util.configVolumeMount('dashboards-%d' % shard, '/grafana/dashboards/%d' % shard)
              for shard in std.range(0, self._config.dashboard_config_maps - 1)
            ],
            {}
          )
      ) +
      k.util.podPriority('critical'),

    service:
      k.util.serviceFor(self.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },

  // withDashboards sets the list of dashboards. Dashboards is an object where the
  // key should be the filename.
  withDashboards(dashboards={}):: { _dashboards:: dashboards },

  // withDataSources sets the list of datasources. Datasources can be created
  // using datasources.libsonnet.
  withDataSources(datasources=[]):: { _datasources:: datasources },
}
