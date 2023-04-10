import { useEffect, useState } from 'react';

import { ComponentInfo } from '../features/component/types';

/**
 * useComponentInfo retrieves the list of components from the API.
 *
 * @param fromComponent The component requesting component info. Required for
 * determining the proper list of components from the context of a module.
 */
export const useComponentInfo = (fromComponent?: string): ComponentInfo[] => {
  const [components, setComponents] = useState<ComponentInfo[]>([]);

  useEffect(
    function () {
      const worker = async () => {
        const fragments = (fromComponent || '').split('/');

        const infoPath =
          fragments.length === 1
            ? './api/v0/web/components'
            : `./api/v0/component/${fragments.slice(0, fragments.length - 1).join('/')}/components`;

        // Request is relative to the <base> tag inside of <head>.
        const resp = await fetch(infoPath, {
          cache: 'no-cache',
          credentials: 'same-origin',
        });
        setComponents(await resp.json());
      };

      worker().catch(console.error);
    },
    [fromComponent]
  );

  return components;
};
