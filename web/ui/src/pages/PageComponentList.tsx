import { faCubes } from '@fortawesome/free-solid-svg-icons';

import ComponentList from '../features/component/ComponentList';
import Page from '../features/layout/Page';
import { useComponentInfo } from '../hooks/componentInfo';

function PageComponentList() {
  const components = useComponentInfo('');

  return (
    <Page name="Components" desc="List of defined components" icon={faCubes}>
      <ComponentList components={components} />
    </Page>
  );
}

export default PageComponentList;
