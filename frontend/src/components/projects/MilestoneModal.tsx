'use client';

import React, { useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  Select,
  NumberInput,
  NumberInputField,
  FormErrorMessage,
  VStack,
  HStack,
  useToast,
} from '@chakra-ui/react';
import { useForm } from 'react-hook-form';
import { useLanguage } from '@/contexts/LanguageContext';

export interface MilestoneFormData {
  title: string;
  description: string;
  work_area: string;
  priority: string;
  assigned_team: string;
  target_date: string;
}

interface MilestoneModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: number;
  milestone?: any; // Existing milestone for edit mode
  onSuccess: () => void;
}

const MilestoneModal: React.FC<MilestoneModalProps> = ({
  isOpen,
  onClose,
  projectId,
  milestone,
  onSuccess,
}) => {
  const { t } = useLanguage();
  const toast = useToast();
  const isEditMode = !!milestone;

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<MilestoneFormData>({
    defaultValues: {
      title: '',
      description: '',
      work_area: '',
      priority: 'medium',
      assigned_team: '',
      target_date: '',
    },
  });

  // Reset form when modal opens or milestone changes
  useEffect(() => {
    if (isOpen) {
      if (milestone) {
        // Edit mode - populate form with existing data
        setValue('title', milestone.title || '');
        setValue('description', milestone.description || '');
        setValue('work_area', milestone.work_area || '');
        setValue('priority', milestone.priority || 'medium');
        setValue('assigned_team', milestone.assigned_team || '');
        setValue('target_date', milestone.target_date ? milestone.target_date.split('T')[0] : '');
      } else {
        // Add mode - reset to defaults
        reset({
          title: '',
          description: '',
          work_area: '',
          priority: 'medium',
          assigned_team: '',
          target_date: '',
        });
      }
    }
  }, [isOpen, milestone, setValue, reset]);

  const onSubmit = async (data: MilestoneFormData) => {
    try {
      const url = isEditMode
        ? `/api/v1/projects/${projectId}/milestones/${milestone.id}`
        : `/api/v1/projects/${projectId}/milestones`;

      const method = isEditMode ? 'PUT' : 'POST';

      // Prepare payload - convert date string to ISO format
      // Date input gives us YYYY-MM-DD format, we need to ensure proper parsing
      let targetDateISO;
      if (data.target_date) {
        // Parse date string as YYYY-MM-DD and convert to ISO
        const dateParts = data.target_date.split('-');
        if (dateParts.length === 3) {
          // Create date with explicit year, month (0-indexed), day
          const dateObj = new Date(parseInt(dateParts[0]), parseInt(dateParts[1]) - 1, parseInt(dateParts[2]));
          targetDateISO = dateObj.toISOString();
        } else {
          // Fallback to direct parsing
          targetDateISO = new Date(data.target_date).toISOString();
        }
      } else {
        targetDateISO = new Date().toISOString();
      }

      const payload = {
        title: data.title,
        description: data.description || '',
        work_area: data.work_area || '',
        priority: data.priority || 'medium',
        assigned_team: data.assigned_team || '',
        target_date: targetDateISO,
        project_id: projectId,
      };

      console.log('Submitting milestone:', JSON.stringify(payload, null, 2));

      const token = localStorage.getItem('token');
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorData = await response.json();
        console.error('Backend error response:', errorData);
        throw new Error(errorData.error || errorData.details || 'Failed to save milestone');
      }

      toast({
        title: isEditMode ? t('milestones.updateSuccess') : t('milestones.createSuccess'),
        description: isEditMode
          ? t('milestones.updateSuccessDesc')
          : t('milestones.createSuccessDesc'),
        status: 'success',
        duration: 3000,
        isClosable: true,
      });

      onSuccess();
      onClose();
    } catch (error: any) {
      console.error('Error saving milestone:', error);
      toast({
        title: t('milestones.error'),
        description: error.message || t('milestones.saveError'),
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={handleClose} 
      size="xl" 
      scrollBehavior="inside"
      isCentered
    >
      <ModalOverlay 
        bg="blackAlpha.600" 
        backdropFilter="blur(4px)" 
      />
      <ModalContent 
        bg="white" 
        maxH="90vh"
        borderRadius="lg"
        boxShadow="xl"
      >
        <ModalHeader bg="white" borderTopRadius="lg" borderBottomWidth="1px" borderColor="gray.200">
          {isEditMode ? 'Edit Milestone' : 'Add New Milestone'}
        </ModalHeader>
        <ModalCloseButton />

        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody 
            bg="white" 
            overflowY="auto" 
            maxH="calc(90vh - 140px)"
            py={6}
          >
            <VStack spacing={4} align="stretch">
              {/* Milestone Title */}
              <FormControl isInvalid={!!errors.title} isRequired>
                <FormLabel>Milestone Title</FormLabel>
                <Input
                  {...register('title', {
                    required: 'Milestone title is required',
                    minLength: {
                      value: 3,
                      message: 'Title must be at least 3 characters',
                    },
                  })}
                  placeholder="e.g., Electrical Installation"
                  bg="white"
                />
                <FormErrorMessage>{errors.title?.message}</FormErrorMessage>
              </FormControl>

              {/* Work Area/Phase */}
              <FormControl isInvalid={!!errors.work_area}>
                <FormLabel>Work Area/Phase</FormLabel>
                <Select {...register('work_area')} placeholder="Select work area" bg="white">
                  <option value="Site Preparation">Site Preparation</option>
                  <option value="Foundation Work">Foundation Work</option>
                  <option value="Structural Work">Structural Work</option>
                  <option value="Roofing">Roofing</option>
                  <option value="Wall Installation">Wall Installation</option>
                  <option value="Ceiling Installation">Ceiling Installation</option>
                  <option value="Electrical Installation">Electrical Installation</option>
                  <option value="Clean Water Installation">Clean Water Installation</option>
                  <option value="Gray Water Installation">Gray Water Installation</option>
                  <option value="Flooring Installation">Flooring Installation</option>
                  <option value="HVAC Installation">HVAC Installation</option>
                  <option value="Kitchen Equipment Installation">Kitchen Equipment Installation</option>
                  <option value="Furniture Installation">Furniture Installation</option>
                  <option value="Utensils Installation">Utensils Installation</option>
                </Select>
                <FormErrorMessage>{errors.work_area?.message}</FormErrorMessage>
              </FormControl>

              {/* Priority */}
              <FormControl isInvalid={!!errors.priority}>
                <FormLabel>Priority</FormLabel>
                <Select {...register('priority')} bg="white">
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                </Select>
                <FormErrorMessage>{errors.priority?.message}</FormErrorMessage>
              </FormControl>

              {/* Target Date */}
              <FormControl isInvalid={!!errors.target_date} isRequired>
                <FormLabel>Target Date</FormLabel>
                <Input
                  type="date"
                  {...register('target_date', {
                    required: 'Target date is required',
                  })}
                  placeholder="dd/mm/yyyy"
                  bg="white"
                />
                <FormErrorMessage>{errors.target_date?.message}</FormErrorMessage>
              </FormControl>

              {/* Assigned Team */}
              <FormControl isInvalid={!!errors.assigned_team}>
                <FormLabel>Assigned Team</FormLabel>
                <Input
                  {...register('assigned_team')}
                  placeholder="e.g., Electrical Team A"
                  bg="white"
                />
                <FormErrorMessage>{errors.assigned_team?.message}</FormErrorMessage>
              </FormControl>

              {/* Description */}
              <FormControl isInvalid={!!errors.description}>
                <FormLabel>Description</FormLabel>
                <Textarea
                  {...register('description')}
                  placeholder="Detailed milestone description..."
                  rows={4}
                  bg="white"
                />
                <FormErrorMessage>{errors.description?.message}</FormErrorMessage>
              </FormControl>
            </VStack>
          </ModalBody>

          <ModalFooter 
            bg="white" 
            borderBottomRadius="lg" 
            borderTopWidth="1px" 
            borderColor="gray.200"
          >
            <Button variant="ghost" mr={3} onClick={handleClose} isDisabled={isSubmitting}>
              Cancel
            </Button>
            <Button colorScheme="blue" type="submit" isLoading={isSubmitting}>
              {isEditMode ? 'Update' : 'Add Milestone'}
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default MilestoneModal;

