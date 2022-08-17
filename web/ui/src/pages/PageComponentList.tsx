import { faCubes } from '@fortawesome/free-solid-svg-icons';
import Page from '../components/Page';
import ComponentList from '../components/ComponentList';
import { ComponentHealthType, ComponentInfo } from '../features/component/types';

function PageComponentList() {
  const mockComponents: ComponentInfo[] = [
    {
      id: 'local.file.api_key',
      name: 'local.file',
      label: 'api_key',
      health: {
        type: ComponentHealthType.HEALTHY,
      },
      inReferences: [],
      outReferences: [],
    },
    {
      id: 'discovery.k8s.pods',
      name: 'discovery.k8s',
      label: 'pods',
      health: {
        type: ComponentHealthType.UNHEALTHY,
      },
      inReferences: [],
      outReferences: [],
    },
    {
      id: 'metrics.scrape.k8s_pods',
      name: 'metrics.scrape',
      label: 'k8ds_pods',
      health: {
        type: ComponentHealthType.UNKNOWN,
      },
      inReferences: [],
      outReferences: [],
    },
    {
      id: 'metrics.remote_write.default',
      name: 'metrics.remote_write',
      label: 'default',
      health: {
        type: ComponentHealthType.EXITED,
      },
      inReferences: [],
      outReferences: [],
    },
  ];

  return (
    <Page name="Components" desc="List of known components" icon={faCubes}>
      <ComponentList components={mockComponents} />
    </Page>
  );
}

export default PageComponentList;
