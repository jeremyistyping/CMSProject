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
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiFileText,
  FiCheckSquare,
  FiArrowRight,
  FiPlus,
  FiClock,
  FiCheckCircle,
  FiXCircle,
  FiPackage,
  FiAlertCircle,
} from 'react-icons/fi';
import { useRouter } from 'next/navigation';
import purchaseRequestService from '@/services/purchaseRequestService';
import { materialTrackingService } from '@/services/materialTrackingService';

interface PRStats {
  total: number;
  pending: number;
  approved: number;
  rejected: number;
}

interface RecentPR {
  id: number;
  code: string;
  project_name: string;
  total_amount: number;
  status: string;
  created_at: string;
}

export const PurchasingDashboard = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [prStats, setPrStats] = useState<PRStats>({ total: 0, pending: 0, approved: 0, rejected: 0 });
  const [recentPRs, setRecentPRs] = useState<RecentPR[]>([]);
  const [lowStockCount, setLowStockCount] = useState(0);

  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const prs = await purchaseRequestService.getAll();
      
      // Calculate stats
      const stats: PRStats = {
        total: prs.length,
        pending: prs.filter(pr => pr.status?.toLowerCase() === 'pending').length,
        approved: prs.filter(pr => pr.status?.toLowerCase() === 'approved').length,
        rejected: prs.filter(pr => pr.status?.toLowerCase() === 'rejected').length,
      };
      setPrStats(stats);

      // Get recent PRs (last 5)
      const sorted = [...prs].sort((a, b) => 
        new Date(b.created_at || '').getTime() - new Date(a.created_at || '').getTime()
      );
      setRecentPRs(sorted.slice(0, 5).map(pr => ({
        id: pr.id,
        code: pr.code || `PR-${pr.id}`,
        project_name: pr.project?.project_name || '-',
        total_amount: pr.total_amount || 0,
        status: pr.status || 'pending',
        created_at: pr.created_at || '',
      })));

    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'approved': return 'green';
      case 'pending': return 'yellow';
      case 'rejected': return 'red';
      default: return 'gray';
    }
  };

  if (loading) {
    return (
      <Flex justify="center" align="center" minH="60vh">
        <Spinner size="xl" color="blue.500" thickness="4px" />
      </Flex>
    );
  }

  return (
    <Box>
      <Heading as="h2" size="lg" color={useColorModeValue('gray.800', 'white')} mb={6}>
        Dashboard Purchasing
      </Heading>

      {/* Stats Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Total PR</StatLabel>
              <StatNumber color="blue.500">{prStats.total}</StatNumber>
              <StatHelpText>Semua purchase request</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Menunggu Approval</StatLabel>
              <StatNumber color="yellow.500">{prStats.pending}</StatNumber>
              <StatHelpText>
                <Icon as={FiClock} mr={1} />
                Pending review
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Disetujui</StatLabel>
              <StatNumber color="green.500">{prStats.approved}</StatNumber>
              <StatHelpText>
                <Icon as={FiCheckCircle} mr={1} />
                Approved
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        <Card bg={cardBg}>
          <CardBody>
            <Stat>
              <StatLabel>Ditolak</StatLabel>
              <StatNumber color="red.500">{prStats.rejected}</StatNumber>
              <StatHelpText>
                <Icon as={FiXCircle} mr={1} />
                Rejected
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        {/* Recent Purchase Requests */}
        <Card bg={cardBg}>
          <CardHeader>
            <HStack justify="space-between">
              <Heading size="md" display="flex" alignItems="center">
                <Icon as={FiFileText} mr={2} color="blue.500" />
                Purchase Request Terbaru
              </Heading>
              <Button
                size="sm"
                rightIcon={<FiArrowRight />}
                variant="ghost"
                colorScheme="blue"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Lihat Semua
              </Button>
            </HStack>
          </CardHeader>
          <CardBody pt={0}>
            {recentPRs.length > 0 ? (
              <Table size="sm" variant="simple">
                <Thead>
                  <Tr>
                    <Th>No. PR</Th>
                    <Th>Proyek</Th>
                    <Th>Status</Th>
                    <Th isNumeric>Nilai</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {recentPRs.map((pr) => (
                    <Tr
                      key={pr.id}
                      cursor="pointer"
                      _hover={{ bg: hoverBg }}
                      onClick={() => router.push('/cost-control/purchase-requests')}
                    >
                      <Td fontWeight="medium">{pr.code}</Td>
                      <Td><Text noOfLines={1}>{pr.project_name}</Text></Td>
                      <Td>
                        <Badge colorScheme={getStatusColor(pr.status)} size="sm">
                          {pr.status}
                        </Badge>
                      </Td>
                      <Td isNumeric fontSize="sm">{formatCurrency(pr.total_amount)}</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            ) : (
              <Flex justify="center" align="center" py={8} color="gray.500">
                <Icon as={FiAlertCircle} mr={2} />
                <Text>Belum ada purchase request</Text>
              </Flex>
            )}
          </CardBody>
        </Card>

        {/* Quick Actions */}
        <Card bg={cardBg}>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiPlus} mr={2} color="green.500" />
              Akses Cepat
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Button
                leftIcon={<FiPlus />}
                colorScheme="green"
                size="lg"
                onClick={() => router.push('/cost-control/purchase-requests?action=create')}
              >
                Buat Purchase Request Baru
              </Button>
              <Button
                leftIcon={<FiCheckSquare />}
                colorScheme="blue"
                variant="outline"
                onClick={() => router.push('/cost-control/purchase-requests')}
              >
                Kelola Purchase Request
              </Button>
              <Button
                leftIcon={<FiPackage />}
                colorScheme="purple"
                variant="outline"
                onClick={() => router.push('/cost-control/material-tracking')}
              >
                Lihat Material Tracking
              </Button>
            </VStack>
          </CardBody>
        </Card>
      </SimpleGrid>
    </Box>
  );
};
