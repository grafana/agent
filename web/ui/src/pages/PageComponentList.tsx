import { faCubes } from '@fortawesome/free-solid-svg-icons';
import Page from '../features/layout/Page';
import ComponentList from '../features/component/ComponentList';
import { useEffect, useState } from 'react';
import { usePathPrefix } from '../contexts/PathPrefixContext';
import { ComponentInfo } from '../features/component/types';

function PageComponentList() {
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
    <Page name="Components" desc="List of defined components" icon={faCubes}>
      <ComponentList components={components} />
    </Page>
  );
}

export default PageComponentList;
