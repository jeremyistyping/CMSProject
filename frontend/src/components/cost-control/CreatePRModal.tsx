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
    Textarea,
    Select,
    Box,
    Badge,
} from '@chakra-ui/react';
import { FiPlus, FiTrash2, FiPackage } from 'react-icons/fi';
import { useForm, useFieldArray, Controller } from 'react-hook-form';
import purchaseRequestService from '../../services/purchaseRequestService';
import projectService from '../../services/projectService';
import { materialService, vendorService, coaService, Material, Vendor, UnitOfMeasure, COAAccount } from '../../services/masterDataService';
import { CreatePRData, PurchaseRequest } from '../../types/purchaseRequest';
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
    const [materials, setMaterials] = useState<Material[]>([]);
    const [vendors, setVendors] = useState<Vendor[]>([]);
    const [uoms, setUoms] = useState<UnitOfMeasure[]>([]);
    const [coaAccounts, setCoaAccounts] = useState<COAAccount[]>([]);
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
            fetchInitialData();
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

    const fetchInitialData = async () => {
        try {
            const [projectsData, materialsData, vendorsData, uomsData, coaData] = await Promise.all([
                projectService.getAllProjects(),
                materialService.getAll({ is_active: true }),
                vendorService.getAll({ is_active: true }),
                materialService.getUoM(),
                coaService.getAll({ is_active: true }),
            ]);
            setProjects(projectsData);
            setMaterials(materialsData.data || []);
            setVendors(vendorsData.data || []);
            setUoms(uomsData.data || []);
            setCoaAccounts(coaData.data || []);
            
            // Debug: Log materials to check if COA is included
            console.log('Materials loaded:', materialsData.data?.slice(0, 2));
            console.log('COA Accounts loaded:', coaData.data?.slice(0, 2));
        } catch (error) {
            console.error('Error fetching initial data:', error);
        }
    };

    const handleMaterialSelect = (index: number, materialId: string) => {
        if (!materialId) return;
        const material = materials.find(m => m.id === parseInt(materialId));
        if (material) {
            setValue(`items.${index}.item_name`, material.name);
            setValue(`items.${index}.unit`, material.unit);
            setValue(`items.${index}.estimated_price`, material.unit_price);
            setValue(`items.${index}.material_id`, material.id);
            
            // Set COA from material if available
            if (material.coa_account_id) {
                setValue(`items.${index}.coa_account_id` as any, material.coa_account_id);
            }
            
            // Debug log
            console.log('Material selected:', material);
        }
    };

    // Helper to get COA info for an item
    const getItemCOAInfo = (index: number) => {
        const item = watchedItems[index];
        
        // First try to get from selected COA
        if (item?.coa_account_id) {
            const coa = coaAccounts.find(c => c.id === item.coa_account_id);
            if (coa) return coa;
        }
        
        // Then try to get from material
        if (item?.material_id) {
            const material = materials.find(m => m.id === item.material_id);
            if (material?.coa_account_id) {
                const coa = coaAccounts.find(c => c.id === material.coa_account_id);
                if (coa) return coa;
            }
            // If material has coa_account object directly
            if (material?.coa_account) {
                return material.coa_account;
            }
        }
        
        return null;
    };



    const onSubmit = async (data: CreatePRData) => {
        try {
            setIsLoading(true);
            // Calculate total price for each item
            const itemsWithTotal = data.items.map(item => ({
                ...item,
                total_price: (item.quantity || 0) * (item.estimated_price || 0)
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
                await purchaseRequestService.update(prToEdit.id, formattedData as any);
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
                            <FormControl isRequired flex={2}>
                                <FormLabel>Project</FormLabel>
                                <Select placeholder="Select project" {...register('project_id', { required: true, valueAsNumber: true })}>
                                    {projects.map((project) => (
                                        <option key={project.id} value={project.id}>
                                            {project.project_name}
                                        </option>
                                    ))}
                                </Select>
                            </FormControl>
                            <FormControl flex={1}>
                                <FormLabel>Vendor (Optional)</FormLabel>
                                <Select placeholder="Select vendor" {...register('vendor_id', { valueAsNumber: true })}>
                                    {vendors.map((vendor) => (
                                        <option key={vendor.id} value={vendor.id}>
                                            {vendor.name}
                                        </option>
                                    ))}
                                </Select>
                            </FormControl>
                        </HStack>
                        <HStack spacing={4}>
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

                        <HStack justify="space-between" mt={4}>
                            <Text fontWeight="bold">Items</Text>
                            <Badge colorScheme="blue" fontSize="xs">
                                <FiPackage style={{ display: 'inline', marginRight: 4 }} />
                                Pilih dari Master Material atau input manual
                            </Badge>
                        </HStack>
                        <Box overflowX="auto">
                            <Table size="sm" variant="simple">
                                <Thead>
                                    <Tr>
                                        <Th width="200px">Material</Th>
                                        <Th>Item Name</Th>
                                        <Th width="150px">COA / Budget Category</Th>
                                        <Th width="100px">Qty</Th>
                                        <Th width="80px">Unit</Th>
                                        <Th width="130px">Est. Price</Th>
                                        <Th width="130px">Total</Th>
                                        <Th width="40px"></Th>
                                    </Tr>
                                </Thead>
                                <Tbody>
                                    {fields.map((field, index) => {
                                        const coaInfo = getItemCOAInfo(index);
                                        return (
                                            <Tr key={field.id}>
                                                <Td>
                                                    <Select
                                                        size="sm"
                                                        placeholder="Pilih material..."
                                                        onChange={(e) => handleMaterialSelect(index, e.target.value)}
                                                    >
                                                        {materials.map((mat) => (
                                                            <option key={mat.id} value={mat.id}>
                                                                {mat.code} - {mat.name}
                                                            </option>
                                                        ))}
                                                    </Select>
                                                </Td>
                                                <Td>
                                                    <Input {...register(`items.${index}.item_name` as const, { required: true })} placeholder="Item name" size="sm" />
                                                </Td>
                                                <Td>
                                                    <VStack align="start" spacing={1} width="100%">
                                                        <Select
                                                            size="xs"
                                                            placeholder="Pilih COA..."
                                                            {...register(`items.${index}.coa_account_id` as any, { valueAsNumber: true })}
                                                            value={watchedItems[index]?.coa_account_id || ''}
                                                        >
                                                            {coaAccounts.map((coa) => (
                                                                <option key={coa.id} value={coa.id}>
                                                                    {coa.code} - {coa.name}
                                                                </option>
                                                            ))}
                                                        </Select>
                                                        {coaInfo && (
                                                            <HStack spacing={1} flexWrap="wrap">
                                                                <Badge colorScheme="green" fontSize="xx-small">
                                                                    {coaInfo.budget_category?.replace('_', ' ')}
                                                                </Badge>
                                                                {coaInfo.work_package && (
                                                                    <Badge colorScheme="blue" fontSize="xx-small">
                                                                        {coaInfo.work_package}
                                                                    </Badge>
                                                                )}
                                                            </HStack>
                                                        )}
                                                    </VStack>
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
                                                            </NumberInput>
                                                        )}
                                                    />
                                                </Td>
                                                <Td>
                                                    <Select {...register(`items.${index}.unit` as const)} size="sm" placeholder="-">
                                                        {uoms.map((u) => (
                                                            <option key={u.id} value={u.code}>{u.code}</option>
                                                        ))}
                                                    </Select>
                                                </Td>
                                                <Td>
                                                    <Input type="number" step="100" {...register(`items.${index}.estimated_price` as const, { valueAsNumber: true })} size="sm" />
                                                </Td>
                                                <Td>
                                                    <Text fontSize="sm" fontWeight="medium">
                                                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(
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
                                                        variant="ghost"
                                                        onClick={() => remove(index)}
                                                    />
                                                </Td>
                                            </Tr>
                                        );
                                    })}
                                </Tbody>
                            </Table>
                        </Box>

                        <Button leftIcon={<FiPlus />} onClick={() => append({ item_name: '', quantity: 1, unit: '', estimated_price: 0, total_price: 0, notes: '', material_id: undefined })} size="sm" alignSelf="start" colorScheme="blue" variant="outline">
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
