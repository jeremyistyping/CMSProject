import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  VStack,
  HStack,
  Text,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  useToast,
  Divider,
  Box,
  Spinner,
  NumberInput,
  NumberInputField,
  Alert,
  AlertIcon,
} from '@chakra-ui/react';
import { PurchaseRequest } from '@/types/purchaseRequest';
import purchaseOrderService, { PurchaseOrder, CreateGRRequest } from '@/services/purchaseOrderService';

interface GoodsReceiptModalProps {
  isOpen: boolean;
  onClose: () => void;
  pr: PurchaseRequest | null;
  onSuccess: () => void;
}

interface GRItemInput {
  po_item_id: number;
  item_name: string;
  ordered_qty: number;
  received_qty: number;
  received_quantity: number;
  accepted_quantity: number;
  rejected_quantity: number;
  rejection_reason: string;
}

const GoodsReceiptModal: React.FC<GoodsReceiptModalProps> = ({ isOpen, onClose, pr, onSuccess }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [loadingPO, setLoadingPO] = useState(false);
  const [purchaseOrder, setPurchaseOrder] = useState<PurchaseOrder | null>(null);
  const [receiptDate, setReceiptDate] = useState(new Date().toISOString().split('T')[0]);
  const [notes, setNotes] = useState('');
  const [items, setItems] = useState<GRItemInput[]>([]);

  const toast = useToast();

  useEffect(() => {
    if (isOpen && pr) {
      fetchPurchaseOrder();
    }
  }, [isOpen, pr]);

  const fetchPurchaseOrder = async () => {
    if (!pr) return;
    
    setLoadingPO(true);
    try {
      const pos = await purchaseOrderService.getByPRId(pr.id);
      if (pos && pos.length > 0) {
        // Get the active PO (not cancelled)
        const activePO = pos.find(po => po.status !== 'CANCELLED');
        if (activePO) {
          const fullPO = await purchaseOrderService.getById(activePO.id);
          setPurchaseOrder(fullPO);
          
          // Initialize items for receipt
          const grItems: GRItemInput[] = fullPO.items.map(item => ({
            po_item_id: item.id!,
            item_name: item.item_name,
            ordered_qty: item.quantity,
            received_qty: item.received_quantity || 0,
            received_quantity: item.quantity - (item.received_quantity || 0), // Default to remaining qty
            accepted_quantity: item.quantity - (item.received_quantity || 0),
            rejected_quantity: 0,
            rejection_reason: '',
          }));
          setItems(grItems);
        }
      }
    } catch (error) {
      console.error('Error fetching PO:', error);
      toast({
        title: 'Error',
        description: 'Failed to load Purchase Order',
        status: 'error',
        duration: 3000,
      });
    } finally {
      setLoadingPO(false);
    }
  };

  const handleItemChange = (index: number, field: keyof GRItemInput, value: number | string) => {
    const newItems = [...items];
    (newItems[index] as any)[field] = value;
    
    // Auto-calculate accepted if received changes
    if (field === 'received_quantity') {
      newItems[index].accepted_quantity = Number(value) - newItems[index].rejected_quantity;
    }
    if (field === 'rejected_quantity') {
      newItems[index].accepted_quantity = newItems[index].received_quantity - Number(value);
    }
    
    setItems(newItems);
  };

  const handleSubmit = async () => {
    if (!purchaseOrder) return;

    // Validate
    const hasItems = items.some(item => item.received_quantity > 0);
    if (!hasItems) {
      toast({
        title: 'Error',
        description: 'Please enter received quantity for at least one item',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    setIsLoading(true);
    try {
      const grRequest: CreateGRRequest = {
        purchase_order_id: purchaseOrder.id,
        receipt_date: receiptDate,
        notes: notes,
        items: items
          .filter(item => item.received_quantity > 0)
          .map(item => ({
            po_item_id: item.po_item_id,
            received_quantity: item.received_quantity,
            accepted_quantity: item.accepted_quantity,
            rejected_quantity: item.rejected_quantity,
            rejection_reason: item.rejection_reason,
          })),
      };

      await purchaseOrderService.createGoodsReceipt(grRequest);
      
      toast({
        title: 'Success',
        description: 'Goods Receipt created successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      onSuccess();
      onClose();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to create Goods Receipt',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (!pr) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent maxW="900px">
        <ModalHeader>
          Receive Goods
          <Text fontSize="sm" fontWeight="normal" color="gray.500">
            PR: {pr.code}
          </Text>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody>
          {loadingPO ? (
            <Box textAlign="center" py={10}>
              <Spinner size="xl" />
              <Text mt={4}>Loading Purchase Order...</Text>
            </Box>
          ) : !purchaseOrder ? (
            <Alert status="warning">
              <AlertIcon />
              No Purchase Order found for this Purchase Request
            </Alert>
          ) : (
            <VStack spacing={4} align="stretch">
              {/* PO Info */}
              <Box bg="blue.50" p={4} borderRadius="md">
                <HStack justify="space-between" mb={2}>
                  <Text fontWeight="bold">Purchase Order: {purchaseOrder.code}</Text>
                  <Badge colorScheme={purchaseOrder.status === 'SENT' ? 'green' : 'blue'}>
                    {purchaseOrder.status}
                  </Badge>
                </HStack>
                <HStack spacing={8}>
                  <Box>
                    <Text fontSize="sm" color="gray.600">Vendor</Text>
                    <Text fontWeight="medium">{purchaseOrder.vendor?.name || 'N/A'}</Text>
                  </Box>
                  <Box>
                    <Text fontSize="sm" color="gray.600">Total Amount</Text>
                    <Text fontWeight="medium">
                      {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(purchaseOrder.total_amount)}
                    </Text>
                  </Box>
                </HStack>
              </Box>

              <Divider />

              <HStack spacing={4}>
                <FormControl>
                  <FormLabel>Receipt Date</FormLabel>
                  <Input
                    type="date"
                    value={receiptDate}
                    onChange={(e) => setReceiptDate(e.target.value)}
                  />
                </FormControl>
              </HStack>

              <FormControl>
                <FormLabel>Notes</FormLabel>
                <Textarea
                  placeholder="Receipt notes..."
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  rows={2}
                />
              </FormControl>

              <Divider />

              {/* Items */}
              <Box>
                <Text fontWeight="bold" mb={2}>Items to Receive</Text>
                <Table size="sm" variant="simple">
                  <Thead bg="gray.50">
                    <Tr>
                      <Th>Item</Th>
                      <Th isNumeric>Ordered</Th>
                      <Th isNumeric>Already Received</Th>
                      <Th isNumeric>Receive Now</Th>
                      <Th isNumeric>Accepted</Th>
                      <Th isNumeric>Rejected</Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {items.map((item, idx) => {
                      const remaining = item.ordered_qty - item.received_qty;
                      return (
                        <Tr key={idx}>
                          <Td>{item.item_name}</Td>
                          <Td isNumeric>{item.ordered_qty}</Td>
                          <Td isNumeric>{item.received_qty}</Td>
                          <Td isNumeric>
                            <NumberInput
                              size="sm"
                              min={0}
                              max={remaining}
                              value={item.received_quantity}
                              onChange={(_, val) => handleItemChange(idx, 'received_quantity', val || 0)}
                            >
                              <NumberInputField w="80px" />
                            </NumberInput>
                          </Td>
                          <Td isNumeric>
                            <NumberInput
                              size="sm"
                              min={0}
                              max={item.received_quantity}
                              value={item.accepted_quantity}
                              onChange={(_, val) => handleItemChange(idx, 'accepted_quantity', val || 0)}
                            >
                              <NumberInputField w="80px" />
                            </NumberInput>
                          </Td>
                          <Td isNumeric>
                            <NumberInput
                              size="sm"
                              min={0}
                              max={item.received_quantity}
                              value={item.rejected_quantity}
                              onChange={(_, val) => handleItemChange(idx, 'rejected_quantity', val || 0)}
                            >
                              <NumberInputField w="80px" />
                            </NumberInput>
                          </Td>
                        </Tr>
                      );
                    })}
                  </Tbody>
                </Table>
              </Box>
            </VStack>
          )}
        </ModalBody>

        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={onClose}>
            Cancel
          </Button>
          <Button
            colorScheme="teal"
            onClick={handleSubmit}
            isLoading={isLoading}
            isDisabled={!purchaseOrder || loadingPO}
          >
            Create Goods Receipt
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default GoodsReceiptModal;
