import React from 'react';
import { NavLink } from 'react-router-dom';

import { HealthLabel } from '../component/HealthLabel';
import { ComponentInfo } from '../component/types';

import Table from './Table';

import styles from './ComponentList.module.css';

interface ComponentListProps {
  components: ComponentInfo[];
  parent?: string;
}

const TABLEHEADERS = ['Health', 'ID'];

const ComponentList = ({ components }: ComponentListProps) => {
  const tableStyles = { width: '100px' };

  /**
   * Custom renderer for table data
   */
  const renderTableData = () => {
    return components.map(({ health, id }) => (
      <tr key={id} style={{ lineHeight: '2' }}>
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
      <Table tableHeaders={TABLEHEADERS} renderTableData={renderTableData} style={tableStyles} />
    </div>
  );
};

export default ComponentList;
