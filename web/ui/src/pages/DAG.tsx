import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';
import ComponentGraph from '../components/ComponentGraph';
import Page from '../components/Page';

function DAG() {
  return (
    <Page name="DAG" desc="Relationships between components" icon={faDiagramProject}>
      <ComponentGraph />
    </Page>
  );
}

export default DAG;
