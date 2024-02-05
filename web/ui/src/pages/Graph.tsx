import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';

import { ComponentGraph } from '../features/graph/ComponentGraph';
import Page from '../features/layout/Page';
import { useAllComponentInfo } from '../hooks/componentInfo';

function Graph() {
  const [components] = useAllComponentInfo('');

  return (
    <Page name="Graph" desc="Relationships between defined components" icon={faDiagramProject}>
      {components.length > 0 && <ComponentGraph components={components} />}
    </Page>
  );
}

export default Graph;
