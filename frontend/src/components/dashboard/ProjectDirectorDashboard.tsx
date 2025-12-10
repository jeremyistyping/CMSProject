'use client';

import React from 'react';
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
  CircularProgress,
  CircularProgressLabel,
} from '@chakra-ui/react';
import {
  FiFolder,
  FiTarget,
  FiTrendingUp,
  FiArrowRight,
  FiDollarSign,
  FiAlertCircle,
  FiCheckCircle,
  FiClock,
  FiBarChart2,
} from 'react-icons/fi';
import { useRouter } from 'next/navigation';
import { DashboardAnalytics } from '@/hooks/useDashboardAnalytics';

interface ProjectDirectorDashboardProps {
  analytics: DashboardAnalytics | null;
}

export const ProjectDirectorDashboard: React.FC<ProjectDirectorDashboardProps> = ({ analytics }) => {
  const router = useRouter();

  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);

  if (!analytics) {
    return (
      <Flex justify="center" align="center" minH="60vh">
        <Spinner size="xl" color="blue.500" thickness="4px" />
      </Flex>
    );
  }

  const budgetUtilization = analytics.totalBudget > 0
    ? ((analytics.totalSpent / analytics.totalBudget) * 100)
    : 0;

  const completionRate = analytics.totalProjects > 0
    ? ((analytics.completedProjects / analytics.totalProjects) * 100)
    : 0;

  const getStatusColor = (status: string) => {
    const s = status?.toLowerCase();
    if (s === 'active' || s === 'ongoing' || s === 'in_progress') return 'green';
    if (s === 'completed') return 'blue';
    if (s === 'on-hold') return 'yellow';
    return 'gray';
  };

  return (
    <Box>
      <Heading as="h2" size="lg" color={textColor} mb={6}>
        Dashboard Direktur Proyek
      </Heading>

      {/* Key Metrics */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total Proyek</StatLabel>
              <StatNumber color="blue.500">{analytics.totalProjects}</StatNumber>
              <StatHelpText>
                <Icon as={FiFolder} mr={1} />
                {analytics.activeProjects} aktif
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Proyek Selesai</StatLabel>
              <StatNumber color="green.500">{analytics.completedProjects}</StatNumber>
              <StatHelpText>
                <Icon as={FiCheckCircle} mr={1} />
                {completionRate.toFixed(0)}% completion rate
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total Budget</StatLabel>
              <StatNumber fontSize="lg" color="purple.500">
                {formatCurrency(analytics.totalBudget)}
              </StatNumber>
              <StatHelpText>Semua proyek</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total Pengeluaran</StatLabel>
              <StatNumber fontSize="lg" color="teal.500">
                {formatCurrency(analytics.totalSpent)}
              </StatNumber>
              <StatHelpText>
                <Icon as={FiTrendingUp} mr={1} />
                {budgetUtilization.toFixed(1)}% terpakai
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 3 }} spacing={6} mb={6}>
        {/* Project Performance Overview */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiTarget} mr={2} color="blue.500" />
              Performa Proyek
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={6}>
              <Box textAlign="center">
                <CircularProgress 
                  value={completionRate} 
                  size="120px" 
                  color="green.400"
                  trackColor="gray.200"
                >
                  <CircularProgressLabel fontWeight="bold">
                    {completionRate.toFixed(0)}%
                  </CircularProgressLabel>
                </CircularProgress>
                <Text mt={2} fontWeight="medium">Completion Rate</Text>
              </Box>
              <SimpleGrid columns={2} spacing={4} w="full">
                <Box textAlign="center" p={3} bg={hoverBg} borderRadius="md">
                  <Text fontSize="2xl" fontWeight="bold" color="green.500">
                    {analytics.activeProjects}
                  </Text>
                  <Text fontSize="sm" color="gray.500">Aktif</Text>
                </Box>
                <Box textAlign="center" p={3} bg={hoverBg} borderRadius="md">
                  <Text fontSize="2xl" fontWeight="bold" color="blue.500">
                    {analytics.completedProjects}
                  </Text>
                  <Text fontSize="sm" color="gray.500">Selesai</Text>
                </Box>
              </SimpleGrid>
            </VStack>
          </CardBody>
        </Card>

        {/* Budget Overview */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiDollarSign} mr={2} color="green.500" />
              Budget Overview
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={6}>
              <Box textAlign="center">
                <CircularProgress 
                  value={Math.min(budgetUtilization, 100)} 
                  size="120px" 
                  color={budgetUtilization > 100 ? 'red.400' : budgetUtilization > 80 ? 'yellow.400' : 'teal.400'}
                  trackColor="gray.200"
                >
                  <CircularProgressLabel fontWeight="bold">
                    {budgetUtilization.toFixed(0)}%
                  </CircularProgressLabel>
                </CircularProgress>
                <Text mt={2} fontWeight="medium">Budget Utilization</Text>
              </Box>
              <Box w="full">
                <HStack justify="space-between" mb={2}>
                  <Text fontSize="sm" color="gray.500">Terpakai</Text>
                  <Text fontSize="sm" fontWeight="medium">{formatCurrency(analytics.totalSpent)}</Text>
                </HStack>
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.500">Sisa</Text>
                  <Text fontSize="sm" fontWeight="medium" color={analytics.totalBudget - analytics.totalSpent < 0 ? 'red.500' : 'green.500'}>
                    {formatCurrency(analytics.totalBudget - analytics.totalSpent)}
                  </Text>
                </HStack>
              </Box>
            </VStack>
          </CardBody>
        </Card>

        {/* Quick Actions */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiBarChart2} mr={2} color="purple.500" />
              Akses Cepat
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              <Button
                leftIcon={<FiFolder />}
                colorScheme="blue"
                onClick={() => router.push('/projects')}
              >
                Kelola Proyek
              </Button>
              <Button
                leftIcon={<FiTrendingUp />}
                colorScheme="green"
                variant="outline"
                onClick={() => router.push('/cost-control/budget-vs-actual')}
              >
                Budget vs Actual
              </Button>
              <Button
                leftIcon={<FiDollarSign />}
                colorScheme="purple"
                variant="outline"
                onClick={() => router.push('/cost-control')}
              >
                Cost Control
              </Button>
              <Button
                leftIcon={<FiClock />}
                colorScheme="orange"
                variant="outline"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Purchase Requests
              </Button>
            </VStack>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Recent Projects Table */}
      <Card bg={cardBg}>
        <CardHeader>
          <HStack justify="space-between">
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiFolder} mr={2} color="blue.500" />
              Proyek Terbaru
            </Heading>
            <Button
              size="sm"
              rightIcon={<FiArrowRight />}
              variant="ghost"
              colorScheme="blue"
              onClick={() => router.push('/projects')}
            >
              Lihat Semua
            </Button>
          </HStack>
        </CardHeader>
        <CardBody pt={0}>
          {analytics.recentProjects && analytics.recentProjects.length > 0 ? (
            <Table size="sm" variant="simple">
              <Thead>
                <Tr>
                  <Th>Nama Proyek</Th>
                  <Th>Status</Th>
                  <Th>Budget</Th>
                  <Th isNumeric>Progress</Th>
                </Tr>
              </Thead>
              <Tbody>
                {analytics.recentProjects.map((project) => (
                  <Tr
                    key={project.id}
                    cursor="pointer"
                    _hover={{ bg: hoverBg }}
                    onClick={() => router.push(`/projects/${project.id}`)}
                  >
                    <Td>
                      <Text fontWeight="medium" noOfLines={1}>{project.name}</Text>
                    </Td>
                    <Td>
                      <Badge colorScheme={getStatusColor(project.status)} size="sm">
                        {project.status}
                      </Badge>
                    </Td>
                    <Td fontSize="sm">{formatCurrency(project.budget)}</Td>
                    <Td isNumeric>
                      <HStack justify="flex-end">
                        <Progress 
                          value={project.progress} 
                          size="sm" 
                          w="60px"
                          colorScheme={project.progress >= 100 ? 'green' : 'blue'}
                          borderRadius="full"
                        />
                        <Text fontSize="sm">{project.progress}%</Text>
                      </HStack>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          ) : (
            <Flex justify="center" align="center" py={8} color="gray.500">
              <Icon as={FiAlertCircle} mr={2} />
              <Text>Belum ada proyek</Text>
            </Flex>
          )}
        </CardBody>
      </Card>
    </Box>
  );
};
