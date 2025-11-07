'use client';

import React from 'react';
import {
  VStack,
  HStack,
  Text,
  Button,
  Box,
  Icon,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Card,
  CardBody,
  Badge,
  Divider,
  useColorModeValue,
  List,
  ListItem,
  ListIcon,
} from '@chakra-ui/react';
import {
  FiFileText,
  FiPlus,
  FiSettings,
  FiDatabase,
  FiAlertTriangle,
  FiCheckCircle,
  FiArrowRight,
} from 'react-icons/fi';

interface BalanceSheetEmptyStateProps {
  onClose?: () => void;
  onCreateSampleData?: () => void;
  onRefresh?: () => void;
}

export const BalanceSheetEmptyState: React.FC<BalanceSheetEmptyStateProps> = ({
  onClose,
  onCreateSampleData,
  onRefresh,
}) => {
  const iconColor = useColorModeValue('gray.400', 'gray.500');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

  const navigateToAccounts = () => {
    window.open('/accounts', '_blank');
  };

  const navigateToJournal = () => {
    window.open('/journal', '_blank');
  };

  const handleCreateSampleData = async () => {
    if (onCreateSampleData) {
      onCreateSampleData();
    }
  };

  return (
    <VStack spacing={6} align="stretch" py={8}>
      {/* Main Icon and Title */}
      <VStack spacing={4} textAlign="center">
        <Box p={4} bg="gray.100" borderRadius="full">
          <Icon as={FiFileText} boxSize={12} color={iconColor} />
        </Box>
        <VStack spacing={2}>
          <Text fontSize="xl" fontWeight="bold" color={textColor}>
            Balance Sheet Is Empty
          </Text>
          <Text fontSize="md" color={textColor} maxW="md">
            No balance sheet data was found for the selected date.
          </Text>
        </VStack>
      </VStack>

      {/* Diagnostic Information */}
      <Alert status="info" borderRadius="md">
        <AlertIcon />
        <VStack align="start" spacing={2} flex="1">
          <AlertTitle>Why is the balance sheet empty?</AlertTitle>
          <AlertDescription>
            <List spacing={1}>
              <ListItem>
                <ListIcon as={FiAlertTriangle} color="orange.500" />
                No accounts have been created with opening balances
              </ListItem>
              <ListItem>
                <ListIcon as={FiAlertTriangle} color="orange.500" />
                No journal entries have been posted to balance sheet accounts
              </ListItem>
              <ListItem>
                <ListIcon as={FiAlertTriangle} color="orange.500" />
                All asset, liability, and equity account balances are zero
              </ListItem>
            </List>
          </AlertDescription>
        </VStack>
      </Alert>

      {/* Solution Steps */}
      <Card bg={cardBg} borderColor={borderColor}>
        <CardBody>
          <VStack spacing={4} align="stretch">
            <HStack>
              <Icon as={FiCheckCircle} color="green.500" />
              <Text fontSize="lg" fontWeight="semibold">
                How to Fix This
              </Text>
            </HStack>
            
            <VStack spacing={3} align="stretch">
              {/* Step 1: Set up accounts */}
              <Card size="sm" variant="outline">
                <CardBody>
                  <HStack justify="space-between">
                    <VStack align="start" spacing={1} flex="1">
                      <HStack>
                        <Badge colorScheme="blue">Step 1</Badge>
                        <Text fontWeight="semibold">Set Up Chart of Accounts</Text>
                      </HStack>
                      <Text fontSize="sm" color={textColor}>
                        Create asset, liability, and equity accounts with opening balances
                      </Text>
                    </VStack>
                    <Button
                      size="sm"
                      leftIcon={<FiSettings />}
                      colorScheme="blue"
                      variant="outline"
                      onClick={navigateToAccounts}
                    >
                      Manage Accounts
                    </Button>
                  </HStack>
                </CardBody>
              </Card>

              {/* Step 2: Create journal entries */}
              <Card size="sm" variant="outline">
                <CardBody>
                  <HStack justify="space-between">
                    <VStack align="start" spacing={1} flex="1">
                      <HStack>
                        <Badge colorScheme="green">Step 2</Badge>
                        <Text fontWeight="semibold">Record Transactions</Text>
                      </HStack>
                      <Text fontSize="sm" color={textColor}>
                        Post journal entries to record business transactions
                      </Text>
                    </VStack>
                    <Button
                      size="sm"
                      leftIcon={<FiPlus />}
                      colorScheme="green"
                      variant="outline"
                      onClick={navigateToJournal}
                    >
                      Create Entry
                    </Button>
                  </HStack>
                </CardBody>
              </Card>

              {/* Step 3: Generate sample data */}
              <Card size="sm" variant="outline">
                <CardBody>
                  <HStack justify="space-between">
                    <VStack align="start" spacing={1} flex="1">
                      <HStack>
                        <Badge colorScheme="purple">Step 3</Badge>
                        <Text fontWeight="semibold">Use Sample Data</Text>
                      </HStack>
                      <Text fontSize="sm" color={textColor}>
                        Generate sample balance sheet data for testing purposes
                      </Text>
                    </VStack>
                    <Button
                      size="sm"
                      leftIcon={<FiDatabase />}
                      colorScheme="purple"
                      variant="outline"
                      onClick={handleCreateSampleData}
                      isDisabled={!onCreateSampleData}
                    >
                      Create Sample Data
                    </Button>
                  </HStack>
                </CardBody>
              </Card>
            </VStack>
          </VStack>
        </CardBody>
      </Card>

      {/* Sample Data Information */}
      <Alert status="warning" borderRadius="md">
        <AlertIcon />
        <VStack align="start" spacing={2} flex="1">
          <AlertTitle>Sample Data Option</AlertTitle>
          <AlertDescription>
            If you're just getting started or testing the system, you can create sample balance sheet data 
            that includes typical accounts like Cash, Bank, Accounts Receivable, Equipment, Accounts Payable, 
            and Capital accounts with realistic balances.
          </AlertDescription>
        </VStack>
      </Alert>

      {/* Action Buttons */}
      <HStack spacing={3} justify="center">
        <Button variant="ghost" onClick={onClose}>
          Close
        </Button>
        <Button
          colorScheme="blue"
          leftIcon={<FiArrowRight />}
          onClick={navigateToAccounts}
        >
          Set Up Accounts
        </Button>
        {onRefresh && (
          <Button
            colorScheme="green"
            variant="outline"
            onClick={onRefresh}
          >
            Refresh Report
          </Button>
        )}
      </HStack>
    </VStack>
  );
};

export default BalanceSheetEmptyState;