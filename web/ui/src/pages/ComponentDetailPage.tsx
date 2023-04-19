import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

import { ComponentView } from '../features/component/ComponentView';
import { ComponentDetail, componentInfoByID } from '../features/component/types';
import { useComponentInfo } from '../hooks/componentInfo';

const ComponentDetailPage: FC = () => {
  const { '*': id } = useParams();

  const components = useComponentInfo(id);
  const infoByID = componentInfoByID(components);

  const [component, setComponent] = useState<ComponentDetail | undefined>(undefined);

  useEffect(
    function () {
      if (id === undefined) {
        return;
      }

      const fragments = id.split('/');

      const infoRoot =
        fragments.length === 1
          ? './api/v0/web/components/'
          : `./api/v0/component/${fragments.slice(0, fragments.length - 1).join('/')}/components/`;

      const worker = async () => {
        // Request is relative to the <base> tag inside of <head>.
        const resp = await fetch(infoRoot + fragments[fragments.length - 1], {
          cache: 'no-cache',
          credentials: 'same-origin',
        });
        const data: ComponentDetail = await resp.json();

        // Set parent.
        if (fragments.length > 1) {
          data.parent = fragments.slice(0, fragments.length - 1).join('/');
        }

        // Get data from the component module API.
        if (data.id.startsWith('module.')) {
          const resp = await fetch(`./api/v0/component/${id}/components`);
          data.moduleInfo = await resp.json();
        }

        setComponent(data);
      };

      worker().catch(console.error);
    },
    [id]
  );

  return component ? <ComponentView component={component} info={infoByID} /> : <div></div>;
};

export default ComponentDetailPage;
