'use client';

import React, { useState } from 'react';
import {
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Text,
  Box,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  VStack,
  HStack,
  useToast,
} from '@chakra-ui/react';
import { FiSend } from 'react-icons/fi';
import purchaseService from '@/services/purchaseService';

interface SubmitApprovalButtonProps {
  purchaseId: number;
  purchaseCode: string;
  totalAmount: number;
  approvalBaseAmount?: number;
  approvalBasis?: 'SUBTOTAL_BEFORE_DISCOUNT' | 'NET_AFTER_DISCOUNT_BEFORE_TAX' | 'GRAND_TOTAL_AFTER_TAX';
  onSubmitted?: () => void;
  disabled?: boolean;
}

const SubmitApprovalButton: React.FC<SubmitApprovalButtonProps> = ({
  purchaseId,
  purchaseCode,
  totalAmount,
  approvalBaseAmount,
  approvalBasis,
  onSubmitted,
  disabled = false,
}) => {
  const [open, setOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const toast = useToast();

  const handleSubmit = async () => {
    setSubmitting(true);
    setError(null);
    
    try {
      await purchaseService.submitForApproval(purchaseId);
      setOpen(false);
      if (onSubmitted) {
        onSubmitted();
      }
      toast({
        title: 'Success',
        description: 'Purchase submitted for approval successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      console.error('Failed to submit purchase for approval:', error);
      setError(error.response?.data?.error || 'Failed to submit purchase for approval');
    } finally {
      setSubmitting(false);
    }
  };

  const getApprovalInfo = () => {
    const base = approvalBaseAmount ?? totalAmount;
    const basisText = approvalBasis?.replaceAll('_', ' ').toLowerCase() || 'subtotal before discount';
    return {
      level: 'Approval evaluation',
      reason: `Base amount for approval: ${formatCurrency(base)} (basis: ${basisText})`,
      color: 'blue.500',
    };
  };

  const approvalInfo = getApprovalInfo();

  return (
    <>
      <Button
        colorScheme="blue"
        leftIcon={<FiSend />}
        onClick={() => setOpen(true)}
        isDisabled={disabled}
        size="sm"
      >
        Submit for Approval
      </Button>

      <Modal isOpen={open} onClose={() => !submitting && setOpen(false)} size="md">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Submit Purchase for Approval</ModalHeader>
          <ModalCloseButton isDisabled={submitting} />
          
          <ModalBody>
            <VStack spacing={4} align="stretch">
              <Box>
                <Text fontSize="lg" fontWeight="semibold" mb={1}>
                  {purchaseCode}
                </Text>
                <Text fontSize="sm" color="gray.600">
                  Amount: {formatCurrency(totalAmount)}
                </Text>
              </Box>

              <Alert status="info">
                <AlertIcon />
                <Box>
                  <AlertTitle>{approvalInfo.level}</AlertTitle>
                  <AlertDescription>{approvalInfo.reason}</AlertDescription>
                </Box>
              </Alert>

              {error && (
                <Alert status="error">
                  <AlertIcon />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <Text fontSize="sm" color="gray.600">
                Once submitted, this purchase will be sent to the appropriate approver based on the amount and approval workflow.
                You will receive notifications about the approval status.
              </Text>
            </VStack>
          </ModalBody>

          <ModalFooter>
            <HStack spacing={3}>
              <Button variant="ghost" onClick={() => setOpen(false)} isDisabled={submitting}>
                Cancel
              </Button>
              <Button
                colorScheme="blue"
                leftIcon={submitting ? <Spinner size="sm" /> : <FiSend />}
                onClick={handleSubmit}
                isLoading={submitting}
                loadingText="Submitting..."
              >
                Submit for Approval
              </Button>
            </HStack>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  );
};

export default SubmitApprovalButton;
