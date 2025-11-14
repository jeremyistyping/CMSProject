'use client';

import React, { useState, useEffect } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import {
  Box,
  Heading,
  Text,
  SimpleGrid,
  Button,
  VStack,
  HStack,
  useToast,
  Card,
  CardBody,
  Icon,
  Badge,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  Select,
  useColorModeValue,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
} from '@chakra-ui/react';
import { 
  FiTrendingUp,
  FiDollarSign,
  FiActivity,
  FiPieChart,
  FiEye,
  FiBarChart2,
} from 'react-icons/fi';
import api from '@/services/api';

// Report definitions
const getAvailableReports = () => [
  {
    id: 'budget-vs-actual',
    name: 'Budget vs Actual by COA Group',
    description: 'Menampilkan total estimasi vs realisasi per akun',
    type: 'PROJECT',
    icon: FiTrendingUp,
    color: 'blue'
  },
  {
    id: 'profitability',
    name: 'Profitability Report per Project',
    description: '(Pendapatan) – (Total Beban Langsung + Operasional)',
    type: 'PROJECT',
    icon: FiDollarSign,
    color: 'green'
  },
  {
    id: 'cash-flow',
    name: 'Cash Flow per Project',
    description: 'Dari kas masuk & kas keluar sesuai COA tipe Asset/Expense',
    type: 'PROJECT',
    icon: FiActivity,
    color: 'purple'
  },
  {
    id: 'cost-summary',
    name: 'Cost Summary Report',
    description: 'Rekap per kategori (Material, Sewa, Labour, dll)',
    type: 'PROJECT',
    icon: FiPieChart,
    color: 'orange'
  },
  {
    id: 'portfolio-budget-vs-actual',
    name: 'Portfolio Budget vs Actual per Project',
    description: 'Ringkasan budget, actual, dan progress fisik per proyek',
    type: 'PORTFOLIO',
    icon: FiBarChart2,
    color: 'teal'
  }
];

const ReportsPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [selectedReport, setSelectedReport] = useState<any>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [reportData, setReportData] = useState<any>(null);
  const [projects, setProjects] = useState<any[]>([]);
  const toast = useToast();

  // Form parameters
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [projectId, setProjectId] = useState('');

  // Color mode values
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const headingColor = useColorModeValue('gray.700', 'white');
  const textColor = useColorModeValue('gray.800', 'white');
  const descriptionColor = useColorModeValue('gray.600', 'gray.300');

  const availableReports = getAvailableReports();

  // Load projects for dropdown
  useEffect(() => {
    const loadProjects = async () => {
      try {
        const response = await api.get('/api/v1/projects');
        setProjects(response.data.data || response.data || []);
      } catch (error) {
        console.error('Failed to load projects:', error);
      }
    };
    loadProjects();

    // Set default dates (current month)
    const now = new Date();
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
    setStartDate(firstDay.toISOString().split('T')[0]);
    setEndDate(now.toISOString().split('T')[0]);
  }, []);

  const handleViewReport = (report: any) => {
    setSelectedReport(report);
    setReportData(null);
    setIsModalOpen(true);
  };

  const handleGenerateReport = async () => {
    if (!selectedReport) return;

    setLoading(true);
    try {
      const params: any = {
        start_date: startDate,
        end_date: endDate,
      };

      if (projectId) {
        params.project_id = projectId;
      }

      const response = await api.get(`/api/v1/project-reports/${selectedReport.id}`, { params });
      
      setReportData(response.data.data);
      
      toast({
        title: 'Success',
        description: 'Report generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to generate report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount || 0);
  };

  const renderReportContent = () => {
    if (!reportData) return null;

    switch (selectedReport.id) {
      case 'budget-vs-actual':
        return (
          <VStack spacing={4} align="stretch">
            <HStack spacing={4}>
              <Stat>
                <StatLabel>Total Budget</StatLabel>
                <StatNumber>{formatCurrency(reportData.total_budget)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Total Actual</StatLabel>
                <StatNumber>{formatCurrency(reportData.total_actual)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Variance</StatLabel>
                <StatNumber color={reportData.total_variance > 0 ? 'red.500' : 'green.500'}>
                  {formatCurrency(reportData.total_variance)}
                </StatNumber>
                <StatHelpText>
                  {reportData.variance_rate > 0 ? '+' : ''}{reportData.variance_rate?.toFixed(2)}%
                </StatHelpText>
              </Stat>
            </HStack>

            <TableContainer>
              <Table size="sm">
                <Thead>
                  <Tr>
                    <Th>COA Code</Th>
                    <Th>COA Name</Th>
                    <Th isNumeric>Budget</Th>
                    <Th isNumeric>Actual</Th>
                    <Th isNumeric>Variance</Th>
                    <Th>Status</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {reportData.coa_groups?.map((group: any, idx: number) => (
                    <Tr key={idx}>
                      <Td>{group.coa_code}</Td>
                      <Td>{group.coa_name}</Td>
                      <Td isNumeric>{formatCurrency(group.budget)}</Td>
                      <Td isNumeric>{formatCurrency(group.actual)}</Td>
                      <Td isNumeric color={group.variance > 0 ? 'red.500' : 'green.500'}>
                        {formatCurrency(group.variance)}
                      </Td>
                      <Td>
                        <Badge colorScheme={
                          group.status === 'OVER_BUDGET' ? 'red' :
                          group.status === 'UNDER_BUDGET' ? 'green' : 'blue'
                        }>
                          {group.status}
                        </Badge>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </TableContainer>
          </VStack>
        );

      case 'portfolio-budget-vs-actual': {
        const projects = reportData.projects || [];
        const totalBudget = projects.reduce((sum: number, p: any) => sum + (p.budget || 0), 0);
        const totalActual = projects.reduce((sum: number, p: any) => sum + (p.actual || 0), 0);

        return (
          <VStack spacing={4} align="stretch">
            <HStack spacing={4}>
              <Stat>
                <StatLabel>Total Projects</StatLabel>
                <StatNumber>{projects.length}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Total Budget</StatLabel>
                <StatNumber>{formatCurrency(totalBudget)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Total Actual</StatLabel>
                <StatNumber>{formatCurrency(totalActual)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Portfolio Variance</StatLabel>
                <StatNumber color={totalBudget - totalActual < 0 ? 'red.500' : 'green.500'}>
                  {formatCurrency(totalBudget - totalActual)}
                </StatNumber>
              </Stat>
            </HStack>

            <TableContainer>
              <Table size="sm">
                <Thead>
                  <Tr>
                    <Th>Project</Th>
                    <Th isNumeric>Budget</Th>
                    <Th isNumeric>Actual</Th>
                    <Th isNumeric>Variance</Th>
                    <Th isNumeric>Cost Progress</Th>
                    <Th isNumeric>Physical Progress</Th>
                    <Th isNumeric>Progress Gap</Th>
                    <Th>Status</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {projects.map((project: any, idx: number) => {
                    const costProgress = typeof project.cost_progress === 'number' ? project.cost_progress : 0;
                    const physicalProgress = typeof project.physical_progress === 'number' ? project.physical_progress : 0;
                    const progressGap = typeof project.progress_gap === 'number' ? project.progress_gap : 0;
                    const isProblem = project.status === 'OVER_BUDGET' || project.status === 'UNDER_UTILIZED';

                    let statusColorScheme: string = 'green';
                    if (project.status === 'OVER_BUDGET') {
                      statusColorScheme = 'red';
                    } else if (project.status === 'UNDER_UTILIZED') {
                      statusColorScheme = 'orange';
                    } else if (project.status === 'NO_BUDGET') {
                      statusColorScheme = 'gray';
                    }

                    return (
                      <Tr key={project.project_id || idx}>
                        <Td>{project.project_name || `Project ${project.project_id}`}</Td>
                        <Td isNumeric>{formatCurrency(project.budget)}</Td>
                        <Td isNumeric>{formatCurrency(project.actual)}</Td>
                        <Td isNumeric color={project.variance < 0 ? 'red.500' : 'green.500'}>
                          {formatCurrency(project.variance)}
                        </Td>
                        <Td isNumeric>{costProgress.toFixed(1)}%</Td>
                        <Td isNumeric>{physicalProgress.toFixed(1)}%</Td>
                        <Td isNumeric color={isProblem ? 'red.500' : 'gray.800'}>
                          {progressGap.toFixed(1)}%
                        </Td>
                        <Td>
                          <Badge colorScheme={statusColorScheme}>{project.status}</Badge>
                        </Td>
                      </Tr>
                    );
                  })}
                </Tbody>
              </Table>
            </TableContainer>
          </VStack>
        );
      }

      case 'profitability':
        return (
          <VStack spacing={4} align="stretch">
            <HStack spacing={4}>
              <Stat>
                <StatLabel>Total Revenue</StatLabel>
                <StatNumber>{formatCurrency(reportData.total_revenue)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Total Cost</StatLabel>
                <StatNumber>{formatCurrency(reportData.total_direct_cost + reportData.total_operational)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Total Profit</StatLabel>
                <StatNumber color={reportData.total_profit > 0 ? 'green.500' : 'red.500'}>
                  {formatCurrency(reportData.total_profit)}
                </StatNumber>
                <StatHelpText>Margin: {reportData.overall_margin?.toFixed(2)}%</StatHelpText>
              </Stat>
            </HStack>

            <TableContainer>
              <Table size="sm">
                <Thead>
                  <Tr>
                    <Th>Project</Th>
                    <Th isNumeric>Revenue</Th>
                    <Th isNumeric>Direct Cost</Th>
                    <Th isNumeric>Operational</Th>
                    <Th isNumeric>Net Profit</Th>
                    <Th isNumeric>Margin %</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {reportData.projects?.map((project: any, idx: number) => (
                    <Tr key={idx}>
                      <Td>{project.project_name}</Td>
                      <Td isNumeric>{formatCurrency(project.revenue)}</Td>
                      <Td isNumeric>{formatCurrency(project.direct_cost)}</Td>
                      <Td isNumeric>{formatCurrency(project.operational_cost)}</Td>
                      <Td isNumeric color={project.net_profit > 0 ? 'green.500' : 'red.500'}>
                        {formatCurrency(project.net_profit)}
                      </Td>
                      <Td isNumeric>{project.net_profit_margin?.toFixed(2)}%</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </TableContainer>
          </VStack>
        );

      case 'cash-flow':
        return (
          <VStack spacing={4} align="stretch">
            <HStack spacing={4}>
              <Stat>
                <StatLabel>Beginning Balance</StatLabel>
                <StatNumber>{formatCurrency(reportData.beginning_balance)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Cash In</StatLabel>
                <StatNumber color="green.500">{formatCurrency(reportData.total_cash_in)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Cash Out</StatLabel>
                <StatNumber color="red.500">{formatCurrency(reportData.total_cash_out)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Ending Balance</StatLabel>
                <StatNumber>{formatCurrency(reportData.ending_balance)}</StatNumber>
              </Stat>
            </HStack>

            {reportData.projects?.map((project: any, idx: number) => (
              <Box key={idx} borderWidth="1px" borderRadius="lg" p={4}>
                <Heading size="sm" mb={2}>{project.project_name}</Heading>
                <HStack spacing={4}>
                  <Stat size="sm">
                    <StatLabel>Cash In</StatLabel>
                    <StatNumber fontSize="md">{formatCurrency(project.cash_in)}</StatNumber>
                  </Stat>
                  <Stat size="sm">
                    <StatLabel>Cash Out</StatLabel>
                    <StatNumber fontSize="md">{formatCurrency(project.cash_out)}</StatNumber>
                  </Stat>
                  <Stat size="sm">
                    <StatLabel>Net</StatLabel>
                    <StatNumber fontSize="md" color={project.net_cash_flow > 0 ? 'green.500' : 'red.500'}>
                      {formatCurrency(project.net_cash_flow)}
                    </StatNumber>
                  </Stat>
                </HStack>
              </Box>
            ))}
          </VStack>
        );

      case 'cost-summary':
        return (
          <VStack spacing={4} align="stretch">
            <HStack spacing={4}>
              <Stat>
                <StatLabel>Total Cost</StatLabel>
                <StatNumber>{formatCurrency(reportData.total_cost)}</StatNumber>
              </Stat>
              <Stat>
                <StatLabel>Largest Category</StatLabel>
                <StatNumber fontSize="lg">{reportData.largest_category}</StatNumber>
                <StatHelpText>{formatCurrency(reportData.largest_amount)}</StatHelpText>
              </Stat>
            </HStack>

            <TableContainer>
              <Table size="sm">
                <Thead>
                  <Tr>
                    <Th>Category</Th>
                    <Th isNumeric>Amount</Th>
                    <Th isNumeric>Percentage</Th>
                    <Th isNumeric>Items</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {reportData.categories?.map((category: any, idx: number) => (
                    <Tr key={idx}>
                      <Td>{category.category_name}</Td>
                      <Td isNumeric>{formatCurrency(category.total_amount)}</Td>
                      <Td isNumeric>{category.percentage?.toFixed(2)}%</Td>
                      <Td isNumeric>{category.item_count}</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </TableContainer>
          </VStack>
        );

      default:
        return <Text>No data available</Text>;
    }
  };

  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'director']}>
      <Box p={8}>
        <VStack spacing={8} align="stretch">
          <Heading as="h1" size="xl" color={headingColor} fontWeight="medium">
            Project Financial Reports
          </Heading>

          <SimpleGrid columns={[1, 2, 2]} spacing={6}>
            {availableReports.map((report) => (
              <Card
                key={report.id}
                bg={cardBg}
                border="1px"
                borderColor={borderColor}
                borderRadius="md"
                _hover={{ shadow: 'md' }}
                transition="all 0.2s"
              >
                <CardBody>
                  <VStack spacing={4} align="stretch">
                    <HStack justify="space-between">
                      <Icon as={report.icon} boxSize={6} color={`${report.color}.500`} />
                      <Badge colorScheme={report.color} fontSize="xs" px={2} py={1}>
                        {report.type}
                      </Badge>
                    </HStack>
                    
                    <Heading size="md" color={textColor}>
                      {report.name}
                    </Heading>
                    
                    <Text fontSize="sm" color={descriptionColor} minH="40px">
                      {report.description}
                    </Text>
                    
                    <Button
                      colorScheme={report.color}
                      leftIcon={<FiEye />}
                      onClick={() => handleViewReport(report)}
                      width="full"
                    >
                      View Report
                    </Button>
                  </VStack>
                </CardBody>
              </Card>
            ))}
          </SimpleGrid>
        </VStack>
      </Box>

      {/* Report Modal */}
      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} size="6xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{selectedReport?.name}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} align="stretch">
              {/* Parameters Form */}
              <HStack spacing={4}>
                <FormControl>
                  <FormLabel>Start Date</FormLabel>
                  <Input 
                    type="date" 
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>End Date</FormLabel>
                  <Input 
                    type="date" 
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>Project (Optional)</FormLabel>
                  <Select 
                    placeholder="All Projects"
                    value={projectId}
                    onChange={(e) => setProjectId(e.target.value)}
                  >
                    {projects.map((project) => (
                      <option key={project.id} value={project.id}>
                        {project.project_name || project.name || `Project ${project.id}`}
                      </option>
                    ))}
                  </Select>
                </FormControl>
                <FormControl>
                  <FormLabel>&nbsp;</FormLabel>
                  <Button 
                    colorScheme="blue" 
                    onClick={handleGenerateReport}
                    isLoading={loading}
                    width="full"
                  >
                    Generate
                  </Button>
                </FormControl>
              </HStack>

              {/* Report Content */}
              {renderReportContent()}
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={() => setIsModalOpen(false)}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </SimpleLayout>
  );
};

export default ReportsPage;
