'use client';

import React from 'react';
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
  Badge,
  Icon,
  useColorModeValue,
  Box,
  Grid,
  GridItem,
  Image,
  SimpleGrid,
  Divider,
} from '@chakra-ui/react';
import {
  FiCalendar,
  FiSun,
  FiCloud,
  FiCloudRain,
  FiUsers,
  FiPackage,
  FiAlertTriangle,
  FiImage,
  FiUser,
  FiTrendingUp,
  FiDatabase,
  FiTarget,
  FiFileText,
  FiTool,
  FiCheckCircle,
  FiXCircle,
} from 'react-icons/fi';
import { DailyUpdate } from '@/types/project';
import { usePermissions } from '@/hooks/usePermissions';
import { projectService } from '@/services/projectService';
import { useToast, Textarea, Collapse } from '@chakra-ui/react';
import { useState } from 'react';

interface DailyUpdateViewModalProps {
  isOpen: boolean;
  onClose: () => void;
  dailyUpdate: DailyUpdate | null;
  dailyUpdate: DailyUpdate | null;
  onEdit?: () => void;
  onStatusChange?: () => void;
}

const DailyUpdateViewModal: React.FC<DailyUpdateViewModalProps> = ({
  isOpen,
  onClose,
  dailyUpdate,
  onEdit,
  onStatusChange,
}) => {
  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const sectionBg = useColorModeValue('gray.50', 'var(--bg-primary)');

  const { canApprove } = usePermissions();
  const toast = useToast();
  const [isProcessing, setIsProcessing] = useState(false);
  const [showRejectionInput, setShowRejectionInput] = useState(false);
  const [rejectionReason, setRejectionReason] = useState('');

  if (!dailyUpdate) return null;

  const handleApprove = async () => {
    if (!dailyUpdate) return;
    try {
      setIsProcessing(true);
      await projectService.approveDailyUpdate(dailyUpdate.project_id, dailyUpdate.id);
      toast({
        title: 'Daily Update Approved',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      if (onStatusChange) onStatusChange();
      onClose();
    } catch (error: any) {
      toast({
        title: 'Error approving update',
        description: error.message || 'Something went wrong',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const handleReject = async () => {
    if (!dailyUpdate) return;
    if (!rejectionReason.trim()) {
      toast({
        title: 'Rejection reason required',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }
    try {
      setIsProcessing(true);
      await projectService.rejectDailyUpdate(dailyUpdate.project_id, dailyUpdate.id, rejectionReason);
      toast({
        title: 'Daily Update Rejected',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      if (onStatusChange) onStatusChange();
      onClose();
    } catch (error: any) {
      toast({
        title: 'Error rejecting update',
        description: error.message || 'Something went wrong',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const getWeatherIcon = (weather: string) => {
    const weatherLower = weather.toLowerCase();
    if (weatherLower.includes('sunny') || weatherLower.includes('clear')) return FiSun;
    if (weatherLower.includes('rain')) return FiCloudRain;
    return FiCloud;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Convert relative URLs to full URLs
  // Photos are served from backend (port 8080), not frontend (port 3000)
  const getFullUrl = (url: string) => {
    if (!url) return '';
    if (url.startsWith('http')) return url;

    // Normalize path - convert Windows backslashes to forward slashes
    let normalizedUrl = url.replace(/\\/g, '/');

    // Ensure leading slash
    if (!normalizedUrl.startsWith('/')) {
      normalizedUrl = '/' + normalizedUrl;
    }

    // Use backend URL for uploaded files
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    return `${backendUrl}${normalizedUrl}`;
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="4xl" isCentered scrollBehavior="inside">
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
          <HStack spacing={3}>
            <Icon as={FiCalendar} color="green.500" boxSize={6} />
            <VStack align="start" spacing={0}>
              <Text fontSize="lg" fontWeight="bold">
                Daily Update Details
              </Text>
              <Text fontSize="sm" fontWeight="normal" color={subtextColor}>
                {formatDate(dailyUpdate.date)}
              </Text>
            </VStack>
            {/* Status Badge */}
            <Box ml="auto">
              {dailyUpdate.status === 'approved' && (
                <Badge colorScheme="green" fontSize="md" px={3} py={1} borderRadius="full">
                  <HStack spacing={1}>
                    <Icon as={FiCheckCircle} />
                    <Text>Approved{dailyUpdate.approved_by && ` by ${dailyUpdate.approved_by === 'Unknown' ? 'Purchasing Admin' : dailyUpdate.approved_by}`}</Text>
                  </HStack>
                </Badge>
              )}
              {dailyUpdate.status === 'rejected' && (
                <Badge colorScheme="red" fontSize="md" px={3} py={1} borderRadius="full">
                  <HStack spacing={1}>
                    <Icon as={FiXCircle} />
                    <Text>Rejected{dailyUpdate.approved_by && ` by ${dailyUpdate.approved_by === 'Unknown' ? 'Purchasing Admin' : dailyUpdate.approved_by}`}</Text>
                  </HStack>
                </Badge>
              )}
              {(!dailyUpdate.status || dailyUpdate.status === 'pending') && (
                <Badge colorScheme="yellow" fontSize="md" px={3} py={1} borderRadius="full">
                  Pending Approval
                </Badge>
              )}
            </Box>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />

        <ModalBody py={6} bg={bgColor}>
          <VStack align="stretch" spacing={6}>
            {/* Date & Weather Section */}
            <Box bg={sectionBg} p={4} borderRadius="md" borderWidth="1px" borderColor={borderColor}>
              <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                <GridItem>
                  <HStack spacing={2}>
                    <Icon as={FiCalendar} color="green.500" boxSize={5} />
                    <VStack align="start" spacing={0}>
                      <Text fontSize="xs" color={subtextColor} fontWeight="semibold">
                        Date
                      </Text>
                      <Text fontSize="md" color={textColor} fontWeight="medium">
                        {formatDate(dailyUpdate.date)}
                      </Text>
                    </VStack>
                  </HStack>
                </GridItem>
                <GridItem>
                  <HStack spacing={2}>
                    <Icon as={getWeatherIcon(dailyUpdate.weather)} color="orange.500" boxSize={5} />
                    <VStack align="start" spacing={0}>
                      <Text fontSize="xs" color={subtextColor} fontWeight="semibold">
                        Weather
                      </Text>
                      <Text fontSize="md" color={textColor} fontWeight="medium">
                        {dailyUpdate.weather}
                      </Text>
                    </VStack>
                  </HStack>
                </GridItem>
              </Grid>
            </Box>

            {/* Workers Section */}
            <Box>
              <HStack spacing={2} mb={3}>
                <Icon as={FiUsers} color="blue.500" boxSize={5} />
                <Text fontSize="sm" fontWeight="bold" color={textColor}>
                  Workers Present
                </Text>
              </HStack>
              <Badge colorScheme="green" fontSize="lg" px={4} py={2} borderRadius="md">
                {dailyUpdate.workers_present} Workers
              </Badge>
            </Box>

            <Divider />

            {/* Project Progress Section */}
            <Box>
              <HStack spacing={2} mb={3}>
                <Icon as={FiTrendingUp} color="green.500" boxSize={5} />
                <Text fontSize="sm" fontWeight="bold" color={textColor}>
                  Project Progress
                </Text>
              </HStack>
              <SimpleGrid columns={{ base: 2, md: 2 }} spacing={4}>
                {/* Foundation Progress */}
                <Box
                  bg={sectionBg}
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                >
                  <HStack spacing={2} mb={2}>
                    <Icon as={FiDatabase} color="orange.500" boxSize={4} />
                    <Text fontSize="xs" color={subtextColor} fontWeight="semibold">
                      Foundation & Structure
                    </Text>
                  </HStack>
                  <Text fontSize="2xl" fontWeight="bold" color="orange.500">
                    {dailyUpdate.foundation_progress || 0}%
                  </Text>
                </Box>

                {/* Utilities Progress */}
                <Box
                  bg={sectionBg}
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                >
                  <HStack spacing={2} mb={2}>
                    <Icon as={FiTarget} color="purple.500" boxSize={4} />
                    <Text fontSize="xs" color={subtextColor} fontWeight="semibold">
                      Utilities Installation
                    </Text>
                  </HStack>
                  <Text fontSize="2xl" fontWeight="bold" color="purple.500">
                    {dailyUpdate.utilities_progress || 0}%
                  </Text>
                </Box>

                {/* Interior Progress */}
                <Box
                  bg={sectionBg}
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                >
                  <HStack spacing={2} mb={2}>
                    <Icon as={FiFileText} color="pink.500" boxSize={4} />
                    <Text fontSize="xs" color={subtextColor} fontWeight="semibold">
                      Interior & Finishes
                    </Text>
                  </HStack>
                  <Text fontSize="2xl" fontWeight="bold" color="pink.500">
                    {dailyUpdate.interior_progress || 0}%
                  </Text>
                </Box>

                {/* Equipment Progress */}
                <Box
                  bg={sectionBg}
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                >
                  <HStack spacing={2} mb={2}>
                    <Icon as={FiTool} color="green.500" boxSize={4} />
                    <Text fontSize="xs" color={subtextColor} fontWeight="semibold">
                      Equipment
                    </Text>
                  </HStack>
                  <Text fontSize="2xl" fontWeight="bold" color="green.500">
                    {dailyUpdate.equipment_progress || 0}%
                  </Text>
                </Box>
              </SimpleGrid>
            </Box>

            <Divider />

            {/* Work Description Section */}
            <Box>
              <HStack spacing={2} mb={3}>
                <Icon as={FiPackage} color="purple.500" boxSize={5} />
                <Text fontSize="sm" fontWeight="bold" color={textColor}>
                  Work Description
                </Text>
              </HStack>
              <Box
                bg={sectionBg}
                p={4}
                borderRadius="md"
                borderWidth="1px"
                borderColor={borderColor}
              >
                <Text fontSize="sm" color={textColor} whiteSpace="pre-wrap">
                  {dailyUpdate.work_description}
                </Text>
              </Box>
            </Box>

            {/* Materials Used Section */}
            {dailyUpdate.materials_used && dailyUpdate.materials_used.trim() !== '' && (
              <Box>
                <HStack spacing={2} mb={3}>
                  <Icon as={FiPackage} color="teal.500" boxSize={5} />
                  <Text fontSize="sm" fontWeight="bold" color={textColor}>
                    Materials Used
                  </Text>
                </HStack>
                <Box
                  bg={sectionBg}
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                >
                  <Text fontSize="sm" color={textColor} whiteSpace="pre-wrap">
                    {dailyUpdate.materials_used}
                  </Text>
                </Box>
              </Box>
            )}

            {/* Issues Section */}
            {dailyUpdate.issues && dailyUpdate.issues.trim() !== '' ? (
              <Box>
                <HStack spacing={2} mb={3}>
                  <Icon as={FiAlertTriangle} color="red.500" boxSize={5} />
                  <Text fontSize="sm" fontWeight="bold" color="red.500">
                    Issues / Problems
                  </Text>
                </HStack>
                <Box
                  bg="red.50"
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor="red.200"
                >
                  <Text fontSize="sm" color="red.700" whiteSpace="pre-wrap">
                    {dailyUpdate.issues}
                  </Text>
                </Box>
              </Box>
            ) : (
              <Box>
                <HStack spacing={2} mb={3}>
                  <Icon as={FiAlertTriangle} color="green.500" boxSize={5} />
                  <Text fontSize="sm" fontWeight="bold" color="green.500">
                    Issues / Problems
                  </Text>
                </HStack>
                <Box
                  bg="green.50"
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor="green.200"
                  textAlign="center"
                >
                  <Text fontSize="sm" color="green.700" fontStyle="italic">
                    ✓ No issues reported
                  </Text>
                </Box>
              </Box>
            )}

            {/* Tomorrow's Plan Section */}
            {dailyUpdate.tomorrows_plan && dailyUpdate.tomorrows_plan.trim() !== '' && (
              <Box>
                <HStack spacing={2} mb={3}>
                  <Icon as={FiTrendingUp} color="blue.500" boxSize={5} />
                  <Text fontSize="sm" fontWeight="bold" color={textColor}>
                    Tomorrow's Plan
                  </Text>
                </HStack>
                <Box
                  bg={sectionBg}
                  p={4}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                >
                  <Text fontSize="sm" color={textColor} whiteSpace="pre-wrap">
                    {dailyUpdate.tomorrows_plan}
                  </Text>
                </Box>
              </Box>
            )}

            {/* Photos Section */}
            <Box>
              <HStack spacing={2} mb={3}>
                <Icon as={FiImage} color="blue.500" boxSize={5} />
                <Text fontSize="sm" fontWeight="bold" color={textColor}>
                  Photos
                </Text>
                {dailyUpdate.photos && dailyUpdate.photos.length > 0 && (
                  <Badge colorScheme="blue" ml={2}>
                    {dailyUpdate.photos.length} {dailyUpdate.photos.length === 1 ? 'photo' : 'photos'}
                  </Badge>
                )}
              </HStack>

              {dailyUpdate.photos && dailyUpdate.photos.length > 0 ? (
                <SimpleGrid columns={{ base: 2, md: 3 }} spacing={4}>
                  {dailyUpdate.photos.map((photo, index) => (
                    <Box
                      key={index}
                      position="relative"
                      borderRadius="md"
                      overflow="hidden"
                      borderWidth="1px"
                      borderColor={borderColor}
                      transition="all 0.3s"
                      _hover={{
                        transform: 'scale(1.05)',
                        boxShadow: 'lg',
                        cursor: 'pointer',
                      }}
                      onClick={() => window.open(getFullUrl(photo), '_blank')}
                    >
                      <Image
                        src={getFullUrl(photo)}
                        alt={`Photo ${index + 1}`}
                        objectFit="cover"
                        w="100%"
                        h="150px"
                        fallbackSrc="https://via.placeholder.com/300x200?text=Loading..."
                      />
                      <Box
                        position="absolute"
                        bottom={0}
                        left={0}
                        right={0}
                        bg="blackAlpha.700"
                        py={1}
                        px={2}
                      >
                        <Text fontSize="xs" color="white" textAlign="center">
                          Photo {index + 1}
                        </Text>
                      </Box>
                    </Box>
                  ))}
                </SimpleGrid>
              ) : (
                <Box
                  bg={sectionBg}
                  p={6}
                  borderRadius="md"
                  borderWidth="1px"
                  borderColor={borderColor}
                  textAlign="center"
                >
                  <Icon as={FiImage} boxSize={10} color="gray.400" mb={2} />
                  <Text fontSize="sm" color={subtextColor} fontStyle="italic">
                    No photos attached
                  </Text>
                </Box>
              )}
            </Box>

            <Divider />

            {/* Created By Section */}
            <Box>
              <HStack spacing={2}>
                <Icon as={FiUser} color="gray.500" boxSize={4} />
                <Text fontSize="xs" color={subtextColor}>
                  Created by:{' '}
                  <Text as="span" fontWeight="semibold" color={textColor}>
                    {dailyUpdate.created_by || 'Unknown'}
                  </Text>
                </Text>
              </HStack>
            </Box>

            {/* Approval Info Section */}
            {(dailyUpdate.status === 'approved' || dailyUpdate.status === 'rejected') && (
              <Box bg={dailyUpdate.status === 'approved' ? 'green.50' : 'red.50'} p={4} borderRadius="md" borderWidth="1px" borderColor={dailyUpdate.status === 'approved' ? 'green.200' : 'red.200'}>
                <VStack align="start" spacing={2}>
                  <Text fontWeight="bold" color={dailyUpdate.status === 'approved' ? 'green.700' : 'red.700'}>
                    {dailyUpdate.status === 'approved' ? 'Approved By' : 'Rejected By'}: {dailyUpdate.approved_by || 'Unknown'}
                  </Text>
                  {dailyUpdate.approved_at && (
                    <Text fontSize="sm" color={dailyUpdate.status === 'approved' ? 'green.600' : 'red.600'}>
                      Date: {new Date(dailyUpdate.approved_at).toLocaleString('id-ID')}
                    </Text>
                  )}
                  {dailyUpdate.status === 'rejected' && dailyUpdate.rejection_reason && (
                    <Box mt={2} w="full">
                      <Text fontWeight="semibold" fontSize="sm" color="red.700">Reason:</Text>
                      <Text fontSize="sm" color="red.600">{dailyUpdate.rejection_reason}</Text>
                    </Box>
                  )}
                </VStack>
              </Box>
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
            <Button variant="ghost" onClick={onClose}>
              Close
            </Button>
            {onEdit && (
              <Button
                colorScheme="green"
                onClick={() => {
                  onClose();
                  onEdit();
                }}
              >
                Edit Update
              </Button>
            )}

            {/* Approval Buttons */}
            {canApprove('daily_updates') && (!dailyUpdate.status || dailyUpdate.status === 'pending') && (
              <>
                <Button
                  colorScheme="red"
                  variant="outline"
                  onClick={() => setShowRejectionInput(!showRejectionInput)}
                  isDisabled={isProcessing}
                >
                  Reject
                </Button>
                <Button
                  colorScheme="green"
                  onClick={handleApprove}
                  isLoading={isProcessing}
                >
                  Approve
                </Button>
              </>
            )}
          </HStack>

          {/* Rejection Input */}
          <Collapse in={showRejectionInput} animateOpacity style={{ width: '100%' }}>
            <Box mt={4} p={4} bg="red.50" borderRadius="md" borderWidth="1px" borderColor="red.200">
              <VStack align="stretch" spacing={3}>
                <Text fontWeight="bold" color="red.700">Reason for Rejection:</Text>
                <Textarea
                  value={rejectionReason}
                  onChange={(e) => setRejectionReason(e.target.value)}
                  placeholder="Please explain why this update is being rejected..."
                  bg="white"
                  borderColor="red.300"
                />
                <HStack justify="flex-end">
                  <Button size="sm" onClick={() => setShowRejectionInput(false)}>Cancel</Button>
                  <Button
                    size="sm"
                    colorScheme="red"
                    onClick={handleReject}
                    isLoading={isProcessing}
                  >
                    Confirm Rejection
                  </Button>
                </HStack>
              </VStack>
            </Box>
          </Collapse>

        </ModalFooter >
      </ModalContent >
    </Modal >
  );
};

export default DailyUpdateViewModal;

