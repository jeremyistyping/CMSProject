'use client';

import React, { useEffect, useMemo, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import {
  AdminDashboard,
  EmployeeDashboard,
  PurchasingDashboard,
  CostControlDashboard,
  GMDashboard,
  ProjectDirectorDashboard,
  ManagingDirectorDashboard
} from '@/components/dashboard';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useDashboardAnalytics } from '@/hooks/useDashboardAnalytics';
import {
  Flex,
  VStack,
  Spinner,
  Text,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useToast,
} from '@chakra-ui/react';

export default function DashboardPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [redirecting, setRedirecting] = useState(false);

  // Use shared analytics hook (real data only, no dummy)
  const { analytics, loading: analyticsLoading, error } = useDashboardAnalytics(user, token);

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!user || !token) {
      router.push('/login');
    }
  }, [user, token, router]);

  // Handle unauthorized role redirect
  useEffect(() => {
    if (user && !['admin', 'employee', 'purchasing', 'cost_control', 'project_director', 'gm', 'managing_director'].includes(user.role)) {
      setRedirecting(true);
      router.push('/unauthorized');
    }
  }, [user, router]);

  const toast = useToast();

  const roleNeedsAnalytics = useMemo(() => {
    const role = user?.role;
    return role === 'admin' || role === 'project_director' || role === 'gm' || role === 'managing_director' || role === 'cost_control';
  }, [user]);

  const isLoading = redirecting || (roleNeedsAnalytics ? analyticsLoading : false);

  const renderDashboardByRole = () => {
    if (isLoading) {
      return (
        <Flex justify="center" align="center" minH="60vh">
          <VStack spacing={4}>
            <Spinner size="xl" color="brand.500" thickness="4px" />
            <Text>{redirecting ? 'Mengalihkan...' : 'Memuat dasbor...'}</Text>
          </VStack>
        </Flex>
      );
    }

    if (error && roleNeedsAnalytics) {
      return (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          <AlertTitle mr={2}>Error!</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      );
    }

    switch (user?.role) {
      case 'admin':
        return <AdminDashboard analytics={analytics} />;
      case 'managing_director':
        return <ManagingDirectorDashboard analytics={analytics} />;
      case 'project_director':
        return <ProjectDirectorDashboard analytics={analytics} />;
      case 'gm':
        return <GMDashboard analytics={analytics} />;
      case 'cost_control':
        return <CostControlDashboard analytics={analytics} />;
      case 'employee':
        return <EmployeeDashboard />;
      case 'purchasing':
        return <PurchasingDashboard />;
      default:
        return (
          <Flex justify="center" align="center" minH="60vh">
            <VStack spacing={4}>
              <Spinner size="xl" color="brand.500" thickness="4px" />
              <Text>Mengalihkan ke halaman yang sesuai...</Text>
            </VStack>
          </Flex>
        );
    }
  };

  return (
    <SimpleLayout allowedRoles={['admin', 'employee', 'purchasing', 'cost_control', 'project_director', 'gm', 'managing_director']}>
      {renderDashboardByRole()}
    </SimpleLayout>
  );
}
