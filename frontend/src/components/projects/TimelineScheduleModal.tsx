'use client';

import React, { useState, useEffect } from 'react';
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
  VStack,
  useToast,
  Grid,
  GridItem,
} from '@chakra-ui/react';
import projectService from '@/services/projectService';
import { TimelineSchedule } from '@/types/project';

interface TimelineScheduleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  projectId: string;
  schedule: TimelineSchedule | null;
}

export default function TimelineScheduleModal({
  isOpen,
  onClose,
  onSuccess,
  projectId,
  schedule,
}: TimelineScheduleModalProps) {
  const [formData, setFormData] = useState({
    work_area: '',
    assigned_team: '',
    start_date: '',
    end_date: '',
    start_time: '08:00',
    end_time: '17:00',
    notes: '',
    status: 'not-started' as 'not-started' | 'in-progress' | 'completed',
  });
  const [loading, setLoading] = useState(false);
  const toast = useToast();

  // Common work areas for suggestions
  const workAreaOptions = [
    'Site Preparation',
    'Foundation Work',
    'Structural Work',
    'Roofing Work',
    'Mechanical Work',
    'Electrical Work',
    'Plumbing Work',
    'Interior Work',
    'Exterior Finishes',
    'Landscaping',
    'Final Inspection',
  ];

  useEffect(() => {
    if (schedule) {
      // Edit mode - populate form with existing data
      setFormData({
        work_area: schedule.work_area || '',
        assigned_team: schedule.assigned_team || '',
        start_date: schedule.start_date ? schedule.start_date.split('T')[0] : '',
        end_date: schedule.end_date ? schedule.end_date.split('T')[0] : '',
        start_time: schedule.start_time || '08:00',
        end_time: schedule.end_time || '17:00',
        notes: schedule.notes || '',
        status: schedule.status || 'not-started',
      });
    } else {
      // Add mode - reset form
      setFormData({
        work_area: '',
        assigned_team: '',
        start_date: '',
        end_date: '',
        start_time: '08:00',
        end_time: '17:00',
        notes: '',
        status: 'not-started',
      });
    }
  }, [schedule, isOpen]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!formData.work_area.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Work area is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!formData.start_date || !formData.end_date) {
      toast({
        title: 'Validation Error',
        description: 'Start date and end date are required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (new Date(formData.end_date) < new Date(formData.start_date)) {
      toast({
        title: 'Validation Error',
        description: 'End date must be after or equal to start date',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setLoading(true);

    try {
      // Prepare data for API
      const scheduleData = {
        work_area: formData.work_area.trim(),
        assigned_team: formData.assigned_team.trim(),
        start_date: `${formData.start_date}T${formData.start_time}:00Z`,
        end_date: `${formData.end_date}T${formData.end_time}:00Z`,
        start_time: formData.start_time,
        end_time: formData.end_time,
        notes: formData.notes.trim(),
        status: formData.status,
      };

      if (schedule) {
        // Update existing schedule
        await projectService.updateTimelineSchedule(projectId, String(schedule.id), scheduleData);
        toast({
          title: 'Success',
          description: 'Schedule updated successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else {
        // Create new schedule
        await projectService.createTimelineSchedule(projectId, scheduleData);
        toast({
          title: 'Success',
          description: 'Schedule created successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }

      onSuccess();
    } catch (error: any) {
      console.error('Error saving schedule:', error);
      toast({
        title: 'Error',
        description: error.message || 'Failed to save schedule',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl" isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>{schedule ? 'Edit Schedule Item' : 'Add New Schedule Item'}</ModalHeader>
        <ModalCloseButton />
        <form onSubmit={handleSubmit}>
          <ModalBody>
            <VStack spacing={4}>
              {/* Work Area */}
              <FormControl isRequired>
                <FormLabel>Work Area</FormLabel>
                <Input
                  name="work_area"
                  value={formData.work_area}
                  onChange={handleChange}
                  placeholder="Select or enter work area"
                  list="work-area-options"
                />
                <datalist id="work-area-options">
                  {workAreaOptions.map((option) => (
                    <option key={option} value={option} />
                  ))}
                </datalist>
              </FormControl>

              {/* Assigned Team */}
              <FormControl>
                <FormLabel>Assigned Team</FormLabel>
                <Input
                  name="assigned_team"
                  value={formData.assigned_team}
                  onChange={handleChange}
                  placeholder="Team or contractor name"
                />
              </FormControl>

              {/* Date Range */}
              <Grid templateColumns="repeat(2, 1fr)" gap={4} width="full">
                <GridItem>
                  <FormControl isRequired>
                    <FormLabel>Start Date</FormLabel>
                    <Input
                      type="date"
                      name="start_date"
                      value={formData.start_date}
                      onChange={handleChange}
                    />
                  </FormControl>
                </GridItem>
                <GridItem>
                  <FormControl isRequired>
                    <FormLabel>End Date</FormLabel>
                    <Input
                      type="date"
                      name="end_date"
                      value={formData.end_date}
                      onChange={handleChange}
                    />
                  </FormControl>
                </GridItem>
              </Grid>

              {/* Time Range */}
              <Grid templateColumns="repeat(2, 1fr)" gap={4} width="full">
                <GridItem>
                  <FormControl>
                    <FormLabel>Start Time</FormLabel>
                    <Input
                      type="time"
                      name="start_time"
                      value={formData.start_time}
                      onChange={handleChange}
                    />
                  </FormControl>
                </GridItem>
                <GridItem>
                  <FormControl>
                    <FormLabel>End Time</FormLabel>
                    <Input
                      type="time"
                      name="end_time"
                      value={formData.end_time}
                      onChange={handleChange}
                    />
                  </FormControl>
                </GridItem>
              </Grid>

              {/* Status */}
              <FormControl>
                <FormLabel>Status</FormLabel>
                <Select name="status" value={formData.status} onChange={handleChange}>
                  <option value="not-started">Not Started</option>
                  <option value="in-progress">In Progress</option>
                  <option value="completed">Completed</option>
                </Select>
              </FormControl>

              {/* Notes */}
              <FormControl>
                <FormLabel>Notes</FormLabel>
                <Textarea
                  name="notes"
                  value={formData.notes}
                  onChange={handleChange}
                  placeholder="Additional notes or requirements"
                  rows={3}
                />
              </FormControl>
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose} isDisabled={loading}>
              Cancel
            </Button>
            <Button colorScheme="blue" type="submit" isLoading={loading}>
              {schedule ? 'Update' : 'Add'} Schedule
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
}

