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
  VStack,
  HStack,
  Text,
  Slider,
  SliderTrack,
  SliderFilledTrack,
  SliderThumb,
  Box,
  Progress,
  useColorModeValue,
  useToast,
  Icon,
  Grid,
  GridItem,
  Divider,
} from '@chakra-ui/react';
import { FiBarChart, FiDatabase, FiTarget, FiFileText, FiClock, FiSave } from 'react-icons/fi';
import projectService from '@/services/projectService';
import { Project } from '@/types/project';

interface UpdateProgressModalProps {
  isOpen: boolean;
  onClose: () => void;
  project: Project;
  onSuccess: () => void; // Callback to refresh project data
}

interface ProgressData {
  overall_progress: number;
  foundation_progress: number;
  utilities_progress: number;
  interior_progress: number;
  equipment_progress: number;
}

const UpdateProgressModal: React.FC<UpdateProgressModalProps> = ({
  isOpen,
  onClose,
  project,
  onSuccess,
}) => {
  const toast = useToast();
  const [loading, setLoading] = useState(false);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const sliderTrackColor = useColorModeValue('gray.200', 'gray.600');
  const cardBgColor = useColorModeValue('gray.50', 'var(--bg-primary)');

  // Initialize progress state from project
  const [progress, setProgress] = useState<ProgressData>({
    overall_progress: 0,
    foundation_progress: 0,
    utilities_progress: 0,
    interior_progress: 0,
    equipment_progress: 0,
  });

  // Update progress state when project changes
  useEffect(() => {
    if (project) {
      setProgress({
        overall_progress: project.overall_progress || 0,
        foundation_progress: project.foundation_progress || 0,
        utilities_progress: project.utilities_progress || 0,
        interior_progress: project.interior_progress || 0,
        equipment_progress: project.equipment_progress || 0,
      });
    }
  }, [project, isOpen]);

  const handleProgressChange = (field: keyof ProgressData, value: number) => {
    setProgress((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSave = async () => {
    try {
      setLoading(true);

      // Call backend API to update progress
      await projectService.updateProgress(project.id.toString(), progress);

      toast({
        title: 'Success',
        description: 'Project progress updated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });

      // Refresh project data
      onSuccess();

      // Close modal
      onClose();
    } catch (error) {
      console.error('Error updating progress:', error);
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to update progress. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    // Reset to original values
    if (project) {
      setProgress({
        overall_progress: project.overall_progress || 0,
        foundation_progress: project.foundation_progress || 0,
        utilities_progress: project.utilities_progress || 0,
        interior_progress: project.interior_progress || 0,
        equipment_progress: project.equipment_progress || 0,
      });
    }
    onClose();
  };

  const getProgressColor = (value: number) => {
    if (value >= 75) return 'green';
    if (value >= 50) return 'green';
    if (value >= 25) return 'orange';
    return 'red';
  };

  const ProgressItem = ({
    label,
    icon,
    color,
    value,
    field,
  }: {
    label: string;
    icon: any;
    color: string;
    value: number;
    field: keyof ProgressData;
  }) => (
    <Box p={4} bg={cardBgColor} borderRadius="md" borderWidth="1px" borderColor={borderColor}>
      <VStack align="stretch" spacing={3}>
        <HStack justify="space-between">
          <HStack spacing={2}>
            <Icon as={icon} color={`${color}.500`} boxSize={5} />
            <Text fontSize="sm" fontWeight="medium" color={textColor}>
              {label}
            </Text>
          </HStack>
          <Text fontSize="xl" fontWeight="bold" color={textColor}>
            {value}%
          </Text>
        </HStack>

        <Slider
          value={value}
          onChange={(val) => handleProgressChange(field, val)}
          min={0}
          max={100}
          step={1}
          isDisabled={loading}
        >
          <SliderTrack bg={sliderTrackColor}>
            <SliderFilledTrack bg={`${color}.500`} />
          </SliderTrack>
          <SliderThumb boxSize={5} bg={`${color}.500`} />
        </Slider>

        <Progress
          value={value}
          size="sm"
          colorScheme={getProgressColor(value)}
          borderRadius="full"
        />
      </VStack>
    </Box>
  );

  return (
    <Modal isOpen={isOpen} onClose={handleCancel} size="2xl" isCentered>
      <ModalOverlay backdropFilter="blur(4px)" />
      <ModalContent bg={bgColor} maxH="90vh" overflowY="auto">
        <ModalHeader color={textColor}>
          <VStack align="start" spacing={1}>
            <Text fontSize="xl" fontWeight="bold">
              Update Project Progress
            </Text>
            <Text fontSize="sm" fontWeight="normal" color={subtextColor}>
              {project.project_name}
            </Text>
          </VStack>
        </ModalHeader>
        <ModalCloseButton isDisabled={loading} />

        <ModalBody pb={6}>
          <VStack spacing={4} align="stretch">
            {/* Overall Progress - Full Width */}
            <Box>
              <Text fontSize="md" fontWeight="semibold" color={textColor} mb={3}>
                Overall Progress
              </Text>
              <ProgressItem
                label="Total Completion"
                icon={FiBarChart}
                color="green"
                value={progress.overall_progress}
                field="overall_progress"
              />
            </Box>

            <Divider />

            {/* Individual Progress Categories */}
            <Box>
              <Text fontSize="md" fontWeight="semibold" color={textColor} mb={3}>
                Progress by Category
              </Text>
              <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                <GridItem>
                  <ProgressItem
                    label="Foundation"
                    icon={FiDatabase}
                    color="orange"
                    value={progress.foundation_progress}
                    field="foundation_progress"
                  />
                </GridItem>
                <GridItem>
                  <ProgressItem
                    label="Utilities"
                    icon={FiTarget}
                    color="purple"
                    value={progress.utilities_progress}
                    field="utilities_progress"
                  />
                </GridItem>
                <GridItem>
                  <ProgressItem
                    label="Interior"
                    icon={FiFileText}
                    color="pink"
                    value={progress.interior_progress}
                    field="interior_progress"
                  />
                </GridItem>
                <GridItem>
                  <ProgressItem
                    label="Equipment"
                    icon={FiClock}
                    color="green"
                    value={progress.equipment_progress}
                    field="equipment_progress"
                  />
                </GridItem>
              </Grid>
            </Box>

            {/* Progress Preview Summary */}
            <Box p={4} bg={cardBgColor} borderRadius="md" borderWidth="1px" borderColor={borderColor}>
              <Text fontSize="sm" fontWeight="semibold" color={textColor} mb={2}>
                Progress Summary
              </Text>
              <VStack spacing={2} align="stretch">
                <HStack justify="space-between" fontSize="xs">
                  <Text color={subtextColor}>Average Progress:</Text>
                  <Text fontWeight="bold" color={textColor}>
                    {Math.round(
                      (progress.foundation_progress +
                        progress.utilities_progress +
                        progress.interior_progress +
                        progress.equipment_progress) /
                        4
                    )}
                    %
                  </Text>
                </HStack>
                <HStack justify="space-between" fontSize="xs">
                  <Text color={subtextColor}>Categories Completed (100%):</Text>
                  <Text fontWeight="bold" color={textColor}>
                    {
                      [
                        progress.foundation_progress,
                        progress.utilities_progress,
                        progress.interior_progress,
                        progress.equipment_progress,
                      ].filter((p) => p === 100).length
                    }{' '}
                    / 4
                  </Text>
                </HStack>
              </VStack>
            </Box>
          </VStack>
        </ModalBody>

        <ModalFooter borderTopWidth="1px" borderColor={borderColor}>
          <HStack spacing={3}>
            <Button variant="ghost" onClick={handleCancel} isDisabled={loading}>
              Cancel
            </Button>
            <Button
              colorScheme="green"
              leftIcon={<FiSave />}
              onClick={handleSave}
              isLoading={loading}
              loadingText="Saving..."
            >
              Save Progress
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default UpdateProgressModal;

