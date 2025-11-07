import React, { useState, useEffect, useMemo } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useAuth } from '@/contexts/AuthContext';
import { useModulePermissions } from '@/hooks/usePermissions';
import { useTranslation } from '@/hooks/useTranslation';
import {
  Box,
  Button,
  Heading,
  Input,
  InputGroup,
  InputLeftElement,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  Flex,
  Select,
  useToast,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Image,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  useDisclosure,
  Text,
  Grid,
  HStack,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Icon
} from '@chakra-ui/react';
import { FiSearch, FiEdit, FiTrash2, FiUpload, FiEye, FiPlus, FiGrid, FiPackage, FiMapPin, FiSettings, FiChevronDown } from 'react-icons/fi';
import ProductService, { Product, Category, WarehouseLocation } from '@/services/productService';
import ProductForm from './ProductForm';
import CategoryManagement from './CategoryManagement';
import UnitManagement from './UnitManagement';
import WarehouseLocationManagement from './WarehouseLocationManagement';
import { ProductUnit } from './UnitForm';
import { formatIDR, formatCurrencyDetailed } from '@/utils/currency';
import { getProductImageUrl, debugImageUrl } from '@/utils/imageUrl';

const ProductCatalog: React.FC = () => {
  const { user } = useAuth();
  const { t } = useTranslation();
  const { 
    canView, 
    canCreate, 
    canEdit, 
    canDelete, 
    canExport 
  } = useModulePermissions('products');
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [warehouseLocations, setWarehouseLocations] = useState<WarehouseLocation[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [warehouseLocationFilter, setWarehouseLocationFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [sortBy, setSortBy] = useState('name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [pendingUpload, setPendingUpload] = useState<{productId: number, file: File} | null>(null);
  const { isOpen: isAlertOpen, onOpen: onAlertOpen, onClose: onAlertClose } = useDisclosure();
  const { isOpen: isDetailOpen, onOpen: onDetailOpen, onClose: onDetailClose } = useDisclosure();
  const { isOpen: isManagementModalOpen, onOpen: onManagementModalOpen, onClose: onManagementModalClose } = useDisclosure();
  const [detailProduct, setDetailProduct] = useState<Product | null>(null);
  const toast = useToast();

  // Tooltip descriptions for product page
  const tooltips = {
    search: 'Cari produk berdasarkan nama, kode, atau deskripsi',
    category: 'Kategori produk untuk pengelompokan dan pelaporan',
    unit: 'Satuan unit produk (contoh: Pcs, Kg, Liter, Box)',
    warehouse: 'Lokasi gudang/warehouse tempat produk disimpan',
    stock: 'Jumlah stok tersedia saat ini',
    minStock: 'Stok minimum sebagai peringatan untuk reorder',
    costPrice: 'Harga pokok/cost produk (untuk perhitungan COGS)',
    salePrice: 'Harga jual standar kepada customer',
    productCode: 'Kode unik produk (SKU atau Product Code)',
    barcode: 'Barcode produk untuk scanning',
    description: 'Deskripsi detail produk',
    isActive: 'Status produk: Active (dijual) atau Inactive (tidak dijual)',
    trackInventory: 'Aktifkan tracking inventory untuk produk ini',
    revenueAccount: 'Akun pendapatan di chart of accounts untuk produk ini',
    expenseAccount: 'Akun biaya/expense untuk pembelian produk ini',
  };

  useEffect(() => {
    fetchProducts();
    fetchCategories();
    fetchWarehouseLocations();
  }, []);

  useEffect(() => {
    fetchProducts();
  }, [searchTerm, categoryFilter, warehouseLocationFilter, statusFilter]);

  const fetchProducts = async () => {
    try {
      const params: any = {};
      if (searchTerm) params.search = searchTerm;
      if (categoryFilter) params.category = categoryFilter;
      
      const data = await ProductService.getProducts(params);
      setProducts(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch products',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const fetchCategories = async () => {
    try {
      const data = await ProductService.getCategories();
      setCategories(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch categories',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const fetchWarehouseLocations = async () => {
    try {
      const data = await ProductService.getWarehouseLocations();
      setWarehouseLocations(data.data);
      
      // Show info message if using mock data
      if (data.message && data.message.includes('mock')) {
        console.info('Using mock warehouse locations data - implement backend API for full functionality');
      }
    } catch (error) {
      console.error('Failed to fetch warehouse locations:', error);
      // Set empty array instead of showing error to user
      setWarehouseLocations([]);
    }
  };

const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleSearch = () => {
    fetchProducts();
  };

  const handleAddProductClick = () => {
    if (!canCreate) return;
    setSelectedProduct(null);
    setIsModalOpen(true);
  };

  const handleSaveProduct = (product: Product) => {
    if (selectedProduct) {
      // Update existing product in list
      setProducts(prevProducts => 
        prevProducts.map(p => p.id === product.id ? product : p)
      );
    } else {
      // Add new product to list
      setProducts(prevProducts => [...prevProducts, product]);
    }
    setIsModalOpen(false);
    setSelectedProduct(null);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedProduct(null);
  };

  const handleEditProduct = (product: Product) => {
    setSelectedProduct(product);
    setIsModalOpen(true);
  };

  const handleViewDetails = (product: Product) => {
    setDetailProduct(product);
    onDetailOpen();
  };

  const handleDeleteProduct = async (product: Product) => {
    if (!window.confirm(`Are you sure you want to delete "${product.name}"?`)) {
      return;
    }

    try {
      await ProductService.deleteProduct(product.id!);
      setProducts(prevProducts => prevProducts.filter(p => p.id !== product.id));
      toast({
        title: 'Product deleted successfully',
        status: 'success',
        isClosable: true,
      });
    } catch (error) {
      toast({
        title: 'Failed to delete product',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>, productId: number) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      const product = products.find(p => p.id === productId);
      
      if (product && product.image_path) {
        // Product already has an image, show confirmation
        setPendingUpload({ productId, file });
        onAlertOpen();
      } else {
        // No existing image, upload directly
        handleUpload(productId, file);
      }
    }
  };

  const confirmImageUpdate = () => {
    if (pendingUpload) {
      handleUpload(pendingUpload.productId, pendingUpload.file);
      setPendingUpload(null);
    }
    onAlertClose();
  };

  const handleUpload = async (productId: number, file: File) => {
    try {
      const response = await ProductService.uploadProductImage(productId, file);
      
      // Update the product in the list with the new image path
      setProducts(prevProducts => 
        prevProducts.map(p => 
          p.id === productId 
            ? { ...p, image_path: response.path }
            : p
        )
      );
      
      toast({
        title: 'Image uploaded successfully',
        status: 'success',
        isClosable: true,
      });
      
      // Reset file input
      const fileInput = document.getElementById(`file-upload-${productId}`) as HTMLInputElement;
      if (fileInput) {
        fileInput.value = '';
      }
    } catch (error) {
      toast({
        title: 'Failed to upload image',
        status: 'error',
        isClosable: true,
      });
    }
  };

  // Management handlers
  const handleOpenManagement = () => {
    if (!canCreate) return;
    onManagementModalOpen();
  };

  // Filtered and sorted products using useMemo for performance

  const filteredAndSortedProducts = useMemo(() => {
    return products
      .filter(product => {
        const matchesSearch = searchTerm ? 
          product.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
          product.code.toLowerCase().includes(searchTerm.toLowerCase()) : true;
        const matchesCategory = categoryFilter ? 
          product.category?.id === Number(categoryFilter) : true;
        const matchesWarehouseLocation = warehouseLocationFilter ? 
          product.warehouse_location?.id === Number(warehouseLocationFilter) : true;
        const matchesStatus = statusFilter ? 
          (statusFilter === 'active' ? product.is_active : !product.is_active) : true;
        return matchesSearch && matchesCategory && matchesWarehouseLocation && matchesStatus;
      })
      .sort((a, b) => {
        let comparison = 0;
        if (sortBy === 'name' || sortBy === 'code' || sortBy === 'category') {
          const aValue = sortBy === 'category' ? a.category?.name || '' : a[sortBy as keyof Product] as string;
          const bValue = sortBy === 'category' ? b.category?.name || '' : b[sortBy as keyof Product] as string;
          comparison = (aValue < bValue ? -1 : (aValue > bValue ? 1 : 0)) * (sortOrder === 'asc' ? 1 : -1);
        } else if (sortBy === 'stock' || sortBy === 'sale_price') {
          const aValue = a[sortBy as keyof Product] as number;
          const bValue = b[sortBy as keyof Product] as number;
          comparison = (aValue < bValue ? -1 : (aValue > bValue ? 1 : 0)) * (sortOrder === 'asc' ? 1 : -1);
        }
        return comparison;
      });
  }, [products, searchTerm, categoryFilter, warehouseLocationFilter, statusFilter, sortBy, sortOrder]);

  return (
    <SimpleLayout allowedRoles={['admin', 'inventory_manager', 'employee', 'finance', 'director']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Box>
            <Heading as="h1" size="xl" mb={2}>{t('products.productCatalog')}</Heading>
          </Box>
          
          {/* Management Buttons */}
          <HStack spacing={3}>
            {canCreate && (
              <>
                <Button 
                  leftIcon={<FiSettings />} 
                  rightIcon={<FiChevronDown />}
                  colorScheme="teal" 
                  size="lg" 
                  onClick={handleOpenManagement}
                  variant="outline"
                >
                  Manage Categories, Units & Locations
                </Button>
                <Button 
                  leftIcon={<FiPlus />} 
                  colorScheme="brand" 
                  size="lg" 
                  onClick={handleAddProductClick}
                >
                  {t('products.addProduct')}
                </Button>
              </>
            )}
          </HStack>
        </Flex>

        {/* Search and Filters */}
        <Box mb={6}>
          <Flex gap={4} mb={4} flexWrap="wrap">
            {/* Search */}
            <InputGroup maxW="400px">
              <InputLeftElement pointerEvents="none">
                <FiSearch color="gray.300" />
              </InputLeftElement>
              <Input
                placeholder={t('products.searchProducts')}
                value={searchTerm}
                onChange={handleSearchChange}
              />
            </InputGroup>
            
            {/* Category Filter */}
            <Select
              placeholder="All Categories"
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
              maxW="200px"
            >
              {categories.map(category => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </Select>

            {/* Warehouse Location Filter */}
            <Select
              placeholder="All Locations"
              value={warehouseLocationFilter}
              onChange={(e) => setWarehouseLocationFilter(e.target.value)}
              maxW="200px"
            >
              {warehouseLocations.map(location => (
                <option key={location.id} value={location.id}>
                  {location.name}
                </option>
              ))}
            </Select>
            
            {/* Status Filter */}
            <Select
              placeholder="All Status"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              maxW="150px"
            >
              <option value="active">{t('products.active')}</option>
              <option value="inactive">{t('products.inactive')}</option>
            </Select>
            
            {/* Sort Options */}
            <Select
              value={`${sortBy}-${sortOrder}`}
              onChange={(e) => {
                const [field, order] = e.target.value.split('-');
                setSortBy(field);
                setSortOrder(order as 'asc' | 'desc');
              }}
              maxW="180px"
            >
              <option value="name-asc">Name A-Z</option>
              <option value="name-desc">Name Z-A</option>
              <option value="code-asc">Code A-Z</option>
              <option value="code-desc">Code Z-A</option>
              <option value="category-asc">Category A-Z</option>
              <option value="stock-desc">Stock High-Low</option>
              <option value="stock-asc">Stock Low-High</option>
              <option value="sale_price-desc">Price High-Low</option>
              <option value="sale_price-asc">Price Low-High</option>
            </Select>
            
            {/* Clear Filters */}
            <Button
              onClick={() => {
                setSearchTerm('');
                setCategoryFilter('');
                setWarehouseLocationFilter('');
                setStatusFilter('');
                setSortBy('name');
                setSortOrder('asc');
              }}
              variant="outline"
              size="md"
            >
              Clear Filters
            </Button>
          </Flex>
          
          {/* Results Summary */}
          <Text fontSize="sm" color="gray.600">
Showing {filteredAndSortedProducts.length} product{filteredAndSortedProducts.length !== 1 ? 's' : ''}
{(searchTerm || categoryFilter || warehouseLocationFilter || statusFilter) ? ' (filtered)' : ''}
          </Text>
        </Box>

        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Product ID</Th>
              <Th>Name</Th>
              <Th>Category</Th>
              <Th>Warehouse Location</Th>
              <Th>Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {filteredAndSortedProducts.map((product) => (
              <Tr key={product.id}>
                <Td>{product.code}</Td>
                <Td>{product.name}</Td>
                <Td>{product.category?.name}</Td>
                <Td>{product.warehouse_location?.name || 'No Location'}</Td>
                <Td>
                  <Button 
                    size="sm" 
                    variant="ghost" 
                    leftIcon={<FiEye />} 
                    mr={2}
                    onClick={() => handleViewDetails(product)}
                  >
                    View
                  </Button>
                  {canEdit && (
                    <>
                      <Button 
                        size="sm" 
                        variant="ghost" 
                        leftIcon={<FiEdit />} 
                        mr={2}
                        onClick={() => handleEditProduct(product)}
                      >
                        Edit
                      </Button>
                      <Button 
                        size="sm" 
                        variant="ghost" 
                        colorScheme="red" 
                        leftIcon={<FiTrash2 />} 
                        mr={2}
                        onClick={() => handleDeleteProduct(product)}
                      >
                        Delete
                      </Button>
                      <Input
                        type="file"
                        accept="image/*"
                        onChange={(e) => handleFileChange(e, product.id!)}
                        style={{ display: 'none' }}
                        id={`file-upload-${product.id}`}
                      />
                      <Button
                        size="sm"
                        variant="ghost"
                        leftIcon={<FiUpload />}
                        as="label"
                        htmlFor={`file-upload-${product.id}`}
                        cursor="pointer"
                      >
                        {product.image_path ? 'Update Image' : 'Upload Image'}
                      </Button>
                    </>
                  )}
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>

        {/* Add/Edit Product Modal */}
        {canEdit && (
          <Modal isOpen={isModalOpen || !!selectedProduct} onClose={handleCloseModal} size="6xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              {selectedProduct ? "Edit Product" : "Add Product"}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              <ProductForm 
                product={selectedProduct || undefined} 
                onSave={handleSaveProduct} 
                onCancel={handleCloseModal} 
              />
            </ModalBody>
          </ModalContent>
          </Modal>
        )}
        
        {/* Product Details Modal */}
        <Modal isOpen={isDetailOpen} onClose={onDetailClose} size="4xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              Product Details - {detailProduct?.name}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {detailProduct && (
                <Box>
                  {/* Product Image */}
                  <Flex justify="center" mb={6}>
                    {detailProduct.image_path ? (
                      <Image 
                        src={getProductImageUrl(detailProduct.image_path) || ''} 
                        alt={detailProduct.name}
                        maxH="250px"
                        maxW="350px"
                        objectFit="contain"
                        borderRadius="lg"
                        border="2px"
                        borderColor="gray.300"
                        boxShadow="md"
                        bg="white"
                        p={2}
                        fallbackSrc="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 100 100'%3E%3Crect width='100' height='100' fill='%23f0f0f0'/%3E%3Ctext x='50' y='50' text-anchor='middle' dy='.3em' font-family='Arial, sans-serif' font-size='14' fill='%23999'%3ENo Image%3C/text%3E%3C/svg%3E"
                        onError={(e) => {
                          console.error('Image failed to load:', detailProduct.image_path);
                          console.error('Attempted URL:', getProductImageUrl(detailProduct.image_path));
                          debugImageUrl(detailProduct.image_path);
                        }}
                      />
                    ) : (
                      <Box 
                        w="350px" 
                        h="250px" 
                        bg="gray.50" 
                        borderRadius="lg" 
                        border="2px"
                        borderColor="gray.200"
                        display="flex" 
                        alignItems="center" 
                        justifyContent="center"
                        boxShadow="sm"
                      >
                        <Text color="gray.400" fontSize="lg">Tidak ada gambar produk</Text>
                      </Box>
                    )}
                  </Flex>

                  {/* Basic Information */}
                  <Box mb={4}>
                    <Text fontSize="lg" fontWeight="bold" mb={2} color="blue.600">Basic Information</Text>
                    <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Product Code:</Text>
                        <Text fontSize="md">{detailProduct.code}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Product Name:</Text>
                        <Text fontSize="md">{detailProduct.name}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Category:</Text>
                        <Text fontSize="md">{detailProduct.category?.name || 'No Category'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Unit:</Text>
                        <Text fontSize="md">{detailProduct.unit}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Warehouse Location:</Text>
                        <Text fontSize="md">{detailProduct.warehouse_location?.name || 'No Location Assigned'}</Text>
                      </Box>
                    </Grid>
                    {detailProduct.description && (
                      <Box mt={3}>
                        <Text fontWeight="semibold" color="gray.600">Description:</Text>
                        <Text fontSize="md">{detailProduct.description}</Text>
                      </Box>
                    )}
                  </Box>

                  {/* Product Details */}
                  <Box mb={4}>
                    <Text fontSize="lg" fontWeight="bold" mb={2} color="blue.600">Product Details</Text>
                    <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Brand:</Text>
                        <Text fontSize="md">{detailProduct.brand || 'Not specified'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Model:</Text>
                        <Text fontSize="md">{detailProduct.model || 'Not specified'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">SKU:</Text>
                        <Text fontSize="md">{detailProduct.sku || 'Not specified'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Barcode:</Text>
                        <Text fontSize="md">{detailProduct.barcode || 'Not specified'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Weight:</Text>
                        <Text fontSize="md">{detailProduct.weight} kg</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Dimensions:</Text>
                        <Text fontSize="md">{detailProduct.dimensions || 'Not specified'}</Text>
                      </Box>
                    </Grid>
                  </Box>

                  {/* Pricing */}
                  <Box mb={4}>
                    <Text fontSize="lg" fontWeight="bold" mb={2} color="blue.600">Pricing</Text>
                    <Grid templateColumns="repeat(3, 1fr)" gap={4}>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Purchase Price:</Text>
                        <Text fontSize="md" color="green.600" fontWeight="bold">
                          {formatCurrencyDetailed(detailProduct.purchase_price || 0)}
                        </Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Sale Price:</Text>
                        <Text fontSize="md" color="blue.600" fontWeight="bold">
                          {formatCurrencyDetailed(detailProduct.sale_price || 0)}
                        </Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Pricing Tier:</Text>
                        <Text fontSize="md">{detailProduct.pricing_tier || 'Standard'}</Text>
                      </Box>
                    </Grid>
                  </Box>

                  {/* Inventory */}
                  <Box mb={4}>
                    <Text fontSize="lg" fontWeight="bold" mb={2} color="blue.600">Inventory</Text>
                    <Grid templateColumns="repeat(4, 1fr)" gap={4}>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Current Stock:</Text>
                        <Text fontSize="md" fontWeight="bold">{detailProduct.stock}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Min Stock:</Text>
                        <Text fontSize="md">{detailProduct.min_stock}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Max Stock:</Text>
                        <Text fontSize="md">{detailProduct.max_stock}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Reorder Level:</Text>
                        <Text fontSize="md">{detailProduct.reorder_level}</Text>
                      </Box>
                    </Grid>
                  </Box>

                  {/* Settings */}
                  <Box mb={4}>
                    <Text fontSize="lg" fontWeight="bold" mb={2} color="blue.600">Settings</Text>
                    <Grid templateColumns="repeat(3, 1fr)" gap={4}>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Status:</Text>
                        <Text fontSize="md" color={detailProduct.is_active ? 'green.600' : 'red.600'}>
                          {detailProduct.is_active ? 'Active' : 'Inactive'}
                        </Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Service Product:</Text>
                        <Text fontSize="md">{detailProduct.is_service ? 'Yes' : 'No'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="semibold" color="gray.600">Taxable:</Text>
                        <Text fontSize="md">{detailProduct.taxable ? 'Yes' : 'No'}</Text>
                      </Box>
                    </Grid>
                  </Box>

                  {/* Notes */}
                  {detailProduct.notes && (
                    <Box>
                      <Text fontSize="lg" fontWeight="bold" mb={2} color="blue.600">Notes</Text>
                      <Text fontSize="md" p={3} bg="gray.50" borderRadius="md">
                        {detailProduct.notes}
                      </Text>
                    </Box>
                  )}
                </Box>
              )}
            </ModalBody>
          </ModalContent>
        </Modal>

        {/* Management Modal with Tabs */}
        {canCreate && (
          <Modal isOpen={isManagementModalOpen} onClose={onManagementModalClose} size="6xl">
            <ModalOverlay />
            <ModalContent maxH="90vh">
              <ModalHeader>
                Manage Categories, Units & Warehouse Locations
              </ModalHeader>
              <ModalCloseButton />
              <ModalBody pb={6} overflowY="auto">
                <Tabs colorScheme="teal" variant="enclosed">
                  <TabList>
                    <Tab>
                      <Icon as={FiGrid} mr={2} />
                      Categories
                    </Tab>
                    <Tab>
                      <Icon as={FiPackage} mr={2} />
                      Units
                    </Tab>
                    <Tab>
                      <Icon as={FiMapPin} mr={2} />
                      Warehouse Locations
                    </Tab>
                  </TabList>

                  <TabPanels>
                    <TabPanel>
                      <CategoryManagement />
                    </TabPanel>
                    <TabPanel>
                      <UnitManagement />
                    </TabPanel>
                    <TabPanel>
                      <WarehouseLocationManagement />
                    </TabPanel>
                  </TabPanels>
                </Tabs>
              </ModalBody>
            </ModalContent>
          </Modal>
        )}

        {/* Image Update Confirmation Dialog */}
        <AlertDialog
          isOpen={isAlertOpen}
          leastDestructiveRef={React.useRef(null)}
          onClose={onAlertClose}
        >
          <AlertDialogOverlay>
            <AlertDialogContent>
              <AlertDialogHeader fontSize="lg" fontWeight="bold">
                Update Product Image
              </AlertDialogHeader>

              <AlertDialogBody>
                This product already has an image. Are you sure you want to replace it with the new image?
              </AlertDialogBody>

              <AlertDialogFooter>
                <Button onClick={onAlertClose}>
                  Cancel
                </Button>
                <Button colorScheme="blue" onClick={confirmImageUpdate} ml={3}>
                  Update Image
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialogOverlay>
        </AlertDialog>
      </Box>
    </SimpleLayout>
  );
};

export default ProductCatalog;
