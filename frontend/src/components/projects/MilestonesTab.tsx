'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  HStack,
  VStack,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Select,
  Input,
  InputGroup,
  InputLeftElement,
  SimpleGrid,
  Text,
  Spinner,
  useDisclosure,
  useToast,
  IconButton,
  Badge,
  Progress,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiPlus,
  FiSearch,
  FiFilter,
  FiList,
  FiTrendingUp,
  FiRefreshCw,
} from 'react-icons/fi';
import { useLanguage } from '@/contexts/LanguageContext';
import MilestoneCard from './MilestoneCard';
import MilestoneModal from './MilestoneModal';
import MilestoneTimeline from './MilestoneTimeline';

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
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [viewMode, setViewMode] = useState<'grid' | 'timeline'>('grid');

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    completed: 0,
    inProgress: 0,
    pending: 0,
    delayed: 0,
    completionRate: 0,
    weightedProgress: 0,
  });

  // Colors
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

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
      const milestonesData = data.data || [];
      setMilestones(milestonesData);
      calculateStats(milestonesData);
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

  // Calculate statistics
  const calculateStats = (milestonesData: any[]) => {
    const total = milestonesData.length;
    const completed = milestonesData.filter((m) => m.status === 'completed').length;
    const inProgress = milestonesData.filter((m) => m.status === 'in-progress').length;
    const pending = milestonesData.filter((m) => m.status === 'pending').length;
    const delayed = milestonesData.filter((m) => m.status === 'delayed').length;
    const completionRate = total > 0 ? (completed / total) * 100 : 0;

    // Calculate weighted progress
    const completedMilestones = milestonesData.filter((m) => m.status === 'completed');
    const weightedProgress = completedMilestones.reduce(
      (sum, m) => sum + (m.weight_percentage || 0),
      0
    );

    setStats({
      total,
      completed,
      inProgress,
      pending,
      delayed,
      completionRate,
      weightedProgress,
    });
  };

  // Filter milestones
  useEffect(() => {
    let filtered = [...milestones];

    // Filter by status
    if (statusFilter !== 'all') {
      filtered = filtered.filter((m) => m.status === statusFilter);
    }

    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (m) =>
          m.name.toLowerCase().includes(query) ||
          (m.description && m.description.toLowerCase().includes(query))
      );
    }

    setFilteredMilestones(filtered);
  }, [milestones, statusFilter, searchQuery]);

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
          method: 'PUT',
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
    <VStack align="stretch" spacing={4}>
      {/* Statistics Cards */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4}>
        <Box p={4} bg={bgColor} borderRadius="md" border="1px" borderColor={borderColor}>
          <Stat>
            <StatLabel>{t('milestones.total')}</StatLabel>
            <StatNumber>{stats.total}</StatNumber>
          </Stat>
        </Box>

        <Box p={4} bg={bgColor} borderRadius="md" border="1px" borderColor={borderColor}>
          <Stat>
            <StatLabel>{t('milestones.completed')}</StatLabel>
            <StatNumber color="green.500">{stats.completed}</StatNumber>
            <StatHelpText>{stats.completionRate.toFixed(0)}% {t('milestones.complete')}</StatHelpText>
          </Stat>
        </Box>

        <Box p={4} bg={bgColor} borderRadius="md" border="1px" borderColor={borderColor}>
          <Stat>
            <StatLabel>{t('milestones.inProgress')}</StatLabel>
            <StatNumber color="blue.500">{stats.inProgress}</StatNumber>
          </Stat>
        </Box>

        <Box p={4} bg={bgColor} borderRadius="md" border="1px" borderColor={borderColor}>
          <Stat>
            <StatLabel>{t('milestones.weightedProgress')}</StatLabel>
            <StatNumber color="purple.500">{stats.weightedProgress.toFixed(1)}%</StatNumber>
          </Stat>
        </Box>
      </SimpleGrid>

      {/* Overall Progress */}
      <Box p={4} bg={bgColor} borderRadius="md" border="1px" borderColor={borderColor}>
        <VStack align="stretch" spacing={2}>
          <HStack justify="space-between">
            <Text fontWeight="semibold">{t('milestones.overallProgress')}</Text>
            <Badge colorScheme="blue" fontSize="md">
              {stats.completionRate.toFixed(0)}%
            </Badge>
          </HStack>
          <Progress value={stats.completionRate} colorScheme="blue" size="lg" hasStripe />
        </VStack>
      </Box>

      {/* Toolbar */}
      <HStack justify="space-between" wrap="wrap" spacing={4}>
        <HStack spacing={4} flex={1}>
          {/* Search */}
          <InputGroup maxW="300px">
            <InputLeftElement pointerEvents="none">
              <FiSearch color="gray.300" />
            </InputLeftElement>
            <Input
              placeholder={t('milestones.search')}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </InputGroup>

          {/* Status Filter */}
          <Select
            maxW="200px"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            icon={<FiFilter />}
          >
            <option value="all">{t('milestones.allStatuses')}</option>
            <option value="pending">{t('milestones.statusPending')}</option>
            <option value="in-progress">{t('milestones.statusInProgress')}</option>
            <option value="completed">{t('milestones.statusCompleted')}</option>
            <option value="delayed">{t('milestones.statusDelayed')}</option>
          </Select>

          {/* Refresh Button */}
          <IconButton
            aria-label="Refresh"
            icon={<FiRefreshCw />}
            onClick={fetchMilestones}
            isLoading={loading}
          />
        </HStack>

        {/* Actions */}
        <HStack>
          {/* View Mode Toggle */}
          <HStack borderRadius="md" border="1px" borderColor={borderColor}>
            <Button
              size="sm"
              leftIcon={<FiList />}
              variant={viewMode === 'grid' ? 'solid' : 'ghost'}
              onClick={() => setViewMode('grid')}
            >
              {t('milestones.gridView')}
            </Button>
            <Button
              size="sm"
              leftIcon={<FiTrendingUp />}
              variant={viewMode === 'timeline' ? 'solid' : 'ghost'}
              onClick={() => setViewMode('timeline')}
            >
              {t('milestones.timelineView')}
            </Button>
          </HStack>

          {/* Add Milestone Button */}
          <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddMilestone}>
            {t('milestones.addMilestone')}
          </Button>
        </HStack>
      </HStack>

      {/* Content */}
      {loading ? (
        <Box py={12} textAlign="center">
          <Spinner size="xl" />
        </Box>
      ) : filteredMilestones.length === 0 ? (
        <Box
          p={12}
          textAlign="center"
          bg={bgColor}
          borderRadius="md"
          border="1px"
          borderColor={borderColor}
        >
          <Text color="gray.500" fontSize="lg">
            {searchQuery || statusFilter !== 'all'
              ? t('milestones.noFilteredMilestones')
              : t('milestones.noMilestones')}
          </Text>
          {!searchQuery && statusFilter === 'all' && (
            <Button mt={4} leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddMilestone}>
              {t('milestones.addFirstMilestone')}
            </Button>
          )}
        </Box>
      ) : viewMode === 'grid' ? (
        <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={4}>
          {filteredMilestones.map((milestone) => (
            <MilestoneCard
              key={milestone.id}
              milestone={milestone}
              onEdit={handleEditMilestone}
              onDelete={handleDeleteMilestone}
              onComplete={handleCompleteMilestone}
            />
          ))}
        </SimpleGrid>
      ) : (
        <Box
          p={6}
          bg={bgColor}
          borderRadius="md"
          border="1px"
          borderColor={borderColor}
        >
          <MilestoneTimeline
            milestones={filteredMilestones}
            onMilestoneClick={handleEditMilestone}
          />
        </Box>
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

