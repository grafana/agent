import { faCubes } from '@fortawesome/free-solid-svg-icons';

import ComponentList from '../features/component/ComponentList';
import { SortOrder } from '../features/component/types';
import Page from '../features/layout/Page';
import { useComponentInfo } from '../hooks/componentInfo';

function PageComponentList() {
  const [components, setComponents] = useComponentInfo('');

  // TODO: make this sorting logic reusable
  const handleSorting = (sortField: string, sortOrder: SortOrder): void => {
    if (!sortField || !sortOrder) return;
    const sorted = [...components].sort((a, b) => {
      const sortValueA = sortField === 'Health' ? a.health.state.toString() : a.localID;
      const sortValueB = sortField === 'Health' ? b.health.state.toString() : b.localID;
      if (sortValueA === null) return 1;
      if (sortValueB === null) return -1;
      return (
        sortValueA.localeCompare(sortValueB, 'en', {
          numeric: true,
        }) * (sortOrder === SortOrder.ASC ? 1 : -1)
      );
    });
    setComponents(sorted);
  };

  return (
    <Page name="Components" desc="List of defined components" icon={faCubes}>
      <ComponentList components={components} handleSorting={handleSorting} />
    </Page>
  );
}

export default PageComponentList;
