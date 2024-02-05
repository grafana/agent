import { useEffect, useState } from 'react';

import { ComponentInfo } from '../features/component/types';

/**
 * useComponentInfo retrieves the list of components from the API.
 *
 * @param fromComponent The component requesting component info. Required for
 * determining the proper list of components from the context of a module.
 */
export const useComponentInfo = (
  moduleID: string
): [ComponentInfo[], React.Dispatch<React.SetStateAction<ComponentInfo[]>>] => {
  const [components, setComponents] = useState<ComponentInfo[]>([]);

  useEffect(
    function () {
      const worker = async () => {
        const infoPath = moduleID === '' ? './api/v0/web/components' : `./api/v0/web/modules/${moduleID}/components`;

        // Request is relative to the <base> tag inside of <head>.
        const resp = await fetch(infoPath, {
          cache: 'no-cache',
          credentials: 'same-origin',
        });
        setComponents(await resp.json());
      };

      worker().catch(console.error);
    },
    [moduleID]
  );

  return [components, setComponents];
};

/**
 * useAllComponentInfo retrieves the list of all components (non-module and ) from the API.
 *
 * @param fromComponent The component requesting component info. Required for
 * determining the proper list of components from the context of a module.
 */
export const useAllComponentInfo = (
  moduleID: string
): [ComponentInfo[], React.Dispatch<React.SetStateAction<ComponentInfo[]>>] => {
  const [components, setComponents] = useState<ComponentInfo[]>([]);

  useEffect(
    function () {
      const worker = async () => {
        // Request is relative to the <base> tag inside of <head>.
        const totalComponents: ComponentInfo[] = [];

        const resp = await fetch('./api/v0/web/components', {
          cache: 'no-cache',
          credentials: 'same-origin',
        });

        let fetchedComponents: ComponentInfo[] = await resp.json();
        let moduleComponentIds = new Array<ComponentInfo>();
        for (const component of fetchedComponents) {
          const { localID: moduleId } = component;
          if (moduleId.startsWith('module')) {
            moduleComponentIds.push(component);
          }
          totalComponents.push(component);
        }

        // Fetch sub-components in `module.*` components
        while (moduleComponentIds.length > 0) {
          const newModuleComponentIds: ComponentInfo[] = [];
          for (const { localID: moduleId } of moduleComponentIds) {
            const resp = await fetch(`./api/v0/web/modules/${moduleId}/components`, {
              cache: 'no-cache',
              credentials: 'same-origin',
            });
            fetchedComponents = (await resp.json()) as ComponentInfo[];
            for (const component of fetchedComponents) {
              const { localID: moduleId } = component;
              if (moduleId.startsWith('module')) {
                newModuleComponentIds.push(component);
              }
              totalComponents.push(component);
            }
          }
          moduleComponentIds = newModuleComponentIds;
        }

        setComponents(totalComponents);
      };
      worker().catch(console.error);
    },
    [moduleID]
  );

  return [components, setComponents];
};
