import { faDiagramProject } from '@fortawesome/free-solid-svg-icons';
import { useEffect, useState } from 'react';
import { usePathPrefix } from '../contexts/PathPrefixContext';
import { ComponentGraph } from '../features/component/ComponentGraph';
import { ComponentInfo } from '../features/component/types';
import Page from '../features/layout/Page';

function DAG() {
  const pathPrefix = usePathPrefix();

  const [components, setComponents] = useState<ComponentInfo[]>([]);

  useEffect(
    function () {
      const worker = async () => {
        const resp = await fetch(pathPrefix + 'api/v0/web/components');
        const jsonResp = await resp.json();

        console.log(jsonResp);
        setComponents(jsonResp);
      };

      worker().catch(console.error);
    },
    [pathPrefix]
  );

  return (
    <Page name="DAG" desc="Relationships between defined components" icon={faDiagramProject}>
      {components.length > 0 && <ComponentGraph components={components} />}
    </Page>
  );
}

export default DAG;
