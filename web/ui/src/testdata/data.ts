import { ComponentDetail, ComponentHealthState, ComponentInfo } from '../features/component/types';
import { StmtType, ValueType } from '../features/river-js/types';

export const testComponents: ComponentInfo[] = [
  {
    id: 'local.file.api_key',
    name: 'local.file',
    label: 'api_key',
    health: {
      state: ComponentHealthState.HEALTHY,
    },
    referencedBy: ['metrics.remote_write.default'],
    referencesTo: [],
  },
  {
    id: 'discovery.k8s.pods',
    name: 'discovery.k8s',
    label: 'pods',
    health: {
      state: ComponentHealthState.UNHEALTHY,
    },
    referencedBy: ['metrics.scrape.k8s_pods'],
    referencesTo: [],
  },
  {
    id: 'metrics.scrape.k8s_pods',
    name: 'metrics.scrape',
    label: 'k8ds_pods',
    health: {
      state: ComponentHealthState.UNKNOWN,
    },
    referencedBy: [],
    referencesTo: ['metrics.remote_write.default', 'discovery.k8s.pods'],
  },
  {
    id: 'metrics.remote_write.default',
    name: 'metrics.remote_write',
    label: 'default',
    health: {
      state: ComponentHealthState.EXITED,
    },
    referencedBy: ['metrics.scrape.k8s_pods'],
    referencesTo: [],
  },
];

export const testComponentDetail: ComponentDetail = {
  id: 'metrics.scrape.k8s_pods',
  name: 'metrics.scrape',
  label: 'k8s_pods',
  health: {
    state: ComponentHealthState.UNKNOWN,
    message: 'Component updated',
    updatedTime: '2020-08-17',
  },
  referencedBy: [],
  referencesTo: ['metrics.remote_write.default', 'discovery.k8s.pods'],

  rawConfig: `metrics.scrape "k8s_pods" {
  targets    = discovery.k8s.pods.targets
  forward_to = [metrics.remote_write.default.receiver]

  scrape_config {
    job_name = "default"
  }
}`,

  arguments: [
    {
      type: StmtType.ATTR,
      name: 'targets',
      value: {
        type: ValueType.ARRAY,
        value: [
          {
            type: ValueType.OBJECT,
            value: [
              {
                key: '__address__',
                value: { type: ValueType.STRING, value: 'demo.robustperception.io:9090' },
              },
              {
                key: 'other_label',
                value: { type: ValueType.STRING, value: 'foobar' },
              },
            ],
          },
        ],
      },
    },
    {
      type: StmtType.ATTR,
      name: 'forward_to',
      value: {
        type: ValueType.ARRAY,
        value: [
          {
            type: ValueType.CAPSULE,
            value: 'capsule("metrics.Receiver")',
          },
        ],
      },
    },
    {
      type: StmtType.BLOCK,
      name: 'scrape_config',
      body: [
        {
          type: StmtType.ATTR,
          name: 'job_name',
          value: {
            type: ValueType.STRING,
            value: 'default',
          },
        },
      ],
    },
  ],
};
