'use client';

import { useState, useEffect } from 'react';
import {
    Modal, ModalOverlay, ModalContent, ModalHeader, ModalBody, ModalFooter, ModalCloseButton,
    Button, FormControl, FormLabel, Input, Select, Textarea, VStack, HStack, useToast,
    NumberInput, NumberInputField, NumberInputStepper, NumberIncrementStepper, NumberDecrementStepper,
} from '@chakra-ui/react';
import { coaService, COAAccount } from '@/services/masterDataService';
import { expenseTransactionService, CreateExpenseTransactionDTO, ExpenseTransaction } from '@/services/expenseTransactionService';

interface ExpenseTransactionFormProps {
    isOpen: boolean;
    onClose: () => void;
    projectId: number;
    expense?: ExpenseTransaction | null;
    onSuccess: () => void;
}

const TRANSACTION_TYPES = [
    { value: 'LABOUR', label: 'Tenaga Kerja' },
    { value: 'MATERIAL', label: 'Material' },
    { value: 'OPERATIONAL', label: 'Operasional' },
    { value: 'OTHER', label: 'Lainnya' },
];

const UNITS = [
    { value: 'ls', label: 'Lump Sum (ls)' },
    { value: 'pcs', label: 'Pieces (pcs)' },
    { value: 'm2', label: 'Meter Persegi (m²)' },
    { value: 'm3', label: 'Meter Kubik (m³)' },
    { value: 'kg', label: 'Kilogram (kg)' },
    { value: 'ton', label: 'Ton' },
    { value: 'liter', label: 'Liter' },
    { value: 'unit', label: 'Unit' },
];

export default function ExpenseTransactionForm({ isOpen, onClose, projectId, expense, onSuccess }: ExpenseTransactionFormProps) {
    const [loading, setLoading] = useState(false);
    const [coaAccounts, setCoaAccounts] = useState<COAAccount[]>([]);
    const [formData, setFormData] = useState<CreateExpenseTransactionDTO>({
        project_id: projectId,
        transaction_date: new Date().toISOString().split('T')[0],
        coa_account_id: 0,
        description: '',
        amount: 0,
        unit: 'ls',
        quantity: 1,
        transaction_type: 'OPERATIONAL',
        reference_no: '',
        notes: '',
    });

    const toast = useToast();

    useEffect(() => {
        if (isOpen) {
            loadCOAAccounts();
            if (expense) {
                setFormData({
                    project_id: expense.project_id,
                    transaction_date: expense.transaction_date.split('T')[0],
                    coa_account_id: expense.coa_account_id,
                    description: expense.description,
                    amount: expense.amount,
                    unit: expense.unit,
                    quantity: expense.quantity,
                    transaction_type: expense.transaction_type,
                    reference_no: expense.reference_no || '',
                    notes: expense.notes || '',
                });
            } else {
                resetForm();
            }
        }
    }, [isOpen, expense]);

    const loadCOAAccounts = async () => {
        try {
            const result = await coaService.getAll({ is_active: true, is_header: false });
            setCoaAccounts(result.data || []);
        } catch (error) {
            console.error('Error loading COA accounts:', error);
            toast({
                title: 'Gagal memuat COA',
                status: 'error',
                duration: 3000,
            });
        }
    };

    const resetForm = () => {
        setFormData({
            project_id: projectId,
            transaction_date: new Date().toISOString().split('T')[0],
            coa_account_id: 0,
            description: '',
            amount: 0,
            unit: 'ls',
            quantity: 1,
            transaction_type: 'OPERATIONAL',
            reference_no: '',
            notes: '',
        });
    };

    const handleSubmit = async () => {
        if (!formData.coa_account_id) {
            toast({
                title: 'COA harus dipilih',
                status: 'warning',
                duration: 3000,
            });
            return;
        }

        if (!formData.description.trim()) {
            toast({
                title: 'Deskripsi harus diisi',
                status: 'warning',
                duration: 3000,
            });
            return;
        }

        if (formData.amount <= 0) {
            toast({
                title: 'Jumlah harus lebih dari 0',
                status: 'warning',
                duration: 3000,
            });
            return;
        }

        setLoading(true);
        try {
            if (expense) {
                await expenseTransactionService.update(expense.id, formData);
                toast({
                    title: 'Transaksi berhasil diupdate',
                    status: 'success',
                    duration: 3000,
                });
            } else {
                await expenseTransactionService.create(projectId, formData);
                toast({
                    title: 'Transaksi berhasil ditambahkan',
                    status: 'success',
                    duration: 3000,
                });
            }
            onSuccess();
            onClose();
        } catch (error: any) {
            toast({
                title: 'Gagal menyimpan transaksi',
                description: error.response?.data?.error || error.message,
                status: 'error',
                duration: 5000,
            });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="xl">
            <ModalOverlay />
            <ModalContent>
                <ModalHeader>{expense ? 'Edit Transaksi Biaya' : 'Tambah Transaksi Biaya'}</ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <VStack spacing={4}>
                        <HStack w="full" spacing={4}>
                            <FormControl isRequired>
                                <FormLabel>Tanggal</FormLabel>
                                <Input
                                    type="date"
                                    value={formData.transaction_date}
                                    onChange={(e) => setFormData({ ...formData, transaction_date: e.target.value })}
                                />
                            </FormControl>
                            <FormControl isRequired>
                                <FormLabel>Tipe Transaksi</FormLabel>
                                <Select
                                    value={formData.transaction_type}
                                    onChange={(e) => setFormData({ ...formData, transaction_type: e.target.value })}
                                >
                                    {TRANSACTION_TYPES.map(t => (
                                        <option key={t.value} value={t.value}>{t.label}</option>
                                    ))}
                                </Select>
                            </FormControl>
                        </HStack>

                        <FormControl isRequired>
                            <FormLabel>COA Account</FormLabel>
                            <Select
                                placeholder="Pilih COA"
                                value={formData.coa_account_id}
                                onChange={(e) => setFormData({ ...formData, coa_account_id: parseInt(e.target.value) })}
                            >
                                {coaAccounts.map(coa => (
                                    <option key={coa.id} value={coa.id}>
                                        {coa.code} - {coa.name}
                                        {coa.budget_category && ` (${coa.budget_category})`}
                                    </option>
                                ))}
                            </Select>
                        </FormControl>

                        <FormControl isRequired>
                            <FormLabel>Deskripsi</FormLabel>
                            <Input
                                placeholder="Contoh: Bensin untuk kendaraan proyek"
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            />
                        </FormControl>

                        <HStack w="full" spacing={4}>
                            <FormControl isRequired flex={2}>
                                <FormLabel>Jumlah (Rp)</FormLabel>
                                <NumberInput
                                    min={0}
                                    value={formData.amount}
                                    onChange={(_, value) => setFormData({ ...formData, amount: value })}
                                >
                                    <NumberInputField />
                                    <NumberInputStepper>
                                        <NumberIncrementStepper />
                                        <NumberDecrementStepper />
                                    </NumberInputStepper>
                                </NumberInput>
                            </FormControl>

                            <FormControl flex={1}>
                                <FormLabel>Quantity</FormLabel>
                                <NumberInput
                                    min={0}
                                    value={formData.quantity}
                                    onChange={(_, value) => setFormData({ ...formData, quantity: value })}
                                >
                                    <NumberInputField />
                                    <NumberInputStepper>
                                        <NumberIncrementStepper />
                                        <NumberDecrementStepper />
                                    </NumberInputStepper>
                                </NumberInput>
                            </FormControl>

                            <FormControl flex={1}>
                                <FormLabel>Unit</FormLabel>
                                <Select
                                    value={formData.unit}
                                    onChange={(e) => setFormData({ ...formData, unit: e.target.value })}
                                >
                                    {UNITS.map(u => (
                                        <option key={u.value} value={u.value}>{u.label}</option>
                                    ))}
                                </Select>
                            </FormControl>
                        </HStack>

                        <FormControl>
                            <FormLabel>No. Referensi</FormLabel>
                            <Input
                                placeholder="PR-001, PO-001, atau referensi lainnya"
                                value={formData.reference_no}
                                onChange={(e) => setFormData({ ...formData, reference_no: e.target.value })}
                            />
                        </FormControl>

                        <FormControl>
                            <FormLabel>Catatan</FormLabel>
                            <Textarea
                                placeholder="Catatan tambahan (opsional)"
                                value={formData.notes}
                                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                                rows={3}
                            />
                        </FormControl>
                    </VStack>
                </ModalBody>
                <ModalFooter>
                    <Button variant="ghost" mr={3} onClick={onClose}>
                        Batal
                    </Button>
                    <Button colorScheme="blue" onClick={handleSubmit} isLoading={loading}>
                        {expense ? 'Update' : 'Simpan'}
                    </Button>
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
}
