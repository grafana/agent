import { useEffect, useState } from 'react';
import { ComponentInfo } from '../features/component/types';

/**
 * useComponentInfo retrieves the list of components from the API.
 */
export const useComponentInfo = (): ComponentInfo[] => {
  const [components, setComponents] = useState<ComponentInfo[]>([]);

  useEffect(function () {
    const worker = async () => {
      // Request is relative to the <base> tag inside of <head>.
      const resp = await fetch('./api/v0/web/components', {
        cache: 'no-cache',
        credentials: 'same-origin',
      });
      setComponents(await resp.json());
    };

    worker().catch(console.error);
  }, []);

  return components;
};
