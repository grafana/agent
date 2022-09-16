import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';
import { ComponentGraph } from '../features/component/ComponentGraph';
import Page from '../features/layout/Page';
import { useComponentInfo } from '../hooks/componentInfo';

function DAG() {
  const components = useComponentInfo();

  return (
    <Page name="DAG" desc="Relationships between defined components" icon={faDiagramProject}>
      {components.length > 0 && <ComponentGraph components={components} />}
    </Page>
  );
}

export default DAG;
