'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Text,
  VStack,
  HStack,
  Badge,
  Divider,
  Button,
  Icon,
  useToast,
  Alert,
  AlertIcon,
  Flex,
  Spacer,
  Tag,
  TagLabel,
  TagLeftIcon,
  useColorModeValue,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
} from '@chakra-ui/react';
import {
  FiDollarSign,
  FiCalendar,
  FiUser,
  FiFileText,
  FiRefreshCw,
  FiArrowRight,
  FiArrowDown,
  FiExternalLink,
  FiRotateCcw,
} from 'react-icons/fi';
import { paymentService } from '@/services/paymentService';
import JournalEntryPreview from './JournalEntryPreview';
import type {
  PaymentWithJournalInfo,
  AccountBalanceUpdate,
  JournalEntry,
  PaymentJournalResult,
} from '@/services/paymentService';

interface PaymentDetailsViewProps {
  paymentId: string;
  onEdit?: (paymentId: string) => void;
  onReverse?: (paymentId: string) => void;
}

// Currency formatter
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount);
};

const PaymentDetailsView: React.FC<PaymentDetailsViewProps> = ({
  paymentId,
  onEdit,
  onReverse
}) => {
  const [paymentInfo, setPaymentInfo] = useState<PaymentWithJournalInfo | null>(null);
  const [accountBalances, setAccountBalances] = useState<AccountBalanceUpdate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isReversing, setIsReversing] = useState(false);
  const [reverseResult, setReverseResult] = useState<PaymentJournalResult | null>(null);

  const { isOpen: isReverseModalOpen, onOpen: onReverseModalOpen, onClose: onReverseModalClose } = useDisclosure();
  const toast = useToast();

  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const headerBg = useColorModeValue('gray.50', 'gray.700');
  const textSecondary = useColorModeValue('gray.600', 'gray.400');
  const textPrimary = useColorModeValue('gray.700', 'gray.200');

  useEffect(() => {
    loadPaymentDetails();
  }, [paymentId]);

  const loadPaymentDetails = async () => {
    setIsLoading(true);
    try {
      const details = await paymentService.getPaymentWithJournal(paymentId);
      setPaymentInfo(details);
      
      // Load account balance updates if available
      if (details.journal_info?.journal_entry_id) {
        try {
          const balances = await paymentService.getAccountBalanceUpdates(paymentId);
          setAccountBalances(balances);
        } catch (error) {
          console.warn('Failed to load account balances:', error);
        }
      }
    } catch (error: any) {
      toast({
        title: 'Load Error',
        description: error.message || 'Failed to load payment details',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleReversePayment = async () => {
    setIsReversing(true);
    try {
      const result = await paymentService.reversePayment(paymentId, 'Manual reversal from UI');
      
      setReverseResult(result);
      
      toast({
        title: 'Payment Reversed',
        description: 'Payment and journal entries have been reversed successfully',
        status: 'success',
        duration: 5000,
        isClosable: true
      });

      // Reload payment details to show updated status
      await loadPaymentDetails();
      
      if (onReverse) {
        onReverse(paymentId);
      }
    } catch (error: any) {
      toast({
        title: 'Reverse Failed',
        description: error.message || 'Failed to reverse payment',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    } finally {
      setIsReversing(false);
      onReverseModalClose();
    }
  };

  const refreshBalances = async () => {
    try {
      await paymentService.refreshAccountBalances();
      const balances = await paymentService.getAccountBalanceUpdates(paymentId);
      setAccountBalances(balances);
      
      toast({
        title: 'Balances Refreshed',
        description: 'Account balances have been updated',
        status: 'success',
        duration: 3000,
        isClosable: true
      });
    } catch (error: any) {
      toast({
        title: 'Refresh Failed',
        description: error.message || 'Failed to refresh balances',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    }
  };

  if (isLoading) {
    return (
      <Box p={6}>
        <Card bg={bgColor}>
          <CardBody>
            <VStack spacing={4} py={8}>
              <Icon as={FiRefreshCw} boxSize={8} color="gray.400" />
              <Text color="gray.500">Loading payment details...</Text>
            </VStack>
          </CardBody>
        </Card>
      </Box>
    );
  }

  if (!paymentInfo) {
    return (
      <Box p={6}>
        <Alert status="error">
          <AlertIcon />
          Payment not found or failed to load
        </Alert>
      </Box>
    );
  }

  const { payment, journal_info: journalInfo } = paymentInfo;

  const StatusBadge = ({ status }: { status: string }) => {
    const colorScheme = status === 'COMPLETED' ? 'green' : 
                       status === 'PENDING' ? 'yellow' : 
                       status === 'FAILED' ? 'red' :
                       status === 'REVERSED' ? 'purple' : 'gray';
    
    return (
      <Badge colorScheme={colorScheme} variant="solid">
        {status}
      </Badge>
    );
  };

  const PaymentTypeIcon = ({ type }: { type: string }) => {
    return type === 'RECEIVE' ? 
      <Icon as={FiArrowDown} color="green.500" /> : 
      <Icon as={FiArrowRight} color="blue.500" />;
  };

  return (
    <Box maxW="6xl" mx="auto" p={6}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Card bg={bgColor}>
          <CardHeader bg={headerBg}>
            <Flex align="center">
              <HStack spacing={3}>
                <PaymentTypeIcon type={payment.payment_type} />
                <VStack align="start" spacing={0}>
                  <Heading size="md" color={textPrimary}>
                    Payment {payment.payment_code}
                  </Heading>
                  <Text fontSize="sm" color={textSecondary}>
                    Created {new Date(payment.created_at).toLocaleString()}
                  </Text>
                </VStack>
              </HStack>
              <Spacer />
              <HStack>
                <StatusBadge status={payment.status} />
                {onEdit && payment.status !== 'REVERSED' && (
                  <Button size="sm" variant="outline" onClick={() => onEdit(paymentId)}>
                    Edit
                  </Button>
                )}
                {payment.status === 'COMPLETED' && (
                  <Button
                    size="sm"
                    colorScheme="red"
                    variant="outline"
                    leftIcon={<Icon as={FiRotateCcw} />}
                    onClick={onReverseModalOpen}
                  >
                    Reverse
                  </Button>
                )}
              </HStack>
            </Flex>
          </CardHeader>
        </Card>

        <Tabs colorScheme="blue">
          <TabList>
            <Tab>Payment Details</Tab>
            <Tab>Journal Entry</Tab>
            {accountBalances.length > 0 && <Tab>Account Balances</Tab>}
          </TabList>

          <TabPanels>
            {/* Payment Details Tab */}
            <TabPanel px={0}>
              <Card bg={bgColor}>
                <CardBody>
                  <VStack spacing={6} align="stretch">
                    {/* Payment Summary */}
                    <HStack spacing={8} align="start">
                      <Stat>
                        <StatLabel>Amount</StatLabel>
                        <StatNumber fontSize="2xl">
                          {formatCurrency(payment.amount)}
                        </StatNumber>
                        <StatHelpText>{payment.currency}</StatHelpText>
                      </Stat>
                      <Stat>
                        <StatLabel>Payment Method</StatLabel>
                        <StatNumber fontSize="lg">
                          {payment.payment_method.replace('_', ' ')}
                        </StatNumber>
                      </Stat>
                      <Stat>
                        <StatLabel>Type</StatLabel>
                        <StatNumber fontSize="lg" color={payment.payment_type === 'RECEIVE' ? 'green.500' : 'blue.500'}>
                          {payment.payment_type}
                        </StatNumber>
                      </Stat>
                    </HStack>

                    <Divider />

                    {/* Payment Details Grid */}
                    <VStack spacing={4} align="stretch">
                      {payment.description && (
                        <Box>
                          <Text fontWeight="medium" color={textPrimary}>Description</Text>
                          <Text color={textSecondary}>{payment.description}</Text>
                        </Box>
                      )}

                      <HStack spacing={8} align="start">
                        {payment.reference_id && (
                          <Box>
                            <Text fontWeight="medium" color={textPrimary}>Reference ID</Text>
                            <Text color={textSecondary}>{payment.reference_id}</Text>
                          </Box>
                        )}
                        
                        {payment.customer_id && (
                          <Box>
                            <Text fontWeight="medium" color={textPrimary}>Customer ID</Text>
                            <Text color={textSecondary}>{payment.customer_id}</Text>
                          </Box>
                        )}

                        {payment.vendor_id && (
                          <Box>
                            <Text fontWeight="medium" color={textPrimary}>Vendor ID</Text>
                            <Text color={textSecondary}>{payment.vendor_id}</Text>
                          </Box>
                        )}

                        {payment.invoice_id && (
                          <Box>
                            <Text fontWeight="medium" color={textPrimary}>Invoice ID</Text>
                            <Text color={textSecondary}>{payment.invoice_id}</Text>
                          </Box>
                        )}
                      </HStack>

                      <HStack spacing={8}>
                        <Box>
                          <Text fontWeight="medium" color={textPrimary}>Created</Text>
                          <Text color={textSecondary}>
                            {new Date(payment.created_at).toLocaleString()}
                          </Text>
                        </Box>
                        
                        <Box>
                          <Text fontWeight="medium" color={textPrimary}>Updated</Text>
                          <Text color={textSecondary}>
                            {new Date(payment.updated_at).toLocaleString()}
                          </Text>
                        </Box>
                      </HStack>
                    </VStack>
                  </VStack>
                </CardBody>
              </Card>
            </TabPanel>

            {/* Journal Entry Tab */}
            <TabPanel px={0}>
              {journalInfo ? (
                <VStack spacing={4} align="stretch">
                  {/* Journal Info Summary */}
                  <Card bg={bgColor} borderWidth="1px" borderColor={borderColor}>
                    <CardBody>
                      <HStack spacing={4} justify="space-between">
                        <VStack align="start" spacing={1}>
                          <Text fontWeight="medium" color={textPrimary}>Journal Entry</Text>
                          <Tag size="md" colorScheme="blue">
                            <TagLeftIcon as={FiFileText} />
                            <TagLabel>{journalInfo.journal_entry_number}</TagLabel>
                          </Tag>
                        </VStack>
                        
                        <VStack align="end" spacing={1}>
                          <Text fontWeight="medium" color={textPrimary}>Status</Text>
                          <StatusBadge status={journalInfo.status} />
                        </VStack>

                        <Button
                          size="sm"
                          variant="outline"
                          leftIcon={<Icon as={FiExternalLink} />}
                        >
                          View Full Journal
                        </Button>
                      </HStack>
                    </CardBody>
                  </Card>

                  {/* Journal Entry Details */}
                  <JournalEntryPreview
                    journalResult={{
                      success: true,
                      message: '',
                      journal_entry: {
                        id: journalInfo.journal_entry_id,
                        entry_number: journalInfo.journal_entry_number,
                        description: `Payment ${payment.payment_code}`,
                        status: journalInfo.status,
                        total_debit: journalInfo.total_debit || 0,
                        total_credit: journalInfo.total_credit || 0,
                        is_balanced: (journalInfo.total_debit || 0) === (journalInfo.total_credit || 0),
                        lines: [],
                        created_at: payment.created_at,
                        updated_at: payment.updated_at,
                      } as JournalEntry,
                      account_updates: accountBalances
                    }}
                    title="Journal Entry Details"
                    isPreview={false}
                    showAccountUpdates={false}
                  />
                </VStack>
              ) : (
                <Card bg={bgColor}>
                  <CardBody>
                    <VStack spacing={4} py={8}>
                      <Icon as={FiFileText} boxSize={8} color="gray.400" />
                      <Text color="gray.500">No journal entry associated with this payment</Text>
                    </VStack>
                  </CardBody>
                </Card>
              )}
            </TabPanel>

            {/* Account Balances Tab */}
            {accountBalances.length > 0 && (
              <TabPanel px={0}>
                <Card bg={bgColor}>
                  <CardHeader>
                    <Flex align="center">
                      <HStack>
                        <Icon as={FiDollarSign} color="blue.500" />
                        <Heading size="sm" color={textPrimary}>Account Balance Changes</Heading>
                      </HStack>
                      <Spacer />
                      <Button
                        size="sm"
                        variant="outline"
                        leftIcon={<Icon as={FiRefreshCw} />}
                        onClick={refreshBalances}
                      >
                        Refresh Balances
                      </Button>
                    </Flex>
                  </CardHeader>
                  <CardBody>
                    <VStack spacing={3} align="stretch">
                      {accountBalances.map((update, index) => (
                        <Box key={update.account_id || index} p={4} borderWidth="1px" borderColor={borderColor} borderRadius="md">
                          <HStack justify="space-between" mb={2}>
                            <VStack align="start" spacing={0}>
                              <Text fontWeight="medium" color={textPrimary}>
                                {update.account_code} - {update.account_name}
                              </Text>
                              <Text fontSize="sm" color={textSecondary}>
                                Account ID: {update.account_id}
                              </Text>
                            </VStack>
                            <Badge colorScheme={update.change_type === 'INCREASE' ? 'green' : 'red'}>
                              {update.change_type === 'INCREASE' ? '+' : '-'} {formatCurrency(Math.abs(update.change))}
                            </Badge>
                          </HStack>
                          <HStack spacing={4}>
                            <Text fontSize="sm" color={textSecondary}>
                              Previous: {formatCurrency(update.old_balance)}
                            </Text>
                            <Icon as={FiArrowRight} color="gray.400" boxSize={3} />
                            <Text fontSize="sm" color={textSecondary}>
                              New: {formatCurrency(update.new_balance)}
                            </Text>
                          </HStack>
                        </Box>
                      ))}
                    </VStack>
                  </CardBody>
                </Card>
              </TabPanel>
            )}
          </TabPanels>
        </Tabs>
      </VStack>

      {/* Reverse Payment Modal */}
      <Modal isOpen={isReverseModalOpen} onClose={onReverseModalClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Reverse Payment</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} align="start">
              <Alert status="warning">
                <AlertIcon />
                <Text fontSize="sm">
                  This will reverse the payment and create reversing journal entries. This action cannot be undone.
                </Text>
              </Alert>
              <Text>
                Are you sure you want to reverse payment <strong>{payment.payment_code}</strong> for{' '}
                <strong>{formatCurrency(payment.amount)}</strong>?
              </Text>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onReverseModalClose}>
              Cancel
            </Button>
            <Button
              colorScheme="red"
              onClick={handleReversePayment}
              isLoading={isReversing}
              loadingText="Reversing..."
            >
              Reverse Payment
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Reverse Result Modal */}
      {reverseResult && (
        <Modal isOpen={!!reverseResult} onClose={() => setReverseResult(null)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Payment Reversal Complete</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <JournalEntryPreview
                journalResult={reverseResult}
                title="Reversal Journal Entry"
                isPreview={false}
                showAccountUpdates={true}
              />
            </ModalBody>
            <ModalFooter>
              <Button onClick={() => setReverseResult(null)}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      )}
    </Box>
  );
};

export default PaymentDetailsView;