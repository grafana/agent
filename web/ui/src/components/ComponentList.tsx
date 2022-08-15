import { FC } from 'react';
import { NavLink } from 'react-router-dom';
import styles from './ComponentList.module.css';

export enum ComponentHealth {
  HEALTHY = 'healthy',
  UNHEALTHY = 'unhealthy',
  UNKNOWN = 'unknown',
  EXITED = 'exited',
}

export interface Component {
  id: string;
  health: ComponentHealth;
}

interface ComponentListProps {
  components: Component[];
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
        const healthMappings = {
          [ComponentHealth.HEALTHY]: `${styles.health} ${styles['state-ok']}`,
          [ComponentHealth.UNHEALTHY]: `${styles.health} ${styles['state-error']}`,
          [ComponentHealth.UNKNOWN]: `${styles.health} ${styles['state-warn']}`,
          [ComponentHealth.EXITED]: `${styles.health} ${styles['state-error']}`,
        };
        const healthClass = healthMappings[component.health];

        return (
          <ul>
            <li>
              <span className={healthClass}>{component.health}</span>
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
