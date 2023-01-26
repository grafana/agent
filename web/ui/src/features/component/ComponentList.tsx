import React from 'react';
import { NavLink } from 'react-router-dom';

import { HealthLabel } from './HealthLabel';
import { ComponentInfo } from './types';

import styles from './ComponentList.module.css';

interface ComponentListProps {
  components: ComponentInfo[];
}

const ComponentList = ({ components }: ComponentListProps) => {
  return (
    <div className={styles.list}>
      <table className={styles.table}>
        <tr>
          <th>Health</th>
          <th>ID</th>
        </tr>
        {components.map((component) => {
          return (
            <tr>
              <td>
                <HealthLabel health={component.health.state} />
              </td>
              <td>
                {component.id}
                <NavLink to={'/component/' + component.id} className={styles.viewButton}>
                  View
                </NavLink>
              </td>
            </tr>
          );
        })}
      </table>
    </div>
  );
};

export default ComponentList;
