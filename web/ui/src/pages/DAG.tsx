import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';
import ComponentGraph from '../features/component/ComponentGraph';
import Page from '../features/layout/Page';

function DAG() {
  return (
    <Page name="DAG" desc="Relationships between components" icon={faDiagramProject}>
      <ComponentGraph />
    </Page>
  );
}

export default DAG;
