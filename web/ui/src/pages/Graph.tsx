import { useParams } from 'react-router-dom';
import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';

import { ComponentGraph } from '../features/graph/ComponentGraph';
import Page from '../features/layout/Page';
import { useComponentInfo } from '../hooks/componentInfo';
import { parseID } from '../utils/id';

export function Graph() {
  const [components] = useComponentInfo('');

  return (
    <Page name="Graph" desc="Relationships between defined components" icon={faDiagramProject}>
      {components.length > 0 && <ComponentGraph components={components} />}
    </Page>
  );
}

export function ModuleGraph() {
  const { '*': id } = useParams();

  const { localID } = parseID(id || '');
  const [components] = useComponentInfo(localID);

  return (
    <Page name="Graph" desc={`Relationships between defined components in ${localID}`} icon={faDiagramProject}>
      {components.length > 0 && <ComponentGraph components={components} />}
    </Page>
  );
}
