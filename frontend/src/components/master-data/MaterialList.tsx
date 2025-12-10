'use client';

import { useState, useEffect, useCallback } from 'react';
import {
    Box, Button, Table, Thead, Tbody, Tr, Th, Td, Badge, IconButton, HStack, VStack,
    Input, Select, useDisclosure, useToast, Spinner, Text, InputGroup, InputLeftElement,
    Modal, ModalOverlay, ModalContent, ModalHeader, ModalBody, ModalFooter, ModalCloseButton,
    FormControl, FormLabel, NumberInput, NumberInputField, Switch, Stat, StatLabel, StatNumber,
    StatGroup, useColorModeValue, SimpleGrid
} from '@chakra-ui/react';
import { FiPlus, FiEdit2, FiTrash2, FiSearch, FiPackage, FiAlertTriangle } from 'react-icons/fi';
import { materialService, Material, MaterialCategory, MaterialSummary, UnitOfMeasure } from '@/services/masterDataService';

export default function MaterialList() {
    const [materials, setMaterials] = useState<Material[]>([]);
    const [categories, setCategories] = useState<MaterialCategory[]>([]);
    const [uoms, setUoms] = useState<UnitOfMeasure[]>([]);
    const [summary, setSummary] = useState<MaterialSummary | null>(null);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [categoryFilter, setCategoryFilter] = useState('');
    const [lowStockOnly, setLowStockOnly] = useState(false);
    const [editingMaterial, setEditingMaterial] = useState<Material | null>(null);
    const [formData, setFormData] = useState({
        code: '', name: '', description: '', category_id: '', unit: '', unit_price: 0, min_stock: 0, max_stock: 0
    });

    const { isOpen, onOpen, onClose } = useDisclosure();
    const toast = useToast();
    const bgColor = useColorModeValue('white', 'gray.800');

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const filter: Record<string, any> = {};
            if (search) filter.search = search;
            if (categoryFilter) filter.category_id = categoryFilter;
            if (lowStockOnly) filter.low_stock = true;

            const [materialsRes, categoriesRes, uomsRes, summaryRes] = await Promise.all([
                materialService.getAll(filter),
                materialService.getCategories(),
                materialService.getUoM(),
                materialService.getSummary()
            ]);

            setMaterials(materialsRes.data || []);
            setCategories(categoriesRes.data || []);
            setUoms(uomsRes.data || []);
            setSummary(summaryRes);
        } catch (error) {
            console.error('Error fetching materials:', error);
            toast({ title: 'Gagal memuat data material', status: 'error', duration: 3000 });
        } finally {
            setLoading(false);
        }
    }, [search, categoryFilter, lowStockOnly, toast]);

    useEffect(() => { fetchData(); }, [fetchData]);


    const handleOpenModal = (material?: Material) => {
        if (material) {
            setEditingMaterial(material);
            setFormData({
                code: material.code,
                name: material.name,
                description: material.description || '',
                category_id: material.category_id?.toString() || '',
                unit: material.unit,
                unit_price: material.unit_price,
                min_stock: material.min_stock,
                max_stock: material.max_stock,
            });
        } else {
            setEditingMaterial(null);
            setFormData({ code: '', name: '', description: '', category_id: '', unit: '', unit_price: 0, min_stock: 0, max_stock: 0 });
        }
        onOpen();
    };

    const handleSubmit = async () => {
        try {
            const data = {
                ...formData,
                category_id: formData.category_id ? parseInt(formData.category_id) : undefined,
            };
            if (editingMaterial) {
                await materialService.update(editingMaterial.id, data);
                toast({ title: 'Material berhasil diupdate', status: 'success', duration: 3000 });
            } else {
                await materialService.create(data);
                toast({ title: 'Material berhasil ditambahkan', status: 'success', duration: 3000 });
            }
            onClose();
            fetchData();
        } catch (error: any) {
            toast({ title: error.response?.data?.error || 'Gagal menyimpan material', status: 'error', duration: 3000 });
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Yakin ingin menghapus material ini?')) return;
        try {
            await materialService.delete(id);
            toast({ title: 'Material berhasil dihapus', status: 'success', duration: 3000 });
            fetchData();
        } catch (error: any) {
            toast({ title: error.response?.data?.error || 'Gagal menghapus material', status: 'error', duration: 3000 });
        }
    };

    const formatCurrency = (value: number) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value);

    return (
        <Box>
            {/* Summary Stats */}
            {summary && (
                <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Total Material</StatLabel>
                        <StatNumber>{summary.total_materials}</StatNumber>
                    </Stat>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Material Aktif</StatLabel>
                        <StatNumber color="green.500">{summary.active_materials}</StatNumber>
                    </Stat>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Stok Rendah</StatLabel>
                        <StatNumber color="orange.500">{summary.low_stock_count}</StatNumber>
                    </Stat>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Nilai Stok</StatLabel>
                        <StatNumber fontSize="md">{formatCurrency(summary.total_stock_value)}</StatNumber>
                    </Stat>
                </SimpleGrid>
            )}

            {/* Filters */}
            <HStack mb={4} justify="space-between" wrap="wrap" gap={2}>
                <HStack flex={1} minW="300px" wrap="wrap" gap={2}>
                    <InputGroup maxW="250px">
                        <InputLeftElement><FiSearch /></InputLeftElement>
                        <Input placeholder="Cari material..." value={search} onChange={(e) => setSearch(e.target.value)} />
                    </InputGroup>
                    <Select maxW="180px" value={categoryFilter} onChange={(e) => setCategoryFilter(e.target.value)} placeholder="Semua Kategori">
                        {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                    </Select>
                    <HStack>
                        <Switch isChecked={lowStockOnly} onChange={(e) => setLowStockOnly(e.target.checked)} />
                        <Text fontSize="sm">Stok Rendah</Text>
                    </HStack>
                </HStack>
                <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={() => handleOpenModal()}>Tambah Material</Button>
            </HStack>

            {/* Table */}
            {loading ? (
                <Box textAlign="center" py={10}><Spinner size="lg" /></Box>
            ) : (
                <Box overflowX="auto">
                    <Table size="sm">
                        <Thead>
                            <Tr>
                                <Th>Kode</Th>
                                <Th>Nama</Th>
                                <Th>Kategori</Th>
                                <Th>Satuan</Th>
                                <Th isNumeric>Harga</Th>
                                <Th isNumeric>Stok</Th>
                                <Th>Status</Th>
                                <Th>Aksi</Th>
                            </Tr>
                        </Thead>
                        <Tbody>
                            {materials.length === 0 ? (
                                <Tr><Td colSpan={8} textAlign="center" py={8} color="gray.500">Tidak ada data material</Td></Tr>
                            ) : materials.map(mat => {
                                const isLowStock = mat.current_stock <= mat.min_stock;
                                return (
                                    <Tr key={mat.id}>
                                        <Td fontFamily="mono">{mat.code}</Td>
                                        <Td>{mat.name}</Td>
                                        <Td>{mat.category?.name || '-'}</Td>
                                        <Td>{mat.unit}</Td>
                                        <Td isNumeric>{formatCurrency(mat.unit_price)}</Td>
                                        <Td isNumeric>
                                            <HStack justify="flex-end">
                                                {isLowStock && <FiAlertTriangle color="orange" />}
                                                <Text color={isLowStock ? 'orange.500' : undefined}>{mat.current_stock}</Text>
                                            </HStack>
                                        </Td>
                                        <Td><Badge colorScheme={mat.is_active ? 'green' : 'gray'}>{mat.is_active ? 'Aktif' : 'Nonaktif'}</Badge></Td>
                                        <Td>
                                            <HStack spacing={1}>
                                                <IconButton aria-label="Edit" icon={<FiEdit2 />} size="sm" variant="ghost" onClick={() => handleOpenModal(mat)} />
                                                <IconButton aria-label="Delete" icon={<FiTrash2 />} size="sm" variant="ghost" colorScheme="red" onClick={() => handleDelete(mat.id)} />
                                            </HStack>
                                        </Td>
                                    </Tr>
                                );
                            })}
                        </Tbody>
                    </Table>
                </Box>
            )}

            {/* Modal Form */}
            <Modal isOpen={isOpen} onClose={onClose} size="xl">
                <ModalOverlay />
                <ModalContent>
                    <ModalHeader>{editingMaterial ? 'Edit Material' : 'Tambah Material'}</ModalHeader>
                    <ModalCloseButton />
                    <ModalBody>
                        <VStack spacing={4}>
                            <HStack w="full" spacing={4}>
                                <FormControl isRequired>
                                    <FormLabel>Kode</FormLabel>
                                    <Input value={formData.code} onChange={(e) => setFormData({ ...formData, code: e.target.value })} placeholder="MAT-001" />
                                </FormControl>
                                <FormControl isRequired>
                                    <FormLabel>Satuan</FormLabel>
                                    <Select value={formData.unit} onChange={(e) => setFormData({ ...formData, unit: e.target.value })} placeholder="Pilih satuan">
                                        {uoms.map(u => <option key={u.id} value={u.code}>{u.name} ({u.symbol || u.code})</option>)}
                                    </Select>
                                </FormControl>
                            </HStack>
                            <FormControl isRequired>
                                <FormLabel>Nama Material</FormLabel>
                                <Input value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} placeholder="Besi Beton D10" />
                            </FormControl>
                            <HStack w="full" spacing={4}>
                                <FormControl>
                                    <FormLabel>Kategori</FormLabel>
                                    <Select value={formData.category_id} onChange={(e) => setFormData({ ...formData, category_id: e.target.value })} placeholder="Pilih kategori">
                                        {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                                    </Select>
                                </FormControl>
                                <FormControl>
                                    <FormLabel>Harga Satuan</FormLabel>
                                    <NumberInput value={formData.unit_price} onChange={(_, val) => setFormData({ ...formData, unit_price: val || 0 })} min={0}>
                                        <NumberInputField />
                                    </NumberInput>
                                </FormControl>
                            </HStack>
                            <HStack w="full" spacing={4}>
                                <FormControl>
                                    <FormLabel>Stok Minimum</FormLabel>
                                    <NumberInput value={formData.min_stock} onChange={(_, val) => setFormData({ ...formData, min_stock: val || 0 })} min={0}>
                                        <NumberInputField />
                                    </NumberInput>
                                </FormControl>
                                <FormControl>
                                    <FormLabel>Stok Maksimum</FormLabel>
                                    <NumberInput value={formData.max_stock} onChange={(_, val) => setFormData({ ...formData, max_stock: val || 0 })} min={0}>
                                        <NumberInputField />
                                    </NumberInput>
                                </FormControl>
                            </HStack>
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
