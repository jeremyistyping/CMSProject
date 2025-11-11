'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  Icon,
  useColorModeValue,
  useToast,
  IconButton,
  Input,
  Card,
  CardBody,
  Spinner,
  Center,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  useDisclosure,
} from '@chakra-ui/react';
import {
  FiPlus,
  FiCalendar,
  FiEdit,
  FiTrash2,
  FiMoreVertical,
  FiEye,
  FiSun,
  FiCloud,
  FiCloudRain,
} from 'react-icons/fi';
import projectService from '@/services/projectService';
import { DailyUpdate } from '@/types/project';
import DailyUpdateModal from './DailyUpdateModal';

interface DailyUpdatesTabProps {
  projectId: string;
}

const DailyUpdatesTab: React.FC<DailyUpdatesTabProps> = ({ projectId }) => {
  const toast = useToast();
  const { isOpen: isModalOpen, onOpen: onModalOpen, onClose: onModalClose } = useDisclosure();
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const cancelRef = React.useRef<HTMLButtonElement>(null);

  const [dailyUpdates, setDailyUpdates] = useState<DailyUpdate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedUpdate, setSelectedUpdate] = useState<DailyUpdate | null>(null);
  const [updateToDelete, setUpdateToDelete] = useState<string | null>(null);
  const [dateFilter, setDateFilter] = useState('');

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'var(--bg-primary)');

  useEffect(() => {
    fetchDailyUpdates();
  }, [projectId]);

  const fetchDailyUpdates = async () => {
    try {
      setLoading(true);
      const data = await projectService.getDailyUpdates(projectId);
      setDailyUpdates(data || []);
    } catch (error: any) {
      console.error('Error fetching daily updates:', error);
      
      // Backend not ready yet - show empty state instead of error
      if (error?.response?.status === 404) {
        console.log('Daily Updates endpoint not yet implemented in backend');
        setDailyUpdates([]);
      } else {
        toast({
          title: 'Error',
          description: 'Failed to load daily updates',
          status: 'error',
          duration: 3000,
        });
        setDailyUpdates([]);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleAddNew = () => {
    setSelectedUpdate(null);
    onModalOpen();
  };

  const handleEdit = (update: DailyUpdate) => {
    setSelectedUpdate(update);
    onModalOpen();
  };

  const handleDeleteClick = (updateId: string) => {
    setUpdateToDelete(updateId);
    onDeleteOpen();
  };

  const handleDeleteConfirm = async () => {
    if (!updateToDelete) return;

    try {
      await projectService.deleteDailyUpdate(projectId, updateToDelete);
      toast({
        title: 'Success',
        description: 'Daily update deleted successfully',
        status: 'success',
        duration: 3000,
      });
      fetchDailyUpdates();
    } catch (error: any) {
      // Backend not ready
      if (error?.response?.status === 404) {
        toast({
          title: 'Backend Not Ready',
          description: 'Daily Updates API not yet implemented. Please wait for backend setup.',
          status: 'warning',
          duration: 5000,
        });
      } else {
        toast({
          title: 'Error',
          description: 'Failed to delete daily update',
          status: 'error',
          duration: 3000,
        });
      }
    } finally {
      onDeleteClose();
      setUpdateToDelete(null);
    }
  };

  const handleModalSuccess = () => {
    fetchDailyUpdates();
    onModalClose();
  };

  const getWeatherIcon = (weather: string) => {
    const weatherLower = weather.toLowerCase();
    if (weatherLower.includes('sunny') || weatherLower.includes('clear')) return FiSun;
    if (weatherLower.includes('rain')) return FiCloudRain;
    return FiCloud;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const filteredUpdates = dateFilter
    ? dailyUpdates.filter((update) => update.date.startsWith(dateFilter))
    : dailyUpdates;

  if (loading) {
    return (
      <Center h="400px">
        <Spinner size="xl" color="green.500" />
      </Center>
    );
  }

  return (
    <Box>
      <VStack align="stretch" spacing={4}>
        {/* Header with Add Button and Date Filter */}
        <HStack justify="space-between" flexWrap="wrap" gap={4}>
          <HStack spacing={3}>
            <Icon as={FiCalendar} boxSize={5} color="green.500" />
            <Text fontSize="lg" fontWeight="semibold" color={textColor}>
              Daily Updates ({filteredUpdates.length})
            </Text>
          </HStack>

          <HStack spacing={3} flexWrap="wrap">
            <Input
              type="date"
              value={dateFilter}
              onChange={(e) => setDateFilter(e.target.value)}
              placeholder="Filter by date"
              w="200px"
              size="sm"
              bg={bgColor}
              borderColor={borderColor}
            />
            <Button
              leftIcon={<FiPlus />}
              colorScheme="green"
              size="sm"
              onClick={handleAddNew}
            >
              Add Daily Update
            </Button>
          </HStack>
        </HStack>

        {/* Daily Updates Table */}
        {filteredUpdates.length === 0 ? (
          <Card borderWidth="1px" borderColor={borderColor}>
            <CardBody>
              <Center h="300px">
                <VStack spacing={4}>
                  <Icon as={FiCalendar} boxSize={12} color="gray.400" />
                  <Text color={subtextColor} fontSize="lg">
                    No daily updates yet
                  </Text>
                  <Text color={subtextColor} fontSize="sm">
                    Click "Add Daily Update" to create your first entry
                  </Text>
                </VStack>
              </Center>
            </CardBody>
          </Card>
        ) : (
          <Card borderWidth="1px" borderColor={borderColor}>
            <Box overflowX="auto">
              <Table variant="simple">
                <Thead>
                  <Tr>
                    <Th color={subtextColor}>Date</Th>
                    <Th color={subtextColor}>Weather</Th>
                    <Th color={subtextColor}>Workers</Th>
                    <Th color={subtextColor}>Work Description</Th>
                    <Th color={subtextColor}>Issues</Th>
                    <Th color={subtextColor}>Actions</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {filteredUpdates.map((update) => (
                    <Tr key={update.id} _hover={{ bg: hoverBg }} transition="all 0.2s">
                      <Td>
                        <VStack align="start" spacing={0}>
                          <Text fontSize="sm" fontWeight="medium" color={textColor}>
                            {formatDate(update.date)}
                          </Text>
                          <Text fontSize="xs" color={subtextColor}>
                            {update.created_by || 'Unknown'}
                          </Text>
                        </VStack>
                      </Td>
                      <Td>
                        <HStack spacing={2}>
                          <Icon as={getWeatherIcon(update.weather)} color="orange.500" />
                          <Text fontSize="sm" color={textColor}>
                            {update.weather}
                          </Text>
                        </HStack>
                      </Td>
                      <Td>
                        <Badge colorScheme="green" fontSize="sm">
                          {update.workers_present} workers
                        </Badge>
                      </Td>
                      <Td maxW="300px">
                        <Text fontSize="sm" color={textColor} noOfLines={2}>
                          {update.work_description}
                        </Text>
                      </Td>
                      <Td maxW="200px">
                        {update.issues ? (
                          <Text fontSize="sm" color="red.500" noOfLines={2}>
                            {update.issues}
                          </Text>
                        ) : (
                          <Text fontSize="sm" color="green.500">
                            No issues
                          </Text>
                        )}
                      </Td>
                      <Td>
                        <Menu>
                          <MenuButton
                            as={IconButton}
                            icon={<FiMoreVertical />}
                            variant="ghost"
                            size="sm"
                            aria-label="Actions"
                          />
                          <MenuList>
                            <MenuItem
                              icon={<FiEdit />}
                              onClick={() => handleEdit(update)}
                            >
                              Edit
                            </MenuItem>
                            <MenuItem
                              icon={<FiTrash2 />}
                              color="red.500"
                              onClick={() => handleDeleteClick(update.id)}
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
            </Box>
          </Card>
        )}
      </VStack>

      {/* Daily Update Modal */}
      <DailyUpdateModal
        isOpen={isModalOpen}
        onClose={onModalClose}
        projectId={projectId}
        dailyUpdate={selectedUpdate}
        onSuccess={handleModalSuccess}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        isOpen={isDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent bg={bgColor}>
            <AlertDialogHeader fontSize="lg" fontWeight="bold" color={textColor}>
              Delete Daily Update
            </AlertDialogHeader>

            <AlertDialogBody color={textColor}>
              Are you sure? This action cannot be undone.
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteClose}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={handleDeleteConfirm} ml={3}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </Box>
  );
};

export default DailyUpdatesTab;

