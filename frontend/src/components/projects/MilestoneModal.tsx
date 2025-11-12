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
  name: string;
  description: string;
  target_date: string;
  status: string;
  order_number: number;
  weight_percentage: number;
  actual_completion_date?: string;
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
      name: '',
      description: '',
      target_date: '',
      status: 'pending',
      order_number: 1,
      weight_percentage: 0,
      actual_completion_date: '',
    },
  });

  // Reset form when modal opens or milestone changes
  useEffect(() => {
    if (isOpen) {
      if (milestone) {
        // Edit mode - populate form with existing data
        setValue('name', milestone.name || '');
        setValue('description', milestone.description || '');
        setValue('target_date', milestone.target_date ? milestone.target_date.split('T')[0] : '');
        setValue('status', milestone.status || 'pending');
        setValue('order_number', milestone.order_number || 1);
        setValue('weight_percentage', milestone.weight_percentage || 0);
        setValue('actual_completion_date', milestone.actual_completion_date ? milestone.actual_completion_date.split('T')[0] : '');
      } else {
        // Add mode - reset to defaults
        reset({
          name: '',
          description: '',
          target_date: '',
          status: 'pending',
          order_number: 1,
          weight_percentage: 0,
          actual_completion_date: '',
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

      // Prepare payload
      const payload = {
        ...data,
        project_id: projectId,
        // Convert empty string to null for optional fields
        actual_completion_date: data.actual_completion_date || null,
      };

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
        throw new Error(errorData.error || 'Failed to save milestone');
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
    <Modal isOpen={isOpen} onClose={handleClose} size="xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          {isEditMode ? t('milestones.editMilestone') : t('milestones.addMilestone')}
        </ModalHeader>
        <ModalCloseButton />

        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody>
            <VStack spacing={4} align="stretch">
              {/* Name */}
              <FormControl isInvalid={!!errors.name} isRequired>
                <FormLabel>{t('milestones.name')}</FormLabel>
                <Input
                  {...register('name', {
                    required: t('milestones.nameRequired'),
                    minLength: {
                      value: 3,
                      message: t('milestones.nameMinLength'),
                    },
                    maxLength: {
                      value: 100,
                      message: t('milestones.nameMaxLength'),
                    },
                  })}
                  placeholder={t('milestones.namePlaceholder')}
                />
                <FormErrorMessage>{errors.name?.message}</FormErrorMessage>
              </FormControl>

              {/* Description */}
              <FormControl isInvalid={!!errors.description}>
                <FormLabel>{t('milestones.description')}</FormLabel>
                <Textarea
                  {...register('description', {
                    maxLength: {
                      value: 500,
                      message: t('milestones.descriptionMaxLength'),
                    },
                  })}
                  placeholder={t('milestones.descriptionPlaceholder')}
                  rows={3}
                />
                <FormErrorMessage>{errors.description?.message}</FormErrorMessage>
              </FormControl>

              <HStack spacing={4}>
                {/* Order Number */}
                <FormControl isInvalid={!!errors.order_number} isRequired flex={1}>
                  <FormLabel>{t('milestones.orderNumber')}</FormLabel>
                  <NumberInput min={1} max={100} defaultValue={1}>
                    <NumberInputField
                      {...register('order_number', {
                        required: t('milestones.orderRequired'),
                        valueAsNumber: true,
                        min: {
                          value: 1,
                          message: t('milestones.orderMin'),
                        },
                        max: {
                          value: 100,
                          message: t('milestones.orderMax'),
                        },
                      })}
                    />
                  </NumberInput>
                  <FormErrorMessage>{errors.order_number?.message}</FormErrorMessage>
                </FormControl>

                {/* Weight Percentage */}
                <FormControl isInvalid={!!errors.weight_percentage} isRequired flex={1}>
                  <FormLabel>{t('milestones.weight')} (%)</FormLabel>
                  <NumberInput min={0} max={100} defaultValue={0}>
                    <NumberInputField
                      {...register('weight_percentage', {
                        required: t('milestones.weightRequired'),
                        valueAsNumber: true,
                        min: {
                          value: 0,
                          message: t('milestones.weightMin'),
                        },
                        max: {
                          value: 100,
                          message: t('milestones.weightMax'),
                        },
                      })}
                    />
                  </NumberInput>
                  <FormErrorMessage>{errors.weight_percentage?.message}</FormErrorMessage>
                </FormControl>
              </HStack>

              <HStack spacing={4}>
                {/* Status */}
                <FormControl isInvalid={!!errors.status} isRequired flex={1}>
                  <FormLabel>{t('milestones.status')}</FormLabel>
                  <Select {...register('status', { required: t('milestones.statusRequired') })}>
                    <option value="pending">{t('milestones.statusPending')}</option>
                    <option value="in-progress">{t('milestones.statusInProgress')}</option>
                    <option value="completed">{t('milestones.statusCompleted')}</option>
                    <option value="delayed">{t('milestones.statusDelayed')}</option>
                  </Select>
                  <FormErrorMessage>{errors.status?.message}</FormErrorMessage>
                </FormControl>

                {/* Target Date */}
                <FormControl isInvalid={!!errors.target_date} isRequired flex={1}>
                  <FormLabel>{t('milestones.targetDate')}</FormLabel>
                  <Input
                    type="date"
                    {...register('target_date', {
                      required: t('milestones.targetDateRequired'),
                    })}
                  />
                  <FormErrorMessage>{errors.target_date?.message}</FormErrorMessage>
                </FormControl>
              </HStack>

              {/* Actual Completion Date - Optional */}
              <FormControl isInvalid={!!errors.actual_completion_date}>
                <FormLabel>{t('milestones.actualCompletionDate')}</FormLabel>
                <Input type="date" {...register('actual_completion_date')} />
                <FormErrorMessage>{errors.actual_completion_date?.message}</FormErrorMessage>
              </FormControl>
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleClose} isDisabled={isSubmitting}>
              {t('common.cancel')}
            </Button>
            <Button colorScheme="blue" type="submit" isLoading={isSubmitting}>
              {isEditMode ? t('common.update') : t('common.create')}
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default MilestoneModal;

