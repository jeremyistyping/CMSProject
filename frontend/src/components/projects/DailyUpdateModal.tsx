'use client';

import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  Select,
  VStack,
  HStack,
  useToast,
  useColorModeValue,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Text,
} from '@chakra-ui/react';
import { FiSave, FiUpload, FiX, FiImage } from 'react-icons/fi';
import projectService from '@/services/projectService';
import { DailyUpdate } from '@/types/project';
import PhotoUpload from './PhotoUpload';

interface DailyUpdateModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: string;
  dailyUpdate: DailyUpdate | null;
  onSuccess: () => void;
}

interface DailyUpdateFormData {
  date: string;
  weather: string;
  workers_present: number;
  work_description: string;
  materials_used: string;
  issues: string;
}

interface PhotoPreview {
  file: File;
  preview: string;
}

const DailyUpdateModal: React.FC<DailyUpdateModalProps> = ({
  isOpen,
  onClose,
  projectId,
  dailyUpdate,
  onSuccess,
}) => {
  const toast = useToast();
  const [loading, setLoading] = useState(false);
  const [photoFiles, setPhotoFiles] = useState<PhotoPreview[]>([]);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const inputBgColor = useColorModeValue('white', 'gray.700');
  const inputTextColor = useColorModeValue('gray.800', 'white');
  const placeholderColor = useColorModeValue('gray.400', 'gray.400');

  const [formData, setFormData] = useState<DailyUpdateFormData>({
    date: new Date().toISOString().split('T')[0],
    weather: 'Sunny',
    workers_present: 0,
    work_description: '',
    materials_used: '',
    issues: '',
  });

  useEffect(() => {
    if (dailyUpdate) {
      // Edit mode - populate form
      setFormData({
        date: dailyUpdate.date.split('T')[0],
        weather: dailyUpdate.weather,
        workers_present: dailyUpdate.workers_present,
        work_description: dailyUpdate.work_description,
        materials_used: dailyUpdate.materials_used,
        issues: dailyUpdate.issues,
      });
      // Clear photos on edit mode (for now, we don't support editing photos)
      setPhotoFiles([]);
    } else {
      // Add mode - reset form
      setFormData({
        date: new Date().toISOString().split('T')[0],
        weather: 'Sunny',
        workers_present: 0,
        work_description: '',
        materials_used: '',
        issues: '',
      });
      setPhotoFiles([]);
    }
  }, [dailyUpdate, isOpen]);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleWorkersChange = (valueString: string) => {
    setFormData((prev) => ({
      ...prev,
      workers_present: parseInt(valueString) || 0,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!projectId) {
      toast({
        title: 'Error',
        description: 'Project ID is missing',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    if (!formData.date) {
      toast({
        title: 'Validation Error',
        description: 'Date is required',
        status: 'warning',
        duration: 3000,
      });
      return;
    }

    if (!formData.work_description.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Work description is required',
        status: 'warning',
        duration: 3000,
      });
      return;
    }

    console.log('Submitting daily update for project:', projectId);

    try {
      setLoading(true);

      const updateData = {
        ...formData,
        date: new Date(formData.date).toISOString(),
        photos: [], // Will be handled separately for multipart upload
        created_by: 'Current User', // Should be from auth context
      };

      console.log('Daily update data:', updateData);
      console.log('Photos to upload:', photoFiles.length);

      if (dailyUpdate) {
        // Update existing (no photo upload support for now)
        await projectService.updateDailyUpdate(projectId, dailyUpdate.id, updateData);
        toast({
          title: 'Success',
          description: 'Daily update updated successfully',
          status: 'success',
          duration: 3000,
        });
      } else {
        // Create new with photos
        const photos = photoFiles.map(p => p.file);
        await projectService.createDailyUpdate(projectId, updateData, photos);
        toast({
          title: 'Success',
          description: 'Daily update created successfully',
          status: 'success',
          duration: 3000,
        });
      }

      onSuccess();
    } catch (error: any) {
      console.error('Error saving daily update:', error);
      console.error('Error response:', error?.response?.data);
      console.error('Error status:', error?.response?.status);
      console.error('Project ID:', projectId);
      
      // Backend not ready - show friendly message
      if (error?.response?.status === 404) {
        toast({
          title: 'Backend Not Ready',
          description: 'Daily Updates API is not yet implemented in the backend. This feature will be available after backend setup.',
          status: 'warning',
          duration: 5000,
          isClosable: true,
        });
      } else if (error?.response?.status === 400) {
        // Bad request - likely validation error
        const errorMessage = error?.response?.data?.error || 'Invalid data. Please check all required fields.';
        toast({
          title: 'Validation Error',
          description: errorMessage,
          status: 'error',
          duration: 5000,
          isClosable: true,
        });
      } else {
        const errorMessage = error?.response?.data?.error || error?.message || 'Failed to save daily update';
        toast({
          title: 'Error',
          description: errorMessage,
          status: 'error',
          duration: 5000,
          isClosable: true,
        });
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleCancel} size="3xl" isCentered>
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(4px)" />
      <ModalContent 
        bg={bgColor} 
        maxH="90vh" 
        borderRadius="lg"
        boxShadow="2xl"
      >
        <ModalHeader 
          color={textColor} 
          borderBottomWidth="1px" 
          borderColor={borderColor}
          bg={bgColor}
          borderTopRadius="lg"
        >
          {dailyUpdate ? 'Edit Daily Update' : 'Add Daily Update'}
        </ModalHeader>
        <ModalCloseButton isDisabled={loading} />

        <form onSubmit={handleSubmit}>
          <ModalBody overflowY="auto" py={6} bg={bgColor}>
            <VStack spacing={5} align="stretch">
              {/* Date and Weather */}
              <HStack spacing={4} align="flex-start">
                <FormControl isRequired flex={1}>
                  <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
                    Date <Text as="span" color="red.500">*</Text>
                  </FormLabel>
                  <Input
                    type="date"
                    name="date"
                    value={formData.date}
                    onChange={handleInputChange}
                    bg={inputBgColor}
                    color={inputTextColor}
                    borderColor={borderColor}
                    borderWidth="1px"
                    isDisabled={loading}
                    size="md"
                    _focus={{
                      borderColor: 'green.500',
                      boxShadow: '0 0 0 1px var(--accent-color)',
                    }}
                    _placeholder={{ color: placeholderColor }}
                  />
                </FormControl>

                <FormControl isRequired flex={1}>
                  <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
                    Weather <Text as="span" color="red.500">*</Text>
                  </FormLabel>
                  <Select
                    name="weather"
                    value={formData.weather}
                    onChange={handleInputChange}
                    bg={inputBgColor}
                    color={inputTextColor}
                    borderColor={borderColor}
                    borderWidth="1px"
                    isDisabled={loading}
                    size="md"
                    _focus={{
                      borderColor: 'green.500',
                      boxShadow: '0 0 0 1px var(--accent-color)',
                    }}
                  >
                    <option value="Sunny">☀️ Sunny</option>
                    <option value="Cloudy">☁️ Cloudy</option>
                    <option value="Rainy">🌧️ Rainy</option>
                    <option value="Stormy">⛈️ Stormy</option>
                    <option value="Partly Cloudy">⛅ Partly Cloudy</option>
                  </Select>
                </FormControl>
              </HStack>

              {/* Workers Present */}
              <FormControl isRequired>
                <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
                  Number of Workers Present <Text as="span" color="red.500">*</Text>
                </FormLabel>
                <NumberInput
                  value={formData.workers_present}
                  onChange={handleWorkersChange}
                  min={0}
                  max={500}
                  isDisabled={loading}
                >
                  <NumberInputField
                    bg={inputBgColor}
                    color={inputTextColor}
                    borderColor={borderColor}
                    borderWidth="1px"
                    placeholder="e.g. 15"
                    _focus={{
                      borderColor: 'green.500',
                      boxShadow: '0 0 0 1px var(--accent-color)',
                    }}
                    _placeholder={{ color: placeholderColor }}
                  />
                  <NumberInputStepper>
                    <NumberIncrementStepper borderColor={borderColor} />
                    <NumberDecrementStepper borderColor={borderColor} />
                  </NumberInputStepper>
                </NumberInput>
                <Text fontSize="xs" color="gray.500" mt={1}>
                  Total workers on site today
                </Text>
              </FormControl>

              {/* Work Description */}
              <FormControl isRequired>
                <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
                  Work Description <Text as="span" color="red.500">*</Text>
                </FormLabel>
                <Textarea
                  name="work_description"
                  value={formData.work_description}
                  onChange={handleInputChange}
                  placeholder="Example: Foundation pouring completed, electrical wiring started on 2nd floor, plumbing installation ongoing..."
                  rows={4}
                  bg={inputBgColor}
                  color={inputTextColor}
                  borderColor={borderColor}
                  borderWidth="1px"
                  isDisabled={loading}
                  resize="vertical"
                  _focus={{
                    borderColor: 'green.500',
                    boxShadow: '0 0 0 1px var(--accent-color)',
                  }}
                  _placeholder={{ color: placeholderColor }}
                />
                <Text fontSize="xs" color="gray.500" mt={1}>
                  Describe the work completed today
                </Text>
              </FormControl>

              {/* Materials Used */}
              <FormControl>
                <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
                  Materials Used
                </FormLabel>
                <Textarea
                  name="materials_used"
                  value={formData.materials_used}
                  onChange={handleInputChange}
                  placeholder="Example: 50 bags cement, 10 cubic meters sand, 200 bricks..."
                  rows={3}
                  bg={inputBgColor}
                  color={inputTextColor}
                  borderColor={borderColor}
                  borderWidth="1px"
                  isDisabled={loading}
                  resize="vertical"
                  _focus={{
                    borderColor: 'green.500',
                    boxShadow: '0 0 0 1px var(--accent-color)',
                  }}
                  _placeholder={{ color: placeholderColor }}
                />
                <Text fontSize="xs" color="gray.500" mt={1}>
                  List materials consumed or delivered (optional)
                </Text>
              </FormControl>

              {/* Issues/Problems */}
              <FormControl>
                <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
                  Issues / Problems
                </FormLabel>
                <Textarea
                  name="issues"
                  value={formData.issues}
                  onChange={handleInputChange}
                  placeholder="Example: Material delivery delayed, equipment breakdown... (Leave blank if no issues)"
                  rows={3}
                  bg={inputBgColor}
                  color={inputTextColor}
                  borderColor={borderColor}
                  borderWidth="1px"
                  isDisabled={loading}
                  resize="vertical"
                  _focus={{
                    borderColor: 'green.500',
                    boxShadow: '0 0 0 1px var(--accent-color)',
                  }}
                  _placeholder={{ color: placeholderColor }}
                />
                <Text fontSize="xs" color="gray.500" mt={1}>
                  Report any problems encountered today (optional)
                </Text>
              </FormControl>

              {/* Photo Upload */}
              {!dailyUpdate && (
                <PhotoUpload
                  photos={photoFiles}
                  onPhotosChange={setPhotoFiles}
                  isDisabled={loading}
                  maxPhotos={10}
                />
              )}
            </VStack>
          </ModalBody>

          <ModalFooter 
            borderTopWidth="1px" 
            borderColor={borderColor}
            bg={bgColor}
            borderBottomRadius="lg"
          >
            <HStack spacing={3}>
              <Button 
                variant="ghost" 
                onClick={handleCancel} 
                isDisabled={loading}
                size="md"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                colorScheme="green"
                leftIcon={<FiSave />}
                isLoading={loading}
                loadingText={dailyUpdate ? 'Updating...' : 'Creating...'}
                size="md"
              >
                {dailyUpdate ? 'Update' : 'Create'} Daily Update
              </Button>
            </HStack>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default DailyUpdateModal;

