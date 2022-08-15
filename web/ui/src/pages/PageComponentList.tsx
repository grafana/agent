import { faCubes } from '@fortawesome/free-solid-svg-icons';
import Page from '../components/Page';
import ComponentList, { ComponentHealth } from '../components/ComponentList';

function PageComponentList() {
  const mockComponents = [
    { id: 'local.file.api_key', health: ComponentHealth.HEALTHY },
    { id: 'discovery.k8s.pods', health: ComponentHealth.UNHEALTHY },
    { id: 'metrics.scrape.k8s_pods', health: ComponentHealth.UNKNOWN },
    { id: 'metrics.remote_write.default', health: ComponentHealth.EXITED },
  ];

  return (
    <Page name="Components" desc="List of known components" icon={faCubes}>
      <ComponentList components={mockComponents} />
    </Page>
  );
}

export default PageComponentList;
