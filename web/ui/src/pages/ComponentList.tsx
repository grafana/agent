import CComponentList, { ComponentHealth } from '../components/ComponentList';

function ComponentList() {
  const mockComponents = [
    { id: 'local.file.api_key', health: ComponentHealth.HEALTHY },
    { id: 'discovery.k8s.pods', health: ComponentHealth.UNHEALTHY },
    { id: 'metrics.scrape.k8s_pods', health: ComponentHealth.UNKNOWN },
    { id: 'metrics.remote_write.default', health: ComponentHealth.EXITED },
  ];

  return <CComponentList components={mockComponents} />;
}

export default ComponentList;
