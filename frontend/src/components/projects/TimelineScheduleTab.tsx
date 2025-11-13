'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Badge,
  Card,
  CardBody,
  Icon,
  IconButton,
  useToast,
  Center,
  Spinner,
  useColorModeValue,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  SimpleGrid,
  Heading,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react';
import {
  FiPlus,
  FiEdit2,
  FiTrash2,
  FiMoreVertical,
  FiCalendar,
  FiList,
  FiClock,
  FiUsers,
} from 'react-icons/fi';
import projectService from '@/services/projectService';
import { TimelineSchedule } from '@/types/project';
import TimelineScheduleModal from './TimelineScheduleModal';
import { format, parseISO, eachWeekOfInterval, startOfYear, endOfYear, startOfWeek, endOfWeek } from 'date-fns';

interface TimelineScheduleTabProps {
  projectId: string;
}

export default function TimelineScheduleTab({ projectId }: TimelineScheduleTabProps) {
  const [schedules, setSchedules] = useState<TimelineSchedule[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedSchedule, setSelectedSchedule] = useState<TimelineSchedule | null>(null);
  const [viewMode, setViewMode] = useState<'list' | 'calendar'>('list');
  
  const toast = useToast();
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');

  useEffect(() => {
    fetchSchedules();
  }, [projectId]);

  const fetchSchedules = async () => {
    try {
      setLoading(true);
      const data = await projectService.getTimelineSchedules(projectId);
      setSchedules(data || []);
    } catch (error: any) {
      console.error('Error fetching schedules:', error);
      toast({
        title: 'Error',
        description: error.message || 'Failed to load schedules',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleAddSchedule = () => {
    setSelectedSchedule(null);
    setIsModalOpen(true);
  };

  const handleEditSchedule = (schedule: TimelineSchedule) => {
    setSelectedSchedule(schedule);
    setIsModalOpen(true);
  };

  const handleDeleteSchedule = async (scheduleId: string | number) => {
    if (!confirm('Are you sure you want to delete this schedule?')) return;

    try {
      await projectService.deleteTimelineSchedule(projectId, String(scheduleId));
      toast({
        title: 'Success',
        description: 'Schedule deleted successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      fetchSchedules();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to delete schedule',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setSelectedSchedule(null);
  };

  const handleModalSuccess = () => {
    fetchSchedules();
    handleModalClose();
  };

  const getStatusBadgeColor = (status: string) => {
    switch (status) {
      case 'not-started':
        return 'gray';
      case 'in-progress':
        return 'yellow';
      case 'completed':
        return 'green';
      default:
        return 'gray';
    }
  };

  const formatDate = (dateString: string) => {
    try {
      return format(parseISO(dateString), 'MMM dd, yyyy');
    } catch {
      return dateString;
    }
  };

  const renderListView = () => (
    <VStack spacing={4} align="stretch">
      {schedules.length === 0 ? (
        <Center py={10}>
          <VStack spacing={4}>
            <Icon as={FiCalendar} boxSize={16} color="gray.400" />
            <Text color="gray.500" fontSize="lg">
              No schedule items yet
            </Text>
            <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddSchedule}>
              Add First Schedule Item
            </Button>
          </VStack>
        </Center>
      ) : (
        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Work Area</Th>
              <Th>Assigned Team</Th>
              <Th>Schedule</Th>
              <Th>Duration</Th>
              <Th>Status</Th>
              <Th width="80px">Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {schedules.map((schedule) => (
              <Tr key={schedule.id}>
                <Td fontWeight="medium">{schedule.work_area}</Td>
                <Td>
                  <HStack>
                    <Icon as={FiUsers} color="blue.500" />
                    <Text>{schedule.assigned_team || 'Not Assigned'}</Text>
                  </HStack>
                </Td>
                <Td>
                  <VStack align="start" spacing={0}>
                    <Text fontSize="sm">
                      {formatDate(schedule.start_date)} - {formatDate(schedule.end_date)}
                    </Text>
                    <Text fontSize="xs" color="gray.500">
                      {schedule.start_time} - {schedule.end_time}
                    </Text>
                  </VStack>
                </Td>
                <Td>{schedule.duration ? `${schedule.duration} days` : '-'}</Td>
                <Td>
                  <Badge colorScheme={getStatusBadgeColor(schedule.status)}>
                    {schedule.status.replace('-', ' ').toUpperCase()}
                  </Badge>
                </Td>
                <Td>
                  <Menu>
                    <MenuButton
                      as={IconButton}
                      icon={<FiMoreVertical />}
                      variant="ghost"
                      size="sm"
                    />
                    <MenuList>
                      <MenuItem icon={<FiEdit2 />} onClick={() => handleEditSchedule(schedule)}>
                        Edit
                      </MenuItem>
                      <MenuItem
                        icon={<FiTrash2 />}
                        color="red.500"
                        onClick={() => handleDeleteSchedule(schedule.id)}
                      >
                        Delete
                      </MenuItem>
                    </MenuList>
                  </Menu>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      )}
    </VStack>
  );

  const renderCalendarView = () => {
    const currentYear = new Date().getFullYear();
    const weeks = eachWeekOfInterval({
      start: startOfYear(new Date(currentYear, 0, 1)),
      end: endOfYear(new Date(currentYear, 11, 31))
    }, { weekStartsOn: 1 }).slice(0, 12);

    return (
      <Box>
        <Box overflowX="auto">
          <Box minW="800px">
            {/* Header */}
            <HStack spacing={0} mb={4}>
              <Box width="200px" fontWeight="bold" px={4} py={2}>
                Phase
              </Box>
              {weeks.map((week, index) => (
                <Box key={index} flex="1" textAlign="center" fontSize="xs" px={1} py={2}>
                  Week {index + 1}
                  <br />
                  {format(week, 'MM/dd')}
                </Box>
              ))}
            </HStack>

            {/* Schedule Items */}
            {schedules.length === 0 ? (
              <Center py={10}>
                <VStack spacing={4}>
                  <Icon as={FiCalendar} boxSize={16} color="gray.400" />
                  <Text color="gray.500" fontSize="lg">
                    No schedule data available
                  </Text>
                  <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddSchedule}>
                    Add Schedule Item
                  </Button>
                </VStack>
              </Center>
            ) : (
              <VStack spacing={2} align="stretch">
                {schedules.map((schedule) => (
                  <HStack key={schedule.id} spacing={0} borderWidth="1px" borderColor={borderColor}>
                    <Box width="200px" px={4} py={2}>
                      <Text fontWeight="medium" fontSize="sm">
                        {schedule.work_area}
                      </Text>
                      <Badge size="sm" colorScheme={getStatusBadgeColor(schedule.status)}>
                        {schedule.status.replace('-', ' ')}
                      </Badge>
                    </Box>
                    <Box flex="1" position="relative" height="40px">
                      {/* This is a simplified calendar bar - in real implementation, 
                          calculate the actual position based on dates */}
                      <Box
                        position="absolute"
                        left="10%"
                        width="30%"
                        height="24px"
                        top="8px"
                        bg={schedule.status === 'completed' ? 'green.400' : schedule.status === 'in-progress' ? 'yellow.400' : 'gray.300'}
                        borderRadius="md"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                      >
                        <Text fontSize="xs" color="white" fontWeight="medium">
                          {schedule.duration} days
                        </Text>
                      </Box>
                    </Box>
                  </HStack>
                ))}
              </VStack>
            )}

            {/* Legend */}
            {schedules.length > 0 && (
              <HStack spacing={6} mt={6} justify="center">
                <HStack>
                  <Box width="16px" height="16px" bg="gray.300" borderRadius="sm" />
                  <Text fontSize="sm">Not Started</Text>
                </HStack>
                <HStack>
                  <Box width="16px" height="16px" bg="yellow.400" borderRadius="sm" />
                  <Text fontSize="sm">In Progress</Text>
                </HStack>
                <HStack>
                  <Box width="16px" height="16px" bg="green.400" borderRadius="sm" />
                  <Text fontSize="sm">Completed</Text>
                </HStack>
              </HStack>
            )}
          </Box>
        </Box>
      </Box>
    );
  };

  if (loading) {
    return (
      <Center h="400px">
        <Spinner size="xl" color="blue.500" />
      </Center>
    );
  }

  return (
    <Box>
      {/* Header */}
      <HStack justify="space-between" mb={6}>
        <Heading size="md">
          Timeline Schedule{' '}
          <Badge colorScheme="blue" fontSize="md">
            {schedules.length} {schedules.length === 1 ? 'item' : 'items'}
          </Badge>
        </Heading>
        <HStack>
          <HStack spacing={0} borderWidth="1px" borderRadius="md" borderColor={borderColor}>
            <IconButton
              icon={<FiList />}
              aria-label="List view"
              onClick={() => setViewMode('list')}
              variant={viewMode === 'list' ? 'solid' : 'ghost'}
              colorScheme={viewMode === 'list' ? 'blue' : 'gray'}
              borderRadius="0"
              borderRightWidth="1px"
            />
            <IconButton
              icon={<FiCalendar />}
              aria-label="Calendar view"
              onClick={() => setViewMode('calendar')}
              variant={viewMode === 'calendar' ? 'solid' : 'ghost'}
              colorScheme={viewMode === 'calendar' ? 'blue' : 'gray'}
              borderRadius="0"
            />
          </HStack>
          <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={handleAddSchedule}>
            Add Schedule Item
          </Button>
        </HStack>
      </HStack>

      {/* Content */}
      <Card bg={bgColor} borderColor={borderColor} borderWidth="1px">
        <CardBody>
          {viewMode === 'list' ? renderListView() : renderCalendarView()}
        </CardBody>
      </Card>

      {/* Modal */}
      <TimelineScheduleModal
        isOpen={isModalOpen}
        onClose={handleModalClose}
        onSuccess={handleModalSuccess}
        projectId={projectId}
        schedule={selectedSchedule}
      />
    </Box>
  );
}

