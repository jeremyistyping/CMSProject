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
  Input,
  FormControl,
  FormLabel,
  useToast,
  Alert,
  AlertIcon,
  AlertDescription,
  Badge,
  Divider,
  SimpleGrid,
  Spinner,
  Icon,
} from '@chakra-ui/react';
import { FiMapPin, FiNavigation, FiCheck, FiMousePointer, FiMap } from 'react-icons/fi';

interface MapPickerProps {
  isOpen: boolean;
  onClose: () => void;
  onLocationSelect: (coordinates: string) => void;
  currentCoordinates?: string;
  title?: string;
}

interface GeolocationCoords {
  latitude: number;
  longitude: number;
}

const MapPicker: React.FC<MapPickerProps> = ({
  isOpen,
  onClose,
  onLocationSelect,
  currentCoordinates,
  title = "Pick Asset Location"
}) => {
  const [selectedCoordinates, setSelectedCoordinates] = useState<string>(currentCoordinates || '');
  const [manualInput, setManualInput] = useState<string>(currentCoordinates || '');
  const [isGettingLocation, setIsGettingLocation] = useState(false);
  const [isSelectingOnMap, setIsSelectingOnMap] = useState(false);
  const [selectedAddress, setSelectedAddress] = useState<string>('');
  const toast = useToast();

  useEffect(() => {
    if (currentCoordinates) {
      setSelectedCoordinates(currentCoordinates);
      setManualInput(currentCoordinates);
    }
  }, [currentCoordinates]);

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
        setManualInput(coordinates);
        setIsGettingLocation(false);
        
        toast({
          title: 'Location detected',
          description: 'Current location has been set',
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

  // Validate coordinate format
  const validateCoordinates = (coords: string): boolean => {
    const coordRegex = /^-?\d+\.?\d*,-?\d+\.?\d*$/;
    return coordRegex.test(coords.trim());
  };

  // Handle manual input change
  const handleManualInputChange = (value: string) => {
    setManualInput(value);
    if (validateCoordinates(value)) {
      setSelectedCoordinates(value.trim());
    }
  };

  // Handle confirm location
  const handleConfirm = () => {
    if (!selectedCoordinates) {
      toast({
        title: 'No location selected',
        description: 'Please select a location or enter coordinates',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!validateCoordinates(selectedCoordinates)) {
      toast({
        title: 'Invalid coordinates',
        description: 'Please enter valid coordinates in format: latitude,longitude',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    onLocationSelect(selectedCoordinates);
    onClose();
  };

  // Open in Google Maps for preview
  const openInGoogleMaps = () => {
    if (selectedCoordinates && validateCoordinates(selectedCoordinates)) {
      const mapsUrl = `https://www.google.com/maps?q=${selectedCoordinates}`;
      window.open(mapsUrl, '_blank');
    }
  };

  // Open Google Maps in selection mode
  const openMapSelector = () => {
    setIsSelectingOnMap(true);
    const defaultLocation = selectedCoordinates || '-6.200000,106.816666'; // Default to Jakarta
    const mapsUrl = `https://www.google.com/maps?q=${defaultLocation}&z=15`;
    
    // Open in new window with instructions
    const mapWindow = window.open(mapsUrl, '_blank', 'width=800,height=600,scrollbars=yes,resizable=yes');
    
    if (mapWindow) {
      toast({
        title: 'Map Opened',
        description: 'Right-click on your desired location and copy the coordinates to paste here',
        status: 'info',
        duration: 8000,
        isClosable: true,
      });
    }
  };

  // Predefined popular locations for quick selection
  const popularLocations = [
    { name: 'Jakarta Pusat', coordinates: '-6.200000,106.816666', address: 'Central Jakarta, DKI Jakarta' },
    { name: 'Surabaya', coordinates: '-7.250445,112.768845', address: 'Surabaya, East Java' },
    { name: 'Bandung', coordinates: '-6.917464,107.619123', address: 'Bandung, West Java' },
    { name: 'Medan', coordinates: '3.595196,98.672226', address: 'Medan, North Sumatra' },
    { name: 'Semarang', coordinates: '-6.966667,110.416664', address: 'Semarang, Central Java' },
    { name: 'Yogyakarta', coordinates: '-7.795580,110.369492', address: 'Yogyakarta Special Region' },
  ];

  // Select from popular locations
  const selectPopularLocation = (location: typeof popularLocations[0]) => {
    setSelectedCoordinates(location.coordinates);
    setManualInput(location.coordinates);
    setSelectedAddress(location.address);
    toast({
      title: 'Location Selected',
      description: `Selected ${location.name}`,
      status: 'success',
      duration: 3000,
      isClosable: true,
    });
  };

  // Reset to original coordinates
  const handleReset = () => {
    setSelectedCoordinates(currentCoordinates || '');
    setManualInput(currentCoordinates || '');
  };

  // Clear coordinates
  const handleClear = () => {
    setSelectedCoordinates('');
    setManualInput('');
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <HStack>
            <FiMapPin />
            <Text>{title}</Text>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody>
          <VStack spacing={4}>
            {/* Location selection methods */}
            <Box w="full">
              <Text fontSize="sm" color="gray.600" mb={3}>
                Choose how you want to set the location:
              </Text>
              <SimpleGrid columns={1} spacing={3}>
                {/* Current location detector */}
                <Button
                  leftIcon={<FiNavigation />}
                  colorScheme="blue"
                  variant="outline"
                  onClick={getCurrentLocation}
                  isLoading={isGettingLocation}
                  loadingText="Getting location..."
                  w="full"
                  size="md"
                >
                  Use Current Location
                </Button>
                
                {/* Map selector */}
                <Button
                  leftIcon={<FiMap />}
                  colorScheme="green"
                  variant="outline"
                  onClick={openMapSelector}
                  w="full"
                  size="md"
                >
                  Select Location on Map
                </Button>
              </SimpleGrid>
            </Box>
            
            <Divider />
            
            {/* Popular locations quick select */}
            <Box w="full">
              <Text fontSize="sm" fontWeight="medium" color="gray.700" mb={3}>
                🏢 Popular Locations (Quick Select)
              </Text>
              <SimpleGrid columns={2} spacing={2}>
                {popularLocations.map((location) => (
                  <Button
                    key={location.name}
                    size="sm"
                    variant="ghost"
                    onClick={() => selectPopularLocation(location)}
                    leftIcon={<FiMapPin />}
                    justifyContent="flex-start"
                    px={2}
                    h="auto"
                    whiteSpace="normal"
                    textAlign="left"
                  >
                    <Box>
                      <Text fontSize="xs" fontWeight="medium">
                        {location.name}
                      </Text>
                      <Text fontSize="xs" color="gray.500" noOfLines={1}>
                        {location.address}
                      </Text>
                    </Box>
                  </Button>
                ))}
              </SimpleGrid>
            </Box>
            
            <Divider />

            {/* Manual coordinate input */}
            <FormControl>
              <FormLabel>
                <HStack>
                  <Icon as={FiMousePointer} />
                  <Text>Manual Coordinates Entry</Text>
                </HStack>
              </FormLabel>
              <Input
                placeholder="e.g., -6.200000,106.816666"
                value={manualInput}
                onChange={(e) => handleManualInputChange(e.target.value)}
                fontSize="sm"
                bg="gray.50"
                _focus={{ bg: 'white', borderColor: 'blue.400' }}
              />
              <Text fontSize="xs" color="gray.500" mt={1}>
                💡 <strong>Tip:</strong> Right-click on Google Maps → Copy coordinates → Paste here
              </Text>
            </FormControl>

            {/* Validation feedback */}
            {manualInput && !validateCoordinates(manualInput) && (
              <Alert status="warning" size="sm">
                <AlertIcon />
                <AlertDescription fontSize="sm">
                  Invalid coordinate format. Use: latitude,longitude
                </AlertDescription>
              </Alert>
            )}

            {/* Selected coordinates display */}
            {selectedCoordinates && validateCoordinates(selectedCoordinates) && (
              <Box w="full" p={4} bg="green.50" borderRadius="lg" border="1px" borderColor="green.200">
                <VStack spacing={3}>
                  <HStack w="full" justify="space-between">
                    <HStack>
                      <Icon as={FiCheck} color="green.500" />
                      <Text fontSize="sm" fontWeight="medium" color="green.700">
                        ✅ Selected Location
                      </Text>
                    </HStack>
                    <Badge colorScheme="green" variant="solid">Valid</Badge>
                  </HStack>
                  
                  <Box w="full" p={2} bg="white" borderRadius="md" border="1px" borderColor="green.300">
                    <Text fontSize="xs" color="gray.600" mb={1}>Coordinates:</Text>
                    <Text fontSize="sm" fontFamily="mono" color="green.800" fontWeight="medium">
                      📍 {selectedCoordinates}
                    </Text>
                    {selectedAddress && (
                      <>
                        <Text fontSize="xs" color="gray.600" mt={2} mb={1}>Address:</Text>
                        <Text fontSize="sm" color="gray.700">
                          📍 {selectedAddress}
                        </Text>
                      </>
                    )}
                  </Box>
                  
                  <HStack spacing={2}>
                    <Button
                      size="sm"
                      variant="outline"
                      colorScheme="blue"
                      onClick={openInGoogleMaps}
                      leftIcon={<FiMapPin />}
                    >
                      Preview in Maps
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      colorScheme="gray"
                      onClick={() => {
                        navigator.clipboard.writeText(selectedCoordinates);
                        toast({
                          title: 'Copied!',
                          description: 'Coordinates copied to clipboard',
                          status: 'success',
                          duration: 2000,
                        });
                      }}
                    >
                      Copy
                    </Button>
                  </HStack>
                </VStack>
              </Box>
            )}

            {/* Helper text */}
            <Alert status="info" size="sm" borderRadius="md">
              <AlertIcon />
              <AlertDescription fontSize="sm">
                <strong>How to get coordinates:</strong><br />
                1. Click "Select Location on Map" button above<br />
                2. Right-click on your desired location in Google Maps<br />
                3. Click on the coordinates that appear<br />
                4. Paste here in the manual input field
              </AlertDescription>
            </Alert>
            
            {/* Loading state for map selection */}
            {isSelectingOnMap && (
              <Alert status="loading" size="sm" borderRadius="md">
                <Spinner size="sm" mr={2} />
                <AlertDescription fontSize="sm">
                  🗺️ Google Maps opened in new tab. Follow the instructions above to get coordinates.
                </AlertDescription>
              </Alert>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <HStack spacing={3}>
            <Button variant="ghost" onClick={handleClear} size="sm">
              Clear
            </Button>
            {currentCoordinates && (
              <Button variant="ghost" onClick={handleReset} size="sm">
                Reset
              </Button>
            )}
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              colorScheme="blue"
              onClick={handleConfirm}
              isDisabled={!selectedCoordinates || !validateCoordinates(selectedCoordinates)}
            >
              Confirm Location
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default MapPicker;
