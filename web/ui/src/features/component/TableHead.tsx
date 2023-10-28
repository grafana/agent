import { useState } from 'react';

interface Props {
  headers: string[];
  handleSorting?: (header: string, sortOrder: string) => void;
}

const TableHead = ({ headers, handleSorting }: Props) => {
  const [sortField, setSortField] = useState('');
  const [order, setOrder] = useState('asc');

  const handleSortingChange = (header: string) => {
    const sortOrder = header === sortField && order === 'asc' ? 'desc' : 'asc';
    setSortField(header);
    setOrder(sortOrder);
    if (handleSorting !== undefined) {
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
