import React from 'react';

import styles from './Table.module.css';

interface Props {
  tableHeaders: string[];
  style?: React.CSSProperties;
  renderTableData: () => JSX.Element[];
}

const Table = ({ tableHeaders, style = {}, renderTableData }: Props) => {
  return (
    <table className={styles.table}>
      <col span={1} style={style} />
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
