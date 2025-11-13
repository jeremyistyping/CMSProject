'use client';

import React, { useState, useEffect } from 'react';
import {
  Button,
  HStack,
  VStack,
  Select,
  Text,
  Spinner,
  useDisclosure,
  useToast,
  Center,
  Icon,
} from '@chakra-ui/react';
import {
  FiPlus,
  FiTarget,
} from 'react-icons/fi';
import { useLanguage } from '@/contexts/LanguageContext';
import MilestoneCard from './MilestoneCard';
import MilestoneModal from './MilestoneModal';

interface MilestonesTabProps {
  projectId: number;
}

const MilestonesTab: React.FC<MilestonesTabProps> = ({ projectId }) => {
  const { t } = useLanguage();
  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

  // State
  const [milestones, setMilestones] = useState<any[]>([]);
  const [filteredMilestones, setFilteredMilestones] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedMilestone, setSelectedMilestone] = useState<any>(null);
  const [statusFilter, setStatusFilter] = useState('all');
  const [priorityFilter, setPriorityFilter] = useState('all');

  // Fetch milestones
  const fetchMilestones = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/projects/${projectId}/milestones`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch milestones');
      }

      const data = await response.json();
      // Backend returns array directly, not wrapped in { data: ... }
      const milestonesData = Array.isArray(data) ? data : (data.data || []);
      setMilestones(milestonesData);
    } catch (error: any) {
      console.error('Error fetching milestones:', error);
      toast({
        title: t('milestones.error'),
        description: error.message || t('milestones.fetchError'),
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };


  // Filter milestones
  useEffect(() => {
    let filtered = [...milestones];

    // Filter by status
    if (statusFilter !== 'all') {
      filtered = filtered.filter((m) => m.status === statusFilter);
    }

    // Filter by priority
    if (priorityFilter !== 'all') {
      filtered = filtered.filter((m) => m.priority === priorityFilter);
    }

    setFilteredMilestones(filtered);
  }, [milestones, statusFilter, priorityFilter]);

  // Initial fetch
  useEffect(() => {
    fetchMilestones();
  }, [projectId]);

  // Handle add milestone
  const handleAddMilestone = () => {
    setSelectedMilestone(null);
    onOpen();
  };

  // Handle edit milestone
  const handleEditMilestone = (milestone: any) => {
    setSelectedMilestone(milestone);
    onOpen();
  };

  // Handle delete milestone
  const handleDeleteMilestone = async (milestoneId: number) => {
    if (!confirm(t('milestones.confirmDelete'))) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `/api/v1/projects/${projectId}/milestones/${milestoneId}`,
        {
          method: 'DELETE',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Failed to delete milestone');
      }

      toast({
        title: t('milestones.deleteSuccess'),
        description: t('milestones.deleteSuccessDesc'),
        status: 'success',
        duration: 3000,
        isClosable: true,
      });

      fetchMilestones();
    } catch (error: any) {
      console.error('Error deleting milestone:', error);
      toast({
        title: t('milestones.error'),
        description: error.message || t('milestones.deleteError'),
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle complete milestone
  const handleCompleteMilestone = async (milestoneId: number) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `/api/v1/projects/${projectId}/milestones/${milestoneId}/complete`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Failed to complete milestone');
      }

      toast({
        title: t('milestones.completeSuccess'),
        description: t('milestones.completeSuccessDesc'),
        status: 'success',
        duration: 3000,
        isClosable: true,
      });

      fetchMilestones();
    } catch (error: any) {
      console.error('Error completing milestone:', error);
      toast({
        title: t('milestones.error'),
        description: error.message || t('milestones.completeError'),
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle modal success
  const handleModalSuccess = () => {
    fetchMilestones();
  };

  return (
    <VStack align="stretch" spacing={6}>
      {/* Toolbar */}
      <HStack justify="space-between" wrap="wrap" spacing={4}>
        <HStack spacing={3}>
          {/* Status Filter */}
          <Select
            w="200px"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="in-progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="delayed">Delayed</option>
          </Select>

          {/* Priority Filter */}
          <Select
            w="200px"
            value={priorityFilter}
            onChange={(e) => setPriorityFilter(e.target.value)}
          >
            <option value="all">Priority</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </Select>
        </HStack>

        {/* Add Milestone Button */}
        <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddMilestone}>
          Add Milestone
        </Button>
      </HStack>

      {/* Content */}
      {loading ? (
        <Center h="300px">
          <Spinner size="xl" color="blue.500" />
        </Center>
      ) : filteredMilestones.length === 0 ? (
        <Center h="400px" flexDirection="column">
          <Icon as={FiTarget} boxSize={16} color="gray.400" mb={4} />
          <Text color="gray.500" fontSize="lg" mb={2}>
            No milestones match your filter
          </Text>
          <Text color="gray.400" fontSize="sm" mb={6}>
            Try adjusting your filter or add new milestones
          </Text>
          {statusFilter === 'all' && priorityFilter === 'all' && (
            <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddMilestone}>
              Add First Milestone
            </Button>
          )}
        </Center>
      ) : (
        <VStack align="stretch" spacing={3}>
          {filteredMilestones.map((milestone) => (
            <MilestoneCard
              key={milestone.id}
              milestone={milestone}
              onEdit={handleEditMilestone}
              onDelete={handleDeleteMilestone}
              onComplete={handleCompleteMilestone}
            />
          ))}
        </VStack>
      )}

      {/* Milestone Modal */}
      <MilestoneModal
        isOpen={isOpen}
        onClose={onClose}
        projectId={projectId}
        milestone={selectedMilestone}
        onSuccess={handleModalSuccess}
      />
    </VStack>
  );
};

export default MilestonesTab;

