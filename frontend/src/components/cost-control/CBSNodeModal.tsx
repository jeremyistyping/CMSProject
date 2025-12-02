import React, { useEffect } from 'react';
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
    NumberInput,
    NumberInputField,
    NumberInputStepper,
    NumberIncrementStepper,
    NumberDecrementStepper,
    FormErrorMessage,
    useToast,
} from '@chakra-ui/react';
import { useForm } from 'react-hook-form';
import { CBSNode } from '../../types/cbs';
import cbsService from '../../services/cbsService';

interface CBSNodeModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
    projectId: number;
    parentId?: number;
    nodeToEdit?: CBSNode;
}

interface FormData {
    code: string;
    name: string;
    description: string;
    budget_amount: number;
}

const CBSNodeModal: React.FC<CBSNodeModalProps> = ({
    isOpen,
    onClose,
    onSuccess,
    projectId,
    parentId,
    nodeToEdit,
}) => {
    const toast = useToast();
    const isEdit = !!nodeToEdit;

    const {
        register,
        handleSubmit,
        reset,
        setValue,
        formState: { errors, isSubmitting },
    } = useForm<FormData>();

    useEffect(() => {
        if (isOpen) {
            if (nodeToEdit) {
                setValue('code', nodeToEdit.code);
                setValue('name', nodeToEdit.name);
                setValue('description', nodeToEdit.description || '');
                setValue('budget_amount', nodeToEdit.budget_amount);
            } else {
                reset({
                    code: '',
                    name: '',
                    description: '',
                    budget_amount: 0,
                });
            }
        }
    }, [isOpen, nodeToEdit, reset, setValue]);

    const onSubmit = async (data: FormData) => {
        try {
            const payload = {
                ...data,
                project_id: projectId,
                parent_id: parentId,
            };

            if (isEdit && nodeToEdit) {
                await cbsService.updateCBSNode(nodeToEdit.id, payload);
                toast({
                    title: 'Success',
                    description: 'CBS Node updated successfully',
                    status: 'success',
                    duration: 3000,
                    isClosable: true,
                });
            } else {
                await cbsService.createCBSNode(payload);
                toast({
                    title: 'Success',
                    description: 'CBS Node created successfully',
                    status: 'success',
                    duration: 3000,
                    isClosable: true,
                });
            }
            onSuccess();
            onClose();
        } catch (error: any) {
            console.error('Error saving CBS node:', error);
            toast({
                title: 'Error',
                description: error.response?.data?.error || 'Failed to save CBS node',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="lg">
            <ModalOverlay />
            <ModalContent>
                <ModalHeader>{isEdit ? 'Edit CBS Node' : 'Add New CBS Node'}</ModalHeader>
                <ModalCloseButton />
                <form onSubmit={handleSubmit(onSubmit)}>
                    <ModalBody>
                        <VStack spacing={4}>
                            <FormControl isInvalid={!!errors.code} isRequired>
                                <FormLabel>Cost Code</FormLabel>
                                <Input
                                    placeholder="e.g. 1.1.1"
                                    {...register('code', { required: 'Cost code is required' })}
                                />
                                <FormErrorMessage>{errors.code?.message}</FormErrorMessage>
                            </FormControl>

                            <FormControl isInvalid={!!errors.name} isRequired>
                                <FormLabel>Name</FormLabel>
                                <Input
                                    placeholder="e.g. Site Preparation"
                                    {...register('name', { required: 'Name is required' })}
                                />
                                <FormErrorMessage>{errors.name?.message}</FormErrorMessage>
                            </FormControl>

                            <FormControl>
                                <FormLabel>Description</FormLabel>
                                <Textarea
                                    placeholder="Optional description..."
                                    {...register('description')}
                                />
                            </FormControl>

                            <FormControl isInvalid={!!errors.budget_amount}>
                                <FormLabel>Budget Amount (IDR)</FormLabel>
                                <NumberInput min={0}>
                                    <NumberInputField
                                        {...register('budget_amount', {
                                            valueAsNumber: true,
                                            min: { value: 0, message: 'Budget cannot be negative' },
                                        })}
                                    />
                                    <NumberInputStepper>
                                        <NumberIncrementStepper />
                                        <NumberDecrementStepper />
                                    </NumberInputStepper>
                                </NumberInput>
                                <FormErrorMessage>{errors.budget_amount?.message}</FormErrorMessage>
                            </FormControl>
                        </VStack>
                    </ModalBody>

                    <ModalFooter>
                        <Button variant="ghost" mr={3} onClick={onClose}>
                            Cancel
                        </Button>
                        <Button colorScheme="blue" type="submit" isLoading={isSubmitting}>
                            {isEdit ? 'Update' : 'Create'}
                        </Button>
                    </ModalFooter>
                </form>
            </ModalContent>
        </Modal>
    );
};

export default CBSNodeModal;
