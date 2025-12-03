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
    VStack,
    HStack,
    Text,
    Box,
    Select,
    NumberInput,
    NumberInputField,
    FormControl,
    FormLabel,
    Textarea,
    useToast,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Badge,
    Alert,
    AlertIcon,
    IconButton,
    Divider,
} from '@chakra-ui/react';
import { FiPlus, FiTrash2 } from 'react-icons/fi';
import { PurchaseRequest } from '../../types/purchaseRequest';
import { CBSNode, PRCBSMapping } from '../../types/cbs';
import cbsService from '../../services/cbsService';

interface PRVerificationModalProps {
    isOpen: boolean;
    onClose: () => void;
    pr: PurchaseRequest | null;
    cbsNodes: CBSNode[];
    onSuccess: () => void;
}

interface ItemAllocation {
    id: string; // Unique ID for React key
    cbs_node_id: number;
    allocated_amount: number;
}

interface ItemAllocations {
    [itemId: number]: ItemAllocation[];
}

const PRVerificationModal: React.FC<PRVerificationModalProps> = ({
    isOpen,
    onClose,
    pr,
    cbsNodes,
    onSuccess,
}) => {
    const toast = useToast();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [itemAllocations, setItemAllocations] = useState<ItemAllocations>({});
    const [notes, setNotes] = useState('');
    const [flattenedNodes, setFlattenedNodes] = useState<CBSNode[]>([]);

    // Flatten CBS tree for select dropdown
    useEffect(() => {
        const flatten = (nodes: CBSNode[], level = 0): CBSNode[] => {
            if (!nodes) return [];
            let result: CBSNode[] = [];
            nodes.forEach(node => {
                result.push({ ...node, level });
                if (node.children) {
                    result = result.concat(flatten(node.children, level + 1));
                }
            });
            return result;
        };
        setFlattenedNodes(flatten(cbsNodes));
    }, [cbsNodes]);

    // Initialize allocations when PR changes
    useEffect(() => {
        if (pr && isOpen) {
            const initialAllocations: ItemAllocations = {};
            pr.items.forEach(item => {
                initialAllocations[item.id] = [
                    {
                        id: `${item.id}-0`,
                        cbs_node_id: 0,
                        allocated_amount: item.total_price,
                    }
                ];
            });
            setItemAllocations(initialAllocations);
            setNotes('');
        }
    }, [pr, isOpen]);

    const addAllocation = (itemId: number) => {
        setItemAllocations(prev => ({
            ...prev,
            [itemId]: [
                ...prev[itemId],
                {
                    id: `${itemId}-${Date.now()}`,
                    cbs_node_id: 0,
                    allocated_amount: 0,
                }
            ]
        }));
    };

    const removeAllocation = (itemId: number, allocationId: string) => {
        setItemAllocations(prev => ({
            ...prev,
            [itemId]: prev[itemId].filter(a => a.id !== allocationId)
        }));
    };

    const updateAllocation = (itemId: number, allocationId: string, field: 'cbs_node_id' | 'allocated_amount', value: number) => {
        setItemAllocations(prev => ({
            ...prev,
            [itemId]: prev[itemId].map(a =>
                a.id === allocationId ? { ...a, [field]: value } : a
            )
        }));
    };

    const getItemTotal = (itemId: number): number => {
        return itemAllocations[itemId]?.reduce((sum, a) => sum + (a.allocated_amount || 0), 0) || 0;
    };

    const getItemExpected = (itemId: number): number => {
        return pr?.items.find(i => i.id === itemId)?.total_price || 0;
    };

    const isItemValid = (itemId: number): boolean => {
        const allocations = itemAllocations[itemId] || [];
        const total = getItemTotal(itemId);
        const expected = getItemExpected(itemId);
        const allAssigned = allocations.every(a => a.cbs_node_id !== 0);
        return Math.abs(total - expected) <= 100 && allAssigned;
    };

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(amount);
    };

    const handleSubmit = async () => {
        if (!pr) return;

        // Validate all items
        const allValid = pr.items.every(item => isItemValid(item.id));
        if (!allValid) {
            toast({
                title: 'Validation Error',
                description: 'All items must be fully allocated and CBS nodes must be selected',
                status: 'error',
                duration: 5000,
                isClosable: true,
            });
            return;
        }

        // Flatten all allocations into single array
        const allMappings: Partial<PRCBSMapping>[] = [];
        pr.items.forEach(item => {
            const allocations = itemAllocations[item.id] || [];
            allocations.forEach(allocation => {
                allMappings.push({
                    pr_item_id: item.id,
                    cbs_node_id: allocation.cbs_node_id,
                    allocated_amount: allocation.allocated_amount,
                });
            });
        });

        setIsSubmitting(true);
        try {
            await cbsService.verifyPR(pr.id, allMappings, notes);
            toast({
                title: 'Success',
                description: 'Purchase Request verified successfully',
                status: 'success',
                duration: 3000,
                isClosable: true,
            });
            onSuccess();
            onClose();
        } catch (error: any) {
            console.error('Error verifying PR:', error);
            toast({
                title: 'Error',
                description: error.response?.data?.error || 'Failed to verify PR',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsSubmitting(false);
        }
    };

    if (!pr) return null;

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="6xl">
            <ModalOverlay />
            <ModalContent>
                <ModalHeader>Verify Purchase Request: {pr.code}</ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <VStack spacing={6} align="stretch">
                        {/* PR Summary */}
                        <Box p={4} bg="gray.50" borderRadius="md">
                            <HStack justify="space-between" mb={2}>
                                <Text fontWeight="bold">Project: {pr.project?.project_name}</Text>
                                <Badge colorScheme="blue">{pr.status}</Badge>
                            </HStack>
                            <Text>Total Amount: <strong>{formatCurrency(pr.total_amount)}</strong></Text>
                            <Text fontSize="sm" color="gray.600" mt={2}>{pr.notes}</Text>
                        </Box>

                        {/* Cost Allocation by Item */}
                        <Box>
                            <Text fontWeight="bold" mb={3}>Cost Allocation</Text>
                            <VStack spacing={4} align="stretch">
                                {pr.items.map((item) => {
                                    const allocations = itemAllocations[item.id] || [];
                                    const itemTotal = getItemTotal(item.id);
                                    const itemExpected = getItemExpected(item.id);
                                    const variance = itemTotal - itemExpected;
                                    const isValid = isItemValid(item.id);

                                    return (
                                        <Box key={item.id} p={4} borderWidth="1px" borderRadius="md" borderColor={isValid ? 'green.200' : 'red.200'}>
                                            {/* Item Header */}
                                            <HStack justify="space-between" mb={3}>
                                                <Box>
                                                    <Text fontWeight="bold">{item.item_name}</Text>
                                                    <Text fontSize="sm" color="gray.600">
                                                        {item.quantity} {item.unit} × {formatCurrency(item.estimated_price)} = {formatCurrency(item.total_price)}
                                                    </Text>
                                                </Box>
                                                <VStack align="end" spacing={0}>
                                                    <Text fontSize="sm" fontWeight="bold">
                                                        Allocated: {formatCurrency(itemTotal)}
                                                    </Text>
                                                    {Math.abs(variance) > 100 && (
                                                        <Text fontSize="xs" color={variance > 0 ? 'red.500' : 'orange.500'}>
                                                            {variance > 0 ? '+' : ''}{formatCurrency(variance)}
                                                        </Text>
                                                    )}
                                                </VStack>
                                            </HStack>

                                            {/* Allocations Table */}
                                            <Table size="sm" variant="simple">
                                                <Thead>
                                                    <Tr>
                                                        <Th>CBS Node</Th>
                                                        <Th isNumeric>Amount</Th>
                                                        <Th width="60px"></Th>
                                                    </Tr>
                                                </Thead>
                                                <Tbody>
                                                    {allocations.map((allocation) => (
                                                        <Tr key={allocation.id}>
                                                            <Td>
                                                                <Select
                                                                    size="sm"
                                                                    placeholder="Select CBS Node"
                                                                    value={allocation.cbs_node_id || ''}
                                                                    onChange={(e) => updateAllocation(item.id, allocation.id, 'cbs_node_id', Number(e.target.value))}
                                                                >
                                                                    {flattenedNodes.map(node => (
                                                                        <option key={node.id} value={node.id}>
                                                                            {'\u00A0'.repeat((node.level || 0) * 4)} {node.code} - {node.name}
                                                                        </option>
                                                                    ))}
                                                                </Select>
                                                            </Td>
                                                            <Td isNumeric>
                                                                <NumberInput
                                                                    size="sm"
                                                                    value={allocation.allocated_amount}
                                                                    onChange={(val) => updateAllocation(item.id, allocation.id, 'allocated_amount', Number(val))}
                                                                    min={0}
                                                                >
                                                                    <NumberInputField textAlign="right" />
                                                                </NumberInput>
                                                            </Td>
                                                            <Td>
                                                                {allocations.length > 1 && (
                                                                    <IconButton
                                                                        aria-label="Remove allocation"
                                                                        icon={<FiTrash2 />}
                                                                        size="sm"
                                                                        colorScheme="red"
                                                                        variant="ghost"
                                                                        onClick={() => removeAllocation(item.id, allocation.id)}
                                                                    />
                                                                )}
                                                            </Td>
                                                        </Tr>
                                                    ))}
                                                </Tbody>
                                            </Table>

                                            {/* Add Allocation Button */}
                                            <Button
                                                size="sm"
                                                leftIcon={<FiPlus />}
                                                variant="ghost"
                                                colorScheme="blue"
                                                mt={2}
                                                onClick={() => addAllocation(item.id)}
                                            >
                                                Add CBS Allocation
                                            </Button>
                                        </Box>
                                    );
                                })}
                            </VStack>
                        </Box>

                        {/* Overall Validation Alert */}
                        {pr.items.some(item => !isItemValid(item.id)) && (
                            <Alert status="warning" borderRadius="md">
                                <AlertIcon />
                                Some items have allocation mismatches or unassigned CBS nodes
                            </Alert>
                        )}

                        {/* Verification Notes */}
                        <FormControl>
                            <FormLabel>Verification Notes</FormLabel>
                            <Textarea
                                placeholder="Add notes about budget verification..."
                                value={notes}
                                onChange={(e) => setNotes(e.target.value)}
                            />
                        </FormControl>
                    </VStack>
                </ModalBody>

                <ModalFooter>
                    <Button variant="ghost" mr={3} onClick={onClose}>
                        Cancel
                    </Button>
                    <Button
                        colorScheme="green"
                        onClick={handleSubmit}
                        isLoading={isSubmitting}
                        isDisabled={pr.items.some(item => !isItemValid(item.id))}
                    >
                        Verify & Submit
                    </Button>
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
};

export default PRVerificationModal;
