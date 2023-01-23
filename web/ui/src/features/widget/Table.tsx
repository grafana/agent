import React from 'react';

import styles from './Table.module.css';

interface Props {
  tableHeaders: string[];
  style?: React.CSSProperties;
  renderTableData: () => JSX.Element[];
}

/**
 * Simple table component that accept a custom header, and custom render
 * function for the table data
 */
const Table = ({ tableHeaders, style = {}, renderTableData }: Props) => {
  const renderHeader = () => {
    return (
      <tr>
        {tableHeaders.map((header) => (
          <th key={header}>{header}</th>
        ))}
      </tr>
    );
  };

  return (
    <table className={styles.table}>
      <colgroup span={1} style={style} />
      <tbody>
        {tableHeaders && tableHeaders.length !== 0 ? renderHeader() : null}
        {renderTableData()}
      </tbody>
    </table>
  );
};

export default Table;
