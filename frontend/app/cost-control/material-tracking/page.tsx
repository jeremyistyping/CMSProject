'use client';

import React, { useEffect, useState } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useModulePermissions } from '@/hooks/usePermissions';
import {
  Box,
  Heading,
  Text,
  VStack,
  HStack,
  Spinner,
  Alert,
  AlertIcon,
  useColorModeValue,
  Select,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  Card,
  CardBody,
  Button,
} from '@chakra-ui/react';
import projectService from '@/services/projectService';
import { materialTrackingService, MaterialSummaryStats, MaterialItemSummary, MaterialMovement } from '@/services/materialTrackingService';
import MaterialMovementTable from '@/components/cost-control/MaterialMovementTable';
import { Project } from '@/types/project';
import RecordUsageModal from '@/components/cost-control/RecordUsageModal';

const MaterialTrackingPage: React.FC = () => {
  const { canView, loading } = useModulePermissions('cost_control');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const cardBg = useColorModeValue('white', 'gray.700');

  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string>('');

  const [summary, setSummary] = useState<MaterialSummaryStats | null>(null);
  const [items, setItems] = useState<MaterialItemSummary[]>([]);
  const [movements, setMovements] = useState<MaterialMovement[]>([]);
  const [isLoadingData, setIsLoadingData] = useState(false);

  const [isRecordUsageModalOpen, setIsRecordUsageModalOpen] = useState(false);
  const [preSelectedProductId, setPreSelectedProductId] = useState<number | undefined>(undefined);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // Load projects on mount
  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const data = await projectService.getActiveProjects();
        setProjects(Array.isArray(data) ? data : []);
        if (Array.isArray(data) && data.length > 0) {
          setSelectedProjectId(String(data[0].id));
        }
      } catch (error) {
        console.error('Failed to load projects:', error);
      }
    };
    fetchProjects();
  }, []);

  // Load data when project changes
  useEffect(() => {
    if (!selectedProjectId) return;

    const fetchData = async () => {
      setIsLoadingData(true);
      try {
        const projectId = parseInt(selectedProjectId);
        const [summaryData, itemsData, movementsData] = await Promise.all([
          materialTrackingService.getSummary(projectId),
          materialTrackingService.getItems(projectId),
          materialTrackingService.getMovements(projectId),
        ]);

        setSummary(summaryData);
        setItems(Array.isArray(itemsData) ? itemsData : []);
        setMovements(Array.isArray(movementsData) ? movementsData : []);
      } catch (error) {
        console.error('Failed to load material data:', error);
        // Reset to empty arrays on error
        setSummary(null);
        setItems([]);
        setMovements([]);
      } finally {
        setIsLoadingData(false);
      }
    };

    fetchData();
  }, [selectedProjectId, refreshTrigger]);

  const handleRecordUsage = (productId?: number) => {
    setPreSelectedProductId(productId);
    setIsRecordUsageModalOpen(true);
  };

  const handleSuccess = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount || 0);

  if (loading) {
    return (
      <SimpleLayout>
        <Box display="flex" alignItems="center" justifyContent="center" minH="60vh">
          <Spinner size="xl" />
        </Box>
      </SimpleLayout>
    );
  }

  if (!canView) {
    return (
      <SimpleLayout>
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          Access Denied
        </Alert>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout>
      <VStack align="stretch" spacing={6}>
        <Box>
          <Heading size="lg" color={headingColor} mb={2}>Material Tracking</Heading>
          <Text color={textColor}>Monitor material usage, stock, and movements per project.</Text>
        </Box>

        {/* Filter Section */}
        <Card bg={cardBg} variant="outline">
          <CardBody>
            <HStack spacing={4} justify="space-between" align="flex-end">
              <Box minW="300px">
                <Text mb={2} fontWeight="medium">Select Project</Text>
                <Select
                  value={selectedProjectId}
                  onChange={(e) => setSelectedProjectId(e.target.value)}
                  placeholder="Select a project"
                >
                  {(projects || []).map((p) => (
                    <option key={p.id} value={p.id}>{p.project_name}</option>
                  ))}
                </Select>
              </Box>
              <Button
                colorScheme="blue"
                onClick={() => handleRecordUsage()}
                isDisabled={!selectedProjectId}
              >
                Record Usage
              </Button>
            </HStack>
          </CardBody>
        </Card>

        {isLoadingData ? (
          <Box display="flex" justifyContent="center" py={10}>
            <Spinner size="xl" />
          </Box>
        ) : !selectedProjectId ? (
          <Alert status="info">
            <AlertIcon />
            Please select a project to view material tracking data.
          </Alert>
        ) : (
          <>
            {/* Summary Cards */}
            <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
              <Card bg={cardBg} variant="outline">
                <CardBody>
                  <Stat>
                    <StatLabel>Total Purchased</StatLabel>
                    <StatNumber color="blue.500">{formatCurrency(summary?.total_purchased_value || 0)}</StatNumber>
                    <StatHelpText>{summary?.total_items || 0} Items</StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
              <Card bg={cardBg} variant="outline">
                <CardBody>
                  <Stat>
                    <StatLabel>Total Used</StatLabel>
                    <StatNumber color="orange.500">{formatCurrency(summary?.total_used_value || 0)}</StatNumber>
                    <StatHelpText>Consumed in field</StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
              <Card bg={cardBg} variant="outline">
                <CardBody>
                  <Stat>
                    <StatLabel>Remaining Value</StatLabel>
                    <StatNumber color="green.500">{formatCurrency(summary?.total_remaining_value || 0)}</StatNumber>
                    <StatHelpText>Project Stock</StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
              <Card bg={cardBg} variant="outline">
                <CardBody>
                  <Stat>
                    <StatLabel>Low Stock Items</StatLabel>
                    <StatNumber color={summary?.low_stock_items ? "red.500" : "gray.500"}>
                      {summary?.low_stock_items || 0}
                    </StatNumber>
                    <StatHelpText>Need attention</StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </SimpleGrid>

            {/* Tabs for Details */}
            <Card bg={cardBg} variant="outline">
              <CardBody>
                <Tabs variant="enclosed">
                  <TabList>
                    <Tab>Material Items</Tab>
                    <Tab>Movements History</Tab>
                  </TabList>
                  <TabPanels>
                    <TabPanel px={0}>
                      <Box overflowX="auto">
                        <Table variant="simple" size="sm">
                          <Thead>
                            <Tr>
                              <Th>Item Name</Th>
                              <Th>Category</Th>
                              <Th isNumeric>Purchased</Th>
                              <Th isNumeric>Used</Th>
                              <Th isNumeric>Remaining</Th>
                              <Th isNumeric>Avg Cost</Th>
                              <Th isNumeric>Total Value</Th>
                              <Th>Status</Th>
                              <Th>Actions</Th>
                            </Tr>
                          </Thead>
                          <Tbody>
                            {(items || []).map((item) => (
                              <Tr key={item.product_id}>
                                <Td>
                                  <Text fontWeight="medium">{item.product_name}</Text>
                                  <Text fontSize="xs" color="gray.500">{item.product_code}</Text>
                                </Td>
                                <Td>{item.category}</Td>
                                <Td isNumeric>{item.purchased_qty} {item.unit}</Td>
                                <Td isNumeric>{item.used_qty} {item.unit}</Td>
                                <Td isNumeric fontWeight="bold">{item.remaining_qty} {item.unit}</Td>
                                <Td isNumeric>{formatCurrency(item.avg_unit_cost)}</Td>
                                <Td isNumeric>{formatCurrency(item.total_value)}</Td>
                                <Td>
                                  <Badge
                                    colorScheme={
                                      item.status === 'CRITICAL' ? 'red' :
                                        item.status === 'LOW' ? 'orange' : 'green'
                                    }
                                  >
                                    {item.status}
                                  </Badge>
                                </Td>
                                <Td>
                                  <Button
                                    size="xs"
                                    colorScheme="blue"
                                    variant="outline"
                                    onClick={() => handleRecordUsage(item.product_id)}
                                  >
                                    Record
                                  </Button>
                                </Td>
                              </Tr>
                            ))}
                            {items.length === 0 && (
                              <Tr>
                                <Td colSpan={8} textAlign="center" py={4} color="gray.500">
                                  No material items found for this project.
                                </Td>
                              </Tr>
                            )}
                          </Tbody>
                        </Table>
                      </Box>
                    </TabPanel>
                    <TabPanel px={0}>
                      <MaterialMovementTable movements={movements} />
                    </TabPanel>
                  </TabPanels>
                </Tabs>
              </CardBody>
            </Card>
          </>
        )}

        <RecordUsageModal
          isOpen={isRecordUsageModalOpen}
          onClose={() => setIsRecordUsageModalOpen(false)}
          projectId={parseInt(selectedProjectId)}
          onSuccess={handleSuccess}
          preSelectedProductId={preSelectedProductId}
        />
      </VStack>
    </SimpleLayout>
  );
};

export default MaterialTrackingPage;
