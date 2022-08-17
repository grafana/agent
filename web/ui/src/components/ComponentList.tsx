import { FC } from 'react';
import { NavLink } from 'react-router-dom';
import { HealthLabel } from '../features/component/HealthLabel';
import { ComponentInfo } from '../features/component/types';
import styles from './ComponentList.module.css';

interface ComponentListProps {
  components: ComponentInfo[];
}

const ComponentList: FC<ComponentListProps> = ({ components }) => {
  return (
    <div className={styles.list}>
      <header>
        <ul>
          <li>Health</li>
          <li>ID</li>
        </ul>
      </header>
      {components.map((component) => {
        return (
          <ul>
            <li>
              <HealthLabel health={component.health.type} />
            </li>
            <li className={styles.text}>{component.id}</li>
            <li>
              <NavLink to={'/component/' + component.id} className={styles.viewButton}>
                View
              </NavLink>
            </li>
          </ul>
        );
      })}
    </div>
  );
};

export default ComponentList;
