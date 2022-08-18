import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { usePathPrefix } from '../contexts/PathPrefixContext';
import { ComponentView } from '../features/component/ComponentView';
import { ComponentDetail, componentInfoByID } from '../features/component/types';
import { useComponentInfo } from '../hooks/componentInfo';

export const ComponentDetailPage: FC = () => {
  const { id } = useParams();

  const pathPrefix = usePathPrefix();

  const components = useComponentInfo();
  const infoByID = componentInfoByID(components);

  const [component, setComponent] = useState<ComponentDetail | undefined>(undefined);

  useEffect(
    function () {
      const worker = async () => {
        const resp = await fetch(pathPrefix + 'api/v0/web/components/' + id);
        const content = await resp.json();

        console.log(content);
        setComponent(content);
      };

      worker().catch(console.error);
    },
    [id, pathPrefix]
  );

  return component ? <ComponentView component={component} info={infoByID} /> : <div></div>;
};
