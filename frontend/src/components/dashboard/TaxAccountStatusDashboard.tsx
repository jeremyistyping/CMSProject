'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Text,
  Progress,
  Badge,
  Alert,
  AlertIcon,
  AlertDescription,
  Button,
  SimpleGrid,
  VStack,
  HStack,
  Icon,
  List,
  ListItem,
  ListIcon,
  Divider,
  useColorModeValue,
  Flex,
  Spinner
} from '@chakra-ui/react';
import {
  FiCheckCircle,
  FiAlertTriangle,
  FiXCircle,
  FiSettings,
  FiTrendingUp,
  FiDollarSign,
  FiCreditCard,
  FiBriefcase
} from 'react-icons/fi';
import { useRouter } from 'next/navigation';

interface AccountInfo {
  id: number;
  code: string;
  name: string;
  type: string;
  is_active: boolean;
}

interface AccountStatus {
  is_configured: boolean;
  account?: AccountInfo;
  warnings?: string[];
}

interface TaxAccountDashboard {
  is_fully_configured: boolean;
  health_score: number;
  sales_receivable: AccountStatus;
  sales_cash: AccountStatus;
  sales_bank: AccountStatus;
  sales_revenue: AccountStatus;
  sales_output_vat: AccountStatus;
  system_warnings: string[];
  last_updated?: string;
  updated_by?: {
    name: string;
  };
}

export const TaxAccountStatusDashboard: React.FC = () => {
  const [dashboard, setDashboard] = useState<TaxAccountDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');

  useEffect(() => {
    fetchDashboard();
  }, []);

  const fetchDashboard = async () => {
    try {
      const response = await fetch('/api/v1/settings/tax-accounts/status');
      const data = await response.json();

      if (data.success) {
        setDashboard(data.data);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: AccountStatus) => {
    if (!status.is_configured) {
      return <Badge colorScheme="red">Not Configured</Badge>;
    }

    if (status.warnings && status.warnings.length > 0) {
      return <Badge colorScheme="yellow">Warning</Badge>;
    }

    return <Badge colorScheme="green">Configured</Badge>;
  };

  const getStatusIcon = (status: AccountStatus) => {
    if (!status.is_configured) {
      return <Icon as={FiXCircle} color="red.500" boxSize={4} />;
    }

    if (status.warnings && status.warnings.length > 0) {
      return <Icon as={FiAlertTriangle} color="yellow.500" boxSize={4} />;
    }

    return <Icon as={FiCheckCircle} color="green.500" boxSize={4} />;
  };

  const getHealthScoreColor = (score: number) => {
    if (score >= 80) return 'green';
    if (score >= 60) return 'yellow';
    return 'red';
  };

  if (loading) {
    return (
      <Card>
        <CardBody display="flex" justifyContent="center" alignItems="center" p={8}>
          <Spinner size="lg" color="blue.500" />
        </CardBody>
      </Card>
    );
  }

  if (!dashboard) {
    return (
      <Card>
        <CardBody p={8}>
          <Alert status="error">
            <AlertIcon />
            <AlertDescription>
              Unable to load tax account configuration status.
            </AlertDescription>
          </Alert>
        </CardBody>
      </Card>
    );
  }

  return (
    <VStack spacing={6} align="stretch">
      {/* Health Score Overview */}
      <Card>
        <CardHeader pb={0}>
          <Heading size="md" display="flex" alignItems="center">
            <Icon as={FiSettings} mr={2} />
            Tax Account Configuration Health
          </Heading>
        </CardHeader>
        <CardBody>
          <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
            <Box>
              <Flex justify="space-between" mb={2}>
                <Text fontSize="sm" fontWeight="medium">Configuration Health</Text>
                <Text fontSize="sm" color="gray.500">{dashboard.health_score}%</Text>
              </Flex>
              <Progress
                value={dashboard.health_score}
                colorScheme={getHealthScoreColor(dashboard.health_score)}
                size="sm"
                borderRadius="full"
              />
              <Text mt={2} fontSize="xs" color="gray.500">
                {dashboard.is_fully_configured ? 'All required accounts configured' : 'Some accounts need configuration'}
              </Text>
            </Box>

            <VStack align="stretch" spacing={2}>
              <Flex justify="space-between" align="center">
                <Text fontSize="sm">Status:</Text>
                <Badge colorScheme={dashboard.is_fully_configured ? 'green' : 'red'}>
                  {dashboard.is_fully_configured ? 'Complete' : 'Incomplete'}
                </Badge>
              </Flex>
              {dashboard.last_updated && (
                <Text fontSize="xs" color="gray.500">
                  Last updated: {dashboard.last_updated} by {dashboard.updated_by?.name || 'System'}
                </Text>
              )}
            </VStack>
          </SimpleGrid>

          {/* System Warnings */}
          {dashboard.system_warnings.length > 0 && (
            <Alert status="warning" mt={4} borderRadius="md">
              <AlertIcon />
              <Box flex="1">
                <AlertDescription display="block">
                  <Text fontWeight="medium" mb={1}>Configuration Issues:</Text>
                  <List styleType="disc" pl={4} spacing={1}>
                    {dashboard.system_warnings.map((warning, index) => (
                      <ListItem key={index} fontSize="sm">{warning}</ListItem>
                    ))}
                  </List>
                </AlertDescription>
              </Box>
            </Alert>
          )}
        </CardBody>
      </Card>

      {/* Account Configuration Details */}
      <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
        {/* Sales Transaction Accounts */}
        <Card>
          <CardHeader pb={2}>
            <Heading size="sm" display="flex" alignItems="center">
              <Icon as={FiCreditCard} mr={2} />
              Sales Transaction Accounts
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              {[
                { label: 'Receivable Account', status: dashboard.sales_receivable },
                { label: 'Cash Account', status: dashboard.sales_cash },
                { label: 'Bank Account', status: dashboard.sales_bank }
              ].map((item, idx) => (
                <Flex key={idx} justify="space-between" align="center" p={2} borderWidth="1px" borderRadius="md" borderColor={borderColor}>
                  <HStack spacing={3}>
                    {getStatusIcon(item.status)}
                    <Box>
                      <Text fontSize="sm" fontWeight="medium">{item.label}</Text>
                      {item.status.account ? (
                        <Text fontSize="xs" color="gray.500">
                          [{item.status.account.code}] {item.status.account.name}
                        </Text>
                      ) : (
                        <Text fontSize="xs" color="red.500">Not configured</Text>
                      )}
                    </Box>
                  </HStack>
                  {getStatusBadge(item.status)}
                </Flex>
              ))}
            </VStack>
          </CardBody>
        </Card>

        {/* Revenue & Tax Accounts */}
        <Card>
          <CardHeader pb={2}>
            <Heading size="sm" display="flex" alignItems="center">
              <Icon as={FiTrendingUp} mr={2} />
              Revenue & Tax Accounts
            </Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              {[
                { label: 'Revenue Account', status: dashboard.sales_revenue },
                { label: 'Output VAT Account', status: dashboard.sales_output_vat }
              ].map((item, idx) => (
                <Flex key={idx} justify="space-between" align="center" p={2} borderWidth="1px" borderRadius="md" borderColor={borderColor}>
                  <HStack spacing={3}>
                    {getStatusIcon(item.status)}
                    <Box>
                      <Text fontSize="sm" fontWeight="medium">{item.label}</Text>
                      {item.status.account ? (
                        <Text fontSize="xs" color="gray.500">
                          [{item.status.account.code}] {item.status.account.name}
                        </Text>
                      ) : (
                        <Text fontSize="xs" color="red.500">Not configured</Text>
                      )}
                    </Box>
                  </HStack>
                  {getStatusBadge(item.status)}
                </Flex>
              ))}

              {/* Action Button */}
              <Box pt={4} borderTopWidth="1px" borderColor={borderColor}>
                <Button
                  variant="outline"
                  size="sm"
                  w="full"
                  leftIcon={<Icon as={FiSettings} />}
                  onClick={() => router.push('/settings/tax-accounts')}
                >
                  Configure Accounts
                </Button>
              </Box>
            </VStack>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Account Usage Information */}
      <Card>
        <CardHeader pb={2}>
          <Heading size="md" display="flex" alignItems="center">
            <Icon as={FiBriefcase} mr={2} />
            How These Accounts Are Used
          </Heading>
        </CardHeader>
        <CardBody>
          <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4} fontSize="sm">
            <Box>
              <Text fontWeight="medium" mb={2}>Sales Transactions:</Text>
              <List spacing={1} color="gray.600">
                <ListItem>• <strong>Receivable Account:</strong> Records credit sales (customer owes money)</ListItem>
                <ListItem>• <strong>Cash Account:</strong> Records immediate cash payments</ListItem>
                <ListItem>• <strong>Bank Account:</strong> Records bank transfers and checks</ListItem>
              </List>
            </Box>
            <Box>
              <Text fontWeight="medium" mb={2}>Revenue & Tax:</Text>
              <List spacing={1} color="gray.600">
                <ListItem>• <strong>Revenue Account:</strong> Records sales income</ListItem>
                <ListItem>• <strong>Output VAT Account:</strong> Records tax obligations (PPN Keluaran)</ListItem>
              </List>
            </Box>
          </SimpleGrid>

          <Box mt={4} p={3} bg="blue.50" borderRadius="lg">
            <HStack align="start" spacing={2}>
              <Icon as={FiDollarSign} color="blue.600" mt={1} />
              <Box fontSize="sm">
                <Text fontWeight="medium" color="blue.900">Example Transaction:</Text>
                <Text color="blue.700">
                  When you make a sale for $1,000 + $110 VAT:
                  <br />• Revenue Account gets credited $1,000
                  <br />• Output VAT Account gets credited $110
                  <br />• Cash/Bank/Receivable Account gets debited $1,110
                </Text>
              </Box>
            </HStack>
          </Box>
        </CardBody>
      </Card>
    </VStack>
  );
};

export default TaxAccountStatusDashboard;