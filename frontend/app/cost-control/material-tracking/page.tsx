'use client';

import React from 'react';
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
  Badge,
} from '@chakra-ui/react';

const MaterialTrackingPage: React.FC = () => {
  const { canView, loading } = useModulePermissions('cost_control');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const boxBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');

  if (loading) {
    return (
      <SimpleLayout>
        <Box display="flex" alignItems="center" justifyContent="center" minH="60vh">
          <HStack spacing={3}>
            <Spinner />
            <Text>Checking permissions...</Text>
          </HStack>
        </Box>
      </SimpleLayout>
    );
  }

  if (!canView) {
    return (
      <SimpleLayout>
        <Box maxW="xl">
          <Alert status="error" borderRadius="md">
            <AlertIcon />
            <Box>
              <Heading size="sm" mb={1}>Access Denied</Heading>
              <Text fontSize="sm">Anda tidak memiliki akses ke modul Cost Control. Silakan hubungi administrator.</Text>
            </Box>
          </Alert>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout>
      <Box>
        <VStack align="start" spacing={4} mb={6}>
          <Heading size="lg" color={headingColor}>Material Tracking</Heading>
          <Text fontSize="sm" color={textColor} maxW="3xl">
            Halaman ini akan digunakan untuk memonitor pemakaian dan pergerakan material per proyek:
            penerimaan, pemakaian ke CBS / pekerjaan, dan sisa stok terkait proyek.
          </Text>
        </VStack>

        <Box
          bg={boxBg}
          borderWidth="1px"
          borderColor={borderColor}
          borderRadius="lg"
          p={6}
        >
          <VStack align="start" spacing={3}>
            <Badge colorScheme="green" variant="subtle">Placeholder</Badge>
            <Text fontSize="sm" color={textColor}>
              Struktur halaman dan route sudah siap. Pada tahap berikutnya, di area ini akan ditambahkan:
            </Text>
            <Box as="ul" pl={5} fontSize="sm" color={textColor}>
              <Box as="li">Filter proyek dan lokasi gudang</Box>
              <Box as="li">Ringkasan total material (masuk, keluar, saldo)</Box>
              <Box as="li">Tabel movement material per item / per proyek</Box>
              <Box as="li">Integrasi dengan modul Inventory / Purchase</Box>
            </Box>
          </VStack>
        </Box>
      </Box>
    </SimpleLayout>
  );
};

export default MaterialTrackingPage;

