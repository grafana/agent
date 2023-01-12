import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

import { ComponentView } from '../features/component/ComponentView';
import { ComponentDetail, componentInfoByID } from '../features/component/types';
import { useComponentInfo } from '../hooks/componentInfo';

const ComponentDetailPage: FC = () => {
  const { id } = useParams();

  const components = useComponentInfo();
  const infoByID = componentInfoByID(components);

  const [component, setComponent] = useState<ComponentDetail | undefined>(undefined);

  useEffect(
    function () {
      const worker = async () => {
        // Request is relative to the <base> tag inside of <head>.
        const resp = await fetch('./api/v0/web/components/' + id, {
          cache: 'no-cache',
          credentials: 'same-origin',
        });
        setComponent(await resp.json());
      };

      worker().catch(console.error);
    },
    [id]
  );

  return component ? <ComponentView component={component} info={infoByID} /> : <div></div>;
};

export default ComponentDetailPage;
