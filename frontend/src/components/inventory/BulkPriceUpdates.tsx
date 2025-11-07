import React, { useState } from 'react';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Button,
  Heading,
  Input,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  Flex,
  useToast,
} from '@chakra-ui/react';
import ProductService, { BulkPriceUpdate } from '@/services/productService';

const BulkPriceUpdates: React.FC = () => {
  const [priceUpdates, setPriceUpdates] = useState<BulkPriceUpdate['updates']>([
    { product_id: 1, purchase_price: undefined, sale_price: undefined },
  ]);
  const toast = useToast();

  const handleUpdateChange = (
    index: number,
    field: 'purchase_price' | 'sale_price',
    value: string
  ) => {
    const updates = [...priceUpdates];
    updates[index][field] = value !== '' ? parseFloat(value) : undefined;
    setPriceUpdates(updates);
  };

  const handleAddRow = () => {
    setPriceUpdates([
      ...priceUpdates,
      { product_id: priceUpdates.length + 1, purchase_price: undefined, sale_price: undefined },
    ]);
  };

  const handleSubmit = async () => {
    try {
      const data = await ProductService.bulkPriceUpdate({ updates: priceUpdates });
      toast({
        title: 'Bulk price update successful',
        status: 'success',
        isClosable: true,
      });
    } catch (error) {
      toast({
        title: 'Bulk price update failed',
        status: 'error',
        isClosable: true,
      });
    }
  };

  return (
    <Layout allowedRoles={['ADMIN', 'INVENTORY_MANAGER']}>
      <Box>
        <Heading as="h1" size="xl" mb={6}>
          Bulk Price Updates
        </Heading>

        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Product ID</Th>
              <Th>Purchase Price</Th>
              <Th>Sale Price</Th>
            </Tr>
          </Thead>
          <Tbody>
            {priceUpdates.map((update, index) => (
              <Tr key={index}>
                <Td>{update.product_id}</Td>
                <Td>
                  <Input
                    size="sm"
                    type="number"
                    value={update.purchase_price || ''}
                    onChange={(e) =>
                      handleUpdateChange(index, 'purchase_price', e.target.value)
                    }
                  />
                </Td>
                <Td>
                  <Input
                    size="sm"
                    type="number"
                    value={update.sale_price || ''}
                    onChange={(e) =>
                      handleUpdateChange(index, 'sale_price', e.target.value)
                    }
                  />
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>

        <Flex justify="space-between" align="center" mt={4}>
          <Button onClick={handleAddRow} colorScheme="blue">
            Add Row
          </Button>
          <Button onClick={handleSubmit} colorScheme="green">
            Submit
          </Button>
        </Flex>
      </Box>
    </Layout>
  );
};

export default BulkPriceUpdates;
