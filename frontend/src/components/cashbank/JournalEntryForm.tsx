'use client';

import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  FormControl,
  FormLabel,
  Select,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Input,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconButton,
  Tooltip,
  Switch,
  FormErrorMessage,
  useToast,
  Spinner,
} from '@chakra-ui/react';
import { FiPlus, FiTrash2, FiInfo, FiActivity, FiCheckCircle } from 'react-icons/fi';
import { ManualJournalEntry } from '@/services/cashbankService';
import { Account } from '@/types/account';
import { accountService } from '@/services/accountService';
import { useAuth } from '@/contexts/AuthContext';
// Websocket removed - using standard refresh instead

interface JournalEntryFormProps {
  journalEntries: ManualJournalEntry[];
  onChange: (entries: ManualJournalEntry[]) => void;
  isEnabled: boolean;
  onToggle: (enabled: boolean) => void;
  transactionAmount: number;
  mode: 'deposit' | 'withdrawal';
  cashBankAccountId?: number;
  errors?: { [key: number]: string };
}

const JournalEntryForm: React.FC<JournalEntryFormProps> = ({
  journalEntries,
  onChange,
  isEnabled,
  onToggle,
  transactionAmount,
  mode,
  cashBankAccountId,
  errors = {}
}) => {
  const { token } = useAuth();
  const toast = useToast();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isEnabled && token) {
      loadAccounts();
    }
  }, [isEnabled, token]);
  
  // Websocket connection removed - using standard refresh

  const loadAccounts = async () => {
    try {
      setLoading(true);
      const data = await accountService.getAccounts(token!);
      // Filter out the cash/bank account being used in the transaction
      const filteredAccounts = data.filter(acc => acc.id !== cashBankAccountId);
      setAccounts(filteredAccounts);
    } catch (error) {
      console.error('Error loading accounts:', error);
      
      // Enhanced error handling with user-friendly messages
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      
      toast({
        title: 'Failed to Load Accounts',
        description: `Unable to load account list: ${errorMessage}`,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      
      // Set empty accounts array on error
      setAccounts([]);
    } finally {
      setLoading(false);
    }
  };

  const addEntry = () => {
    const newEntry: ManualJournalEntry = {
      account_id: 0,
      description: '',
      debit_amount: 0,
      credit_amount: 0,
    };
    onChange([...journalEntries, newEntry]);
  };

  const removeEntry = (index: number) => {
    const updatedEntries = journalEntries.filter((_, i) => i !== index);
    onChange(updatedEntries);
  };

  const updateEntry = (index: number, field: keyof ManualJournalEntry, value: any) => {
    const updatedEntries = journalEntries.map((entry, i) => {
      if (i === index) {
        return { ...entry, [field]: value };
      }
      return entry;
    });
    onChange(updatedEntries);
  };

  const calculateTotals = () => {
    const totalDebit = journalEntries.reduce((sum, entry) => sum + (entry.debit_amount || 0), 0);
    const totalCredit = journalEntries.reduce((sum, entry) => sum + (entry.credit_amount || 0), 0);
    return { totalDebit, totalCredit };
  };

  const { totalDebit, totalCredit } = calculateTotals();
  const isBalanced = Math.abs(totalDebit - totalCredit) < 0.01; // Allow for minor floating point differences
  const hasValidEntries = journalEntries.length > 0 && journalEntries.every(entry => 
    entry.account_id > 0 && 
    entry.description.trim() !== '' &&
    (entry.debit_amount > 0 || entry.credit_amount > 0)
  );

  const getAccountById = (accountId: number) => {
    return accounts.find(acc => acc.id === accountId);
  };

  return (
    <Box>
      <HStack justify="space-between" align="center" mb={4}>
        <HStack spacing={3}>
          <Text fontSize="md" fontWeight="semibold" color="gray.700">
            📝 Manual Journal Entries
          </Text>
          <Switch
            isChecked={isEnabled}
            onChange={(e) => onToggle(e.target.checked)}
            colorScheme="purple"
            size="md"
          />
          <Badge
            colorScheme={isEnabled ? 'purple' : 'gray'}
            variant={isEnabled ? 'solid' : 'subtle'}
            fontSize="xs"
          >
            {isEnabled ? 'MANUAL MODE' : 'AUTOMATIC MODE'}
          </Badge>
          
          {/* Real-time Connection Status */}
          {isEnabled && (
            <Tooltip 
              label={isConnectedToBalanceService ? 'Connected to real-time balance updates' : 'Real-time balance updates unavailable'} 
              fontSize="xs"
            >
              <Badge
                colorScheme={isConnectedToBalanceService ? 'green' : 'yellow'}
                variant="subtle"
                fontSize="xs"
                display="flex"
                alignItems="center"
                gap={1}
              >
                {isConnectedToBalanceService ? (
                  <>
                    <FiActivity size={10} />
                    LIVE
                  </>
                ) : (
                  <>
                    <FiInfo size={10} />
                    OFFLINE
                  </>
                )}
              </Badge>
            </Tooltip>
          )}
        </HStack>
        
        {isEnabled && (
          <Tooltip label="Add journal entry line" fontSize="xs">
            <IconButton
              aria-label="Add entry"
              icon={<FiPlus />}
              size="sm"
              colorScheme="purple"
              variant="outline"
              onClick={addEntry}
            />
          </Tooltip>
        )}
      </HStack>

      {!isEnabled && (
        <Alert status="info" borderRadius="md" mb={4}>
          <AlertIcon />
          <Box>
            <AlertTitle fontSize="sm">Automatic Journal Entries</AlertTitle>
            <AlertDescription fontSize="xs">
              {mode === 'deposit' 
                ? 'The system will automatically create journal entries debiting the cash/bank account and crediting "Other Income".'
                : 'The system will automatically create journal entries debiting "General Expense" and crediting the cash/bank account.'
              }
            </AlertDescription>
          </Box>
        </Alert>
      )}

      {isEnabled && (
        <Box>
          <Alert status="warning" borderRadius="md" mb={4}>
            <AlertIcon />
            <Box>
              <AlertTitle fontSize="sm">Manual Double-Entry Mode</AlertTitle>
              <AlertDescription fontSize="xs">
                You must specify the complete journal entries. The cash/bank account entry will be handled automatically.
                Make sure your debits equal credits for proper double-entry bookkeeping.
              </AlertDescription>
            </Box>
          </Alert>

          {journalEntries.length > 0 ? (
            <Box>
              <Box overflowX="auto" mb={4}>
                <Table size="sm" variant="simple">
                  <Thead>
                    <Tr bg="gray.50">
                      <Th width="35%">Account</Th>
                      <Th width="25%">Description</Th>
                      <Th width="15%" textAlign="right">Debit</Th>
                      <Th width="15%" textAlign="right">Credit</Th>
                      <Th width="10%" textAlign="center">Action</Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {journalEntries.map((entry, index) => (
                      <Tr key={index} _hover={{ bg: 'gray.25' }}>
                        <Td>
                          <FormControl isInvalid={!!errors[index] && errors[index].includes('account')}>
                            <Select
                              size="sm"
                              value={entry.account_id}
                              onChange={(e) => updateEntry(index, 'account_id', parseInt(e.target.value))}
                              placeholder="Select account..."
                              isDisabled={loading}
                            >
                              {accounts.map((account) => (
                                <option key={account.id} value={account.id}>
                                  {account.code} - {account.name}
                                </option>
                              ))}
                            </Select>
                            {errors[index] && errors[index].includes('account') && (
                              <FormErrorMessage fontSize="xs">
                                {errors[index]}
                              </FormErrorMessage>
                            )}
                          </FormControl>
                        </Td>
                        <Td>
                          <FormControl isInvalid={!!errors[index] && errors[index].includes('description')}>
                            <Input
                              size="sm"
                              value={entry.description}
                              onChange={(e) => updateEntry(index, 'description', e.target.value)}
                              placeholder="Entry description"
                            />
                            {errors[index] && errors[index].includes('description') && (
                              <FormErrorMessage fontSize="xs">
                                {errors[index]}
                              </FormErrorMessage>
                            )}
                          </FormControl>
                        </Td>
                        <Td>
                          <NumberInput
                            size="sm"
                            value={entry.debit_amount}
                            onChange={(_, value) => {
                              updateEntry(index, 'debit_amount', value || 0);
                              // Clear credit when entering debit
                              if (value && value > 0) {
                                updateEntry(index, 'credit_amount', 0);
                              }
                            }}
                            min={0}
                            precision={2}
                          >
                            <NumberInputField textAlign="right" />
                            <NumberInputStepper>
                              <NumberIncrementStepper />
                              <NumberDecrementStepper />
                            </NumberInputStepper>
                          </NumberInput>
                        </Td>
                        <Td>
                          <NumberInput
                            size="sm"
                            value={entry.credit_amount}
                            onChange={(_, value) => {
                              updateEntry(index, 'credit_amount', value || 0);
                              // Clear debit when entering credit
                              if (value && value > 0) {
                                updateEntry(index, 'debit_amount', 0);
                              }
                            }}
                            min={0}
                            precision={2}
                          >
                            <NumberInputField textAlign="right" />
                            <NumberInputStepper>
                              <NumberIncrementStepper />
                              <NumberDecrementStepper />
                            </NumberInputStepper>
                          </NumberInput>
                        </Td>
                        <Td textAlign="center">
                          <Tooltip label="Remove entry" fontSize="xs">
                            <IconButton
                              aria-label="Remove entry"
                              icon={<FiTrash2 />}
                              size="xs"
                              colorScheme="red"
                              variant="ghost"
                              onClick={() => removeEntry(index)}
                            />
                          </Tooltip>
                        </Td>
                      </Tr>
                    ))}
                  </Tbody>
                </Table>
              </Box>

              {/* Balance Summary */}
              <Box bg="gray.50" p={4} borderRadius="md" mb={4}>
                <VStack spacing={3} align="stretch">
                  <HStack justify="space-between">
                    <Text fontSize="sm" fontWeight="medium">Journal Entry Totals:</Text>
                  </HStack>
                  
                  <HStack justify="space-between">
                    <Text fontSize="sm">Total Debits:</Text>
                    <Text fontSize="sm" fontWeight="bold" fontFamily="mono">
                      IDR {totalDebit.toLocaleString('id-ID')}
                    </Text>
                  </HStack>
                  
                  <HStack justify="space-between">
                    <Text fontSize="sm">Total Credits:</Text>
                    <Text fontSize="sm" fontWeight="bold" fontFamily="mono">
                      IDR {totalCredit.toLocaleString('id-ID')}
                    </Text>
                  </HStack>
                  
                  <HStack justify="space-between">
                    <Text fontSize="sm">Difference:</Text>
                    <Text 
                      fontSize="sm" 
                      fontWeight="bold" 
                      fontFamily="mono"
                      color={isBalanced ? 'green.600' : 'red.500'}
                    >
                      IDR {Math.abs(totalDebit - totalCredit).toLocaleString('id-ID')}
                    </Text>
                  </HStack>

                  <Alert status={isBalanced ? 'success' : 'warning'} py={2}>
                    <AlertIcon boxSize="16px" />
                    <Text fontSize="xs">
                      {isBalanced 
                        ? '✅ Journal entries are balanced' 
                        : '⚠️ Journal entries are not balanced - debits must equal credits'
                      }
                    </Text>
                  </Alert>

                  {/* Transaction Amount Validation */}
                  {transactionAmount > 0 && (
                    <Alert status="info" py={2}>
                      <AlertIcon boxSize="16px" />
                      <Text fontSize="xs">
                        💡 Transaction amount: IDR {transactionAmount.toLocaleString('id-ID')}. 
                        Your journal entries should correspond to this amount.
                      </Text>
                    </Alert>
                  )}
                </VStack>
              </Box>
            </Box>
          ) : (
            <Box>
              <Alert status="info" borderRadius="md" mb={4}>
                <AlertIcon />
                <Box>
                  <AlertTitle fontSize="sm">No Journal Entries</AlertTitle>
                  <AlertDescription fontSize="xs">
                    Click the "+" button to add journal entry lines. Remember that the cash/bank account entry 
                    will be created automatically based on your transaction.
                  </AlertDescription>
                </Box>
              </Alert>
            </Box>
          )}

          {journalEntries.length > 0 && (
            <HStack spacing={3} justify="center">
              <Button
                size="sm"
                leftIcon={<FiPlus />}
                colorScheme="purple"
                variant="outline"
                onClick={addEntry}
              >
                Add Entry Line
              </Button>
            </HStack>
          )}
        </Box>
      )}
    </Box>
  );
};

export default JournalEntryForm;
