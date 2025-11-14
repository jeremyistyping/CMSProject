'use client';

import React from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useModulePermissions } from '@/hooks/usePermissions';
import {
  Box,
  Heading,
  Text,
  SimpleGrid,
  Card,
  CardBody,
  Icon,
  Badge,
  VStack,
  HStack,
  Spinner,
  Alert,
  AlertIcon,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiTrendingUp, FiPackage, FiLayers, FiCheckSquare, FiArrowRight } from 'react-icons/fi';
import Link from 'next/link';

const CostControlOverviewPage: React.FC = () => {
  const { canView, loading } = useModulePermissions('cost_control');
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textColor = useColorModeValue('gray.600', 'gray.300');

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

  const modules = [
    {
      id: 'budget-vs-actual',
      name: 'Budget vs Actual per Project',
      description: 'Bandingkan budget dan realisasi biaya per proyek secara ringkas.',
      href: '/cost-control/budget-vs-actual',
      icon: FiTrendingUp,
      badge: 'Budget'
    },
    {
      id: 'material-tracking',
      name: 'Material Tracking',
      description: 'Monitoring pemakaian dan pergerakan material per proyek.',
      href: '/cost-control/material-tracking',
      icon: FiPackage,
      badge: 'Material'
    },
    {
      id: 'cbs',
      name: 'Cost Breakdown Structure (CBS)',
      description: 'Struktur breakdown biaya (cost code) untuk tiap paket pekerjaan.',
      href: '/cost-control/cbs',
      icon: FiLayers,
      badge: 'Structure'
    },
    {
      id: 'purchase-requests',
      name: 'Purchase Request Management',
      description: 'Permintaan pembelian (PR) dan alur approval terkait proyek.',
      href: '/cost-control/purchase-requests',
      icon: FiCheckSquare,
      badge: 'PR'
    },
  ];

  return (
    <SimpleLayout>
      <Box>
        <VStack align="start" spacing={6} mb={8}>
          <Heading size="lg" color={headingColor}>Cost Control</Heading>
          <Text fontSize="sm" color={textColor} maxW="3xl">
            Modul Cost Control mengelola seluruh aspek pengendalian biaya proyek: mulai dari budget vs actual,
            material, struktur biaya (CBS), hingga permintaan pembelian.
          </Text>
        </VStack>

        <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
          {modules.map((mod) => (
            <Link key={mod.id} href={mod.href}>
              <Card
                bg={cardBg}
                borderWidth="1px"
                borderColor={borderColor}
                _hover={{ borderColor: 'green.500', transform: 'translateY(-2px)', boxShadow: 'lg' }}
                transition="all 0.2s ease"
                cursor="pointer"
              >
                <CardBody>
                  <HStack justify="space-between" mb={4}>
                    <HStack spacing={3}>
                      <Box
                        w={10}
                        h={10}
                        borderRadius="full"
                        bg="green.500"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                      >
                        <Icon as={mod.icon} color="white" />
                      </Box>
                      <VStack align="start" spacing={1}>
                        <Heading size="sm">{mod.name}</Heading>
                        <Badge colorScheme="green" variant="subtle">{mod.badge}</Badge>
                      </VStack>
                    </HStack>
                    <Icon as={FiArrowRight} color="green.400" />
                  </HStack>
                  <Text fontSize="sm" color={textColor}>{mod.description}</Text>
                </CardBody>
              </Card>
            </Link>
          ))}
        </SimpleGrid>
      </Box>
    </SimpleLayout>
  );
};

export default CostControlOverviewPage;

