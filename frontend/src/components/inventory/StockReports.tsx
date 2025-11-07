import React, { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Button,
  Heading,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useToast,
  Text,
} from '@chakra-ui/react';
import ProductService from '@/services/productService';

const StockReports: React.FC = () => {
  const [stockReport, setStockReport] = useState<any[]>([]);
  const [totalValue, setTotalValue] = useState<number>(0);
  const toast = useToast();

  useEffect(() => {
    fetchStockReport();
  }, []);

  const fetchStockReport = async () => {
    try {
      const data = await ProductService.getStockReport();
      setStockReport(data.data);
      setTotalValue(data.total_value);
    } catch (error) {
      toast({
        title: 'Failed to fetch stock report',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);
  };

  return (
    <Layout allowedRoles={['ADMIN', 'INVENTORY_MANAGER']}>
      <Box>
        <Heading as="h1" size="xl" mb={6}>
          Stock Reports
        </Heading>

        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Product ID</Th>
              <Th>Name</Th>
              <Th>Current Stock</Th>
              <Th>Unit Cost</Th>
              <Th>Total Value</Th>
              <Th>Status</Th>
            </Tr>
          </Thead>
          <Tbody>
            {stockReport.map((item, index) => (
              <Tr key={index}>
                <Td>{item.product_id}</Td>
                <Td>{item.name}</Td>
                <Td>{item.current_stock}</Td>
                <Td>{formatCurrency(item.unit_cost)}</Td>
                <Td>{formatCurrency(item.total_value)}</Td>
                <Td>{item.status}</Td>
              </Tr>
            ))}
          </Tbody>
        </Table>

        <Box mt={4}>
          <Text fontSize="lg" fontWeight="bold">
            Total Inventory Value: {formatCurrency(totalValue)}
          </Text>
        </Box>
      </Box>
    </Layout>
  );
};

export default StockReports;
