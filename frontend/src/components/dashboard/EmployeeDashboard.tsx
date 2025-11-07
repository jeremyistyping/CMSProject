'use client';
import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import api from '@/services/api';
import { API_ENDPOINTS } from '@/config/api';
import { 
  Box, 
  Heading, 
  Text, 
  Card,
  CardHeader,
  CardBody,
  Button,
  HStack,
  Icon,
  List,
  ListItem,
  ListIcon,
  Badge,
  Flex,
  Spinner
} from '@chakra-ui/react';
import {
  FiUser,
  FiPlus,
  FiFileText,
  FiActivity,
  FiBell
} from 'react-icons/fi';

interface DashboardSummary {
  statistics?: Record<string, any>;
  recent_activities?: Array<{
    id: number;
    action: string;
    table_name?: string;
    record_id?: number;
    user_id?: number;
    created_at?: string;
  }>;
  unread_notifications?: number;
  min_stock_alerts_count?: number;
}

interface EmployeeDashboardData {
  pending_approvals?: Array<{
    id: number;
    title: string;
    description: string;
    status: string;
    created_at: string;
    urgency_level?: string;
    days_pending?: number;
  }>;
  approval_notifications?: Array<{
    id: number;
    title: string;
    message: string;
    type: string;
    status: string;
    created_at: string;
    urgency_level?: string;
  }>;
  purchase_requests?: Array<{
    id: number;
    title: string;
    status: string;
    total_amount: number;
    approval_step: string;
    days_pending?: number;
    urgency_level?: string;
  }>;
  workflows?: Array<{
    id: number;
    name: string;
    status: string;
    total_steps: number;
    current_step: number;
  }>;
  summary?: {
    total_pending_approvals: number;
    total_notifications: number;
    urgent_items: number;
  };
}

export const EmployeeDashboard = () => {
  const router = useRouter();
  const [employeeData, setEmployeeData] = useState<EmployeeDashboardData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchEmployeeDashboard = async () => {
      try {
        // Use employee dashboard endpoint
        const res = await api.get(API_ENDPOINTS.DASHBOARD_EMPLOYEE);
        setEmployeeData(res.data?.data || res.data || {});
        setError(null);
      } catch (e: any) {
        console.error('Employee dashboard fetch error:', e);
        setError(e?.response?.data?.error || e?.message || 'Gagal memuat dashboard karyawan');
      } finally {
        setLoading(false);
      }
    };
    fetchEmployeeDashboard();
  }, []);
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Saya
      </Heading>

      {loading ? (
        <Flex justify="center" align="center" minH="120px">
          <Spinner size="lg" color="brand.500" thickness="4px" />
        </Flex>
      ) : (
        <>
          {error && (
            <Box bg="red.50" p={4} borderRadius="lg" borderLeft="4px solid" borderColor="red.500" mb={4}>
              <Text color="red.700">{error}</Text>
            </Box>
          )}

          {/* Approval Notifications */}
          {employeeData?.approval_notifications && employeeData.approval_notifications.length > 0 && (
            <Card mt={6}>
              <CardHeader>
                <Heading size="md" display="flex" alignItems="center">
                  <Icon as={FiBell} mr={2} color="orange.500" />
                  Notifikasi Approval
                </Heading>
              </CardHeader>
              <CardBody>
                <List spacing={3}>
                  {employeeData.approval_notifications.slice(0, 5).map((notif) => (
                    <ListItem key={notif.id} display="flex" alignItems="center" p={3} bg="gray.50" borderRadius="md">
                      <ListIcon as={FiBell} color={notif.urgency_level === 'urgent' ? 'red.500' : 'orange.500'} />
                      <Box flex="1">
                        <Text fontWeight="medium">{notif.title}</Text>
                        <Text fontSize="sm" color="gray.600">
                          {notif.message}
                        </Text>
                        <Text fontSize="xs" color="gray.500">
                          {new Date(notif.created_at).toLocaleString('id-ID')}
                        </Text>
                      </Box>
                      <Badge 
                        colorScheme={notif.urgency_level === 'urgent' ? 'red' : 'orange'}
                        size="sm"
                      >
                        {notif.urgency_level || notif.status}
                      </Badge>
                    </ListItem>
                  ))}
                </List>
              </CardBody>
            </Card>
          )}

          {/* Purchase Requests */}
          {employeeData?.purchase_requests && employeeData.purchase_requests.length > 0 && (
            <Card mt={6}>
              <CardHeader>
                <Heading size="md" display="flex" alignItems="center">
                  <Icon as={FiFileText} mr={2} color="blue.500" />
                  Purchase Requests Saya
                </Heading>
              </CardHeader>
              <CardBody>
                <List spacing={3}>
                  {employeeData.purchase_requests.slice(0, 5).map((req) => (
                    <ListItem key={req.id} display="flex" alignItems="center" p={3} bg="gray.50" borderRadius="md">
                      <ListIcon as={FiFileText} color="blue.500" />
                      <Box flex="1">
                        <Text fontWeight="medium">{req.title}</Text>
                        <Text fontSize="sm" color="gray.600">
                          {req.approval_step} • Rp {req.total_amount?.toLocaleString('id-ID') || '0'}
                        </Text>
                        <Text fontSize="xs" color="gray.500">
                          {req.days_pending && `${req.days_pending} hari tertunda`}
                        </Text>
                      </Box>
                      <Badge 
                        colorScheme={req.status === 'approved' ? 'green' : req.status === 'rejected' ? 'red' : 'yellow'}
                        size="sm"
                      >
                        {req.status}
                      </Badge>
                    </ListItem>
                  ))}
                </List>
              </CardBody>
            </Card>
          )}

          {/* Approval Workflows */}
          {employeeData?.workflows && employeeData.workflows.length > 0 && (
            <Card mt={6}>
              <CardHeader>
                <Heading size="md" display="flex" alignItems="center">
                  <Icon as={FiActivity} mr={2} color="purple.500" />
                  Approval Workflows
                </Heading>
              </CardHeader>
              <CardBody>
                <List spacing={3}>
                  {employeeData.workflows.slice(0, 5).map((workflow) => (
                    <ListItem key={workflow.id} display="flex" alignItems="center" p={3} bg="gray.50" borderRadius="md">
                      <ListIcon as={FiActivity} color="purple.500" />
                      <Box flex="1">
                        <Text fontWeight="medium">{workflow.name}</Text>
                        <Text fontSize="sm" color="gray.600">
                          Step {workflow.current_step} dari {workflow.total_steps}
                        </Text>
                      </Box>
                      <Badge 
                        colorScheme={workflow.status === 'completed' ? 'green' : 'blue'}
                        size="sm"
                      >
                        {workflow.status}
                      </Badge>
                    </ListItem>
                  ))}
                </List>
              </CardBody>
            </Card>
          )}

          {/* Akses Cepat - Employee Features */}
          <Card mt={6}>
            <CardHeader>
              <Heading size="md" display="flex" alignItems="center">
                <Icon as={FiPlus} mr={2} color="blue.500" />
                Akses Cepat
              </Heading>
            </CardHeader>
            <CardBody>
              <Text mb={4} color="gray.600">
                Akses fitur-fitur employee dashboard dan approval workflow.
              </Text>
              <HStack spacing={4} flexWrap="wrap">
                <Button
                  leftIcon={<FiFileText />}
                  colorScheme="blue"
                  variant="outline"
                  onClick={() => router.push('/purchases')}
                  size="md"
                >
                  Purchase Requests
                </Button>
                <Button
                  leftIcon={<FiActivity />}
                  colorScheme="purple"
                  variant="outline"
                  onClick={() => window.location.reload()}
                  size="md"
                >
                  Refresh Dashboard
                </Button>
              </HStack>
            </CardBody>
          </Card>
        </>
      )}
    </Box>
  );
};
