import { FC } from 'react';
import { ComponentView } from '../features/component/ComponentView';
import { ComponentDetail, ComponentHealthType } from '../features/component/types';
import { StmtType, ValueType } from '../features/river-js/types';

const testComponentDetail: ComponentDetail = {
  id: 'metrics.scrape.k8s_pods',
  name: 'metrics.scrape',
  label: 'k8s_pods',
  health: {
    type: ComponentHealthType.HEALTHY,
  },
  inReferences: [],
  outReferences: ['metrics.remote_write.default', 'discovery.k8s.pods'],

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
  //exports: [],
  //debugInfo: [],
};

export const ComponentDetailPage: FC = () => {
  return <ComponentView component={testComponentDetail} />;
};
