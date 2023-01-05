import React from 'react';

import styles from '../component/ComponentList.module.css';

interface Props {
  tableHeaders: string[];
  renderTableData: () => JSX.Element[];
}

const Table = ({ tableHeaders, renderTableData }: Props) => {
  return (
    <table className={styles.table}>
      <tbody>
        <tr>
          {tableHeaders.map((header) => (
            <th key={header}>{header}</th>
          ))}
        </tr>
        {renderTableData()}
      </tbody>
    </table>
  );
};

export default Table;
