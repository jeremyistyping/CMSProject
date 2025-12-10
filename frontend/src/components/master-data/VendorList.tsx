'use client';

import { useState, useEffect, useCallback } from 'react';
import {
    Box, Button, Table, Thead, Tbody, Tr, Th, Td, Badge, IconButton, HStack, VStack,
    Input, Select, useDisclosure, useToast, Spinner, Text, InputGroup, InputLeftElement,
    Modal, ModalOverlay, ModalContent, ModalHeader, ModalBody, ModalFooter, ModalCloseButton,
    FormControl, FormLabel, Textarea, NumberInput, NumberInputField, Stat, StatLabel, StatNumber,
    useColorModeValue, SimpleGrid, Tabs, TabList, TabPanels, Tab, TabPanel
} from '@chakra-ui/react';
import { FiPlus, FiEdit2, FiTrash2, FiSearch, FiStar, FiPhone, FiMail } from 'react-icons/fi';
import { vendorService, Vendor, VendorCategory, VendorSummary } from '@/services/masterDataService';

function VendorList() {
    const [vendors, setVendors] = useState<Vendor[]>([]);
    const [categories, setCategories] = useState<VendorCategory[]>([]);
    const [summary, setSummary] = useState<VendorSummary | null>(null);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [categoryFilter, setCategoryFilter] = useState('');
    const [editingVendor, setEditingVendor] = useState<Vendor | null>(null);
    const [formData, setFormData] = useState({
        code: '', name: '', contact_person: '', email: '', phone: '', address: '',
        city: '', province: '', postal_code: '', npwp: '', bank_name: '', bank_account: '',
        bank_branch: '', payment_terms: 30, category_id: '', notes: ''
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

            const [vendorsRes, categoriesRes, summaryRes] = await Promise.all([
                vendorService.getAll(filter),
                vendorService.getCategories(),
                vendorService.getSummary()
            ]);

            setVendors(vendorsRes.data || []);
            setCategories(categoriesRes.data || []);
            setSummary(summaryRes);
        } catch (error) {
            console.error('Error fetching vendors:', error);
            toast({ title: 'Gagal memuat data vendor', status: 'error', duration: 3000 });
        } finally {
            setLoading(false);
        }
    }, [search, categoryFilter, toast]);

    useEffect(() => { fetchData(); }, [fetchData]);

    const handleOpenModal = (vendor?: Vendor) => {
        if (vendor) {
            setEditingVendor(vendor);
            setFormData({
                code: vendor.code,
                name: vendor.name,
                contact_person: vendor.contact_person || '',
                email: vendor.email || '',
                phone: vendor.phone || '',
                address: vendor.address || '',
                city: vendor.city || '',
                province: vendor.province || '',
                postal_code: vendor.postal_code || '',
                npwp: vendor.npwp || '',
                bank_name: vendor.bank_name || '',
                bank_account: vendor.bank_account || '',
                bank_branch: vendor.bank_branch || '',
                payment_terms: vendor.payment_terms,
                category_id: vendor.category_id?.toString() || '',
                notes: vendor.notes || '',
            });
        } else {
            setEditingVendor(null);
            setFormData({
                code: '', name: '', contact_person: '', email: '', phone: '', address: '',
                city: '', province: '', postal_code: '', npwp: '', bank_name: '', bank_account: '',
                bank_branch: '', payment_terms: 30, category_id: '', notes: ''
            });
        }
        onOpen();
    };

    const handleSubmit = async () => {
        try {
            const data = {
                ...formData,
                category_id: formData.category_id ? parseInt(formData.category_id) : undefined,
            };
            if (editingVendor) {
                await vendorService.update(editingVendor.id, data);
                toast({ title: 'Vendor berhasil diupdate', status: 'success', duration: 3000 });
            } else {
                await vendorService.create(data);
                toast({ title: 'Vendor berhasil ditambahkan', status: 'success', duration: 3000 });
            }
            onClose();
            fetchData();
        } catch (error: any) {
            toast({ title: error.response?.data?.error || 'Gagal menyimpan vendor', status: 'error', duration: 3000 });
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Yakin ingin menghapus vendor ini?')) return;
        try {
            await vendorService.delete(id);
            toast({ title: 'Vendor berhasil dihapus', status: 'success', duration: 3000 });
            fetchData();
        } catch (error: any) {
            toast({ title: error.response?.data?.error || 'Gagal menghapus vendor', status: 'error', duration: 3000 });
        }
    };

    const renderRating = (rating: number) => {
        return (
            <HStack spacing={0}>
                {[1, 2, 3, 4, 5].map(i => (
                    <FiStar key={i} fill={i <= rating ? 'gold' : 'none'} color={i <= rating ? 'gold' : 'gray'} size={14} />
                ))}
            </HStack>
        );
    };

    return (
        <Box>
            {summary && (
                <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4} mb={6}>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Total Vendor</StatLabel>
                        <StatNumber>{summary.total_vendors}</StatNumber>
                    </Stat>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Vendor Aktif</StatLabel>
                        <StatNumber color="green.500">{summary.active_vendors}</StatNumber>
                    </Stat>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Kategori</StatLabel>
                        <StatNumber>{summary.total_categories}</StatNumber>
                    </Stat>
                    <Stat bg={bgColor} p={4} borderRadius="md" borderWidth="1px">
                        <StatLabel>Rating Rata-rata</StatLabel>
                        <StatNumber>{summary.average_rating.toFixed(1)}</StatNumber>
                    </Stat>
                </SimpleGrid>
            )}

            <HStack mb={4} justify="space-between" wrap="wrap" gap={2}>
                <HStack flex={1} minW="300px">
                    <InputGroup maxW="250px">
                        <InputLeftElement><FiSearch /></InputLeftElement>
                        <Input placeholder="Cari vendor..." value={search} onChange={(e) => setSearch(e.target.value)} />
                    </InputGroup>
                    <Select maxW="180px" value={categoryFilter} onChange={(e) => setCategoryFilter(e.target.value)} placeholder="Semua Kategori">
                        {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                    </Select>
                </HStack>
                <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={() => handleOpenModal()}>Tambah Vendor</Button>
            </HStack>

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
                                <Th>Kontak</Th>
                                <Th>Kota</Th>
                                <Th>Rating</Th>
                                <Th>Status</Th>
                                <Th>Aksi</Th>
                            </Tr>
                        </Thead>
                        <Tbody>
                            {vendors.length === 0 ? (
                                <Tr><Td colSpan={8} textAlign="center" py={8} color="gray.500">Tidak ada data vendor</Td></Tr>
                            ) : vendors.map(vendor => (
                                <Tr key={vendor.id}>
                                    <Td fontFamily="mono">{vendor.code}</Td>
                                    <Td>
                                        <Text fontWeight="medium">{vendor.name}</Text>
                                        {vendor.contact_person && <Text fontSize="xs" color="gray.500">{vendor.contact_person}</Text>}
                                    </Td>
                                    <Td>{vendor.category?.name || '-'}</Td>
                                    <Td>
                                        <VStack align="start" spacing={0}>
                                            {vendor.phone && <HStack spacing={1}><FiPhone size={12} /><Text fontSize="xs">{vendor.phone}</Text></HStack>}
                                            {vendor.email && <HStack spacing={1}><FiMail size={12} /><Text fontSize="xs">{vendor.email}</Text></HStack>}
                                        </VStack>
                                    </Td>
                                    <Td>{vendor.city || '-'}</Td>
                                    <Td>{renderRating(vendor.rating)}</Td>
                                    <Td><Badge colorScheme={vendor.is_active ? 'green' : 'gray'}>{vendor.is_active ? 'Aktif' : 'Nonaktif'}</Badge></Td>
                                    <Td>
                                        <HStack spacing={1}>
                                            <IconButton aria-label="Edit" icon={<FiEdit2 />} size="sm" variant="ghost" onClick={() => handleOpenModal(vendor)} />
                                            <IconButton aria-label="Delete" icon={<FiTrash2 />} size="sm" variant="ghost" colorScheme="red" onClick={() => handleDelete(vendor.id)} />
                                        </HStack>
                                    </Td>
                                </Tr>
                            ))}
                        </Tbody>
                    </Table>
                </Box>
            )}

            <Modal isOpen={isOpen} onClose={onClose} size="xl">
                <ModalOverlay />
                <ModalContent maxW="800px">
                    <ModalHeader>{editingVendor ? 'Edit Vendor' : 'Tambah Vendor'}</ModalHeader>
                    <ModalCloseButton />
                    <ModalBody>
                        <Tabs>
                            <TabList>
                                <Tab>Info Dasar</Tab>
                                <Tab>Alamat</Tab>
                                <Tab>Bank & Pajak</Tab>
                            </TabList>
                            <TabPanels>
                                <TabPanel>
                                    <VStack spacing={4}>
                                        <HStack w="full" spacing={4}>
                                            <FormControl isRequired>
                                                <FormLabel>Kode</FormLabel>
                                                <Input value={formData.code} onChange={(e) => setFormData({ ...formData, code: e.target.value })} placeholder="VND-001" />
                                            </FormControl>
                                            <FormControl>
                                                <FormLabel>Kategori</FormLabel>
                                                <Select value={formData.category_id} onChange={(e) => setFormData({ ...formData, category_id: e.target.value })} placeholder="Pilih kategori">
                                                    {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                                                </Select>
                                            </FormControl>
                                        </HStack>
                                        <FormControl isRequired>
                                            <FormLabel>Nama Vendor</FormLabel>
                                            <Input value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} placeholder="PT. Supplier Material" />
                                        </FormControl>
                                        <HStack w="full" spacing={4}>
                                            <FormControl>
                                                <FormLabel>Contact Person</FormLabel>
                                                <Input value={formData.contact_person} onChange={(e) => setFormData({ ...formData, contact_person: e.target.value })} />
                                            </FormControl>
                                            <FormControl>
                                                <FormLabel>Termin Pembayaran (Hari)</FormLabel>
                                                <NumberInput value={formData.payment_terms} onChange={(_, val) => setFormData({ ...formData, payment_terms: val || 30 })} min={0}>
                                                    <NumberInputField />
                                                </NumberInput>
                                            </FormControl>
                                        </HStack>
                                        <HStack w="full" spacing={4}>
                                            <FormControl>
                                                <FormLabel>Telepon</FormLabel>
                                                <Input value={formData.phone} onChange={(e) => setFormData({ ...formData, phone: e.target.value })} />
                                            </FormControl>
                                            <FormControl>
                                                <FormLabel>Email</FormLabel>
                                                <Input type="email" value={formData.email} onChange={(e) => setFormData({ ...formData, email: e.target.value })} />
                                            </FormControl>
                                        </HStack>
                                    </VStack>
                                </TabPanel>
                                <TabPanel>
                                    <VStack spacing={4}>
                                        <FormControl>
                                            <FormLabel>Alamat</FormLabel>
                                            <Textarea value={formData.address} onChange={(e) => setFormData({ ...formData, address: e.target.value })} />
                                        </FormControl>
                                        <HStack w="full" spacing={4}>
                                            <FormControl>
                                                <FormLabel>Kota</FormLabel>
                                                <Input value={formData.city} onChange={(e) => setFormData({ ...formData, city: e.target.value })} />
                                            </FormControl>
                                            <FormControl>
                                                <FormLabel>Provinsi</FormLabel>
                                                <Input value={formData.province} onChange={(e) => setFormData({ ...formData, province: e.target.value })} />
                                            </FormControl>
                                            <FormControl>
                                                <FormLabel>Kode Pos</FormLabel>
                                                <Input value={formData.postal_code} onChange={(e) => setFormData({ ...formData, postal_code: e.target.value })} />
                                            </FormControl>
                                        </HStack>
                                    </VStack>
                                </TabPanel>
                                <TabPanel>
                                    <VStack spacing={4}>
                                        <FormControl>
                                            <FormLabel>NPWP</FormLabel>
                                            <Input value={formData.npwp} onChange={(e) => setFormData({ ...formData, npwp: e.target.value })} placeholder="00.000.000.0-000.000" />
                                        </FormControl>
                                        <HStack w="full" spacing={4}>
                                            <FormControl>
                                                <FormLabel>Nama Bank</FormLabel>
                                                <Input value={formData.bank_name} onChange={(e) => setFormData({ ...formData, bank_name: e.target.value })} />
                                            </FormControl>
                                            <FormControl>
                                                <FormLabel>No. Rekening</FormLabel>
                                                <Input value={formData.bank_account} onChange={(e) => setFormData({ ...formData, bank_account: e.target.value })} />
                                            </FormControl>
                                        </HStack>
                                        <FormControl>
                                            <FormLabel>Cabang Bank</FormLabel>
                                            <Input value={formData.bank_branch} onChange={(e) => setFormData({ ...formData, bank_branch: e.target.value })} />
                                        </FormControl>
                                    </VStack>
                                </TabPanel>
                            </TabPanels>
                        </Tabs>
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

export default VendorList;
