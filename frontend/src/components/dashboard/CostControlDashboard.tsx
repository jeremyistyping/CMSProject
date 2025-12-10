'use client';

import React, { useEffect, useState } from 'react';
import {
  Box,
  Heading,
  Text,
  Card,
  CardHeader,
  CardBody,
  Button,
  VStack,
  HStack,
  Icon,
  SimpleGrid,
  Flex,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Spinner,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Progress,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiDollarSign,
  FiTrendingUp,
  FiTrendingDown,
  FiPackage,
  FiLayers,
  FiCheckSquare,
  FiArrowRight,
  FiAlertTriangle,
  FiAlertCircle,
} from 'react-icons/fi';
import { useRouter } from 'next/navigation';
import purchaseRequestService from '@/services/purchaseRequestService';
import projectService from '@/services/projectService';
import { DashboardAnalytics } from '@/hooks/useDashboardAnalytics';

interface CostControlDashboardProps {
  analytics: DashboardAnalytics | null;
}

interface ProjectBudgetSummary {
  id: number;
  name: string;
  budget: number;
  actual: number;
  variance: number;
  progress: number;
}

export const CostControlDashboard: React.FC<CostControlDashboardProps> = ({ analytics }) => {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [pendingVerification, setPendingVerification] = useState(0);
  const [projectSummaries, setProjectSummaries] = useState<ProjectBudgetSummary[]>([]);

  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      
      // Get PRs pending verification
      const prs = await purchaseRequestService.getAll();
      const pending = prs.filter(pr => 
        pr.status?.toLowerCase() === 'pending' || 
        pr.status?.toLowerCase() === 'submitted'
      ).length;
      setPendingVerification(pending);

      // Get project summaries
      const projects = await projectService.getAllProjects();
      const summaries: ProjectBudgetSummary[] = projects.slice(0, 5).map(p => ({
        id: Number(p.id),
        name: p.project_name,
        budget: p.budget || 0,
        actual: p.actual_cost || 0,
        variance: (p.budget || 0) - (p.actual_cost || 0),
        progress: p.overall_progress || 0,
      }));
      setProjectSummaries(summaries);

    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);

  const formatNumber = (value: number) =>
    new Intl.NumberFormat('id-ID').format(value);

  if (loading && !analytics) {
    return (
      <Flex justify="center" align="center" minH="60vh">
        <Spinner size="xl" color="blue.500" thickness="4px" />
      </Flex>
    );
  }

  const budgetUtilization = analytics?.totalBudget && analytics.totalBudget > 0
    ? ((analytics.totalSpent / analytics.totalBudget) * 100).toFixed(1)
    : '0';

  const remainingBudget = (analytics?.totalBudget || 0) - (analytics?.totalSpent || 0);
  const isOverBudget = remainingBudget < 0;

  return (
    <Box>
      <Heading as="h2" size="lg" color={textColor} mb={6}>
        Dashboard Cost Control
      </Heading>

      {/* Stats Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total Budget</StatLabel>
              <StatNumber fontSize="lg" color="blue.500">
                {formatCurrency(analytics?.totalBudget || 0)}
              </StatNumber>
              <StatHelpText>Semua proyek</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total Pengeluaran</StatLabel>
              <StatNumber fontSize="lg" color={isOverBudget ? 'red.500' : 'green.500'}>
                {formatCurrency(analytics?.totalSpent || 0)}
              </StatNumber>
              <StatHelpText>
                <Icon as={isOverBudget ? FiTrendingUp : FiTrendingDown} mr={1} />
                {budgetUtilization}% terpakai
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Sisa Budget</StatLabel>
              <StatNumber fontSize="lg" color={isOverBudget ? 'red.500' : 'teal.500'}>
                {formatCurrency(remainingBudget)}
              </StatNumber>
              <StatHelpText>
                {isOverBudget ? 'Over budget!' : 'Tersedia'}
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>PR Perlu Verifikasi</StatLabel>
              <StatNumber color="orange.500">{pendingVerification}</StatNumber>
              <StatHelpText>
                <Icon as={FiCheckSquare} mr={1} />
                Menunggu review
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6} mb={6}>
        {/* Budget vs Actual per Project */}
        <Card bg={cardBg}>
          <CardHeader>
            <HStack justify="space-between">
              <Heading size="md" display="flex" alignItems="center">
                <Icon as={FiTrendingUp} mr={2} color="blue.500" />
                Budget vs Actual per Proyek
              </Heading>
              <Button
                size="sm"
                rightIcon={<FiArrowRight />}
                variant="ghost"
                colorScheme="blue"
                onClick={() => router.push('/cost-control/budget-vs-actual')}
              >
                Detail
              </Button>
            </HStack>
          </CardHeader>
          <CardBody pt={0}>
            {projectSummaries.length > 0 ? (
              <VStack spacing={4} align="stretch">
                {projectSummaries.map((project) => {
                  const utilization = project.budget > 0 
                    ? (project.actual / project.budget) * 100 
                    : 0;
                  const isOver = utilization > 100;
                  
                  return (
                    <Box key={project.id} p={3} borderRadius="md" bg={hoverBg}>
                      <HStack justify="space-between" mb={2}>
                        <Text fontWeight="medium" noOfLines={1}>{project.name}</Text>
                        <Badge colorScheme={isOver ? 'red' : utilization > 80 ? 'yellow' : 'green'}>
                          {utilization.toFixed(0)}%
                        </Badge>
                      </HStack>
                      <Progress 
                        value={Math.min(utilization, 100)} 
                        size="sm" 
                        colorScheme={isOver ? 'red' : utilization > 80 ? 'yellow' : 'green'}
                        borderRadius="full"
                      />
                      <HStack justify="space-between" mt={2} fontSize="xs" color="gray.500">
                        <Text>Actual: {formatCurrency(project.actual)}</Text>
                        <Text>Budget: {formatCurrency(project.budget)}</Text>
                      </HStack>
                    </Box>
                  );
                })}
              </VStack>
            ) : (
              <Flex justify="center" align="center" py={8} color="gray.500">
                <Icon as={FiAlertCircle} mr={2} />
                <Text>Belum ada data proyek</Text>
              </Flex>
            )}
          </CardBody>
        </Card>

        {/* Quick Actions */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiDollarSign} mr={2} color="green.500" />
              Akses Cepat
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Button
                leftIcon={<FiCheckSquare />}
                colorScheme="orange"
                size="lg"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Verifikasi Purchase Request ({pendingVerification})
              </Button>
              <Button
                leftIcon={<FiTrendingUp />}
                colorScheme="blue"
                variant="outline"
                onClick={() => router.push('/cost-control/budget-vs-actual')}
              >
                Budget vs Actual Analysis
              </Button>
              <Button
                leftIcon={<FiPackage />}
                colorScheme="purple"
                variant="outline"
                onClick={() => router.push('/cost-control/material-tracking')}
              >
                Material Tracking
              </Button>
              <Button
                leftIcon={<FiLayers />}
                colorScheme="teal"
                variant="outline"
                onClick={() => router.push('/cost-control/cbs')}
              >
                Cost Breakdown Structure
              </Button>
            </VStack>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Alerts Section */}
      {isOverBudget && (
        <Card bg="red.50" borderColor="red.200" borderWidth={1}>
          <CardBody>
            <HStack>
              <Icon as={FiAlertTriangle} color="red.500" boxSize={6} />
              <Box>
                <Text fontWeight="bold" color="red.700">Peringatan Over Budget</Text>
                <Text color="red.600" fontSize="sm">
                  Total pengeluaran melebihi budget sebesar {formatCurrency(Math.abs(remainingBudget))}
                </Text>
              </Box>
            </HStack>
          </CardBody>
        </Card>
      )}
    </Box>
  );
};
