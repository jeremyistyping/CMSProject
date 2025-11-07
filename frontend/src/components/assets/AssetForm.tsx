'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  VStack,
  HStack,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Text,
  Divider,
  useToast,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Switch,
  Badge,
  Icon,
  ButtonGroup,
  Spinner,
  Alert,
  AlertIcon,
  AlertDescription,
} from '@chakra-ui/react';
import { FiSave, FiX, FiMapPin, FiMap, FiImage } from 'react-icons/fi';
import { Asset, AssetFormData, ASSET_CATEGORIES, ASSET_STATUS, DEPRECIATION_METHODS, DEPRECIATION_METHOD_LABELS } from '@/types/asset';
import InteractiveMapPicker from '@/components/common/InteractiveMapPicker';
import AssetImageUpload from './AssetImageUpload';

interface AssetFormProps {
  asset?: Asset | null;
  onSubmit: (data: AssetFormData) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
  mode?: 'create' | 'edit';
}

const AssetForm: React.FC<AssetFormProps> = ({
  asset,
  onSubmit,
  onCancel,
  isLoading = false,
  mode = 'create',
}) => {
  const toast = useToast();
  const [isMapPickerOpen, setIsMapPickerOpen] = useState(false);

  // Form state
  const [formData, setFormData] = useState<AssetFormData>({
    name: '',
    category: 'Office Equipment',
    status: 'ACTIVE',
    purchaseDate: new Date().toISOString().split('T')[0],
    purchasePrice: 0,
    salvageValue: 0,
    usefulLife: 5,
    depreciationMethod: 'STRAIGHT_LINE',
    isActive: true,
    notes: '',
    location: '',
    coordinates: '',
    serialNumber: '',
    condition: 'Good',
    assetAccountId: undefined,
    depreciationAccountId: undefined,
  });

  // Initialize form data if editing
  useEffect(() => {
    if (asset && mode === 'edit') {
      setFormData({
        code: asset.code,
        name: asset.name,
        category: asset.category,
        status: asset.status,
        purchaseDate: asset.purchase_date.split('T')[0],
        purchasePrice: asset.purchase_price,
        salvageValue: asset.salvage_value,
        usefulLife: asset.useful_life,
        depreciationMethod: asset.depreciation_method,
        isActive: asset.is_active,
        notes: asset.notes || '',
        location: asset.location || '',
        coordinates: asset.coordinates || '',
        serialNumber: asset.serial_number || '',
        condition: asset.condition || 'Good',
        assetAccountId: asset.asset_account_id,
        depreciationAccountId: asset.depreciation_account_id,
      });
    }
  }, [asset, mode]);

  // Handle form input changes
  const handleInputChange = (field: keyof AssetFormData) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const value = event.target.value;
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  // Handle number input changes
  const handleNumberChange = (field: keyof AssetFormData) => (
    valueAsString: string,
    valueAsNumber: number
  ) => {
    setFormData(prev => ({
      ...prev,
      [field]: isNaN(valueAsNumber) ? 0 : valueAsNumber,
    }));
  };

  // Handle switch changes
  const handleSwitchChange = (field: keyof AssetFormData) => (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setFormData(prev => ({
      ...prev,
      [field]: event.target.checked,
    }));
  };

  // Handle location selection from map picker
  const handleLocationSelect = (coordinates: string, address?: string) => {
    setFormData(prev => ({
      ...prev,
      coordinates: coordinates,
      location: address || coordinates,
    }));
    
    toast({
      title: 'Location Selected',
      description: `Asset location set to: ${address || coordinates}`,
      status: 'success',
      duration: 3000,
      isClosable: true,
    });
  };

  // Clear location
  const handleClearLocation = () => {
    setFormData(prev => ({
      ...prev,
      coordinates: '',
      location: '',
    }));
  };

  // Form validation
  const validateForm = (): string[] => {
    const errors: string[] = [];

    if (!formData.name.trim()) {
      errors.push('Asset name is required');
    }

    if (!formData.category) {
      errors.push('Category is required');
    }

    if (!formData.purchaseDate) {
      errors.push('Purchase date is required');
    }

    if (formData.purchasePrice <= 0) {
      errors.push('Purchase price must be greater than 0');
    }

    if (formData.salvageValue && formData.salvageValue < 0) {
      errors.push('Salvage value cannot be negative');
    }

    if (formData.salvageValue && formData.salvageValue >= formData.purchasePrice) {
      errors.push('Salvage value must be less than purchase price');
    }

    if (formData.usefulLife <= 0) {
      errors.push('Useful life must be greater than 0 years');
    }

    return errors;
  };

  // Handle form submission
  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    // Validate form
    const errors = validateForm();
    if (errors.length > 0) {
      toast({
        title: 'Validation Error',
        description: errors.join(', '),
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    try {
      await onSubmit(formData);
    } catch (error) {
      console.error('Form submission error:', error);
    }
  };

  // Calculate estimated book value
  const calculateCurrentBookValue = () => {
    const { purchasePrice, salvageValue, usefulLife } = formData;
    if (purchasePrice <= 0 || usefulLife <= 0) return purchasePrice;

    const annualDepreciation = (purchasePrice - (salvageValue || 0)) / usefulLife;
    const currentYear = new Date().getFullYear();
    const purchaseYear = new Date(formData.purchaseDate).getFullYear();
    const yearsElapsed = Math.max(0, currentYear - purchaseYear);

    const bookValue = purchasePrice - (annualDepreciation * yearsElapsed);
    return Math.max(bookValue, salvageValue || 0);
  };

  const currentBookValue = calculateCurrentBookValue();

  return (
    <Box maxW="4xl" mx="auto" p={6}>
      <Card>
        <CardHeader>
          <HStack justify="space-between" align="center">
            <Box>
              <Heading size="lg" color="gray.700">
                {mode === 'create' ? 'Add New Asset' : 'Edit Asset'}
              </Heading>
              <Text color="gray.500" fontSize="sm" mt={1}>
                {mode === 'create' 
                  ? 'Fill in the details below to register a new asset'
                  : 'Update asset information and location details'
                }
              </Text>
            </Box>
            {formData.coordinates && (
              <Badge colorScheme="green" variant="solid" px={3} py={1}>
                <Icon as={FiMapPin} mr={1} />
                Location Set
              </Badge>
            )}
          </HStack>
        </CardHeader>
        
        <CardBody>
          <form onSubmit={handleSubmit}>
            <VStack spacing={6} align="stretch">
              
              {/* Basic Information */}
              <Box>
                <Heading size="md" mb={4} color="gray.600">
                  📋 Basic Information
                </Heading>
                <VStack spacing={4}>
                  <HStack spacing={4} w="full">
                    <FormControl isRequired>
                      <FormLabel>Asset Name</FormLabel>
                      <Input
                        value={formData.name}
                        onChange={handleInputChange('name')}
                        placeholder="e.g., Dell Laptop, Office Desk"
                        bg="gray.50"
                        _focus={{ bg: 'white' }}
                      />
                    </FormControl>

                    <FormControl isRequired>
                      <FormLabel>Category</FormLabel>
                      <Select
                        value={formData.category}
                        onChange={handleInputChange('category')}
                        bg="gray.50"
                      >
                        {ASSET_CATEGORIES.map((category) => (
                          <option key={category} value={category}>
                            {category}
                          </option>
                        ))}
                      </Select>
                    </FormControl>
                  </HStack>

                  <HStack spacing={4} w="full">
                    <FormControl>
                      <FormLabel>Serial Number</FormLabel>
                      <Input
                        value={formData.serialNumber}
                        onChange={handleInputChange('serialNumber')}
                        placeholder="e.g., SN123456789"
                        bg="gray.50"
                        _focus={{ bg: 'white' }}
                      />
                    </FormControl>

                    <FormControl>
                      <FormLabel>Status</FormLabel>
                      <Select
                        value={formData.status}
                        onChange={handleInputChange('status')}
                        bg="gray.50"
                      >
                        {ASSET_STATUS.map((status) => (
                          <option key={status} value={status}>
                            {status}
                          </option>
                        ))}
                      </Select>
                    </FormControl>

                    <FormControl>
                      <FormLabel>Condition</FormLabel>
                      <Select
                        value={formData.condition}
                        onChange={handleInputChange('condition')}
                        bg="gray.50"
                      >
                        <option value="Excellent">Excellent</option>
                        <option value="Good">Good</option>
                        <option value="Fair">Fair</option>
                        <option value="Poor">Poor</option>
                      </Select>
                    </FormControl>
                  </HStack>
                </VStack>
              </Box>

              <Divider />

              {/* Financial Information */}
              <Box>
                <Heading size="md" mb={4} color="gray.600">
                  💰 Financial Information
                </Heading>
                <VStack spacing={4}>
                  <HStack spacing={4} w="full">
                    <FormControl isRequired>
                      <FormLabel>Purchase Date</FormLabel>
                      <Input
                        type="date"
                        value={formData.purchaseDate}
                        onChange={handleInputChange('purchaseDate')}
                        bg="gray.50"
                        _focus={{ bg: 'white' }}
                      />
                    </FormControl>

                    <FormControl isRequired>
                      <FormLabel>Purchase Price (IDR)</FormLabel>
                      <NumberInput
                        value={formData.purchasePrice}
                        onChange={handleNumberChange('purchasePrice')}
                        min={0}
                        precision={0}
                        step={100000}
                      >
                        <NumberInputField bg="gray.50" _focus={{ bg: 'white' }} />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>
                  </HStack>

                  <HStack spacing={4} w="full">
                    <FormControl>
                      <FormLabel>Salvage Value (IDR)</FormLabel>
                      <NumberInput
                        value={formData.salvageValue}
                        onChange={handleNumberChange('salvageValue')}
                        min={0}
                        precision={0}
                        step={100000}
                      >
                        <NumberInputField bg="gray.50" _focus={{ bg: 'white' }} />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>

                    <FormControl isRequired>
                      <FormLabel>Useful Life (Years)</FormLabel>
                      <NumberInput
                        value={formData.usefulLife}
                        onChange={handleNumberChange('usefulLife')}
                        min={1}
                        max={50}
                        precision={0}
                      >
                        <NumberInputField bg="gray.50" _focus={{ bg: 'white' }} />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>

                    <FormControl>
                      <FormLabel>Depreciation Method</FormLabel>
                      <Select
                        value={formData.depreciationMethod}
                        onChange={handleInputChange('depreciationMethod')}
                        bg="gray.50"
                      >
                        {DEPRECIATION_METHODS.map((method) => (
                          <option key={method} value={method}>
                            {DEPRECIATION_METHOD_LABELS[method]}
                          </option>
                        ))}
                      </Select>
                    </FormControl>
                  </HStack>

                  {/* Book Value Estimation */}
                  {formData.purchasePrice > 0 && (
                    <Alert status="info" borderRadius="md">
                      <AlertIcon />
                      <AlertDescription fontSize="sm">
                        <Text>
                          <strong>Estimated Current Book Value:</strong> {' '}
                          {new Intl.NumberFormat('id-ID', {
                            style: 'currency',
                            currency: 'IDR',
                            minimumFractionDigits: 0,
                          }).format(currentBookValue)}
                        </Text>
                      </AlertDescription>
                    </Alert>
                  )}
                </VStack>
              </Box>

              <Divider />

              {/* Location Information */}
              <Box>
                <Heading size="md" mb={4} color="gray.600">
                  📍 Location Information
                </Heading>
                <VStack spacing={4}>
                  <FormControl>
                    <FormLabel>Location Description</FormLabel>
                    <Input
                      value={formData.location}
                      onChange={handleInputChange('location')}
                      placeholder="e.g., Main Office - Floor 2, Room 201"
                      bg="gray.50"
                      _focus={{ bg: 'white' }}
                    />
                  </FormControl>

                  {/* Map Location Section */}
                  <Box w="full">
                    <FormLabel mb={3}>GPS Coordinates (Optional)</FormLabel>
                    
                    {!formData.coordinates ? (
                      <Alert status="info" borderRadius="lg">
                        <AlertIcon />
                        <Box>
                          <AlertDescription fontSize="sm">
                            <Text mb={3}>
                              Set precise GPS location for this asset to enable location tracking and mapping features.
                            </Text>
                            <Button
                              leftIcon={<FiMap />}
                              colorScheme="blue"
                              variant="outline"
                              onClick={() => setIsMapPickerOpen(true)}
                              size="sm"
                            >
                              Select Location
                            </Button>
                          </AlertDescription>
                        </Box>
                      </Alert>
                    ) : (
                      <Box
                        p={4}
                        bg="green.50"
                        borderRadius="lg"
                        border="1px"
                        borderColor="green.200"
                      >
                        <VStack spacing={3}>
                          <HStack w="full" justify="space-between">
                            <HStack>
                              <Icon as={FiMapPin} color="green.500" />
                              <Text fontSize="sm" fontWeight="medium" color="green.700">
                                ✅ Location Confirmed
                              </Text>
                            </HStack>
                            <Badge colorScheme="green" variant="solid">
                              GPS Set
                            </Badge>
                          </HStack>

                          <Box w="full" p={3} bg="white" borderRadius="md">
                            <Text fontSize="xs" color="gray.600" mb={1}>
                              GPS Coordinates:
                            </Text>
                            <Text fontSize="sm" fontFamily="mono" color="green.800" fontWeight="medium">
                              📍 {formData.coordinates}
                            </Text>
                          </Box>

                          <HStack spacing={2}>
                            <Button
                              size="sm"
                              variant="outline"
                              colorScheme="blue"
                              leftIcon={<FiMap />}
                              onClick={() => {
                                const mapsUrl = `https://www.google.com/maps?q=${formData.coordinates}&z=15`;
                                window.open(mapsUrl, '_blank');
                              }}
                            >
                              View on Maps
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              colorScheme="green"
                              leftIcon={<FiMapPin />}
                              onClick={() => setIsMapPickerOpen(true)}
                            >
                              Update Location
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              colorScheme="gray"
                              onClick={handleClearLocation}
                            >
                              Clear
                            </Button>
                          </HStack>
                        </VStack>
                      </Box>
                    )}
                  </Box>
                </VStack>
              </Box>

              <Divider />

              {/* Asset Image */}
              <Box>
                <Heading size="md" mb={4} color="gray.600">
                  📸 Asset Image
                </Heading>
                <VStack spacing={4}>
                  {mode === 'edit' && asset?.id ? (
                    <AssetImageUpload
                      asset={asset}
                      size="md"
                      showLabel={false}
                      onImageUpload={() => {
                        // Refresh the asset data after image upload
                        toast({
                          title: 'Image Updated',
                          description: 'Asset image has been successfully updated.',
                          status: 'success',
                          duration: 3000,
                          isClosable: true,
                        });
                      }}
                    />
                  ) : (
                    <Alert status="info" borderRadius="lg">
                      <AlertIcon />
                      <Box>
                        <AlertDescription fontSize="sm">
                          <Text mb={2}>
                            💡 <strong>Save Asset First to Upload Image</strong>
                          </Text>
                          <Text>
                            You can upload an image after creating this asset. Click "Create Asset" button to save, then edit the asset to add an image.
                          </Text>
                        </AlertDescription>
                      </Box>
                    </Alert>
                  )}
                </VStack>
              </Box>

              <Divider />

              {/* Additional Information */}
              <Box>
                <Heading size="md" mb={4} color="gray.600">
                  📝 Additional Information
                </Heading>
                <VStack spacing={4}>
                  <FormControl>
                    <FormLabel>Notes</FormLabel>
                    <Textarea
                      value={formData.notes}
                      onChange={handleInputChange('notes')}
                      placeholder="Any additional notes about this asset..."
                      rows={3}
                      bg="gray.50"
                      _focus={{ bg: 'white' }}
                    />
                  </FormControl>

                  <HStack w="full" justify="space-between">
                    <FormControl display="flex" alignItems="center">
                      <Switch
                        id="is-active"
                        isChecked={formData.isActive}
                        onChange={handleSwitchChange('isActive')}
                        colorScheme="green"
                        mr={3}
                      />
                      <FormLabel htmlFor="is-active" mb="0" fontSize="sm">
                        Asset is active
                      </FormLabel>
                    </FormControl>
                  </HStack>
                </VStack>
              </Box>

              {/* Form Actions */}
              <HStack spacing={4} justify="flex-end" pt={6} borderTop="1px" borderColor="gray.200">
                <Button
                  variant="outline"
                  onClick={onCancel}
                  size="lg"
                  isDisabled={isLoading}
                >
                  <Icon as={FiX} mr={2} />
                  Cancel
                </Button>
                <Button
                  type="submit"
                  colorScheme="blue"
                  size="lg"
                  isLoading={isLoading}
                  loadingText={mode === 'create' ? 'Creating...' : 'Updating...'}
                  leftIcon={isLoading ? undefined : <FiSave />}
                >
                  {mode === 'create' ? 'Create Asset' : 'Update Asset'}
                </Button>
              </HStack>
            </VStack>
          </form>
        </CardBody>
      </Card>

      {/* Interactive Map Picker Modal */}
      <InteractiveMapPicker
        isOpen={isMapPickerOpen}
        onClose={() => setIsMapPickerOpen(false)}
        onLocationSelect={handleLocationSelect}
        currentCoordinates={formData.coordinates}
        title="Select Asset Location"
      />
    </Box>
  );
};

export default AssetForm;
