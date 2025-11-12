'use client';

import React, { useState } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Box,
  Image,
  SimpleGrid,
  IconButton,
  HStack,
  Text,
  useColorModeValue,
  Center,
  VStack,
  Icon,
} from '@chakra-ui/react';
import { FiChevronLeft, FiChevronRight, FiX, FiImage, FiDownload } from 'react-icons/fi';

interface PhotoGalleryProps {
  isOpen: boolean;
  onClose: () => void;
  photos: string[];
  title?: string;
}

const PhotoGallery: React.FC<PhotoGalleryProps> = ({
  isOpen,
  onClose,
  photos,
  title = 'Photos',
}) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isLightboxOpen, setIsLightboxOpen] = useState(false);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');

  const handlePhotoClick = (index: number) => {
    setCurrentIndex(index);
    setIsLightboxOpen(true);
  };

  const handlePrevious = () => {
    setCurrentIndex((prev) => (prev > 0 ? prev - 1 : photos.length - 1));
  };

  const handleNext = () => {
    setCurrentIndex((prev) => (prev < photos.length - 1 ? prev + 1 : 0));
  };

  const handleDownload = (photoUrl: string) => {
    console.log('Downloading photo:', photoUrl);
    const link = document.createElement('a');
    link.href = photoUrl;
    link.download = photoUrl.split('/').pop() || 'photo.jpg';
    link.target = '_blank';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleLightboxClose = () => {
    setIsLightboxOpen(false);
  };

  // Convert relative URLs to full URLs
  // Photos are served from backend (port 8080), not frontend (port 3000)
  const getFullUrl = (url: string) => {
    if (!url) {
      console.warn('Empty photo URL provided');
      return '';
    }
    
    if (url.startsWith('http')) {
      console.log('Photo URL already absolute:', url);
      return url;
    }
    
    // Normalize path - convert Windows backslashes to forward slashes
    let normalizedUrl = url.replace(/\\/g, '/');
    
    // Ensure leading slash
    if (!normalizedUrl.startsWith('/')) {
      normalizedUrl = '/' + normalizedUrl;
    }
    
    // Use backend URL for uploaded files
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    const fullUrl = `${backendUrl}${normalizedUrl}`;
    console.log('Converting relative URL to absolute:', url, '->', fullUrl);
    return fullUrl;
  };

  if (photos.length === 0) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} size="xl" isCentered>
        <ModalOverlay />
        <ModalContent bg={bgColor}>
          <ModalHeader color={textColor} borderBottomWidth="1px" borderColor={borderColor}>
            {title}
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody py={8}>
            <Center h="200px">
              <VStack spacing={3}>
                <Icon as={FiImage} boxSize={12} color="gray.400" />
                <Text color="gray.500">No photos available</Text>
              </VStack>
            </Center>
          </ModalBody>
        </ModalContent>
      </Modal>
    );
  }

  return (
    <>
      {/* Gallery Grid Modal */}
      <Modal isOpen={isOpen && !isLightboxOpen} onClose={onClose} size="4xl" isCentered>
        <ModalOverlay />
        <ModalContent bg={bgColor} maxH="90vh">
          <ModalHeader
            color={textColor}
            borderBottomWidth="1px"
            borderColor={borderColor}
          >
            {title} ({photos.length})
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody py={6} overflowY="auto">
            <SimpleGrid columns={{ base: 2, md: 3, lg: 4 }} spacing={4}>
              {photos.map((photo, index) => (
                <Box
                  key={index}
                  position="relative"
                  cursor="pointer"
                  borderRadius="md"
                  overflow="hidden"
                  borderWidth="1px"
                  borderColor={borderColor}
                  role="group"
                  _hover={{
                    transform: 'scale(1.05)',
                    shadow: 'lg',
                  }}
                  transition="all 0.2s"
                >
                  <Box onClick={() => handlePhotoClick(index)}>
                    <Image
                      src={getFullUrl(photo)}
                      alt={`Photo ${index + 1}`}
                      objectFit="cover"
                      w="full"
                      h="150px"
                      fallbackSrc="https://via.placeholder.com/150?text=Image+Not+Found"
                    />
                  </Box>
                  <Box
                    position="absolute"
                    top={2}
                    right={2}
                    opacity={0}
                    _groupHover={{ opacity: 1 }}
                    transition="opacity 0.2s"
                  >
                    <IconButton
                      icon={<FiDownload />}
                      size="sm"
                      colorScheme="blue"
                      aria-label="Download photo"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDownload(getFullUrl(photo));
                      }}
                      bg="blue.500"
                      _hover={{ bg: 'blue.600' }}
                    />
                  </Box>
                  <Box
                    position="absolute"
                    bottom={0}
                    left={0}
                    right={0}
                    bg="blackAlpha.700"
                    p={2}
                    opacity={0}
                    _groupHover={{ opacity: 1 }}
                    transition="opacity 0.2s"
                  >
                    <Text fontSize="xs" color="white" textAlign="center">
                      Click to view
                    </Text>
                  </Box>
                </Box>
              ))}
            </SimpleGrid>
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Lightbox Modal */}
      <Modal
        isOpen={isLightboxOpen}
        onClose={handleLightboxClose}
        size="full"
        isCentered
      >
        <ModalOverlay bg="blackAlpha.900" />
        <ModalContent bg="transparent" shadow="none">
          <ModalCloseButton
            size="lg"
            color="white"
            _hover={{ bg: 'whiteAlpha.200' }}
            zIndex={2}
          />
          <ModalBody p={0}>
            <Center h="100vh" position="relative">
              {/* Main Image */}
              <Box maxW="90vw" maxH="90vh" position="relative">
                <Image
                  src={getFullUrl(photos[currentIndex])}
                  alt={`Photo ${currentIndex + 1}`}
                  objectFit="contain"
                  maxH="90vh"
                  maxW="90vw"
                  fallbackSrc="https://via.placeholder.com/800?text=Image+Not+Found"
                />
              </Box>

              {/* Navigation Buttons */}
              {photos.length > 1 && (
                <>
                  <IconButton
                    icon={<FiChevronLeft />}
                    position="absolute"
                    left={4}
                    top="50%"
                    transform="translateY(-50%)"
                    onClick={handlePrevious}
                    colorScheme="whiteAlpha"
                    size="lg"
                    aria-label="Previous photo"
                    bg="whiteAlpha.300"
                    _hover={{ bg: 'whiteAlpha.400' }}
                  />
                  <IconButton
                    icon={<FiChevronRight />}
                    position="absolute"
                    right={4}
                    top="50%"
                    transform="translateY(-50%)"
                    onClick={handleNext}
                    colorScheme="whiteAlpha"
                    size="lg"
                    aria-label="Next photo"
                    bg="whiteAlpha.300"
                    _hover={{ bg: 'whiteAlpha.400' }}
                  />
                </>
              )}

              {/* Bottom Info Bar */}
              <HStack
                position="absolute"
                bottom={4}
                left="50%"
                transform="translateX(-50%)"
                bg="blackAlpha.700"
                px={6}
                py={3}
                borderRadius="full"
                spacing={4}
              >
                <Text color="white" fontSize="sm" fontWeight="medium">
                  {currentIndex + 1} / {photos.length}
                </Text>
                <IconButton
                  icon={<FiDownload />}
                  size="sm"
                  variant="ghost"
                  colorScheme="whiteAlpha"
                  aria-label="Download photo"
                  onClick={() => handleDownload(getFullUrl(photos[currentIndex]))}
                  _hover={{ bg: 'whiteAlpha.300' }}
                />
              </HStack>
            </Center>
          </ModalBody>
        </ModalContent>
      </Modal>
    </>
  );
};

export default PhotoGallery;

