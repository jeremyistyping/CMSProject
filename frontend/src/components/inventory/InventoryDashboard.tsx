import React, { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  Grid,
  GridItem,
  Heading,
  Select,
  Table,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  Badge,
  Input,
  useToast,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
} from '@chakra-ui/react';
import { FiAlertTriangle, FiTrendingUp, FiTrendingDown, FiPackage } from 'react-icons/fi';
import ProductService, { Product, InventoryMovement } from '@/services/productService';

const InventoryDashboard: React.FC = () => {
  const [lowStockProducts, setLowStockProducts] = useState<Product[]>([]);
  const [inventoryMovements, setInventoryMovements] = useState<InventoryMovement[]>([]);
  const [stockValuation, setStockValuation] = useState<any>(null);
  const [valuationMethod, setValuationMethod] = useState<'FIFO' | 'LIFO' | 'Average'>('FIFO');
  const [dateFilter, setDateFilter] = useState({
    start_date: '',
    end_date: ''
  });
  const toast = useToast();

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      await Promise.all([
        fetchLowStockProducts(),
        fetchInventoryMovements(),
        fetchStockValuation()
      ]);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    }
  };

  const fetchLowStockProducts = async () => {
    try {
      const data = await ProductService.getLowStockProducts();
      setLowStockProducts(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch low stock products',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const fetchInventoryMovements = async () => {
    try {
      const data = await ProductService.getInventoryMovements(dateFilter);
      setInventoryMovements(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch inventory movements',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const fetchStockValuation = async () => {
    try {
      const data = await ProductService.getStockValuation({ method: valuationMethod });
      setStockValuation(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch stock valuation',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleValuationMethodChange = (method: 'FIFO' | 'LIFO' | 'Average') => {
    setValuationMethod(method);
    fetchStockValuation();
  };

  const getMovementTypeColor = (type: 'IN' | 'OUT') => {
    return type === 'IN' ? 'green' : 'red';
  };

  const getStockStatusColor = (stock: number, minStock: number) => {
    if (stock <= 0) return 'red';
    if (stock <= minStock) return 'orange';
    return 'green';
  };

  return (
    <Layout allowedRoles={['ADMIN', 'INVENTORY_MANAGER', 'EMPLOYEE']}>
      <Box>
        <Heading as="h1" size="xl" mb={6}>
          Inventory Dashboard
        </Heading>

        {/* Key Metrics */}
        <Grid templateColumns="repeat(4, 1fr)" gap={6} mb={8}>
          <GridItem>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Total Stock Value</StatLabel>
                  <StatNumber>
                    ${stockValuation?.total_value?.toLocaleString() || '0'}
                  </StatNumber>
                  <StatHelpText>
                    <StatArrow type="increase" />
                    Method: {valuationMethod}
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
          
          <GridItem>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Low Stock Items</StatLabel>
                  <StatNumber color="orange.500">
                    {lowStockProducts.length}
                  </StatNumber>
                  <StatHelpText>
                    <FiAlertTriangle />
                    Requires attention
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>

          <GridItem>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Recent Movements</StatLabel>
                  <StatNumber>{inventoryMovements.length}</StatNumber>
                  <StatHelpText>Last 30 days</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>

          <GridItem>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Active Products</StatLabel>
                  <StatNumber>{stockValuation?.details?.length || 0}</StatNumber>
                  <StatHelpText>
                    <FiPackage />
                    In inventory
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
        </Grid>

        <Grid templateColumns="repeat(2, 1fr)" gap={6}>
          {/* Low Stock Alert */}
          <GridItem>
            <Card>
              <CardHeader>
                <Flex justify="space-between" align="center">
                  <Heading size="md">Low Stock Alert</Heading>
                  <FiAlertTriangle color="orange" />
                </Flex>
              </CardHeader>
              <CardBody>
                {lowStockProducts.length > 0 ? (
                  <Table variant="simple" size="sm">
                    <Thead>
                      <Tr>
                        <Th>Product</Th>
                        <Th>Current Stock</Th>
                        <Th>Min Stock</Th>
                        <Th>Status</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {lowStockProducts.slice(0, 5).map((product) => (
                        <Tr key={product.id}>
                          <Td>
                            <Text fontWeight="medium">{product.name}</Text>
                            <Text fontSize="sm" color="gray.500">{product.code}</Text>
                          </Td>
                          <Td>{product.stock}</Td>
                          <Td>{product.min_stock}</Td>
                          <Td>
                            <Badge 
                              colorScheme={getStockStatusColor(product.stock, product.min_stock)}
                              variant="subtle"
                            >
                              {product.stock <= 0 ? 'Out of Stock' : 'Low Stock'}
                            </Badge>
                          </Td>
                        </Tr>
                      ))}
                    </Tbody>
                  </Table>
                ) : (
                  <Text color="gray.500">No low stock items</Text>
                )}
              </CardBody>
            </Card>
          </GridItem>

          {/* Stock Valuation */}
          <GridItem>
            <Card>
              <CardHeader>
                <Flex justify="space-between" align="center">
                  <Heading size="md">Stock Valuation</Heading>
                  <Select 
                    size="sm" 
                    value={valuationMethod}
                    onChange={(e) => handleValuationMethodChange(e.target.value as any)}
                    width="120px"
                  >
                    <option value="FIFO">FIFO</option>
                    <option value="LIFO">LIFO</option>
                    <option value="Average">Average</option>
                  </Select>
                </Flex>
              </CardHeader>
              <CardBody>
                <Text fontSize="2xl" fontWeight="bold" color="blue.500">
                  ${stockValuation?.total_value?.toLocaleString() || '0'}
                </Text>
                <Text fontSize="sm" color="gray.500" mt={2}>
                  Total inventory value using {valuationMethod} method
                </Text>
                {stockValuation?.details && (
                  <Text fontSize="sm" color="gray.600" mt={2}>
                    Based on {stockValuation.details.length} active products
                  </Text>
                )}
              </CardBody>
            </Card>
          </GridItem>
        </Grid>

        {/* Recent Inventory Movements */}
        <Card mt={6}>
          <CardHeader>
            <Flex justify="space-between" align="center">
              <Heading size="md">Recent Inventory Movements</Heading>
              <Flex gap={2}>
                <Input
                  type="date"
                  size="sm"
                  value={dateFilter.start_date}
                  onChange={(e) => setDateFilter(prev => ({ ...prev, start_date: e.target.value }))}
                />
                <Input
                  type="date"
                  size="sm"
                  value={dateFilter.end_date}
                  onChange={(e) => setDateFilter(prev => ({ ...prev, end_date: e.target.value }))}
                />
                <Button size="sm" onClick={fetchInventoryMovements}>
                  Filter
                </Button>
              </Flex>
            </Flex>
          </CardHeader>
          <CardBody>
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>Date</Th>
                  <Th>Product</Th>
                  <Th>Type</Th>
                  <Th>Quantity</Th>
                  <Th>Unit Cost</Th>
                  <Th>Total Cost</Th>
                  <Th>Notes</Th>
                </Tr>
              </Thead>
              <Tbody>
                {inventoryMovements.slice(0, 10).map((movement) => (
                  <Tr key={movement.id}>
                    <Td>
                      {new Date(movement.transaction_date).toLocaleDateString()}
                    </Td>
                    <Td>
                      <Text fontWeight="medium">{movement.product.name}</Text>
                      <Text fontSize="sm" color="gray.500">{movement.product.code}</Text>
                    </Td>
                    <Td>
                      <Badge 
                        colorScheme={getMovementTypeColor(movement.type)}
                        variant="subtle"
                      >
                        {movement.type === 'IN' ? (
                          <><FiTrendingUp /> Stock In</>
                        ) : (
                          <><FiTrendingDown /> Stock Out</>
                        )}
                      </Badge>
                    </Td>
                    <Td>{movement.quantity}</Td>
                    <Td>${movement.unit_cost.toFixed(2)}</Td>
                    <Td>${movement.total_cost.toFixed(2)}</Td>
                    <Td>
                      <Text fontSize="sm" noOfLines={1}>
                        {movement.notes || '-'}
                      </Text>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </CardBody>
        </Card>
      </Box>
    </Layout>
  );
};

export default InventoryDashboard;
