import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  VStack,
  HStack,
  Switch,
  useToast,
  Text,
  Grid,
  GridItem
} from '@chakra-ui/react';
import { FiSave, FiX, FiMapPin } from 'react-icons/fi';
import ProductService, { WarehouseLocation } from '@/services/productService';

interface WarehouseLocationFormProps {
  location?: WarehouseLocation;
  onSave: (location: WarehouseLocation) => void;
  onCancel: () => void;
}

const WarehouseLocationForm: React.FC<WarehouseLocationFormProps> = ({ 
  location, 
  onSave, 
  onCancel 
}) => {
  const [formData, setFormData] = useState<Partial<WarehouseLocation>>({
    code: '',
    name: '',
    description: '',
    address: '',
    is_active: true,
  });

  const [isLoading, setIsLoading] = useState(false);
  const toast = useToast();

  useEffect(() => {
    if (location) {
      setFormData(location);
    }
  }, [location]);

  const handleInputChange = (field: keyof WarehouseLocation, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.code || !formData.name) {
      toast({
        title: 'Validation Error',
        description: 'Code and Name are required fields',
        status: 'error',
        isClosable: true,
      });
      return;
    }

    setIsLoading(true);

    try {
      let result;
      if (location?.id) {
        result = await ProductService.updateWarehouseLocation(location.id, formData);
      } else {
        result = await ProductService.createWarehouseLocation(formData as WarehouseLocation);
      }
      
      toast({
        title: `Warehouse Location ${location?.id ? 'updated' : 'created'} successfully`,
        description: result.message?.includes('Mock') ? 'Note: Using mock data - implement backend API for persistence' : undefined,
        status: 'success',
        isClosable: true,
      });
      
      onSave(result.data);
    } catch (error: any) {
      toast({
        title: `Failed to ${location?.id ? 'update' : 'create'} warehouse location`,
        description: error.response?.data?.error || 'An error occurred',
        status: 'error',
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Box as="form" onSubmit={handleSubmit}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Box>
          <Text fontSize="lg" fontWeight="bold" mb={4} display="flex" alignItems="center">
            <FiMapPin style={{ marginRight: '8px' }} />
            {location?.id ? 'Edit Warehouse Location' : 'Add New Warehouse Location'}
          </Text>
        </Box>

        {/* Basic Information */}
        <Grid templateColumns="repeat(2, 1fr)" gap={4}>
          <GridItem>
            <FormControl isRequired>
              <FormLabel>Location Code</FormLabel>
              <Input
                value={formData.code}
                onChange={(e) => handleInputChange('code', e.target.value)}
                placeholder="e.g., WH-001, GDG-A"
                maxLength={20}
              />
            </FormControl>
          </GridItem>
          <GridItem>
            <FormControl isRequired>
              <FormLabel>Location Name</FormLabel>
              <Input
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                placeholder="e.g., Main Warehouse, Storage Room A"
              />
            </FormControl>
          </GridItem>
        </Grid>

        {/* Description */}
        <FormControl>
          <FormLabel>Description</FormLabel>
          <Textarea
            value={formData.description}
            onChange={(e) => handleInputChange('description', e.target.value)}
            placeholder="Optional description of the warehouse location..."
            rows={3}
          />
        </FormControl>

        {/* Address */}
        <FormControl>
          <FormLabel>Address</FormLabel>
          <Textarea
            value={formData.address}
            onChange={(e) => handleInputChange('address', e.target.value)}
            placeholder="Complete address of the warehouse location..."
            rows={3}
          />
        </FormControl>

        {/* Status */}
        <FormControl display="flex" alignItems="center">
          <FormLabel htmlFor="is_active" mb="0" mr={4}>
            Active Status
          </FormLabel>
          <Switch
            id="is_active"
            isChecked={formData.is_active}
            onChange={(e) => handleInputChange('is_active', e.target.checked)}
            colorScheme="green"
          />
          <Text ml={3} fontSize="sm" color="gray.600">
            {formData.is_active ? 'Active' : 'Inactive'}
          </Text>
        </FormControl>

        {/* Form Actions */}
        <HStack justify="flex-end" spacing={4} pt={4}>
          <Button
            leftIcon={<FiX />}
            onClick={onCancel}
            variant="outline"
            size="md"
          >
            Cancel
          </Button>
          <Button
            leftIcon={<FiSave />}
            type="submit"
            colorScheme="blue"
            isLoading={isLoading}
            loadingText={location?.id ? 'Updating...' : 'Creating...'}
            size="md"
          >
            {location?.id ? 'Update Location' : 'Create Location'}
          </Button>
        </HStack>
      </VStack>
    </Box>
  );
};

export default WarehouseLocationForm;
