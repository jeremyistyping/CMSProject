'use client';

import { useState, useEffect, useCallback } from 'react';
import {
    Box, Button, Table, Thead, Tbody, Tr, Th, Td, Badge, IconButton, HStack, VStack,
    Input, Select, useDisclosure, useToast, Spinner, Text, InputGroup, InputLeftElement,
    Modal, ModalOverlay, ModalContent, ModalHeader, ModalBody, ModalFooter, ModalCloseButton,
    FormControl, FormLabel, Textarea, Switch, useColorModeValue
} from '@chakra-ui/react';
import { FiPlus, FiEdit2, FiTrash2, FiSearch, FiChevronRight, FiChevronDown } from 'react-icons/fi';
import { coaService, COAAccount, COATreeNode } from '@/services/masterDataService';

const COA_TYPES = [
    { value: 'ASSET', label: 'Asset', color: 'blue' },
    { value: 'LIABILITY', label: 'Liability', color: 'red' },
    { value: 'EQUITY', label: 'Equity', color: 'purple' },
    { value: 'REVENUE', label: 'Revenue', color: 'green' },
    { value: 'EXPENSE', label: 'Expense', color: 'orange' },
];

const COA_CATEGORIES = [
    { value: 'Material', label: 'Material' },
    { value: 'Labor', label: 'Tenaga Kerja' },
    { value: 'Equipment', label: 'Peralatan' },
    { value: 'Overhead', label: 'Overhead' },
    { value: 'Subcontractor', label: 'Subkontraktor' },
    { value: 'Other', label: 'Lainnya' },
];

const BUDGET_CATEGORIES = [
    { value: 'LABOUR_BUDGET', label: 'Labour Budget' },
    { value: 'OPERASIONAL_BUDGET', label: 'Operasional Budget' },
    { value: 'OTHER', label: 'Lainnya' },
];

interface COAFormData {
    code: string;
    name: string;
    description: string;
    type: string;
    category: string;
    budget_category: string;
    work_package: string;
    parent_id?: number;
    is_header: boolean;
}

export default function COAList() {
    const [accounts, setAccounts] = useState<COAAccount[]>([]);
    const [treeData, setTreeData] = useState<COATreeNode[]>([]);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [typeFilter, setTypeFilter] = useState('');
    const [viewMode, setViewMode] = useState<'list' | 'tree'>('list');
    const [expandedNodes, setExpandedNodes] = useState<Set<number>>(new Set());
    const [editingAccount, setEditingAccount] = useState<COAAccount | null>(null);
    const [formData, setFormData] = useState<COAFormData>({
        code: '', name: '', description: '', type: 'EXPENSE', category: '', budget_category: '', work_package: '', is_header: false
    });


    const { isOpen, onOpen, onClose } = useDisclosure();
    const toast = useToast();
    const bgColor = useColorModeValue('white', 'gray.800');

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            if (viewMode === 'tree') {
                const result = await coaService.getTree();
                setTreeData(result.data || []);
            } else {
                const filter: Record<string, any> = {};
                if (search) filter.search = search;
                if (typeFilter) filter.type = typeFilter;
                const result = await coaService.getAll(filter);
                setAccounts(result.data || []);
            }
        } catch (error) {
            console.error('Error fetching COA:', error);
            toast({ title: 'Gagal memuat data COA', status: 'error', duration: 3000 });
        } finally {
            setLoading(false);
        }
    }, [search, typeFilter, viewMode, toast]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleOpenModal = (account?: COAAccount) => {
        if (account) {
            setEditingAccount(account);
            setFormData({
                code: account.code,
                name: account.name,
                description: account.description || '',
                type: account.type,
                category: account.category || '',
                budget_category: account.budget_category || '',
                work_package: account.work_package || '',
                parent_id: account.parent_id,
                is_header: account.is_header,
            });
        } else {
            setEditingAccount(null);
            setFormData({ code: '', name: '', description: '', type: 'EXPENSE', category: '', budget_category: '', work_package: '', is_header: false });
        }
        onOpen();
    };

    const handleSubmit = async () => {
        try {
            if (editingAccount) {
                await coaService.update(editingAccount.id, formData);
                toast({ title: 'COA berhasil diupdate', status: 'success', duration: 3000 });
            } else {
                await coaService.create(formData);
                toast({ title: 'COA berhasil ditambahkan', status: 'success', duration: 3000 });
            }
            onClose();
            fetchData();
        } catch (error: any) {
            toast({ title: error.response?.data?.error || 'Gagal menyimpan COA', status: 'error', duration: 3000 });
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Yakin ingin menghapus akun ini?')) return;
        try {
            await coaService.delete(id);
            toast({ title: 'COA berhasil dihapus', status: 'success', duration: 3000 });
            fetchData();
        } catch (error: any) {
            toast({ title: error.response?.data?.error || 'Gagal menghapus COA', status: 'error', duration: 3000 });
        }
    };

    const toggleNode = (id: number) => {
        const newExpanded = new Set(expandedNodes);
        if (newExpanded.has(id)) {
            newExpanded.delete(id);
        } else {
            newExpanded.add(id);
        }
        setExpandedNodes(newExpanded);
    };

    const renderTreeNode = (node: COATreeNode, level: number = 0) => {
        const hasChildren = node.children && node.children.length > 0;
        const isExpanded = expandedNodes.has(node.id);
        const typeInfo = COA_TYPES.find(t => t.value === node.type);

        return (
            <Box key={node.id}>
                <HStack py={2} pl={level * 6} borderBottomWidth="1px" _hover={{ bg: 'gray.50' }}>
                    <Box w="24px">
                        {hasChildren && (
                            <IconButton
                                aria-label="Toggle"
                                icon={isExpanded ? <FiChevronDown /> : <FiChevronRight />}
                                size="xs"
                                variant="ghost"
                                onClick={() => toggleNode(node.id)}
                            />
                        )}
                    </Box>
                    <Text fontWeight={node.is_header ? 'bold' : 'normal'} flex={1}>
                        {node.code} - {node.name}
                    </Text>
                    <Badge colorScheme={typeInfo?.color || 'gray'}>{node.type}</Badge>
                    {node.budget_category && <Badge colorScheme="purple" variant="outline">{node.budget_category}</Badge>}
                    {node.category && <Badge variant="outline">{node.category}</Badge>}
                    <HStack spacing={1}>
                        <IconButton aria-label="Edit" icon={<FiEdit2 />} size="sm" variant="ghost" onClick={() => handleOpenModal(node as any)} />
                        <IconButton aria-label="Delete" icon={<FiTrash2 />} size="sm" variant="ghost" colorScheme="red" onClick={() => handleDelete(node.id)} />
                    </HStack>
                </HStack>
                {hasChildren && isExpanded && node.children!.map(child => renderTreeNode(child, level + 1))}
            </Box>
        );
    };

    return (
        <Box>
            <HStack mb={4} justify="space-between" wrap="wrap" gap={2}>
                <HStack flex={1} minW="300px">
                    <InputGroup maxW="300px">
                        <InputLeftElement><FiSearch /></InputLeftElement>
                        <Input placeholder="Cari kode atau nama..." value={search} onChange={(e) => setSearch(e.target.value)} />
                    </InputGroup>
                    <Select maxW="150px" value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)} placeholder="Semua Tipe">
                        {COA_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
                    </Select>
                    <Select maxW="120px" value={viewMode} onChange={(e) => setViewMode(e.target.value as any)}>
                        <option value="list">List</option>
                        <option value="tree">Tree</option>
                    </Select>
                </HStack>
                <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={() => handleOpenModal()}>Tambah COA</Button>
            </HStack>

            {loading ? (
                <Box textAlign="center" py={10}><Spinner size="lg" /></Box>
            ) : viewMode === 'tree' ? (
                <Box borderWidth="1px" borderRadius="md" overflow="hidden">
                    {treeData.length === 0 ? (
                        <Text p={4} color="gray.500">Tidak ada data COA</Text>
                    ) : (
                        treeData.map(node => renderTreeNode(node))
                    )}
                </Box>
            ) : (
                <Table size="sm">
                    <Thead>
                        <Tr>
                            <Th>Kode</Th>
                            <Th>Nama</Th>
                            <Th>Tipe</Th>
                            <Th>Budget Cat.</Th>
                            <Th>Kategori</Th>
                            <Th>Status</Th>
                            <Th>Aksi</Th>
                        </Tr>
                    </Thead>
                    <Tbody>
                        {accounts.map(acc => {
                            const typeInfo = COA_TYPES.find(t => t.value === acc.type);
                            return (
                                <Tr key={acc.id}>
                                    <Td fontWeight={acc.is_header ? 'bold' : 'normal'}>{acc.code}</Td>
                                    <Td>{acc.name}</Td>
                                    <Td><Badge colorScheme={typeInfo?.color || 'gray'}>{acc.type}</Badge></Td>
                                    <Td>{acc.budget_category ? <Badge colorScheme="purple" variant="outline">{acc.budget_category}</Badge> : '-'}</Td>
                                    <Td>{acc.category || '-'}</Td>
                                    <Td><Badge colorScheme={acc.is_active ? 'green' : 'gray'}>{acc.is_active ? 'Aktif' : 'Nonaktif'}</Badge></Td>
                                    <Td>
                                        <HStack spacing={1}>
                                            <IconButton aria-label="Edit" icon={<FiEdit2 />} size="sm" variant="ghost" onClick={() => handleOpenModal(acc)} />
                                            <IconButton aria-label="Delete" icon={<FiTrash2 />} size="sm" variant="ghost" colorScheme="red" onClick={() => handleDelete(acc.id)} />
                                        </HStack>
                                    </Td>
                                </Tr>
                            );
                        })}
                    </Tbody>
                </Table>
            )}

            {/* Modal Form */}
            <Modal isOpen={isOpen} onClose={onClose} size="lg">
                <ModalOverlay />
                <ModalContent>
                    <ModalHeader>{editingAccount ? 'Edit COA' : 'Tambah COA'}</ModalHeader>
                    <ModalCloseButton />
                    <ModalBody>
                        <VStack spacing={4}>
                            <HStack w="full" spacing={4}>
                                <FormControl isRequired>
                                    <FormLabel>Kode</FormLabel>
                                    <Input value={formData.code} onChange={(e) => setFormData({ ...formData, code: e.target.value })} placeholder="5101" />
                                </FormControl>
                                <FormControl isRequired>
                                    <FormLabel>Tipe</FormLabel>
                                    <Select value={formData.type} onChange={(e) => setFormData({ ...formData, type: e.target.value })}>
                                        {COA_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
                                    </Select>
                                </FormControl>
                            </HStack>
                            <FormControl isRequired>
                                <FormLabel>Nama</FormLabel>
                                <Input value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} placeholder="Biaya Material Struktural" />
                            </FormControl>
                            <HStack w="full" spacing={4}>
                                <FormControl>
                                    <FormLabel>Kategori</FormLabel>
                                    <Select value={formData.category} onChange={(e) => setFormData({ ...formData, category: e.target.value })} placeholder="Pilih kategori">
                                        {COA_CATEGORIES.map(c => <option key={c.value} value={c.value}>{c.label}</option>)}
                                    </Select>
                                </FormControl>
                                <FormControl>
                                    <FormLabel>Budget Category</FormLabel>
                                    <Select value={formData.budget_category} onChange={(e) => setFormData({ ...formData, budget_category: e.target.value })} placeholder="Pilih budget category">
                                        {BUDGET_CATEGORIES.map(c => <option key={c.value} value={c.value}>{c.label}</option>)}
                                    </Select>
                                </FormControl>
                            </HStack>
                            <FormControl>
                                <FormLabel>Work Package</FormLabel>
                                <Input value={formData.work_package} onChange={(e) => setFormData({ ...formData, work_package: e.target.value })} placeholder="PEKERJAAN PERSIAPAN" />
                            </FormControl>
                            <FormControl>
                                <FormLabel>Deskripsi</FormLabel>
                                <Textarea value={formData.description} onChange={(e) => setFormData({ ...formData, description: e.target.value })} />
                            </FormControl>
                            <FormControl display="flex" alignItems="center">
                                <FormLabel mb={0}>Header Account</FormLabel>
                                <Switch isChecked={formData.is_header} onChange={(e) => setFormData({ ...formData, is_header: e.target.checked })} />
                            </FormControl>
                        </VStack>
                    </ModalBody>
                    <ModalFooter>
                        <Button variant="ghost" mr={3} onClick={onClose}>Batal</Button>
                        <Button colorScheme="blue" onClick={handleSubmit}>Simpan</Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </Box>
    );
}
