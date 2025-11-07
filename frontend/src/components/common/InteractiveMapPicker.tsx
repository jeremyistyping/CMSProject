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
  Box,
  Text,
  VStack,
  HStack,
  useToast,
  Alert,
  AlertIcon,
  AlertDescription,
  Badge,
  Icon,
  Spinner,
  SimpleGrid,
  FormControl,
  FormLabel,
  Input,
  Textarea,
} from '@chakra-ui/react';
import { FiMapPin, FiNavigation, FiCheck, FiTarget, FiMap } from 'react-icons/fi';

interface LocationData {
  name: string;
  description: string;
  address: string;
  coordinates: string;
}

interface InteractiveMapPickerProps {
  isOpen: boolean;
  onClose: () => void;
  onLocationSelect: (locationData: LocationData) => void;
  currentCoordinates?: string;
  currentLocationData?: Partial<LocationData>;
  title?: string;
}

const InteractiveMapPicker: React.FC<InteractiveMapPickerProps> = ({
  isOpen,
  onClose,
  onLocationSelect,
  currentCoordinates,
  currentLocationData,
  title = "Select Location on Map"
}) => {
  const [selectedCoordinates, setSelectedCoordinates] = useState<string>(currentCoordinates || '');
  const [locationName, setLocationName] = useState<string>(currentLocationData?.name || '');
  const [locationDescription, setLocationDescription] = useState<string>(currentLocationData?.description || '');
  const [locationAddress, setLocationAddress] = useState<string>(currentLocationData?.address || '');
  const [isGettingLocation, setIsGettingLocation] = useState(false);
  const [mapUrl, setMapUrl] = useState<string>('');
  const [isMapLoading, setIsMapLoading] = useState(true);
  const toast = useToast();

  useEffect(() => {
    if (currentCoordinates) {
      setSelectedCoordinates(currentCoordinates);
    }
    if (currentLocationData) {
      setLocationName(currentLocationData.name || '');
      setLocationDescription(currentLocationData.description || '');
      setLocationAddress(currentLocationData.address || '');
    }
  }, [currentCoordinates, currentLocationData]);

  // Generate embedded map URL for location picking
  useEffect(() => {
    if (isOpen) {
      setIsMapLoading(true);
      const defaultLocation = selectedCoordinates || '-6.200000,106.816666'; // Default to Jakarta
      
      // Create interactive Google Maps embed URL
      const embedUrl = `https://www.google.com/maps/embed/v1/place?key=YOUR_API_KEY&q=${defaultLocation}&zoom=15&maptype=roadmap`;
      
      // For demo purposes, we'll use a regular Google Maps URL that opens in iframe
      const demoUrl = `https://www.google.com/maps?q=${defaultLocation}&z=15&output=embed`;
      
      setMapUrl(demoUrl);
      
      // Simulate loading delay
      setTimeout(() => {
        setIsMapLoading(false);
      }, 1000);
    }
  }, [isOpen, selectedCoordinates]);

  // Get current location using HTML5 Geolocation API
  const getCurrentLocation = () => {
    if (!navigator.geolocation) {
      toast({
        title: 'Geolocation not supported',
        description: 'Your browser does not support geolocation.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    setIsGettingLocation(true);
    navigator.geolocation.getCurrentPosition(
      (position: GeolocationPosition) => {
        const { latitude, longitude } = position.coords;
        const coordinates = `${latitude.toFixed(6)},${longitude.toFixed(6)}`;
        
        setSelectedCoordinates(coordinates);
        setIsGettingLocation(false);
        
        // Update map to show current location
        const newMapUrl = `https://www.google.com/maps?q=${coordinates}&z=15&output=embed`;
        setMapUrl(newMapUrl);
        
        toast({
          title: 'Current location found',
          description: 'Map updated to your current location',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      },
      (error: GeolocationPositionError) => {
        setIsGettingLocation(false);
        let errorMessage = 'Unable to get your location';
        
        switch (error.code) {
          case error.PERMISSION_DENIED:
            errorMessage = 'Location access denied by user';
            break;
          case error.POSITION_UNAVAILABLE:
            errorMessage = 'Location information is unavailable';
            break;
          case error.TIMEOUT:
            errorMessage = 'Location request timed out';
            break;
        }

        toast({
          title: 'Location Error',
          description: errorMessage,
          status: 'error',
          duration: 5000,
          isClosable: true,
        });
      },
      {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 60000
      }
    );
  };

  // Open full Google Maps in new window for location selection
  const openMapSelector = () => {
    const defaultLocation = selectedCoordinates || '-6.200000,106.816666';
    const mapsUrl = `https://www.google.com/maps/@${defaultLocation},15z`;
    
    const mapWindow = window.open(mapsUrl, '_blank', 'width=1000,height=700,scrollbars=yes,resizable=yes');
    
    if (mapWindow) {
      toast({
        title: 'Google Maps Opened',
        description: 'Right-click on your desired location → Copy coordinates → Come back and paste here',
        status: 'info',
        duration: 10000,
        isClosable: true,
      });
    }
  };

  // Handle manual coordinate input from copied Google Maps
  const handleCoordinateInput = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    let value = event.target.value.trim();
    
    // Clean up common coordinate formats from Google Maps
    // Google Maps gives: "lat, lng" or "lat,lng"  
    value = value.replace(/[^\d.,-]/g, ''); // Remove non-numeric chars except dots, commas, minus
    
    // Validate coordinate format
    if (validateCoordinates(value)) {
      setSelectedCoordinates(value);
      
      // Update embedded map
      const newMapUrl = `https://www.google.com/maps?q=${value}&z=15&output=embed`;
      setMapUrl(newMapUrl);
      
      toast({
        title: 'Location Updated',
        description: 'Map updated with new coordinates',
        status: 'success',
        duration: 2000,
        isClosable: true,
      });
    }
  };

  // Validate coordinate format
  const validateCoordinates = (coords: string): boolean => {
    const coordRegex = /^-?\d+\.?\d*,-?\d+\.?\d*$/;
    return coordRegex.test(coords.trim());
  };

  // Handle confirm location
  const handleConfirm = () => {
    if (!selectedCoordinates) {
      toast({
        title: 'No location selected',
        description: 'Please select a location on the map first',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!validateCoordinates(selectedCoordinates)) {
      toast({
        title: 'Invalid coordinates',
        description: 'Please select a valid location',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    if (!locationName.trim()) {
      toast({
        title: 'Location name required',
        description: 'Please enter a name for this location',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    const locationData: LocationData = {
      name: locationName.trim(),
      description: locationDescription.trim(),
      address: locationAddress.trim(),
      coordinates: selectedCoordinates
    };

    onLocationSelect(locationData);
    onClose();
  };

  // Clear selection
  const handleClear = () => {
    setSelectedCoordinates('');
    setLocationName('');
    setLocationDescription('');
    setLocationAddress('');
    // Reset to default location
    const defaultMapUrl = `https://www.google.com/maps?q=-6.200000,106.816666&z=15&output=embed`;
    setMapUrl(defaultMapUrl);
  };

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={onClose} 
      size="4xl" 
      scrollBehavior="inside"
      motionPreset="slideInBottom"
      isCentered
    >
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(3px)" />
      <ModalContent 
        maxH="90vh"
        maxW="4xl"
        mx={4}
        my={4}
        borderRadius="xl"
        shadow="2xl"
        bg="white"
        overflow="hidden"
        sx={{
          // Ensure modal content is properly constrained
          position: 'relative',
          display: 'flex',
          flexDirection: 'column',
          // Enhanced scrollbar styling for all browsers
          '& *::-webkit-scrollbar': {
            width: '8px',
            height: '8px',
          },
          '& *::-webkit-scrollbar-track': {
            background: '#f7fafc',
            borderRadius: '4px',
          },
          '& *::-webkit-scrollbar-thumb': {
            background: '#cbd5e0',
            borderRadius: '4px',
            border: '2px solid transparent',
            backgroundClip: 'padding-box',
            '&:hover': {
              background: '#a0aec0',
            },
            '&:active': {
              background: '#718096',
            },
          },
          // Firefox scrollbar styling
          '& *': {
            scrollbarWidth: 'thin',
            scrollbarColor: '#cbd5e0 #f7fafc',
          },
        }}
      >
        {/* Header */}
        <ModalHeader 
          bg="blue.50" 
          borderBottom="1px" 
          borderColor="blue.200" 
          position="sticky" 
          top={0} 
          zIndex={10}
          py={4}
          px={6}
          borderTopRadius="xl"
          flexShrink={0}
        >
          <HStack>
            <Icon as={FiMap} color="blue.500" boxSize={5} />
            <Text fontWeight="semibold">{title}</Text>
          </HStack>
        </ModalHeader>
        <ModalCloseButton zIndex={11} size="lg" />
      <ModalBody 
        p={0}
        flex={1}
        overflowY="auto"
        maxH="calc(90vh - 140px)"
        sx={{
          // Enhanced scrolling behavior
          scrollBehavior: 'smooth',
          WebkitOverflowScrolling: 'touch',
          overscrollBehavior: 'contain',
          // Custom scrollbar styling
          '&::-webkit-scrollbar': {
            width: '8px',
          },
          '&::-webkit-scrollbar-track': {
            background: '#f7fafc',
            borderRadius: '4px',
          },
          '&::-webkit-scrollbar-thumb': {
            background: '#cbd5e0',
            borderRadius: '4px',
            '&:hover': {
              background: '#a0aec0',
            },
          },
          // Firefox
          scrollbarWidth: 'thin',
          scrollbarColor: '#cbd5e0 #f7fafc',
        }}
      >
        <Box
          px={6}
          py={6}
          onWheel={(e) => {
            // Handle mouse wheel scrolling
            e.stopPropagation();
            const container = e.currentTarget.parentElement;
            if (container) {
              const scrollTop = container.scrollTop;
              const scrollHeight = container.scrollHeight;
              const clientHeight = container.clientHeight;
              
              // Allow scrolling if content is scrollable
              if (scrollHeight > clientHeight) {
                // Prevent default only if we're at the boundaries and trying to scroll further
                if ((scrollTop === 0 && e.deltaY < 0) || 
                    (scrollTop + clientHeight >= scrollHeight && e.deltaY > 0)) {
                  e.preventDefault();
                }
              }
            }
          }}
        >
          <VStack spacing={6}>
            {/* Instructions Section */}
            <Box w="full">
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                📋 How to Select Location
              </Text>
              <Alert status="info" borderRadius="lg" bg="blue.50" borderColor="blue.200">
                <AlertIcon color="blue.500" />
                <Box>
                  <AlertDescription fontSize="sm" color="gray.700">
                    <VStack align="start" spacing={1}>
                      <Text><strong>Method 1:</strong> Use your current GPS location</Text>
                      <Text><strong>Method 2:</strong> Pick location from Google Maps:</Text>
                      <VStack align="start" spacing={0} pl={4}>
                        <Text>• Click "Open Google Maps" below</Text>
                        <Text>• Navigate to your desired location</Text>
                        <Text>• Right-click on the exact spot → Click coordinates</Text>
                        <Text>• Copy and paste coordinates in the input below</Text>
                      </VStack>
                    </VStack>
                  </AlertDescription>
                </Box>
              </Alert>
            </Box>

            {/* Quick Actions Section */}
            <Box w="full">
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                🎯 Quick Actions
              </Text>
              <SimpleGrid columns={2} spacing={4} w="full">
                <Button
                  leftIcon={<FiTarget />}
                  colorScheme="blue"
                  variant="outline"
                  onClick={getCurrentLocation}
                  isLoading={isGettingLocation}
                  loadingText="Getting location..."
                  size="lg"
                  h="60px"
                >
                  <VStack spacing={1}>
                    <Text fontWeight="medium">Use Current Location</Text>
                    <Text fontSize="xs" color="gray.500">GPS Auto-detect</Text>
                  </VStack>
                </Button>
                
                <Button
                  leftIcon={<FiMap />}
                  colorScheme="green"
                  onClick={openMapSelector}
                  size="lg"
                  h="60px"
                >
                  <VStack spacing={1}>
                    <Text fontWeight="medium">Open Google Maps</Text>
                    <Text fontSize="xs">Select manually</Text>
                  </VStack>
                </Button>
              </SimpleGrid>
            </Box>

            {/* Location Details Section */}
            <Box w="full">
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                📝 Location Details
              </Text>
              <VStack spacing={4}>
                <FormControl isRequired>
                  <FormLabel fontSize="sm" fontWeight="medium">Location Name</FormLabel>
                  <Input
                    placeholder="Enter location name (e.g., Main Office, Warehouse A)"
                    value={locationName}
                    onChange={(e) => setLocationName(e.target.value)}
                    size="md"
                    bg="white"
                    _focus={{ borderColor: 'blue.400' }}
                  />
                  <Text fontSize="xs" color="gray.500" mt={1}>
                    🏢 Give this location a descriptive name
                  </Text>
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm" fontWeight="medium">Description</FormLabel>
                  <Textarea
                    placeholder="Add description about this location (e.g., IT equipment storage, Production floor)"
                    value={locationDescription}
                    onChange={(e) => setLocationDescription(e.target.value)}
                    rows={3}
                    size="md"
                    bg="white"
                    _focus={{ borderColor: 'blue.400' }}
                    resize="vertical"
                  />
                  <Text fontSize="xs" color="gray.500" mt={1}>
                    📝 Optional: Add more details about this location
                  </Text>
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm" fontWeight="medium">Address</FormLabel>
                  <Textarea
                    placeholder="Enter full address (e.g., Jl. Sudirman No. 123, Jakarta Pusat)"
                    value={locationAddress}
                    onChange={(e) => setLocationAddress(e.target.value)}
                    rows={2}
                    size="md"
                    bg="white"
                    _focus={{ borderColor: 'blue.400' }}
                    resize="vertical"
                  />
                  <Text fontSize="xs" color="gray.500" mt={1}>
                    🏠 Optional: Full address of this location
                  </Text>
                </FormControl>
              </VStack>
            </Box>

            {/* Coordinate Input Section */}
            <Box w="full">
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                📍 GPS Coordinates
              </Text>
              <FormControl isRequired>
                <FormLabel fontSize="sm" fontWeight="medium">Coordinates Input</FormLabel>
                <Textarea
                  placeholder="Paste coordinates from Google Maps here (e.g., -6.200000,106.816666)"
                  value={selectedCoordinates}
                  onChange={handleCoordinateInput}
                  rows={3}
                  fontFamily="mono"
                  fontSize="sm"
                  bg="gray.50"
                  _focus={{ bg: 'white', borderColor: 'blue.400' }}
                  resize="vertical"
                />
                <Text fontSize="xs" color="gray.500" mt={2}>
                  💡 <strong>Pro tip:</strong> Right-click on Google Maps → Click the coordinates → They will be copied automatically
                </Text>
                
                {/* Validation feedback */}
                {selectedCoordinates && !validateCoordinates(selectedCoordinates) && (
                  <Alert status="warning" mt={3} borderRadius="md" size="sm">
                    <AlertIcon />
                    <AlertDescription fontSize="sm">
                      Please enter valid coordinates in format: latitude,longitude (e.g., -6.200000,106.816666)
                    </AlertDescription>
                  </Alert>
                )}
              </FormControl>
            </Box>

            {/* Selected Location Display */}
            {selectedCoordinates && validateCoordinates(selectedCoordinates) && (
              <Box w="full">
                <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                  ✅ Location Summary
                </Text>
                
                {/* Location Info Card */}
                <Box w="full" p={5} bg="green.50" borderRadius="xl" border="1px" borderColor="green.200">
                  <VStack spacing={4}>
                    <HStack w="full" justify="space-between">
                      <HStack>
                        <Icon as={FiCheck} color="green.500" boxSize={5} />
                        <Text fontSize="lg" fontWeight="semibold" color="green.700">
                          Location Ready
                        </Text>
                      </HStack>
                      <Badge colorScheme="green" variant="solid" px={3} py={1} borderRadius="full">
                        ✓ Complete
                      </Badge>
                    </HStack>
                    
                    <Box w="full" p={4} bg="white" borderRadius="lg" border="1px" borderColor="green.300" shadow="sm">
                      <VStack align="stretch" spacing={4}>
                        {/* Location Name */}
                        {locationName && (
                          <Box>
                            <Text fontSize="xs" color="gray.600" mb={1} fontWeight="medium">LOCATION NAME:</Text>
                            <HStack>
                              <Icon as={FiMapPin} color="green.600" boxSize={4} />
                              <Text fontSize="md" color="green.800" fontWeight="bold">
                                {locationName}
                              </Text>
                            </HStack>
                          </Box>
                        )}
                        
                        {/* Description */}
                        {locationDescription && (
                          <Box>
                            <Text fontSize="xs" color="gray.600" mb={1} fontWeight="medium">DESCRIPTION:</Text>
                            <Text fontSize="sm" color="gray.700">
                              {locationDescription}
                            </Text>
                          </Box>
                        )}
                        
                        {/* Address */}
                        {locationAddress && (
                          <Box>
                            <Text fontSize="xs" color="gray.600" mb={1} fontWeight="medium">ADDRESS:</Text>
                            <Text fontSize="sm" color="gray.700">
                              {locationAddress}
                            </Text>
                          </Box>
                        )}
                        
                        {/* GPS Coordinates */}
                        <Box>
                          <Text fontSize="xs" color="gray.600" mb={1} fontWeight="medium">GPS COORDINATES:</Text>
                          <Text fontSize="sm" fontFamily="mono" color="gray.700">
                            {selectedCoordinates}
                          </Text>
                        </Box>
                        
                        <HStack spacing={3}>
                          <Button
                            size="sm"
                            variant="outline"
                            colorScheme="blue"
                            leftIcon={<FiMap />}
                            onClick={() => {
                              const mapsUrl = `https://www.google.com/maps?q=${selectedCoordinates}&z=15`;
                              window.open(mapsUrl, '_blank');
                            }}
                            flex={1}
                          >
                            Preview on Maps
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            colorScheme="gray"
                            onClick={() => {
                              const locationInfo = `${locationName}\n${locationDescription}\n${locationAddress}\n${selectedCoordinates}`;
                              navigator.clipboard.writeText(locationInfo);
                              toast({
                                title: 'Copied!',
                                description: 'Location info copied to clipboard',
                                status: 'success',
                                duration: 2000,
                              });
                            }}
                            flex={1}
                          >
                            Copy Info
                          </Button>
                        </HStack>
                      </VStack>
                    </Box>
                  </VStack>
                </Box>
                
                {/* Map Preview */}
                <Box w="full" mt={6}>
                  <Text fontSize="sm" fontWeight="medium" color="gray.700" mb={3}>
                    🗺️ Location Preview
                  </Text>
                  <Box 
                    w="full" 
                    h="350px" 
                    border="2px solid" 
                    borderColor="green.200" 
                    borderRadius="xl" 
                    overflow="hidden"
                    position="relative"
                    shadow="md"
                    sx={{
                      // Prevent iframe from interfering with scroll
                      pointerEvents: 'auto',
                      touchAction: 'none',
                    }}
                  >
                    {isMapLoading && (
                      <Box 
                        position="absolute" 
                        top="50%" 
                        left="50%" 
                        transform="translate(-50%, -50%)"
                        zIndex={2}
                        bg="white"
                        p={6}
                        borderRadius="lg"
                        shadow="lg"
                      >
                        <VStack>
                          <Spinner size="lg" color="blue.500" thickness="4px" />
                          <Text fontSize="sm" color="gray.600" fontWeight="medium">Loading map preview...</Text>
                        </VStack>
                      </Box>
                    )}
                    
                    <iframe
                      src={mapUrl}
                      width="100%"
                      height="100%"
                      style={{ 
                        border: 0,
                        display: 'block',
                        // Prevent iframe scroll interference
                        pointerEvents: 'auto',
                        touchAction: 'manipulation',
                      }}
                      loading="lazy"
                      referrerPolicy="no-referrer-when-downgrade"
                      onLoad={() => setIsMapLoading(false)}
                      // Prevent focus issues that can affect scrolling
                      tabIndex={-1}
                    />
                  </Box>
                </Box>
              </Box>
            )}
          </VStack>
          </Box>
        </ModalBody>

        <ModalFooter 
          bg="gray.50" 
          borderTop="1px" 
          borderColor="gray.200"
          py={4}
          px={6}
          borderBottomRadius="xl"
          flexShrink={0}
          position="sticky"
          bottom={0}
          zIndex={10}
        >
          <HStack spacing={3} w="full" justify="flex-end">
            <Button 
              variant="outline" 
              onClick={handleClear} 
              size="md"
              isDisabled={!selectedCoordinates}
              colorScheme="gray"
            >
              Clear
            </Button>
            <Button 
              variant="outline" 
              onClick={onClose}
              size="md"
            >
              Cancel
            </Button>
            <Button
              colorScheme="blue"
              onClick={handleConfirm}
              isDisabled={!selectedCoordinates || !validateCoordinates(selectedCoordinates) || !locationName.trim()}
              leftIcon={<FiMapPin />}
              size="md"
              shadow="md"
              _hover={{
                transform: 'translateY(-1px)',
                shadow: 'lg',
              }}
            >
              Confirm Location
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default InteractiveMapPicker;
