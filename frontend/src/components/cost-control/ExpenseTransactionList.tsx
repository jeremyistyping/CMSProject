'use client';

import { useState, useEffect } from 'react';
import {
    Box, Button, Table, Thead, Tbody, Tr, Th, Td, Badge, IconButton, HStack, VStack,
    Input, Select, useDisclosure, useToast, Spinner, Text, InputGroup, InputLeftElement,
    Menu, MenuButton, MenuList, MenuItem, useColorModeValue,
} from '@chakra-ui/react';
import { FiPlus, FiEdit2, FiTrash2, FiSearch, FiMoreVertical, FiDownload } from 'react-icons/fi';
import { expenseTransactionService, ExpenseTransaction } from '@/services/expenseTransactionService';
import ExpenseTransactionForm from './ExpenseTransactionForm';

interface ExpenseTransactionListProps {
    projectId: number;
}

const TRANSACTION_TYPE_COLORS: Record<string, string> = {
    LABOUR: 'purple',
    MATERIAL: 'blue',
    OPERATIONAL: 'green',
    OTHER: 'gray',
};

export default function ExpenseTransactionList({ projectId }: ExpenseTransactionListProps) {
    const [expenses, setExpenses] = useState<ExpenseTransaction[]>([]);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [typeFilter, setTypeFilter] = useState('');
    const [startDate, setStartDate] = useState('');
    const [endDate, setEndDate] = useState('');
    const [editingExpense, setEditingExpense] = useState<ExpenseTransaction | null>(null);

    const { isOpen, onOpen, onClose } = useDisclosure();
    const toast = useToast();
    const bgColor = useColorModeValue('white', 'gray.800');

    useEffect(() => {
        fetchExpenses();
    }, [projectId, search, typeFilter, startDate, endDate]);

    const fetchExpenses = async () => {
        setLoading(true);
        try {
            const filter: Record<string, any> = {};
            if (search) filter.search = search;
            if (typeFilter) filter.transaction_type = typeFilter;
            if (startDate) filter.start_date = startDate;
            if (endDate) filter.end_date = endDate;

            const result = await expenseTransactionService.getByProject(projectId, filter);
            setExpenses(result.data || []);
        } catch (error) {
            console.error('Error fetching expenses:', error);
            toast({
                title: 'Gagal memuat data transaksi',
                status: 'error',
                duration: 3000,
            });
        } finally {
            setLoading(false);
        }
    };

    const handleOpenModal = (expense?: ExpenseTransaction) => {
        setEditingExpense(expense || null);
        onOpen();
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Yakin ingin menghapus transaksi ini?')) return;

        try {
            await expenseTransactionService.delete(id);
            toast({
                title: 'Transaksi berhasil dihapus',
                status: 'success',
                duration: 3000,
            });
            fetchExpenses();
        } catch (error: any) {
            toast({
                title: 'Gagal menghapus transaksi',
                description: error.response?.data?.error || error.message,
                status: 'error',
                duration: 3000,
            });
        }
    };

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0,
        }).format(amount);
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('id-ID', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
        });
    };

    const totalAmount = expenses.reduce((sum, exp) => sum + exp.amount, 0);

    return (
        <Box>
            <VStack align="stretch" spacing={4}>
                {/* Filter Bar */}
                <HStack spacing={4} flexWrap="wrap">
                    <InputGroup maxW="300px">
                        <InputLeftElement><FiSearch /></InputLeftElement>
                        <Input
                            placeholder="Cari deskripsi atau referensi..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                    </InputGroup>

                    <Select
                        maxW="200px"
                        value={typeFilter}
                        onChange={(e) => setTypeFilter(e.target.value)}
                        placeholder="Semua Tipe"
                    >
                        <option value="LABOUR">Tenaga Kerja</option>
                        <option value="MATERIAL">Material</option>
                        <option value="OPERATIONAL">Operasional</option>
                        <option value="OTHER">Lainnya</option>
                    </Select>

                    <Input
                        type="date"
                        maxW="180px"
                        placeholder="Dari Tanggal"
                        value={startDate}
                        onChange={(e) => setStartDate(e.target.value)}
                    />

                    <Input
                        type="date"
                        maxW="180px"
                        placeholder="Sampai Tanggal"
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                    />

                    <Button
                        leftIcon={<FiPlus />}
                        colorScheme="blue"
                        onClick={() => handleOpenModal()}
                        ml="auto"
                    >
                        Tambah Transaksi
                    </Button>
                </HStack>

                {/* Summary */}
                <HStack
                    p={4}
                    bg={bgColor}
                    borderRadius="md"
                    borderWidth="1px"
                    justify="space-between"
                >
                    <Text fontWeight="bold">Total Transaksi: {expenses.length}</Text>
                    <Text fontWeight="bold" fontSize="lg" color="blue.600">
                        Total: {formatCurrency(totalAmount)}
                    </Text>
                </HStack>

                {/* Table */}
                {loading ? (
                    <Box textAlign="center" py={10}>
                        <Spinner size="lg" />
                    </Box>
                ) : expenses.length === 0 ? (
                    <Box textAlign="center" py={10}>
                        <Text color="gray.500">Belum ada transaksi</Text>
                    </Box>
                ) : (
                    <Box overflowX="auto" borderWidth="1px" borderRadius="md">
                        <Table size="sm">
                            <Thead>
                                <Tr>
                                    <Th>Tanggal</Th>
                                    <Th>Deskripsi</Th>
                                    <Th>COA</Th>
                                    <Th>Tipe</Th>
                                    <Th isNumeric>Qty</Th>
                                    <Th>Unit</Th>
                                    <Th isNumeric>Jumlah</Th>
                                    <Th>Ref</Th>
                                    <Th>Aksi</Th>
                                </Tr>
                            </Thead>
                            <Tbody>
                                {expenses.map((expense) => (
                                    <Tr key={expense.id}>
                                        <Td>{formatDate(expense.transaction_date)}</Td>
                                        <Td maxW="250px" isTruncated title={expense.description}>
                                            {expense.description}
                                        </Td>
                                        <Td>
                                            <Text fontSize="xs" fontWeight="bold">
                                                {expense.coa_account?.code}
                                            </Text>
                                            <Text fontSize="xs" color="gray.600">
                                                {expense.coa_account?.name}
                                            </Text>
                                        </Td>
                                        <Td>
                                            <Badge colorScheme={TRANSACTION_TYPE_COLORS[expense.transaction_type] || 'gray'}>
                                                {expense.transaction_type}
                                            </Badge>
                                        </Td>
                                        <Td isNumeric>{expense.quantity}</Td>
                                        <Td>{expense.unit}</Td>
                                        <Td isNumeric fontWeight="bold">
                                            {formatCurrency(expense.amount)}
                                        </Td>
                                        <Td fontSize="xs">{expense.reference_no || '-'}</Td>
                                        <Td>
                                            <Menu>
                                                <MenuButton
                                                    as={IconButton}
                                                    icon={<FiMoreVertical />}
                                                    variant="ghost"
                                                    size="sm"
                                                />
                                                <MenuList>
                                                    <MenuItem
                                                        icon={<FiEdit2 />}
                                                        onClick={() => handleOpenModal(expense)}
                                                    >
                                                        Edit
                                                    </MenuItem>
                                                    <MenuItem
                                                        icon={<FiTrash2 />}
                                                        color="red.500"
                                                        onClick={() => handleDelete(expense.id)}
                                                    >
                                                        Hapus
                                                    </MenuItem>
                                                </MenuList>
                                            </Menu>
                                        </Td>
                                    </Tr>
                                ))}
                            </Tbody>
                        </Table>
                    </Box>
                )}
            </VStack>

            {/* Form Modal */}
            <ExpenseTransactionForm
                isOpen={isOpen}
                onClose={onClose}
                projectId={projectId}
                expense={editingExpense}
                onSuccess={fetchExpenses}
            />
        </Box>
    );
}
