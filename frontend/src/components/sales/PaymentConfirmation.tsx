'use client';

import React from 'react';
import {
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  VStack,
  HStack,
  Text,
  Box,
  Divider,
  Badge,
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton
} from '@chakra-ui/react';
import { Sale } from '@/services/salesService';
import salesService from '@/services/salesService';

interface PaymentConfirmationProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  sale: Sale | null;
  paymentData: {
    date: string;
    amount: number;
    method: string;
    reference?: string;
    account_name?: string;
    notes?: string;
  };
  isLoading?: boolean;
}

const PaymentConfirmation: React.FC<PaymentConfirmationProps> = ({
  isOpen,
  onClose,
  onConfirm,
  sale,
  paymentData,
  isLoading = false
}) => {
  if (!sale || !paymentData) return null;

  const remainingBalance = (sale.outstanding_amount || 0) - paymentData.amount;
  const isFullPayment = remainingBalance <= 0;

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="md">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Confirm Payment</ModalHeader>
        <ModalCloseButton />
        
        <ModalBody>
          <VStack spacing={4} align="stretch">
            {/* Alert Section */}
            <Alert 
              status={isFullPayment ? "success" : "info"} 
              borderRadius="md"
            >
              <AlertIcon />
              <Box>
                <AlertTitle>
                  {isFullPayment ? "Full Payment!" : "Partial Payment"}
                </AlertTitle>
                <AlertDescription>
                  {isFullPayment 
                    ? "This payment will fully settle the invoice." 
                    : `A balance of ${salesService.formatCurrency(remainingBalance)} will remain.`
                  }
                </AlertDescription>
              </Box>
            </Alert>

            {/* Sale Information */}
            <Box p={4} bg="gray.50" borderRadius="md">
              <Text fontSize="sm" fontWeight="bold" mb={3} color="gray.700">
                Invoice Details
              </Text>
              <VStack spacing={2} align="stretch">
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.600">Sale Code:</Text>
                  <Text fontSize="sm" fontWeight="medium">{sale.code}</Text>
                </HStack>
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.600">Customer:</Text>
                  <Text fontSize="sm" fontWeight="medium">{sale.customer?.name}</Text>
                </HStack>
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.600">Outstanding:</Text>
                  <Text fontSize="sm" fontWeight="bold" color="orange.600">
                    {salesService.formatCurrency(sale.outstanding_amount)}
                  </Text>
                </HStack>
              </VStack>
            </Box>

            {/* Payment Details */}
            <Box p={4} bg="blue.50" borderRadius="md" borderLeft="4px" borderColor="blue.400">
              <Text fontSize="sm" fontWeight="bold" mb={3} color="blue.700">
                Payment Details
              </Text>
              <VStack spacing={2} align="stretch">
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.600">Date:</Text>
                  <Text fontSize="sm" fontWeight="medium">
                    {new Date(paymentData.date).toLocaleDateString('id-ID')}
                  </Text>
                </HStack>
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.600">Amount:</Text>
                  <Text fontSize="sm" fontWeight="bold" color="green.600">
                    {salesService.formatCurrency(paymentData.amount)}
                  </Text>
                </HStack>
                <HStack justify="space-between">
                  <Text fontSize="sm" color="gray.600">Method:</Text>
                  <Badge colorScheme="blue" variant="subtle">
                    {paymentData.method}
                  </Badge>
                </HStack>
                {paymentData.account_name && (
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Account:</Text>
                    <Text fontSize="sm">{paymentData.account_name}</Text>
                  </HStack>
                )}
                {paymentData.reference && (
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Reference:</Text>
                    <Text fontSize="sm">{paymentData.reference}</Text>
                  </HStack>
                )}
              </VStack>
            </Box>

            {/* Result Summary */}
            <Box p={4} bg={isFullPayment ? "green.50" : "yellow.50"} borderRadius="md">
              <Text fontSize="sm" fontWeight="bold" mb={2} 
                color={isFullPayment ? "green.700" : "yellow.700"}>
                After Payment
              </Text>
              <HStack justify="space-between">
                <Text fontSize="sm" color="gray.600">Remaining Balance:</Text>
                <Text fontSize="sm" fontWeight="bold" 
                  color={isFullPayment ? "green.600" : "orange.600"}>
                  {salesService.formatCurrency(Math.max(0, remainingBalance))}
                </Text>
              </HStack>
              <HStack justify="space-between" mt={1}>
                <Text fontSize="sm" color="gray.600">Status:</Text>
                <Badge 
                  colorScheme={isFullPayment ? "green" : "orange"} 
                  variant="subtle"
                >
                  {isFullPayment ? "PAID" : "PARTIALLY PAID"}
                </Badge>
              </HStack>
            </Box>

            {paymentData.notes && (
              <Box p={3} bg="gray.100" borderRadius="md">
                <Text fontSize="xs" color="gray.600" mb={1}>Notes:</Text>
                <Text fontSize="sm">{paymentData.notes}</Text>
              </Box>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={onClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button 
            colorScheme="blue" 
            onClick={onConfirm} 
            isLoading={isLoading}
            loadingText="Processing..."
          >
            Confirm Payment
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default PaymentConfirmation;
