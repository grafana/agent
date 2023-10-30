import { useState } from 'react';

import { SortOrder } from './types';

interface Props {
  headers: string[];
  handleSorting?: (header: string, sortOrder: SortOrder) => void;
}

const TableHead = ({ headers, handleSorting }: Props) => {
  const [sortField, setSortField] = useState('');
  const [order, setOrder] = useState(SortOrder.ASC);

  const handleSortingChange = (header: string) => {
    const sortOrder = header === sortField && order === SortOrder.ASC ? SortOrder.DESC : SortOrder.ASC;
    if (handleSorting !== undefined) {
      setSortField(header);
      setOrder(sortOrder);
      handleSorting(header, sortOrder);
    }
  };

  return (
    <tr>
      {headers.map((header) => {
        const sortOrderHeaderAttribute = sortField === header ? order : 'default';
        return (
          <th
            key={header}
            onClick={() => handleSortingChange(header)}
            data-sort-order={handleSorting ? sortOrderHeaderAttribute : undefined}
          >
            {header}
          </th>
        );
      })}
    </tr>
  );
};

export default TableHead;
