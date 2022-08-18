import { FC } from 'react';
import styles from './HealthLabel.module.css';
import { ComponentHealthType } from './types';

interface HealthLabelProps {
  health: ComponentHealthType;
}

export const HealthLabel: FC<HealthLabelProps> = ({ health }) => {
  const healthMappings = {
    [ComponentHealthType.HEALTHY]: `${styles.health} ${styles['state-ok']}`,
    [ComponentHealthType.UNHEALTHY]: `${styles.health} ${styles['state-error']}`,
    [ComponentHealthType.UNKNOWN]: `${styles.health} ${styles['state-warn']}`,
    [ComponentHealthType.EXITED]: `${styles.health} ${styles['state-error']}`,
  };
  const healthClass = healthMappings[health];

  return <span className={healthClass}>{health}</span>;
};
