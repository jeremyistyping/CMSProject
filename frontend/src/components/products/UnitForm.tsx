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
  Switch,
  useToast,
  VStack,
  HStack,
  Text,
} from '@chakra-ui/react';
import { FiSave, FiX } from 'react-icons/fi';
import ProductService, { ProductUnit } from '@/services/productService';

interface UnitFormProps {
  unit?: ProductUnit;
  onSave: (unit: ProductUnit) => void;
  onCancel: () => void;
}

const UnitForm: React.FC<UnitFormProps> = ({ unit, onSave, onCancel }) => {
  const [formData, setFormData] = useState<Partial<ProductUnit>>({
    code: '',
    name: '',
    symbol: '',
    type: '',
    description: '',
    is_active: true,
  });

  const [isLoading, setIsLoading] = useState(false);
  const toast = useToast();

  useEffect(() => {
    if (unit) {
      setFormData(unit);
    }
  }, [unit]);

  const handleInputChange = (field: keyof ProductUnit, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      // Validate required fields
      if (!formData.code || !formData.code.trim()) {
        toast({
          title: 'Validation Error',
          description: 'Kode unit wajib diisi',
          status: 'error',
          isClosable: true,
        });
        setIsLoading(false);
        return;
      }

      if (!formData.name || !formData.name.trim()) {
        toast({
          title: 'Validation Error',
          description: 'Nama unit wajib diisi',
          status: 'error',
          isClosable: true,
        });
        setIsLoading(false);
        return;
      }

      // Clean the form data to ensure all fields are proper types
      const cleanedData: any = {
        code: formData.code.trim().toUpperCase(),
        name: formData.name.trim(),
        is_active: Boolean(formData.is_active)
      };
      
      // Add optional fields only if they have values
      if (formData.symbol && formData.symbol.trim()) {
        cleanedData.symbol = formData.symbol.trim();
      }
      
      if (formData.type && formData.type.trim()) {
        cleanedData.type = formData.type.trim();
      }
      
      if (formData.description && formData.description.trim()) {
        cleanedData.description = formData.description.trim();
      }
      
      let result;
      if (unit?.id) {
        result = await ProductService.updateProductUnit(unit.id, cleanedData);
      } else {
        result = await ProductService.createProductUnit(cleanedData as ProductUnit);
      }
      
      toast({
        title: `Unit ${unit?.id ? 'updated' : 'created'} successfully`,
        status: 'success',
        isClosable: true,
      });
      
      onSave(result.data);
    } catch (error: any) {
      let errorMessage = error.message || 'An error occurred';
      
      // Handle different error responses
      if (error.response?.data) {
        const errorData = error.response.data;
        errorMessage = errorData.error || errorMessage;
        
        // Special handling for conflict errors
        if (error.response.status === 409 && errorData.existing_unit) {
          errorMessage = `${errorData.error}. Existing unit: ${errorData.existing_unit.name} (ID: ${errorData.existing_unit.id})`;
        }
      }
      
      toast({
        title: `Failed to ${unit?.id ? 'update' : 'create'} unit`,
        description: errorMessage,
        status: 'error',
        isClosable: true,
        duration: 5000, // Show longer for conflict errors
      });
    } finally {
      setIsLoading(false);
    }
  };

  const unitTypes = [
    { value: 'Weight', label: 'Weight (Berat)' },
    { value: 'Volume', label: 'Volume' },
    { value: 'Count', label: 'Count (Jumlah)' },
    { value: 'Length', label: 'Length (Panjang)' },
    { value: 'Area', label: 'Area (Luas)' },
    { value: 'Time', label: 'Time (Waktu)' }
  ];

  return (
    <Box as="form" onSubmit={handleSubmit}>
      <VStack spacing={6} align="stretch">
        {/* Basic Information */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Informasi Dasar</Text>
          <Grid templateColumns="repeat(2, 1fr)" gap={4}>
            <GridItem>
              <FormControl isRequired>
                <FormLabel>Kode Unit</FormLabel>
                <Input
                  value={formData.code}
                  onChange={(e) => handleInputChange('code', e.target.value)}
                  placeholder="Masukkan kode unit (contoh: KG, PCS)"
                />
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl isRequired>
                <FormLabel>Nama Unit</FormLabel>
                <Input
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  placeholder="Masukkan nama unit"
                />
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Simbol</FormLabel>
                <Input
                  value={formData.symbol}
                  onChange={(e) => handleInputChange('symbol', e.target.value)}
                  placeholder="Masukkan simbol (contoh: kg, pcs)"
                />
              </FormControl>
            </GridItem>
            <GridItem>
              <FormControl>
                <FormLabel>Tipe Unit</FormLabel>
                <Select
                  value={formData.type}
                  onChange={(e) => handleInputChange('type', e.target.value)}
                  placeholder="Pilih tipe unit"
                >
                  {unitTypes.map(type => (
                    <option key={type.value} value={type.value}>
                      {type.label}
                    </option>
                  ))}
                </Select>
              </FormControl>
            </GridItem>
            <GridItem colSpan={2}>
              <FormControl>
                <FormLabel>Deskripsi</FormLabel>
                <Textarea
                  value={formData.description}
                  onChange={(e) => handleInputChange('description', e.target.value)}
                  placeholder="Masukkan deskripsi unit"
                  rows={3}
                />
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        {/* Settings */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4}>Pengaturan</Text>
          <FormControl display="flex" alignItems="center">
            <FormLabel htmlFor="is_active" mb="0">
              Status Aktif
            </FormLabel>
            <Switch
              id="is_active"
              isChecked={formData.is_active}
              onChange={(e) => handleInputChange('is_active', e.target.checked)}
            />
          </FormControl>
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
            loadingText={unit?.id ? 'Updating...' : 'Creating...'}
          >
            {unit?.id ? 'Update Unit' : 'Create Unit'}
          </Button>
        </HStack>
      </VStack>
    </Box>
  );
};

export default UnitForm;
