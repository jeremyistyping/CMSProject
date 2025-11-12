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
} from 'react-icons/fi';
import { DailyUpdate } from '@/types/project';

interface DailyUpdateViewModalProps {
  isOpen: boolean;
  onClose: () => void;
  dailyUpdate: DailyUpdate | null;
  onEdit?: () => void;
}

const DailyUpdateViewModal: React.FC<DailyUpdateViewModalProps> = ({
  isOpen,
  onClose,
  dailyUpdate,
  onEdit,
}) => {
  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const sectionBg = useColorModeValue('gray.50', 'var(--bg-primary)');

  if (!dailyUpdate) return null;

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
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default DailyUpdateViewModal;

