import React from 'react';
import { NavLink } from 'react-router-dom';

import { HealthLabel } from '../component/HealthLabel';
import { ComponentInfo } from '../component/types';
import Table from '../widget/Table';

import styles from './ComponentList.module.css';

interface ComponentListProps {
  components: ComponentInfo[];
}

const ComponentList = ({ components }: ComponentListProps) => {
  const tableHeaders = ['Health', 'ID'];
  const tableStyles = { width: '100px' };

  const renderTableData = () => {
    return components.map(({ health, id }) => (
      <tr key={id}>
        <td>
          <HealthLabel health={health.state} />
        </td>
        <td>
          <span>{id}</span>
          <NavLink to={'/component/' + id} className={styles.viewButton}>
            View
          </NavLink>
        </td>
      </tr>
    ));
  };

  return (
    <div className={styles.list}>
      <Table tableHeaders={tableHeaders} renderTableData={renderTableData} style={tableStyles} />
    </div>
  );
};

export default ComponentList;
