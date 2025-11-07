import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  Grid,
  GridItem,
  Divider,
  Switch,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  useToast,
  VStack,
  HStack,
  Text,
  Image,
  AspectRatio,
  IconButton,
  SimpleGrid
} from '@chakra-ui/react';
import { FiSave, FiX, FiUpload, FiTrash2 } from 'react-icons/fi';
import ProductService, { Product, Category, ProductUnit, WarehouseLocation } from '@/services/productService';
import CurrencyInput from '@/components/common/CurrencyInput';
import { getProductImageUrl, debugImageUrl } from '@/utils/imageUrl';

interface ProductFormProps {
  product?: Product;
  onSave: (product: Product) => void;
  onCancel: () => void;
}

const ProductForm: React.FC<ProductFormProps> = ({ product, onSave, onCancel }) => {
  const [formData, setFormData] = useState<Partial<Product>>({
    code: '',
    name: '',
    description: '',
    category_id: undefined,
    warehouse_location_id: undefined,
    brand: '',
    model: '',
    unit: '',
    purchase_price: 0,
    cost_price: 0, // ✅ ADDED: Harga Pokok
    sale_price: 0,
    pricing_tier: '',
    stock: 0,
    min_stock: 0,
    max_stock: 0,
    reorder_level: 0,
    barcode: '',
    sku: '',
    weight: 0,
    dimensions: '',
    is_active: true,
    is_service: false,
    taxable: true,
    image_path: '',
    notes: '',
  });

  const [categories, setCategories] = useState<Category[]>([]);
  const [units, setUnits] = useState<ProductUnit[]>([]);
  const [warehouseLocations, setWarehouseLocations] = useState<WarehouseLocation[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const toast = useToast();

  useEffect(() => {
    fetchCategories();
    fetchUnits();
    fetchWarehouseLocations();
    if (product) {
      setFormData(product);
    }
  }, [product]);

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

  const fetchUnits = async () => {
    try {
      const data = await ProductService.getProductUnits();
      setUnits(data.data);
    } catch (error) {
      toast({
        title: 'Failed to fetch units',
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

  const handleInputChange = (field: keyof Product, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setSelectedImage(file);
      
      // Create preview URL
      const reader = new FileReader();
      reader.onload = (e) => {
        setImagePreview(e.target?.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleImageUpload = async () => {
    if (!selectedImage || !formData.id) {
      toast({
        title: 'Please save the product first before uploading an image',
        status: 'warning',
        isClosable: true,
      });
      return;
    }

    try {
      const response = await ProductService.uploadProductImage(formData.id, selectedImage);
      handleInputChange('image_path', response.path);
      setSelectedImage(null);
      setImagePreview(null);
      
      toast({
        title: 'Image uploaded successfully',
        status: 'success',
        isClosable: true,
      });
    } catch (error: any) {
      console.error('Image upload error:', error);
      let errorMessage = 'Failed to upload image';
      let errorDetail = 'An unknown error occurred';
      
      if (error.response?.data?.error) {
        errorDetail = error.response.data.error;
      } else if (error.message) {
        errorDetail = error.message;
      }
      
      toast({
        title: errorMessage,
        description: errorDetail,
        status: 'error',
        isClosable: true,
        duration: 5000,
      });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      let result;
      if (product?.id) {
        result = await ProductService.updateProduct(product.id, formData);
      } else {
        result = await ProductService.createProduct(formData as Product);
      }
      
      toast({
        title: `Product ${product?.id ? 'updated' : 'created'} successfully`,
        status: 'success',
        isClosable: true,
      });
      
      onSave(result.data);
    } catch (error: any) {
      toast({
        title: `Failed to ${product?.id ? 'update' : 'create'} product`,
        description: error.response?.data?.error || 'An error occurred',
        status: 'error',
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const pricingTiers = [
    'Standard', 'Premium', 'VIP', 'Wholesale', 'Retail'
  ];

  return (
    <Box as="form" onSubmit={handleSubmit}>
      <VStack spacing={6} align="stretch">
        {/* Basic Information */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Informasi Dasar</Text>
          <Grid templateColumns="repeat(4, 1fr)" gap={4}>
            <GridItem colSpan={2}>
              <FormControl isRequired>
                <FormLabel>Kode Produk</FormLabel>
                <Input
                  value={formData.code}
                  onChange={(e) => handleInputChange('code', e.target.value)}
                  placeholder="Masukkan kode produk"
                />
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl isRequired>
                <FormLabel>Nama Produk</FormLabel>
                <Input
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  placeholder="Masukkan nama produk"
                />
              </FormControl>
            </GridItem>
            <GridItem colSpan={4}>
              <FormControl>
                <FormLabel>Deskripsi</FormLabel>
                <Textarea
                  value={formData.description}
                  onChange={(e) => handleInputChange('description', e.target.value)}
                  placeholder="Masukkan deskripsi produk"
                />
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Kategori</FormLabel>
                <Select
                  value={formData.category_id || ''}
                  onChange={(e) => handleInputChange('category_id', e.target.value ? Number(e.target.value) : undefined)}
                  placeholder="Pilih kategori"
                >
                  {categories.map(category => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </Select>
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Lokasi Gudang</FormLabel>
                <Select
                  value={formData.warehouse_location_id || ''}
                  onChange={(e) => handleInputChange('warehouse_location_id', e.target.value ? Number(e.target.value) : undefined)}
                  placeholder="Pilih lokasi gudang"
                >
                  {warehouseLocations.map(location => (
                    <option key={location.id} value={location.id}>
                      {location.name} ({location.code})
                    </option>
                  ))}
                </Select>
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl isRequired>
                <FormLabel>Unit</FormLabel>
                <Select
                  value={formData.unit}
                  onChange={(e) => handleInputChange('unit', e.target.value)}
                  placeholder="Pilih unit"
                >
                  {units.map(unit => (
                    <option key={unit.code} value={unit.code}>
                      {unit.name} ({unit.code})
                    </option>
                  ))}
                </Select>
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        <Divider />

        {/* Product Details */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Detail Produk</Text>
          <Grid templateColumns="repeat(4, 1fr)" gap={4}>
            <GridItem colSpan={2}>
              <FormControl>
                <FormLabel>Merek</FormLabel>
                <Input
                  value={formData.brand}
                  onChange={(e) => handleInputChange('brand', e.target.value)}
                  placeholder="Masukkan merek"
                />
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl>
                <FormLabel>Model</FormLabel>
                <Input
                  value={formData.model}
                  onChange={(e) => handleInputChange('model', e.target.value)}
                  placeholder="Masukkan model"
                />
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl>
                <FormLabel>Barcode</FormLabel>
                <Input
                  value={formData.barcode}
                  onChange={(e) => handleInputChange('barcode', e.target.value)}
                  placeholder="Masukkan barcode"
                />
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl>
                <FormLabel>SKU</FormLabel>
                <Input
                  value={formData.sku}
                  onChange={(e) => handleInputChange('sku', e.target.value)}
                  placeholder="Masukkan SKU"
                />
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Berat (kg)</FormLabel>
                <NumberInput
                  value={formData.weight || 0}
                  onChange={(_, value) => handleInputChange('weight', isNaN(value) ? 0 : value)}
                  min={0}
                  precision={3}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Dimensi</FormLabel>
                <Input
                  value={formData.dimensions}
                  onChange={(e) => handleInputChange('dimensions', e.target.value)}
                  placeholder="P x L x T (cm)"
                />
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        <Divider />

        {/* Pricing */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Harga</Text>
          <Grid templateColumns="repeat(3, 1fr)" gap={4}>
            <GridItem>
              <CurrencyInput
                value={formData.purchase_price || 0}
                onChange={(value) => handleInputChange('purchase_price', value)}
                label="Harga Beli"
                placeholder="Contoh: Rp 50.000"
                isRequired={true}
                size="md"
                min={0}
              />
            </GridItem>
            <GridItem>
              <CurrencyInput
                value={formData.cost_price || 0}
                onChange={(value) => handleInputChange('cost_price', value)}
                label="Harga Pokok (COGS)"
                placeholder="Contoh: Rp 300.000"
                isRequired={true}
                size="md"
                min={0}
              />
            </GridItem>
            <GridItem>
              <CurrencyInput
                value={formData.sale_price || 0}
                onChange={(value) => handleInputChange('sale_price', value)}
                label="Harga Jual"
                placeholder="Contoh: Rp 1.200.000"
                isRequired={true}
                size="md"
                min={0}
              />
            </GridItem>
            <GridItem colSpan={3}>
              <FormControl>
                <FormLabel>Tingkat Harga</FormLabel>
                <Select
                  value={formData.pricing_tier}
                  onChange={(e) => handleInputChange('pricing_tier', e.target.value)}
                  placeholder="Pilih tingkat harga"
                >
                  {pricingTiers.map(tier => (
                    <option key={tier} value={tier}>{tier}</option>
                  ))}
                </Select>
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        <Divider />

        {/* Inventory */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Inventaris</Text>
          <Grid templateColumns="repeat(4, 1fr)" gap={4}>
            <GridItem>
              <FormControl>
                <FormLabel>Stok Saat Ini</FormLabel>
                <NumberInput
                  value={formData.stock || 0}
                  onChange={(_, value) => handleInputChange('stock', isNaN(value) ? 0 : value)}
                  min={0}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Stok Minimum</FormLabel>
                <NumberInput
                  value={formData.min_stock || 0}
                  onChange={(_, value) => handleInputChange('min_stock', isNaN(value) ? 0 : value)}
                  min={0}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Stok Maksimum</FormLabel>
                <NumberInput
                  value={formData.max_stock || 0}
                  onChange={(_, value) => handleInputChange('max_stock', isNaN(value) ? 0 : value)}
                  min={0}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Tingkat Pemesanan Ulang</FormLabel>
                <NumberInput
                  value={formData.reorder_level || 0}
                  onChange={(_, value) => handleInputChange('reorder_level', isNaN(value) ? 0 : value)}
                  min={0}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        <Divider />

        {/* Product Image */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Gambar Produk</Text>
          <Grid templateColumns="repeat(3, 1fr)" gap={4}>
            {/* Current Image */}
            <GridItem>
              <FormControl>
                <FormLabel>Gambar Saat Ini</FormLabel>
                {formData.image_path ? (
                  <Image 
                    src={getProductImageUrl(formData.image_path) || ''} 
                    alt={formData.name || 'Product image'}
                    maxH="150px"
                    maxW="200px"
                    objectFit="contain"
                    borderRadius="md"
                    border="1px"
                    borderColor="gray.200"
                    fallbackSrc="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='150' viewBox='0 0 200 150'%3E%3Crect width='200' height='150' fill='%23f0f0f0'/%3E%3Ctext x='100' y='75' text-anchor='middle' dy='.3em' font-family='Arial, sans-serif' font-size='12' fill='%23999'%3EImage not found%3C/text%3E%3C/svg%3E"
                    onError={(e) => {
                      console.error('Image failed to load:', formData.image_path);
                      console.error('Attempted URL:', getProductImageUrl(formData.image_path));
                      debugImageUrl(formData.image_path);
                    }}
                  />
                ) : (
                  <Box 
                    w="200px" 
                    h="150px" 
                    bg="gray.100" 
                    borderRadius="md" 
                    display="flex" 
                    alignItems="center" 
                    justifyContent="center"
                  >
                    <Text color="gray.500">Tidak ada gambar</Text>
                  </Box>
                )}
              </FormControl>
            </GridItem>
            
            {/* Image Upload */}
            <GridItem>
              <FormControl>
                <FormLabel>Upload Gambar Baru</FormLabel>
                <Input
                  type="file"
                  accept="image/*"
                  onChange={handleImageChange}
                  mb={2}
                />
                {imagePreview && (
                  <Image 
                    src={imagePreview} 
                    alt="Preview"
                    maxH="150px"
                    maxW="200px"
                    objectFit="contain"
                    borderRadius="md"
                    border="1px"
                    borderColor="gray.200"
                  />
                )}
              </FormControl>
            </GridItem>
            
            {/* Upload Button */}
            <GridItem>
              <FormControl>
                <FormLabel>Aksi</FormLabel>
                <Button
                  onClick={handleImageUpload}
                  colorScheme="blue"
                  isDisabled={!selectedImage || !formData.id}
                  size="sm"
                >
                  Upload Gambar
                </Button>
                {!formData.id && (
                  <Text fontSize="xs" color="gray.500" mt={2}>
                    Simpan produk terlebih dahulu untuk mengupload gambar
                  </Text>
                )}
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        <Divider />

        {/* Settings */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Pengaturan</Text>
          <VStack spacing={4} align="stretch">
            <HStack>
              <FormControl display="flex" alignItems="center">
                <FormLabel htmlFor="is_active" mb="0">
                  Aktif
                </FormLabel>
                <Switch
                  id="is_active"
                  isChecked={formData.is_active}
                  onChange={(e) => handleInputChange('is_active', e.target.checked)}
                />
              </FormControl>
              <FormControl display="flex" alignItems="center">
                <FormLabel htmlFor="is_service" mb="0">
                  Produk Jasa
                </FormLabel>
                <Switch
                  id="is_service"
                  isChecked={formData.is_service}
                  onChange={(e) => handleInputChange('is_service', e.target.checked)}
                />
              </FormControl>
              <FormControl display="flex" alignItems="center">
                <FormLabel htmlFor="taxable" mb="0">
                  Dapat Dikenakan Pajak
                </FormLabel>
                <Switch
                  id="taxable"
                  isChecked={formData.taxable}
                  onChange={(e) => handleInputChange('taxable', e.target.checked)}
                />
              </FormControl>
            </HStack>
            <FormControl>
              <FormLabel>Catatan</FormLabel>
              <Textarea
                value={formData.notes}
                onChange={(e) => handleInputChange('notes', e.target.value)}
                placeholder="Catatan tambahan tentang produk"
              />
            </FormControl>
          </VStack>
        </Box>

        {/* Form Actions */}
        <HStack justify="flex-end" spacing={4}>
          <Button
            leftIcon={<FiX />}
            onClick={onCancel}
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            leftIcon={<FiSave />}
            type="submit"
            colorScheme="blue"
            isLoading={isLoading}
            loadingText={product?.id ? 'Updating...' : 'Creating...'}
          >
            {product?.id ? 'Update Product' : 'Create Product'}
          </Button>
        </HStack>
      </VStack>
    </Box>
  );
};

export default ProductForm;
