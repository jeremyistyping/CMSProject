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
  Select,
  Divider,
  Box,
  Spinner,
} from '@chakra-ui/react';
import { PurchaseRequest } from '@/types/purchaseRequest';
import purchaseOrderService, { CreatePORequest } from '@/services/purchaseOrderService';
import { vendorService, Vendor } from '@/services/masterDataService';

interface CreatePOModalProps {
  isOpen: boolean;
  onClose: () => void;
  pr: PurchaseRequest | null;
  onSuccess: () => void;
}

const CreatePOModal: React.FC<CreatePOModalProps> = ({ isOpen, onClose, pr, onSuccess }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [loadingVendors, setLoadingVendors] = useState(false);
  
  const [formData, setFormData] = useState<CreatePORequest>({
    purchase_request_id: 0,
    vendor_id: undefined,
    expected_delivery_date: '',
    delivery_address: '',
    payment_terms: '',
    notes: '',
  });

  const toast = useToast();

  useEffect(() => {
    if (isOpen && pr) {
      setFormData({
        purchase_request_id: pr.id,
        vendor_id: pr.vendor_id || undefined,
        expected_delivery_date: pr.required_date ? new Date(pr.required_date).toISOString().split('T')[0] : '',
        delivery_address: '',
        payment_terms: 'Net 30',
        notes: '',
      });
      fetchVendors();
    }
  }, [isOpen, pr]);

  const fetchVendors = async () => {
    setLoadingVendors(true);
    try {
      const response = await vendorService.getAll({ is_active: true });
      setVendors(response.data || []);
    } catch (error) {
      console.error('Error fetching vendors:', error);
    } finally {
      setLoadingVendors(false);
    }
  };

  const handleSubmit = async () => {
    if (!pr) return;

    setIsLoading(true);
    try {
      // Prepare data with proper null handling for dates
      const submitData = {
        ...formData,
        expected_delivery_date: formData.expected_delivery_date || undefined,
      };
      
      await purchaseOrderService.createFromPR(submitData);
      toast({
        title: 'Success',
        description: 'Purchase Order created successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      onSuccess();
      onClose();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to create Purchase Order',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (!pr) return null;

  const totalAmount = pr.items?.reduce((sum, item) => sum + (item.quantity * item.estimated_price), 0) || pr.total_amount;

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent maxW="800px">
        <ModalHeader>
          Create Purchase Order
          <Text fontSize="sm" fontWeight="normal" color="gray.500">
            From PR: {pr.code}
          </Text>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody>
          <VStack spacing={4} align="stretch">
            {/* PR Info */}
            <Box bg="blue.50" p={4} borderRadius="md">
              <HStack justify="space-between" mb={2}>
                <Text fontWeight="bold">Purchase Request Info</Text>
                <Badge colorScheme="green">{pr.status}</Badge>
              </HStack>
              <HStack spacing={8}>
                <Box>
                  <Text fontSize="sm" color="gray.600">Project</Text>
                  <Text fontWeight="medium">{pr.project?.project_name}</Text>
                </Box>
                <Box>
                  <Text fontSize="sm" color="gray.600">Total Amount</Text>
                  <Text fontWeight="medium">
                    {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(totalAmount)}
                  </Text>
                </Box>
              </HStack>
            </Box>

            <Divider />

            {/* PO Form */}
            <FormControl>
              <FormLabel>Vendor</FormLabel>
              {loadingVendors ? (
                <Spinner size="sm" />
              ) : (
                <Select
                  placeholder="Select vendor"
                  value={formData.vendor_id || ''}
                  onChange={(e) => setFormData({ ...formData, vendor_id: e.target.value ? Number(e.target.value) : undefined })}
                >
                  {vendors.map(v => (
                    <option key={v.id} value={v.id}>{v.name}</option>
                  ))}
                </Select>
              )}
            </FormControl>

            <HStack spacing={4}>
              <FormControl>
                <FormLabel>Expected Delivery Date</FormLabel>
                <Input
                  type="date"
                  value={formData.expected_delivery_date}
                  onChange={(e) => setFormData({ ...formData, expected_delivery_date: e.target.value })}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Payment Terms</FormLabel>
                <Select
                  value={formData.payment_terms}
                  onChange={(e) => setFormData({ ...formData, payment_terms: e.target.value })}
                >
                  <option value="Net 30">Net 30</option>
                  <option value="Net 60">Net 60</option>
                  <option value="Net 90">Net 90</option>
                  <option value="COD">Cash on Delivery</option>
                  <option value="Advance">Advance Payment</option>
                </Select>
              </FormControl>
            </HStack>

            <FormControl>
              <FormLabel>Delivery Address</FormLabel>
              <Textarea
                placeholder="Enter delivery address..."
                value={formData.delivery_address}
                onChange={(e) => setFormData({ ...formData, delivery_address: e.target.value })}
                rows={2}
              />
            </FormControl>

            <FormControl>
              <FormLabel>Notes</FormLabel>
              <Textarea
                placeholder="Additional notes..."
                value={formData.notes}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                rows={2}
              />
            </FormControl>

            <Divider />

            {/* Items from PR */}
            <Box>
              <Text fontWeight="bold" mb={2}>Items (from Purchase Request)</Text>
              <Table size="sm" variant="simple">
                <Thead bg="gray.50">
                  <Tr>
                    <Th>Item</Th>
                    <Th isNumeric>Qty</Th>
                    <Th>Unit</Th>
                    <Th isNumeric>Est. Price</Th>
                    <Th isNumeric>Total</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {pr.items?.map((item, idx) => (
                    <Tr key={idx}>
                      <Td>{item.item_name}</Td>
                      <Td isNumeric>{item.quantity}</Td>
                      <Td>{item.unit}</Td>
                      <Td isNumeric>
                        {new Intl.NumberFormat('id-ID').format(item.estimated_price)}
                      </Td>
                      <Td isNumeric>
                        {new Intl.NumberFormat('id-ID').format(item.quantity * item.estimated_price)}
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </Box>
          </VStack>
        </ModalBody>

        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={onClose}>
            Cancel
          </Button>
          <Button
            colorScheme="blue"
            onClick={handleSubmit}
            isLoading={isLoading}
          >
            Create Purchase Order
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default CreatePOModal;
