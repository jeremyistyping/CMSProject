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
} from '@chakra-ui/react';
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

const PRVerificationModal: React.FC<PRVerificationModalProps> = ({
    isOpen,
    onClose,
    pr,
    cbsNodes,
    onSuccess,
}) => {
    const toast = useToast();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [mappings, setMappings] = useState<Partial<PRCBSMapping>[]>([]);
    const [notes, setNotes] = useState('');
    const [flattenedNodes, setFlattenedNodes] = useState<CBSNode[]>([]);

    // Flatten CBS tree for select dropdown
    useEffect(() => {
        const flatten = (nodes: CBSNode[], level = 0): CBSNode[] => {
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

    // Initialize mappings when PR changes
    useEffect(() => {
        if (pr && isOpen) {
            // Default: one mapping per item, initially unassigned
            const initialMappings = pr.items.map(item => ({
                pr_item_id: item.id,
                cbs_node_id: 0, // 0 means unassigned
                allocated_amount: item.total_price,
            }));
            setMappings(initialMappings);
            setNotes('');
        }
    }, [pr, isOpen]);

    const handleMappingChange = (index: number, field: keyof PRCBSMapping, value: any) => {
        const newMappings = [...mappings];
        newMappings[index] = { ...newMappings[index], [field]: value };
        setMappings(newMappings);
    };

    const calculateTotalAllocated = () => {
        return mappings.reduce((sum, m) => sum + (m.allocated_amount || 0), 0);
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

        // Validation
        const totalAllocated = calculateTotalAllocated();
        const prTotal = pr.total_amount;

        // Allow small floating point difference
        if (Math.abs(totalAllocated - prTotal) > 100) {
            toast({
                title: 'Validation Error',
                description: `Total allocated (${formatCurrency(totalAllocated)}) must match PR total (${formatCurrency(prTotal)})`,
                status: 'error',
                duration: 5000,
                isClosable: true,
            });
            return;
        }

        const unassigned = mappings.some(m => !m.cbs_node_id || m.cbs_node_id === 0);
        if (unassigned) {
            toast({
                title: 'Validation Error',
                description: 'All items must be assigned to a CBS node',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        setIsSubmitting(true);
        try {
            await cbsService.verifyPR(pr.id, mappings, notes);
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
        <Modal isOpen={isOpen} onClose={onClose} size="4xl">
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

                        {/* CBS Mapping Table */}
                        <Box>
                            <Text fontWeight="bold" mb={3}>Cost Allocation</Text>
                            <Table size="sm">
                                <Thead>
                                    <Tr>
                                        <Th>Item</Th>
                                        <Th isNumeric>Amount</Th>
                                        <Th>CBS Node</Th>
                                        <Th>Allocation</Th>
                                    </Tr>
                                </Thead>
                                <Tbody>
                                    {pr.items.map((item, index) => (
                                        <Tr key={item.id}>
                                            <Td>
                                                <Text fontWeight="medium">{item.item_name}</Text>
                                                <Text fontSize="xs" color="gray.500">{item.quantity} {item.unit}</Text>
                                            </Td>
                                            <Td isNumeric>{formatCurrency(item.total_price)}</Td>
                                            <Td>
                                                <Select
                                                    size="sm"
                                                    placeholder="Select CBS Node"
                                                    value={mappings[index]?.cbs_node_id || ''}
                                                    onChange={(e) => handleMappingChange(index, 'cbs_node_id', Number(e.target.value))}
                                                >
                                                    {flattenedNodes.map(node => (
                                                        <option key={node.id} value={node.id}>
                                                            {'\u00A0'.repeat((node.level || 0) * 4)} {node.code} - {node.name}
                                                        </option>
                                                    ))}
                                                </Select>
                                            </Td>
                                            <Td>
                                                <NumberInput
                                                    size="sm"
                                                    value={mappings[index]?.allocated_amount}
                                                    onChange={(val) => handleMappingChange(index, 'allocated_amount', Number(val))}
                                                >
                                                    <NumberInputField />
                                                </NumberInput>
                                            </Td>
                                        </Tr>
                                    ))}
                                </Tbody>
                            </Table>
                        </Box>

                        {/* Validation Alert */}
                        {Math.abs(calculateTotalAllocated() - pr.total_amount) > 100 && (
                            <Alert status="warning" borderRadius="md">
                                <AlertIcon />
                                Allocation mismatch: {formatCurrency(calculateTotalAllocated())} vs {formatCurrency(pr.total_amount)}
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
                        isDisabled={Math.abs(calculateTotalAllocated() - pr.total_amount) > 100}
                    >
                        Verify & Submit
                    </Button>
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
};

export default PRVerificationModal;
