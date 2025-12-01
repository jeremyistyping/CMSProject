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
    VStack,
    HStack,
    Text,
    useToast,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    IconButton,
    NumberInput,
    NumberInputField,
    NumberInputStepper,
    NumberIncrementStepper,
    NumberDecrementStepper,
    Textarea,
    Select,
} from '@chakra-ui/react';
import { FiPlus, FiTrash2 } from 'react-icons/fi';
import { useForm, useFieldArray, Control, Controller } from 'react-hook-form';
import purchaseRequestService from '../../services/purchaseRequestService';
import projectService from '../../services/projectService';
import { CreatePRData, PurchaseRequestItem } from '../../types/purchaseRequest';
import { Project } from '../../types/project';

interface CreatePRModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
    prToEdit?: PurchaseRequest | null;
}

const CreatePRModal: React.FC<CreatePRModalProps> = ({ isOpen, onClose, onSuccess, prToEdit }) => {
    const toast = useToast();
    const [projects, setProjects] = useState<Project[]>([]);
    const [isLoading, setIsLoading] = useState(false);

    const { control, register, handleSubmit, reset, watch, setValue } = useForm<CreatePRData>({
        defaultValues: {
            request_date: new Date().toISOString().split('T')[0],
            items: [{ item_name: '', quantity: 1, unit: '', estimated_price: 0, total_price: 0, notes: '' }],
        },
    });

    const { fields, append, remove } = useFieldArray({
        control,
        name: 'items',
    });

    useEffect(() => {
        if (isOpen) {
            fetchProjects();
            if (prToEdit) {
                // Populate form with PR data
                reset({
                    project_id: prToEdit.project_id,
                    request_date: new Date(prToEdit.request_date).toISOString().split('T')[0],
                    required_date: prToEdit.required_date ? new Date(prToEdit.required_date).toISOString().split('T')[0] : '',
                    notes: prToEdit.notes || '',
                    items: prToEdit.items?.map(item => ({
                        item_name: item.item_name,
                        quantity: item.quantity,
                        unit: item.unit,
                        estimated_price: item.estimated_price,
                        total_price: item.total_price,
                        notes: item.notes || ''
                    })) || []
                });
            } else {
                // Reset to default values
                reset({
                    request_date: new Date().toISOString().split('T')[0],
                    items: [{ item_name: '', quantity: 1, unit: '', estimated_price: 0, total_price: 0, notes: '' }],
                });
            }
        }
    }, [isOpen, prToEdit, reset]);

    const fetchProjects = async () => {
        try {
            const data = await projectService.getAllProjects();
            setProjects(data);
        } catch (error) {
            console.error('Error fetching projects:', error);
        }
    };

    const onSubmit = async (data: CreatePRData) => {
        try {
            setIsLoading(true);
            // Calculate total price for each item
            const itemsWithTotal = data.items.map(item => ({
                ...item,
                total_price: item.quantity * item.estimated_price
            }));

            // Format dates to ISO string for backend (RFC3339)
            const formattedData = {
                ...data,
                request_date: new Date(data.request_date).toISOString(),
                required_date: data.required_date ? new Date(data.required_date).toISOString() : undefined,
                items: itemsWithTotal
            };

            if (prToEdit) {
                // Update existing PR
                await purchaseRequestService.update(prToEdit.id, formattedData);
                toast({
                    title: 'Success',
                    description: 'Purchase Request updated successfully',
                    status: 'success',
                    duration: 3000,
                    isClosable: true,
                });
            } else {
                // Create new PR
                await purchaseRequestService.create(formattedData as CreatePRData);
                toast({
                    title: 'Success',
                    description: 'Purchase Request created successfully',
                    status: 'success',
                    duration: 3000,
                    isClosable: true,
                });
            }

            onSuccess();
            onClose();
            reset();
        } catch (error) {
            console.error('Error saving PR:', error);
            toast({
                title: 'Error',
                description: `Failed to ${prToEdit ? 'update' : 'create'} Purchase Request. Please check your input.`,
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsLoading(false);
        }
    };

    // Watch items to calculate total estimated amount
    const watchedItems = watch('items');
    const totalEstimatedAmount = watchedItems?.reduce((sum, item) => {
        return sum + (Number(item.quantity) || 0) * (Number(item.estimated_price) || 0);
    }, 0) || 0;

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="xl">
            <ModalOverlay />
            <ModalContent maxW="900px">
                <ModalHeader>{prToEdit ? 'Edit Purchase Request' : 'Create Purchase Request'}</ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <VStack spacing={4} align="stretch">
                        <HStack spacing={4}>
                            <FormControl isRequired>
                                <FormLabel>Project</FormLabel>
                                <Select placeholder="Select project" {...register('project_id', { required: true, valueAsNumber: true })}>
                                    {projects.map((project) => (
                                        <option key={project.id} value={project.id}>
                                            {project.project_name}
                                        </option>
                                    ))}
                                </Select>
                            </FormControl>
                            <FormControl isRequired>
                                <FormLabel>Request Date</FormLabel>
                                <Input type="date" {...register('request_date', { required: true })} />
                            </FormControl>
                            <FormControl>
                                <FormLabel>Required Date</FormLabel>
                                <Input type="date" {...register('required_date')} />
                            </FormControl>
                        </HStack>

                        <FormControl>
                            <FormLabel>Notes</FormLabel>
                            <Textarea {...register('notes')} placeholder="General notes for this request" />
                        </FormControl>

                        <Text fontWeight="bold" mt={4}>Items</Text>
                        <Table size="sm" variant="simple">
                            <Thead>
                                <Tr>
                                    <Th>Item Name</Th>
                                    <Th width="140px">Qty</Th>
                                    <Th width="100px">Unit</Th>
                                    <Th width="150px">Est. Price</Th>
                                    <Th width="150px">Total</Th>
                                    <Th width="50px"></Th>
                                </Tr>
                            </Thead>
                            <Tbody>
                                {fields.map((field, index) => (
                                    <Tr key={field.id}>
                                        <Td>
                                            <Input {...register(`items.${index}.item_name` as const, { required: true })} placeholder="Item name" size="sm" />
                                        </Td>
                                        <Td>
                                            <Controller
                                                name={`items.${index}.quantity`}
                                                control={control}
                                                rules={{ required: true }}
                                                render={({ field }) => (
                                                    <NumberInput
                                                        size="sm"
                                                        min={0}
                                                        precision={2}
                                                        step={1}
                                                        value={field.value}
                                                        onChange={(val) => field.onChange(Number(val))}
                                                    >
                                                        <NumberInputField />
                                                        <NumberInputStepper>
                                                            <NumberIncrementStepper />
                                                            <NumberDecrementStepper />
                                                        </NumberInputStepper>
                                                    </NumberInput>
                                                )}
                                            />
                                        </Td>
                                        <Td>
                                            <Select {...register(`items.${index}.unit` as const)} size="sm" placeholder="-">
                                                <option value="Pcs">Pcs</option>
                                                <option value="Kg">Kg</option>
                                                <option value="Set">Set</option>
                                                <option value="Ltr">Ltr</option>
                                                <option value="Mtr">Mtr</option>
                                            </Select>
                                        </Td>
                                        <Td>
                                            <Input type="number" step="100" {...register(`items.${index}.estimated_price` as const, { valueAsNumber: true })} size="sm" />
                                        </Td>
                                        <Td>
                                            <Text fontSize="sm">
                                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(
                                                    (watchedItems[index]?.quantity || 0) * (watchedItems[index]?.estimated_price || 0)
                                                )}
                                            </Text>
                                        </Td>
                                        <Td>
                                            <IconButton
                                                aria-label="Remove item"
                                                icon={<FiTrash2 />}
                                                size="xs"
                                                colorScheme="red"
                                                onClick={() => remove(index)}
                                            />
                                        </Td>
                                    </Tr>
                                ))}
                            </Tbody>
                        </Table>

                        <Button leftIcon={<FiPlus />} onClick={() => append({ item_name: '', quantity: 1, unit: '', estimated_price: 0, total_price: 0, notes: '' })} size="sm" alignSelf="start">
                            Add Item
                        </Button>

                        <HStack justify="flex-end" pt={2}>
                            <Text fontWeight="bold">Total Estimated Amount:</Text>
                            <Text fontSize="lg" fontWeight="bold" color="blue.600">
                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(totalEstimatedAmount)}
                            </Text>
                        </HStack>
                    </VStack>
                </ModalBody>

                <ModalFooter>
                    <Button variant="ghost" mr={3} onClick={onClose}>
                        Cancel
                    </Button>
                    <Button colorScheme="blue" onClick={handleSubmit(onSubmit)} isLoading={isLoading}>
                        {prToEdit ? 'Update Request' : 'Create Request'}
                    </Button>
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
};

export default CreatePRModal;
