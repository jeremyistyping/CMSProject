'use client';

import React from 'react';
import {
  Box,
  VStack,
  HStack,
  Skeleton,
  SkeletonText,
  SkeletonCircle,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Card,
  CardHeader,
  CardBody,
  SimpleGrid,
  Divider
} from '@chakra-ui/react';

// Sales List Skeleton
export const SalesListSkeleton: React.FC<{ rows?: number }> = ({ rows = 10 }) => (
  <VStack spacing={4} align="stretch">
    {/* Header skeleton */}
    <HStack justify="space-between">
      <Skeleton height="40px" width="200px" />
      <HStack spacing={3}>
        <Skeleton height="40px" width="120px" />
        <Skeleton height="40px" width="100px" />
      </HStack>
    </HStack>

    {/* Filters skeleton */}
    <HStack spacing={4}>
      <Skeleton height="32px" width="150px" />
      <Skeleton height="32px" width="120px" />
      <Skeleton height="32px" width="200px" />
      <Skeleton height="32px" width="80px" />
    </HStack>

    {/* Table skeleton */}
    <Box border="1px" borderColor="gray.200" borderRadius="md" overflow="hidden">
      <Table variant="simple">
        <Thead bg="gray.50">
          <Tr>
            <Th><Skeleton height="16px" width="80px" /></Th>
            <Th><Skeleton height="16px" width="100px" /></Th>
            <Th><Skeleton height="16px" width="120px" /></Th>
            <Th><Skeleton height="16px" width="80px" /></Th>
            <Th><Skeleton height="16px" width="100px" /></Th>
            <Th><Skeleton height="16px" width="60px" /></Th>
            <Th><Skeleton height="16px" width="80px" /></Th>
          </Tr>
        </Thead>
        <Tbody>
          {Array.from({ length: rows }).map((_, index) => (
            <Tr key={index}>
              <Td><Skeleton height="20px" width="90px" /></Td>
              <Td><Skeleton height="20px" width="110px" /></Td>
              <Td><Skeleton height="20px" width="130px" /></Td>
              <Td>
                <Skeleton height="20px" width="60px" borderRadius="full" />
              </Td>
              <Td><Skeleton height="20px" width="80px" /></Td>
              <Td><Skeleton height="20px" width="70px" /></Td>
              <Td>
                <HStack spacing={2}>
                  <SkeletonCircle size="24px" />
                  <SkeletonCircle size="24px" />
                  <SkeletonCircle size="24px" />
                </HStack>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </Box>

    {/* Pagination skeleton */}
    <HStack justify="space-between" align="center">
      <Skeleton height="20px" width="150px" />
      <HStack spacing={2}>
        <Skeleton height="32px" width="32px" />
        <Skeleton height="32px" width="32px" />
        <Skeleton height="32px" width="32px" />
        <Skeleton height="32px" width="32px" />
        <Skeleton height="32px" width="32px" />
      </HStack>
    </HStack>
  </VStack>
);

// Sales Detail Skeleton
export const SalesDetailSkeleton: React.FC = () => (
  <VStack spacing={6} align="stretch">
    {/* Header */}
    <HStack justify="space-between" align="start">
      <VStack align="start" spacing={2}>
        <Skeleton height="32px" width="200px" />
        <HStack spacing={4}>
          <Skeleton height="24px" width="80px" borderRadius="full" />
          <Skeleton height="20px" width="120px" />
        </HStack>
      </VStack>
      <HStack spacing={3}>
        <Skeleton height="36px" width="100px" />
        <Skeleton height="36px" width="120px" />
        <Skeleton height="36px" width="80px" />
      </HStack>
    </HStack>

    <Divider />

    {/* Sale Info Cards */}
    <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
      {Array.from({ length: 4 }).map((_, index) => (
        <Card key={index} variant="outline">
          <CardHeader pb={2}>
            <Skeleton height="16px" width="80px" />
          </CardHeader>
          <CardBody pt={0}>
            <Skeleton height="24px" width="100px" />
          </CardBody>
        </Card>
      ))}
    </SimpleGrid>

    {/* Main Content Grid */}
    <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
      {/* Left Column */}
      <VStack spacing={6} align="stretch">
        {/* Customer Info */}
        <Card variant="outline">
          <CardHeader>
            <Skeleton height="20px" width="120px" />
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="start">
              <Skeleton height="16px" width="150px" />
              <SkeletonText noOfLines={3} spacing="2" skeletonHeight="14px" />
            </VStack>
          </CardBody>
        </Card>

        {/* Sale Items */}
        <Card variant="outline">
          <CardHeader>
            <Skeleton height="20px" width="100px" />
          </CardHeader>
          <CardBody>
            <VStack spacing={4} align="stretch">
              {Array.from({ length: 3 }).map((_, index) => (
                <Box key={index} p={3} border="1px" borderColor="gray.100" borderRadius="md">
                  <HStack justify="space-between" align="start">
                    <VStack align="start" spacing={1}>
                      <Skeleton height="16px" width="120px" />
                      <Skeleton height="14px" width="80px" />
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Skeleton height="16px" width="80px" />
                      <Skeleton height="14px" width="60px" />
                    </VStack>
                  </HStack>
                </Box>
              ))}
            </VStack>
          </CardBody>
        </Card>
      </VStack>

      {/* Right Column */}
      <VStack spacing={6} align="stretch">
        {/* Financial Summary */}
        <Card variant="outline">
          <CardHeader>
            <Skeleton height="20px" width="140px" />
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              {Array.from({ length: 5 }).map((_, index) => (
                <HStack key={index} justify="space-between">
                  <Skeleton height="16px" width="100px" />
                  <Skeleton height="16px" width="80px" />
                </HStack>
              ))}
            </VStack>
          </CardBody>
        </Card>

        {/* Payment History */}
        <Card variant="outline">
          <CardHeader>
            <Skeleton height="20px" width="120px" />
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              {Array.from({ length: 2 }).map((_, index) => (
                <Box key={index} p={3} border="1px" borderColor="gray.100" borderRadius="md">
                  <HStack justify="space-between">
                    <VStack align="start" spacing={1}>
                      <Skeleton height="14px" width="80px" />
                      <Skeleton height="12px" width="100px" />
                    </VStack>
                    <Skeleton height="16px" width="70px" />
                  </HStack>
                </Box>
              ))}
            </VStack>
          </CardBody>
        </Card>
      </VStack>
    </SimpleGrid>
  </VStack>
);

// Sales Form Skeleton
export const SalesFormSkeleton: React.FC = () => (
  <VStack spacing={6} align="stretch">
    {/* Header */}
    <HStack justify="space-between">
      <Skeleton height="28px" width="200px" />
      <Skeleton height="24px" width="80px" borderRadius="full" />
    </HStack>

    {/* Basic Information */}
    <Box>
      <Skeleton height="20px" width="150px" mb={4} />
      <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
        <VStack align="start" spacing={2}>
          <Skeleton height="14px" width="80px" />
          <Skeleton height="40px" width="100%" />
        </VStack>
        <VStack align="start" spacing={2}>
          <Skeleton height="14px" width="100px" />
          <Skeleton height="40px" width="100%" />
        </VStack>
      </SimpleGrid>
    </Box>

    {/* Sale Items */}
    <Box>
      <Skeleton height="20px" width="100px" mb={4} />
      <VStack spacing={4} align="stretch">
        {Array.from({ length: 3 }).map((_, index) => (
          <Box key={index} p={4} border="1px" borderColor="gray.200" borderRadius="md">
            <SimpleGrid columns={{ base: 1, md: 4 }} spacing={4}>
              <VStack align="start" spacing={2}>
                <Skeleton height="14px" width="60px" />
                <Skeleton height="40px" width="100%" />
              </VStack>
              <VStack align="start" spacing={2}>
                <Skeleton height="14px" width="50px" />
                <Skeleton height="40px" width="100%" />
              </VStack>
              <VStack align="start" spacing={2}>
                <Skeleton height="14px" width="70px" />
                <Skeleton height="40px" width="100%" />
              </VStack>
              <VStack align="start" spacing={2}>
                <Skeleton height="14px" width="60px" />
                <Skeleton height="40px" width="100%" />
              </VStack>
            </SimpleGrid>
          </Box>
        ))}
      </VStack>
    </Box>

    {/* Financial Summary */}
    <Box p={4} bg="gray.50" borderRadius="md">
      <VStack spacing={3} align="stretch">
        {Array.from({ length: 4 }).map((_, index) => (
          <HStack key={index} justify="space-between">
            <Skeleton height="16px" width="120px" />
            <Skeleton height="16px" width="100px" />
          </HStack>
        ))}
      </VStack>
    </Box>
  </VStack>
);

// Analytics Skeleton
export const AnalyticsSkeleton: React.FC = () => (
  <VStack spacing={6} align="stretch">
    {/* Summary Cards */}
    <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
      {Array.from({ length: 4 }).map((_, index) => (
        <Card key={index} variant="outline">
          <CardBody>
            <VStack align="start" spacing={3}>
              <HStack justify="space-between" w="100%">
                <Skeleton height="16px" width="100px" />
                <SkeletonCircle size="40px" />
              </HStack>
              <Skeleton height="24px" width="80px" />
              <Skeleton height="12px" width="60px" />
            </VStack>
          </CardBody>
        </Card>
      ))}
    </SimpleGrid>

    {/* Charts */}
    <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
      <Card variant="outline">
        <CardHeader>
          <Skeleton height="20px" width="150px" />
        </CardHeader>
        <CardBody>
          <Skeleton height="300px" width="100%" />
        </CardBody>
      </Card>

      <Card variant="outline">
        <CardHeader>
          <Skeleton height="20px" width="120px" />
        </CardHeader>
        <CardBody>
          <Skeleton height="300px" width="100%" />
        </CardBody>
      </Card>
    </SimpleGrid>

    {/* Data Table */}
    <Card variant="outline">
      <CardHeader>
        <Skeleton height="20px" width="180px" />
      </CardHeader>
      <CardBody>
        <VStack spacing={3} align="stretch">
          {Array.from({ length: 5 }).map((_, index) => (
            <HStack key={index} justify="space-between">
              <Skeleton height="16px" width="150px" />
              <Skeleton height="16px" width="80px" />
              <Skeleton height="16px" width="100px" />
            </HStack>
          ))}
        </VStack>
      </CardBody>
    </Card>
  </VStack>
);

// Dashboard Summary Skeleton
export const DashboardSummarySkeleton: React.FC = () => (
  <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
    {Array.from({ length: 4 }).map((_, index) => (
      <Card key={index} variant="outline">
        <CardBody>
          <HStack justify="space-between" align="start">
            <VStack align="start" spacing={2}>
              <Skeleton height="14px" width="80px" />
              <Skeleton height="24px" width="100px" />
              <HStack spacing={2}>
                <Skeleton height="12px" width="40px" />
                <Skeleton height="12px" width="20px" />
              </HStack>
            </VStack>
            <Skeleton height="40px" width="40px" borderRadius="md" />
          </HStack>
        </CardBody>
      </Card>
    ))}
  </SimpleGrid>
);

// Payment Form Skeleton
export const PaymentFormSkeleton: React.FC = () => (
  <VStack spacing={4} align="stretch">
    <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
      <VStack align="start" spacing={2}>
        <Skeleton height="14px" width="80px" />
        <Skeleton height="40px" width="100%" />
      </VStack>
      <VStack align="start" spacing={2}>
        <Skeleton height="14px" width="100px" />
        <Skeleton height="40px" width="100%" />
      </VStack>
    </SimpleGrid>
    
    <VStack align="start" spacing={2}>
      <Skeleton height="14px" width="120px" />
      <Skeleton height="40px" width="100%" />
    </VStack>
    
    <VStack align="start" spacing={2}>
      <Skeleton height="14px" width="60px" />
      <Skeleton height="80px" width="100%" />
    </VStack>
  </VStack>
);

// Generic Card Skeleton
export const CardSkeleton: React.FC<{ lines?: number }> = ({ lines = 3 }) => (
  <Card variant="outline">
    <CardHeader>
      <Skeleton height="20px" width="150px" />
    </CardHeader>
    <CardBody>
      <SkeletonText noOfLines={lines} spacing="2" skeletonHeight="16px" />
    </CardBody>
  </Card>
);

export default {
  SalesListSkeleton,
  SalesDetailSkeleton,
  SalesFormSkeleton,
  AnalyticsSkeleton,
  DashboardSummarySkeleton,
  PaymentFormSkeleton,
  CardSkeleton
};
