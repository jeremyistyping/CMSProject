'use client';

import React from 'react';
import {
  Box,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Text,
  HStack,
  Button,
  Input,
  Flex,
  Heading,
  IconButton,
} from '@chakra-ui/react';
import { FiChevronLeft, FiChevronRight, FiSearch } from 'react-icons/fi';

export interface DataTableColumn<T> {
  header: string;
  accessor: keyof T | ((row: T) => React.ReactNode);
}

export interface DataTableProps<T> {
  columns: DataTableColumn<T>[];
  data: T[];
  keyField: keyof T;
  title?: string;
  actions?: (row: T) => React.ReactNode;
  searchable?: boolean;
  pagination?: boolean;
  pageSize?: number;
  currentPage?: number;
  totalPages?: number;
  onPageChange?: (page: number) => void;
}

export const DataTable = <T extends Record<string, any>>({
  columns,
  data,
  keyField,
  title,
  actions,
  searchable = false,
  pagination = false,
  pageSize = 10,
  currentPage = 1,
  totalPages = 1,
  onPageChange,
}: DataTableProps<T>) => {
  const [searchTerm, setSearchTerm] = React.useState('');
  const [filteredData, setFilteredData] = React.useState(data);

  React.useEffect(() => {
    if (!searchable || !searchTerm) {
      setFilteredData(data);
      return;
    }

    const filtered = data.filter((row) =>
      Object.values(row).some((value) =>
        String(value).toLowerCase().includes(searchTerm.toLowerCase())
      )
    );
    setFilteredData(filtered);
  }, [data, searchTerm, searchable]);

  const renderCellContent = (row: T, column: DataTableColumn<T>) => {
    if (typeof column.accessor === 'function') {
      return column.accessor(row);
    }
    return String(row[column.accessor]);
  };

  return (
    <Box>
      {title && (
        <Heading size="md" mb={4} px={4} pt={4}>
          {title}
        </Heading>
      )}

      {searchable && (
        <Box px={4} pb={4}>
          <HStack>
            <Box position="relative" maxW="300px">
              <FiSearch
                style={{
                  position: 'absolute',
                  left: '12px',
                  top: '50%',
                  transform: 'translateY(-50%)',
                  zIndex: 1,
                  color: 'gray',
                }}
              />
              <Input
                placeholder="Search..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                pl="40px"
                size="sm"
              />
            </Box>
          </HStack>
        </Box>
      )}

      <TableContainer>
        <Table variant="simple">
          <Thead>
            <Tr>
              {columns.map((column, index) => (
                <Th key={index}>{column.header}</Th>
              ))}
              {actions && <Th>Actions</Th>}
            </Tr>
          </Thead>
          <Tbody>
            {filteredData.length > 0 ? (
              filteredData.map((row) => (
                <Tr key={String(row[keyField])}>
                  {columns.map((column, index) => (
                    <Td key={index}>{renderCellContent(row, column)}</Td>
                  ))}
                  {actions && <Td>{actions(row)}</Td>}
                </Tr>
              ))
            ) : (
              <Tr>
                <Td colSpan={columns.length + (actions ? 1 : 0)}>
                  <Text textAlign="center" py={8} color="gray.500">
                    No data available
                  </Text>
                </Td>
              </Tr>
            )}
          </Tbody>
        </Table>
      </TableContainer>

      {pagination && totalPages > 1 && (
        <Flex justify="space-between" align="center" px={4} py={4}>
          <Text fontSize="sm" color="gray.600">
            Page {currentPage} of {totalPages}
          </Text>
          <HStack>
            <IconButton
              aria-label="Previous page"
              icon={<FiChevronLeft />}
              size="sm"
              variant="outline"
              isDisabled={currentPage <= 1}
              onClick={() => onPageChange?.(currentPage - 1)}
            />
            <IconButton
              aria-label="Next page"
              icon={<FiChevronRight />}
              size="sm"
              variant="outline"
              isDisabled={currentPage >= totalPages}
              onClick={() => onPageChange?.(currentPage + 1)}
            />
          </HStack>
        </Flex>
      )}
    </Box>
  );
};

export default DataTable;
