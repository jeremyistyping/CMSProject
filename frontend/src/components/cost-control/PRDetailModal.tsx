import React, { useState } from 'react';
import {
    Modal,
    ModalOverlay,
    ModalContent,
    ModalHeader,
    ModalFooter,
    ModalBody,
    ModalCloseButton,
    Button,
    VStack,
    HStack,
    Text,
    Badge,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Divider,
    Box,
    Textarea,
    useToast,
    FormControl,
    FormLabel,
} from '@chakra-ui/react';
import { PurchaseRequest } from '../../types/purchaseRequest';
import purchaseRequestService from '../../services/purchaseRequestService';

interface PRDetailModalProps {
    isOpen: boolean;
    onClose: () => void;
    pr: PurchaseRequest | null;
    onUpdate: () => void;
}

const PRDetailModal: React.FC<PRDetailModalProps> = ({ isOpen, onClose, pr, onUpdate }) => {
    const toast = useToast();
    const [isLoading, setIsLoading] = useState(false);
    const [rejectionReason, setRejectionReason] = useState('');
    const [isRejecting, setIsRejecting] = useState(false);

    if (!pr) return null;

    const handleApprove = async () => {
        try {
            setIsLoading(true);
            await purchaseRequestService.updateStatus(pr.id, 'APPROVED');
            toast({
                title: 'Success',
                description: 'Purchase Request approved',
                status: 'success',
                duration: 3000,
                isClosable: true,
            });
            onUpdate();
            onClose();
        } catch (error) {
            toast({
                title: 'Error',
                description: 'Failed to approve Purchase Request',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleReject = async () => {
        if (!rejectionReason) {
            toast({
                title: 'Error',
                description: 'Please provide a reason for rejection',
                status: 'warning',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        try {
            setIsLoading(true);
            await purchaseRequestService.updateStatus(pr.id, 'REJECTED', rejectionReason);
            toast({
                title: 'Success',
                description: 'Purchase Request rejected',
                status: 'success',
                duration: 3000,
                isClosable: true,
            });
            onUpdate();
            onClose();
            setIsRejecting(false);
            setRejectionReason('');
        } catch (error) {
            toast({
                title: 'Error',
                description: 'Failed to reject Purchase Request',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsLoading(false);
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'APPROVED': return 'green';
            case 'REJECTED': return 'red';
            case 'REVISION': return 'orange';
            case 'PO_CREATED': return 'blue';
            default: return 'yellow';
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="xl">
            <ModalOverlay />
            <ModalContent maxW="800px">
                <ModalHeader>
                    <HStack spacing={4}>
                        <Text>Purchase Request Details</Text>
                        <Badge colorScheme={getStatusColor(pr.status)} fontSize="md">
                            {pr.status.replace('_', ' ')}
                        </Badge>
                    </HStack>
                    <Text fontSize="sm" fontWeight="normal" color="gray.500" mt={1}>
                        {pr.code}
                    </Text>
                </ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <VStack align="stretch" spacing={6}>
                        {/* Header Info */}
                        <Box p={4} bg="gray.50" borderRadius="md">
                            <HStack justify="space-between" align="start" spacing={8}>
                                <VStack align="start" spacing={1}>
                                    <Text fontSize="sm" color="gray.500">Project</Text>
                                    <Text fontWeight="medium">{pr.project?.project_name}</Text>
                                </VStack>
                                <VStack align="start" spacing={1}>
                                    <Text fontSize="sm" color="gray.500">Requester</Text>
                                    <Text fontWeight="medium">{pr.requester?.name}</Text>
                                </VStack>
                                <VStack align="start" spacing={1}>
                                    <Text fontSize="sm" color="gray.500">Request Date</Text>
                                    <Text fontWeight="medium">{new Date(pr.request_date).toLocaleDateString('id-ID')}</Text>
                                </VStack>
                                <VStack align="start" spacing={1}>
                                    <Text fontSize="sm" color="gray.500">Required Date</Text>
                                    <Text fontWeight="medium">
                                        {pr.required_date ? new Date(pr.required_date).toLocaleDateString('id-ID') : '-'}
                                    </Text>
                                </VStack>
                            </HStack>
                        </Box>

                        {/* Notes */}
                        {pr.notes && (
                            <Box>
                                <Text fontWeight="bold" mb={2}>Notes</Text>
                                <Text fontSize="sm" color="gray.700">{pr.notes}</Text>
                            </Box>
                        )}

                        {/* Items Table */}
                        <Box>
                            <Text fontWeight="bold" mb={2}>Items</Text>
                            <Table size="sm" variant="simple">
                                <Thead>
                                    <Tr>
                                        <Th>Item Name</Th>
                                        <Th isNumeric>Qty</Th>
                                        <Th>Unit</Th>
                                        <Th isNumeric>Est. Price</Th>
                                        <Th isNumeric>Total</Th>
                                    </Tr>
                                </Thead>
                                <Tbody>
                                    {pr.items?.map((item) => (
                                        <Tr key={item.id}>
                                            <Td>{item.item_name}</Td>
                                            <Td isNumeric>{item.quantity}</Td>
                                            <Td>{item.unit}</Td>
                                            <Td isNumeric>
                                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(item.estimated_price)}
                                            </Td>
                                            <Td isNumeric>
                                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(item.total_price)}
                                            </Td>
                                        </Tr>
                                    ))}
                                    <Tr fontWeight="bold" bg="gray.50">
                                        <Td colSpan={4} textAlign="right">Total Estimated Amount</Td>
                                        <Td isNumeric>
                                            {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(pr.total_amount)}
                                        </Td>
                                    </Tr>
                                </Tbody>
                            </Table>
                        </Box>

                        {/* Approval Info */}
                        {pr.approved_by && (
                            <Box p={4} bg="green.50" borderRadius="md" borderColor="green.200" borderWidth="1px">
                                <HStack>
                                    <Text fontSize="sm" color="green.800">
                                        Approved by <b>{pr.approver?.name}</b> on {new Date(pr.approved_at!).toLocaleDateString('id-ID')}
                                    </Text>
                                </HStack>
                            </Box>
                        )}

                        {/* Rejection Info */}
                        {pr.status === 'REJECTED' && pr.rejection_reason && (
                            <Box p={4} bg="red.50" borderRadius="md" borderColor="red.200" borderWidth="1px">
                                <Text fontWeight="bold" color="red.800" mb={1}>Rejection Reason:</Text>
                                <Text fontSize="sm" color="red.800">{pr.rejection_reason}</Text>
                            </Box>
                        )}

                        {/* Rejection Input */}
                        {isRejecting && (
                            <FormControl isRequired>
                                <FormLabel>Reason for Rejection</FormLabel>
                                <Textarea
                                    value={rejectionReason}
                                    onChange={(e) => setRejectionReason(e.target.value)}
                                    placeholder="Please explain why this request is being rejected..."
                                />
                            </FormControl>
                        )}
                    </VStack>
                </ModalBody>

                <ModalFooter>
                    <Button variant="ghost" mr={3} onClick={() => {
                        setIsRejecting(false);
                        onClose();
                    }}>
                        Close
                    </Button>

                    {pr.status === 'PENDING' && !isRejecting && (
                        <>
                            <Button colorScheme="red" variant="outline" mr={3} onClick={() => setIsRejecting(true)}>
                                Reject
                            </Button>
                            <Button colorScheme="green" onClick={handleApprove} isLoading={isLoading}>
                                Approve
                            </Button>
                        </>
                    )}

                    {isRejecting && (
                        <Button colorScheme="red" onClick={handleReject} isLoading={isLoading}>
                            Confirm Rejection
                        </Button>
                    )}
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
};

export default PRDetailModal;
