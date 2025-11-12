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
  FiImage,
  FiChevronLeft,
  FiChevronRight,
  FiDownload,
} from 'react-icons/fi';
import projectService from '@/services/projectService';
import { DailyUpdate, Project } from '@/types/project';
import DailyUpdateModal from './DailyUpdateModal';
import DailyUpdateViewModal from './DailyUpdateViewModal';
import PhotoGallery from './PhotoGallery';
import { generateDailyUpdatesPDF } from '@/utils/dailyUpdatesPdfExport';

interface DailyUpdatesTabProps {
  projectId: string;
  project?: Project;
}

const DailyUpdatesTab: React.FC<DailyUpdatesTabProps> = ({ projectId, project }) => {
  const toast = useToast();
  const { isOpen: isModalOpen, onOpen: onModalOpen, onClose: onModalClose } = useDisclosure();
  const { isOpen: isViewOpen, onOpen: onViewOpen, onClose: onViewClose } = useDisclosure();
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const { isOpen: isPhotoGalleryOpen, onOpen: onPhotoGalleryOpen, onClose: onPhotoGalleryClose } = useDisclosure();
  const cancelRef = React.useRef<HTMLButtonElement>(null);

  const [dailyUpdates, setDailyUpdates] = useState<DailyUpdate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedUpdate, setSelectedUpdate] = useState<DailyUpdate | null>(null);
  const [viewUpdate, setViewUpdate] = useState<DailyUpdate | null>(null);
  const [updateToDelete, setUpdateToDelete] = useState<string | null>(null);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [selectedPhotos, setSelectedPhotos] = useState<string[]>([]);
  const scrollContainerRef = React.useRef<HTMLDivElement>(null);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const hoverBg = useColorModeValue('gray.50', 'var(--bg-primary)');

  useEffect(() => {
    fetchDailyUpdates();
  }, [projectId, startDate, endDate]);

  const fetchDailyUpdates = async () => {
    try {
      setLoading(true);
      const data = await projectService.getDailyUpdates(projectId, startDate, endDate);
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

  const handleView = (update: DailyUpdate) => {
    setViewUpdate(update);
    onViewOpen();
  };

  const handleEdit = (update: DailyUpdate) => {
    setSelectedUpdate(update);
    onModalOpen();
  };

  const handleEditFromView = () => {
    if (viewUpdate) {
      setSelectedUpdate(viewUpdate);
      setViewUpdate(null);
      onModalOpen();
    }
  };

  const handleViewPhotos = (photos: string[]) => {
    setSelectedPhotos(photos);
    onPhotoGalleryOpen();
  };

  const scroll = (direction: 'left' | 'right') => {
    if (scrollContainerRef.current) {
      const scrollAmount = 400;
      const newScrollPosition = scrollContainerRef.current.scrollLeft + (direction === 'right' ? scrollAmount : -scrollAmount);
      scrollContainerRef.current.scrollTo({
        left: newScrollPosition,
        behavior: 'smooth'
      });
    }
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

  const handleClearFilters = () => {
    setStartDate('');
    setEndDate('');
  };

  const handleExportPDF = async () => {
    if (!project) {
      toast({
        title: 'Error',
        description: 'Project information not available',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    if (dailyUpdates.length === 0) {
      toast({
        title: 'No Data',
        description: 'No daily updates to export',
        status: 'warning',
        duration: 3000,
      });
      return;
    }

    try {
      await generateDailyUpdatesPDF({
        projectInfo: {
          name: project.project_name,
          location: project.address || project.city,
          client: project.customer,
          status: project.status,
          progress: project.overall_progress || 0,
        },
        dailyUpdates: dailyUpdates,
        dateRange: {
          startDate: startDate,
          endDate: endDate,
        },
      });

      toast({
        title: 'Success',
        description: 'PDF report generated successfully',
        status: 'success',
        duration: 3000,
      });
    } catch (error) {
      console.error('Error generating PDF:', error);
      toast({
        title: 'Error',
        description: 'Failed to generate PDF report',
        status: 'error',
        duration: 3000,
      });
    }
  };

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
        {/* Header with Add Button and Date Range Filter */}
        <HStack justify="space-between" flexWrap="wrap" gap={4}>
          <HStack spacing={3}>
            <Icon as={FiCalendar} boxSize={5} color="green.500" />
            <Text fontSize="lg" fontWeight="semibold" color={textColor}>
              Daily Updates ({dailyUpdates.length})
            </Text>
          </HStack>

          <HStack spacing={3} flexWrap="wrap">
            <HStack spacing={2}>
              <Input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                placeholder="Start date"
                w="150px"
                size="sm"
                bg={bgColor}
                borderColor={borderColor}
              />
              <Text fontSize="sm" color={subtextColor}>to</Text>
              <Input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                placeholder="End date"
                w="150px"
                size="sm"
                bg={bgColor}
                borderColor={borderColor}
              />
              {(startDate || endDate) && (
                <Button
                  size="sm"
                  variant="ghost"
                  colorScheme="gray"
                  onClick={handleClearFilters}
                >
                  Clear
                </Button>
              )}
            </HStack>
            <Button
              leftIcon={<FiDownload />}
              colorScheme="blue"
              size="sm"
              onClick={handleExportPDF}
              isDisabled={dailyUpdates.length === 0}
            >
              Export PDF
            </Button>
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

        {/* Daily Updates Cards with Horizontal Scroll */}
        {dailyUpdates.length === 0 ? (
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
          <Box position="relative">
            {/* Navigation Buttons */}
            {dailyUpdates.length > 1 && (
              <>
                <IconButton
                  icon={<FiChevronLeft />}
                  aria-label="Scroll left"
                  position="absolute"
                  left="-4"
                  top="50%"
                  transform="translateY(-50%)"
                  zIndex={2}
                  size="lg"
                  colorScheme="green"
                  variant="solid"
                  boxShadow="lg"
                  onClick={() => scroll('left')}
                />
                <IconButton
                  icon={<FiChevronRight />}
                  aria-label="Scroll right"
                  position="absolute"
                  right="-4"
                  top="50%"
                  transform="translateY(-50%)"
                  zIndex={2}
                  size="lg"
                  colorScheme="green"
                  variant="solid"
                  boxShadow="lg"
                  onClick={() => scroll('right')}
                />
              </>
            )}

            {/* Scrollable Cards Container */}
            <Box
              ref={scrollContainerRef}
              overflowX="auto"
              overflowY="hidden"
              css={{
                '&::-webkit-scrollbar': {
                  height: '8px',
                },
                '&::-webkit-scrollbar-track': {
                  background: 'transparent',
                },
                '&::-webkit-scrollbar-thumb': {
                  background: '#48BB78',
                  borderRadius: '4px',
                },
                '&::-webkit-scrollbar-thumb:hover': {
                  background: '#38A169',
                },
              }}
              pb={4}
            >
              <HStack spacing={4} align="stretch" minH="320px">
                {dailyUpdates.map((update) => (
                  <Card
                    key={update.id}
                    minW="380px"
                    maxW="380px"
                    borderWidth="1px"
                    borderColor={borderColor}
                    bg={bgColor}
                    transition="all 0.3s"
                    _hover={{
                      boxShadow: 'xl',
                      transform: 'translateY(-4px)',
                      borderColor: 'green.400',
                    }}
                  >
                    <CardBody>
                      <VStack align="stretch" spacing={4}>
                        {/* Header with Date and Actions */}
                        <HStack justify="space-between">
                          <VStack align="start" spacing={1}>
                            <HStack>
                              <Icon as={FiCalendar} color="green.500" />
                              <Text fontSize="md" fontWeight="bold" color={textColor}>
                                {formatDate(update.date)}
                              </Text>
                            </HStack>
                            <Text fontSize="xs" color={subtextColor}>
                              By: {update.created_by || 'Unknown'}
                            </Text>
                          </VStack>
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
                                icon={<FiEye />}
                                onClick={() => handleView(update)}
                              >
                                View Details
                              </MenuItem>
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
                        </HStack>

                        {/* Weather and Workers */}
                        <HStack spacing={4}>
                          <HStack spacing={2} flex={1}>
                            <Icon as={getWeatherIcon(update.weather)} color="orange.500" boxSize={5} />
                            <Text fontSize="sm" color={textColor}>
                              {update.weather}
                            </Text>
                          </HStack>
                          <Badge colorScheme="green" fontSize="sm" px={3} py={1}>
                            {update.workers_present} workers
                          </Badge>
                        </HStack>

                        {/* Work Description */}
                        <Box>
                          <Text fontSize="xs" fontWeight="semibold" color={subtextColor} mb={1}>
                            Work Description:
                          </Text>
                          <Text fontSize="sm" color={textColor} noOfLines={3}>
                            {update.work_description}
                          </Text>
                        </Box>

                        {/* Photos */}
                        <Box>
                          <Text fontSize="xs" fontWeight="semibold" color={subtextColor} mb={1}>
                            Photos:
                          </Text>
                          {update.photos && update.photos.length > 0 ? (
                            <HStack 
                              spacing={2}
                              p={2}
                              borderWidth="1px"
                              borderColor={borderColor}
                              borderRadius="md"
                              cursor="pointer"
                              onClick={() => handleViewPhotos(update.photos)}
                              _hover={{ bg: hoverBg, borderColor: 'blue.400' }}
                              transition="all 0.2s"
                            >
                              <Icon as={FiImage} color="blue.500" boxSize={5} />
                              <Badge colorScheme="blue" fontSize="sm">
                                {update.photos.length} {update.photos.length === 1 ? 'photo' : 'photos'}
                              </Badge>
                              <Text fontSize="xs" color="blue.500" ml="auto">
                                View →
                              </Text>
                            </HStack>
                          ) : (
                            <Text fontSize="sm" color="gray.400" fontStyle="italic">
                              No photos
                            </Text>
                          )}
                        </Box>

                        {/* Issues */}
                        <Box>
                          <Text fontSize="xs" fontWeight="semibold" color={subtextColor} mb={1}>
                            Issues:
                          </Text>
                          {update.issues ? (
                            <Text fontSize="sm" color="red.500" noOfLines={2}>
                              {update.issues}
                            </Text>
                          ) : (
                            <Text fontSize="sm" color="green.500" fontStyle="italic">
                              No issues reported
                            </Text>
                          )}
                        </Box>
                      </VStack>
                    </CardBody>
                  </Card>
                ))}
              </HStack>
            </Box>
          </Box>
        )}
      </VStack>

      {/* Daily Update Modal (Edit/Create) */}
      <DailyUpdateModal
        isOpen={isModalOpen}
        onClose={onModalClose}
        projectId={projectId}
        dailyUpdate={selectedUpdate}
        onSuccess={handleModalSuccess}
      />

      {/* Daily Update View Modal (Read-only) */}
      <DailyUpdateViewModal
        isOpen={isViewOpen}
        onClose={onViewClose}
        dailyUpdate={viewUpdate}
        onEdit={handleEditFromView}
      />

      {/* Photo Gallery Modal */}
      <PhotoGallery
        isOpen={isPhotoGalleryOpen}
        onClose={onPhotoGalleryClose}
        photos={selectedPhotos}
        title="Daily Update Photos"
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

