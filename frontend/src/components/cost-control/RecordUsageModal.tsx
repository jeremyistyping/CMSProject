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
    Select,
    Textarea,
    VStack,
    Text,
    useToast,
    FormErrorMessage,
    NumberInput,
    NumberInputField,
    NumberInputStepper,
    NumberIncrementStepper,
    NumberDecrementStepper,
} from '@chakra-ui/react';
import { materialTrackingService, MaterialItemSummary } from '@/services/materialTrackingService';

interface RecordUsageModalProps {
    isOpen: boolean;
    onClose: () => void;
    projectId: number;
    onSuccess: () => void;
    preSelectedProductId?: number;
}

const RecordUsageModal: React.FC<RecordUsageModalProps> = ({
    isOpen,
    onClose,
    projectId,
    onSuccess,
    preSelectedProductId,
}) => {
    const [items, setItems] = useState<MaterialItemSummary[]>([]);
    const [selectedProductId, setSelectedProductId] = useState<number | ''>('');
    const [quantity, setQuantity] = useState<number>(1);
    const [notes, setNotes] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const toast = useToast();

    useEffect(() => {
        if (isOpen && projectId) {
            loadItems();
        }
    }, [isOpen, projectId]);

    useEffect(() => {
        if (preSelectedProductId) {
            setSelectedProductId(preSelectedProductId);
        }
    }, [preSelectedProductId]);

    const loadItems = async () => {
        setIsLoading(true);
        try {
            const data = await materialTrackingService.getItems(projectId);
            setItems(data || []);
        } catch (error) {
            console.error('Failed to load items:', error);
            toast({
                title: 'Error',
                description: 'Failed to load material items',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsLoading(false);
        }
    };

    const selectedItem = items.find((i) => i.product_id === Number(selectedProductId));
    const maxQuantity = selectedItem ? selectedItem.remaining_qty : 0;

    const handleSubmit = async () => {
        if (!selectedProductId) {
            toast({
                title: 'Error',
                description: 'Please select a material',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        if (quantity <= 0) {
            toast({
                title: 'Error',
                description: 'Quantity must be greater than 0',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        if (quantity > maxQuantity) {
            toast({
                title: 'Error',
                description: `Quantity cannot exceed available stock (${maxQuantity})`,
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        setIsSubmitting(true);
        try {
            await materialTrackingService.recordUsage(
                projectId,
                Number(selectedProductId),
                quantity,
                notes
            );

            toast({
                title: 'Success',
                description: 'Material usage recorded successfully',
                status: 'success',
                duration: 3000,
                isClosable: true,
            });

            onSuccess();
            onClose();
            // Reset form
            setQuantity(1);
            setNotes('');
            if (!preSelectedProductId) setSelectedProductId('');
        } catch (error: any) {
            console.error('Failed to record usage:', error);
            toast({
                title: 'Error',
                description: error.response?.data?.error || 'Failed to record usage',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <ModalOverlay />
            <ModalContent>
                <ModalHeader>Record Material Usage</ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <VStack spacing={4} align="stretch">
                        <FormControl isRequired>
                            <FormLabel>Material Item</FormLabel>
                            <Select
                                placeholder="Select material"
                                value={selectedProductId}
                                onChange={(e) => setSelectedProductId(Number(e.target.value))}
                                isDisabled={isLoading || !!preSelectedProductId}
                            >
                                {items.map((item) => (
                                    <option key={item.product_id} value={item.product_id}>
                                        {item.product_name} ({item.remaining_qty} {item.unit} available)
                                    </option>
                                ))}
                            </Select>
                        </FormControl>

                        {selectedItem && (
                            <Text fontSize="sm" color="gray.500">
                                Available Stock: <b>{selectedItem.remaining_qty} {selectedItem.unit}</b>
                            </Text>
                        )}

                        <FormControl isRequired isInvalid={quantity > maxQuantity}>
                            <FormLabel>Quantity Used</FormLabel>
                            <NumberInput
                                min={1}
                                max={maxQuantity}
                                value={quantity}
                                onChange={(_, val) => setQuantity(val)}
                            >
                                <NumberInputField />
                                <NumberInputStepper>
                                    <NumberIncrementStepper />
                                    <NumberDecrementStepper />
                                </NumberInputStepper>
                            </NumberInput>
                            {quantity > maxQuantity && (
                                <FormErrorMessage>Cannot exceed available stock</FormErrorMessage>
                            )}
                        </FormControl>

                        <FormControl>
                            <FormLabel>Notes</FormLabel>
                            <Textarea
                                placeholder="Describe usage (e.g., used for foundation work)"
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
                        colorScheme="blue"
                        onClick={handleSubmit}
                        isLoading={isSubmitting}
                        isDisabled={!selectedProductId || quantity <= 0 || quantity > maxQuantity}
                    >
                        Record Usage
                    </Button>
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
};

export default RecordUsageModal;
