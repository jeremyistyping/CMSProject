'use client';
import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Box, 
  Flex, 
  Heading, 
  Text, 
  Card,
  CardHeader,
  CardBody,
  Button,
  HStack,
  Icon,
  Spinner,
  Alert,
  AlertIcon
} from '@chakra-ui/react';
import {
  FiDollarSign,
  FiShoppingCart,
  FiTrendingUp,
  FiPlus,
  FiBarChart2
} from 'react-icons/fi';
import api from '../../services/api';
import { API_ENDPOINTS } from '@/config/api';

interface FinanceDashboardData {
  invoices_pending_payment: number;
  invoices_not_paid: number;
  journals_need_posting: number;
  outstanding_receivables: number;
  outstanding_payables: number;
}

export const FinanceDashboard = () => {
  const router = useRouter();
  const [data, setData] = useState<FinanceDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchFinanceDashboardData();
  }, []);

  const fetchFinanceDashboardData = async () => {
    try {
      setLoading(true);
      const response = await api.get(API_ENDPOINTS.DASHBOARD_FINANCE);
      // Exclude bank reconciliation fields if present
      const d = response.data.data || {};
      setData({
        invoices_pending_payment: d.invoices_pending_payment ?? 0,
        invoices_not_paid: d.invoices_not_paid ?? 0,
        journals_need_posting: d.journals_need_posting ?? 0,
        outstanding_receivables: d.outstanding_receivables ?? 0,
        outstanding_payables: d.outstanding_payables ?? 0,
      });
    } catch (error: any) {
      console.error('Error fetching finance dashboard data:', error);
      setError(error.response?.data?.error || 'Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount || 0);
  };

  if (loading) {
    return (
      <Box>
        <Heading as="h2" size="xl" mb={6} color="gray.800">
          Dasbor Keuangan
        </Heading>
        <Flex justify="center" align="center" h="200px">
          <Spinner size="xl" color="blue.500" />
        </Flex>
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Heading as="h2" size="xl" mb={6} color="gray.800">
          Dasbor Keuangan
        </Heading>
        <Alert status="error">
          <AlertIcon />
          {error}
        </Alert>
      </Box>
    );
  }

  if (!data) {
    return (
      <Box>
        <Heading as="h2" size="xl" mb={6} color="gray.800">
          Dasbor Keuangan
        </Heading>
        <Alert status="info">
          <AlertIcon />
          No data available
        </Alert>
      </Box>
    );
  }
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Keuangan
      </Heading>
    
      <Flex gap={4} flexWrap="wrap" mt={4}>
        <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="220px">
          <Heading as="h3" size="sm" mb={2} color="orange.600">Invoice Perlu Dibayar</Heading>
          <Text fontSize="2xl" fontWeight="bold" color="orange.600">{data.invoices_pending_payment}</Text>
          <Text fontSize="sm" color="gray.500" mt={1}>
            Total Piutang: {formatCurrency(data.outstanding_receivables)}
          </Text>
        </Box>
        
        <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="220px">
          <Heading as="h3" size="sm" mb={2} color="red.600">Invoice Belum Lunas</Heading>
          <Text fontSize="2xl" fontWeight="bold" color="red.600">{data.invoices_not_paid}</Text>
          <Text fontSize="sm" color="gray.500" mt={1}>
            Total Utang: {formatCurrency(data.outstanding_payables)}
          </Text>
        </Box>
        
        <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="220px">
          <Heading as="h3" size="sm" mb={2} color="blue.600">Jurnal Perlu di-Posting</Heading>
          <Text fontSize="2xl" fontWeight="bold" color="blue.600">{data.journals_need_posting}</Text>
          <Text fontSize="sm" color="gray.500" mt={1}>Jurnal draft</Text>
        </Box>
      </Flex>

      <Card mt={6}>
        <CardHeader>
          <Heading size="md" display="flex" alignItems="center">
            <Icon as={FiPlus} mr={2} color="blue.500" />
            Akses Cepat
          </Heading>
        </CardHeader>
        <CardBody>
          <HStack spacing={4} flexWrap="wrap">
            <Button
              leftIcon={<FiDollarSign />}
              colorScheme="green"
              variant="outline"
              onClick={() => router.push('/sales')}
              size="md"
            >
              Tambah Penjualan
            </Button>
            <Button
              leftIcon={<FiShoppingCart />}
              colorScheme="orange"
              variant="outline"
              onClick={() => router.push('/purchases')}
              size="md"
            >
              Tambah Pembelian
            </Button>
            <Button
              leftIcon={<FiBarChart2 />}
              colorScheme="purple"
              variant="outline"
              onClick={() => router.push('/reports')}
              size="md"
            >
              Laporan Keuangan
            </Button>
          </HStack>
        </CardBody>
      </Card>
    </Box>
  );
};
