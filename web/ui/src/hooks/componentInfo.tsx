import { useEffect, useState } from 'react';
import { usePathPrefix } from '../contexts/PathPrefixContext';
import { ComponentInfo } from '../features/component/types';

/**
 * useComponentInfo retrieves the list of components from the API.
 */
export const useComponentInfo = (): ComponentInfo[] => {
  const pathPrefix = usePathPrefix();

  const [components, setComponents] = useState<ComponentInfo[]>([]);

  useEffect(
    function () {
      const worker = async () => {
        const resp = await fetch(pathPrefix + 'api/v0/web/components');
        setComponents(await resp.json());
      };

      worker().catch(console.error);
    },
    [pathPrefix]
  );

  return components;
};
