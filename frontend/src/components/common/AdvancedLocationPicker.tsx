'use client';

import React, { useState, useEffect, useCallback } from 'react';
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
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  InputGroup,
  InputLeftElement,
  Select,
  Stack,
} from '@chakra-ui/react';
import { 
  FiMapPin, 
  FiNavigation, 
  FiCheck, 
  FiMousePointer, 
  FiMap, 
  FiSearch,
  FiGlobe,
  FiTarget,
  FiCrosshair
} from 'react-icons/fi';

interface AdvancedLocationPickerProps {
  isOpen: boolean;
  onClose: () => void;
  onLocationSelect: (coordinates: string, address?: string) => void;
  currentCoordinates?: string;
  title?: string;
}

interface LocationResult {
  name: string;
  coordinates: string;
  address: string;
  type: 'search' | 'popular' | 'recent';
}

const AdvancedLocationPicker: React.FC<AdvancedLocationPickerProps> = ({
  isOpen,
  onClose,
  onLocationSelect,
  currentCoordinates,
  title = "Advanced Location Picker"
}) => {
  const [selectedCoordinates, setSelectedCoordinates] = useState<string>(currentCoordinates || '');
  const [selectedAddress, setSelectedAddress] = useState<string>('');
  const [manualInput, setManualInput] = useState<string>(currentCoordinates || '');
  const [isGettingLocation, setIsGettingLocation] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<LocationResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [recentLocations, setRecentLocations] = useState<LocationResult[]>([]);
  const [activeTab, setActiveTab] = useState(0);
  const toast = useToast();

  // Load recent locations from localStorage
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('recent_asset_locations');
      if (saved) {
        try {
          setRecentLocations(JSON.parse(saved));
        } catch (error) {
          console.error('Error loading recent locations:', error);
        }
      }
    }
  }, []);

  // Save location to recent
  const saveToRecent = useCallback((coordinates: string, address: string) => {
    const newLocation: LocationResult = {
      name: address.split(',')[0] || 'Selected Location',
      coordinates,
      address,
      type: 'recent'
    };

    const updatedRecent = [
      newLocation,
      ...recentLocations.filter(loc => loc.coordinates !== coordinates)
    ].slice(0, 10); // Keep only 10 recent locations

    setRecentLocations(updatedRecent);
    
    if (typeof window !== 'undefined') {
      localStorage.setItem('recent_asset_locations', JSON.stringify(updatedRecent));
    }
  }, [recentLocations]);

  useEffect(() => {
    if (currentCoordinates) {
      setSelectedCoordinates(currentCoordinates);
      setManualInput(currentCoordinates);
    }
  }, [currentCoordinates]);

  // Popular Indonesian locations with more details
  const popularLocations: LocationResult[] = [
    { 
      name: 'Monas Jakarta', 
      coordinates: '-6.175110,106.827153', 
      address: 'National Monument, Central Jakarta, DKI Jakarta',
      type: 'popular'
    },
    { 
      name: 'Tugu Pahlawan Surabaya', 
      coordinates: '-7.245543,112.737752', 
      address: 'Heroes Monument, Surabaya, East Java',
      type: 'popular'
    },
    { 
      name: 'Gedung Sate Bandung', 
      coordinates: '-6.902486,107.618682', 
      address: 'Gedung Sate, Bandung, West Java',
      type: 'popular'
    },
    { 
      name: 'Masjid Raya Medan', 
      coordinates: '3.587934,98.672226', 
      address: 'Grand Mosque of Medan, North Sumatra',
      type: 'popular'
    },
    { 
      name: 'Lawang Sewu Semarang', 
      coordinates: '-6.983389,110.409089', 
      address: 'Lawang Sewu Building, Semarang, Central Java',
      type: 'popular'
    },
    { 
      name: 'Keraton Yogyakarta', 
      coordinates: '-7.805220,110.364563', 
      address: 'Sultan Palace, Yogyakarta Special Region',
      type: 'popular'
    },
    { 
      name: 'Danau Toba', 
      coordinates: '2.686667,98.875000', 
      address: 'Lake Toba, North Sumatra',
      type: 'popular'
    },
    { 
      name: 'Borobudur Temple', 
      coordinates: '-7.607874,110.203751', 
      address: 'Borobudur Temple, Magelang, Central Java',
      type: 'popular'
    }
  ];

  // Mock search function (in real app, integrate with Google Places API)
  const searchLocations = useCallback(async (query: string) => {
    if (!query || query.length < 3) return;
    
    setIsSearching(true);
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Mock search results - filter from popular locations
    const results = popularLocations.filter(location =>
      location.name.toLowerCase().includes(query.toLowerCase()) ||
      location.address.toLowerCase().includes(query.toLowerCase())
    ).map(location => ({ ...location, type: 'search' as const }));
    
    setSearchResults(results);
    setIsSearching(false);
  }, []);

  // Debounced search
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (searchQuery) {
        searchLocations(searchQuery);
      } else {
        setSearchResults([]);
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [searchQuery, searchLocations]);

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
        const address = `Current Location (${latitude.toFixed(4)}, ${longitude.toFixed(4)})`;
        
        setSelectedCoordinates(coordinates);
        setManualInput(coordinates);
        setSelectedAddress(address);
        setIsGettingLocation(false);
        
        // Save to recent
        saveToRecent(coordinates, address);
        
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
      setSelectedAddress('Manual Entry');
    }
  };

  // Select location from results
  const selectLocation = (location: LocationResult) => {
    setSelectedCoordinates(location.coordinates);
    setManualInput(location.coordinates);
    setSelectedAddress(location.address);
    
    // Save to recent if not already recent
    if (location.type !== 'recent') {
      saveToRecent(location.coordinates, location.address);
    }
    
    toast({
      title: 'Location Selected',
      description: location.name,
      status: 'success',
      duration: 2000,
      isClosable: true,
    });
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

    onLocationSelect(selectedCoordinates, selectedAddress);
    onClose();
  };

  // Open in Google Maps
  const openInGoogleMaps = () => {
    if (selectedCoordinates && validateCoordinates(selectedCoordinates)) {
      const mapsUrl = `https://www.google.com/maps?q=${selectedCoordinates}&z=15`;
      window.open(mapsUrl, '_blank');
    }
  };

  // Clear selection
  const handleClear = () => {
    setSelectedCoordinates('');
    setManualInput('');
    setSelectedAddress('');
    setSearchQuery('');
    setSearchResults([]);
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl">
      <ModalOverlay />
      <ModalContent maxH="90vh">
        <ModalHeader>
          <HStack>
            <Icon as={FiMap} color="blue.500" />
            <Text>{title}</Text>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody overflowY="auto">
          <VStack spacing={4}>
            {/* Quick Actions */}
            <Box w="full">
              <SimpleGrid columns={2} spacing={3}>
                <Button
                  leftIcon={<FiTarget />}
                  colorScheme="blue"
                  variant="outline"
                  onClick={getCurrentLocation}
                  isLoading={isGettingLocation}
                  loadingText="Detecting..."
                  size="sm"
                >
                  Current Location
                </Button>
                <Button
                  leftIcon={<FiGlobe />}
                  colorScheme="green"
                  variant="outline"
                  onClick={openInGoogleMaps}
                  isDisabled={!selectedCoordinates}
                  size="sm"
                >
                  Open in Maps
                </Button>
              </SimpleGrid>
            </Box>

            <Divider />

            {/* Tab Navigation */}
            <Tabs index={activeTab} onChange={setActiveTab} w="full" variant="enclosed">
              <TabList>
                <Tab fontSize="sm">🔍 Search</Tab>
                <Tab fontSize="sm">🏢 Popular</Tab>
                <Tab fontSize="sm">🕐 Recent</Tab>
                <Tab fontSize="sm">✏️ Manual</Tab>
              </TabList>

              <TabPanels>
                {/* Search Tab */}
                <TabPanel px={0}>
                  <VStack spacing={3}>
                    <FormControl>
                      <InputGroup>
                        <InputLeftElement>
                          <FiSearch />
                        </InputLeftElement>
                        <Input
                          placeholder="Search locations (e.g., Jakarta, Borobudur)"
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                          bg="gray.50"
                        />
                      </InputGroup>
                    </FormControl>

                    {isSearching && (
                      <HStack>
                        <Spinner size="sm" />
                        <Text fontSize="sm" color="gray.600">Searching...</Text>
                      </HStack>
                    )}

                    {searchResults.length > 0 && (
                      <VStack spacing={2} w="full">
                        {searchResults.map((location, index) => (
                          <Button
                            key={index}
                            variant="ghost"
                            size="sm"
                            onClick={() => selectLocation(location)}
                            w="full"
                            h="auto"
                            py={3}
                            justifyContent="flex-start"
                            leftIcon={<FiMapPin />}
                          >
                            <Box textAlign="left" w="full">
                              <Text fontSize="sm" fontWeight="medium">
                                {location.name}
                              </Text>
                              <Text fontSize="xs" color="gray.500" noOfLines={1}>
                                {location.address}
                              </Text>
                            </Box>
                          </Button>
                        ))}
                      </VStack>
                    )}

                    {searchQuery && !isSearching && searchResults.length === 0 && (
                      <Text fontSize="sm" color="gray.500">
                        No results found for "{searchQuery}"
                      </Text>
                    )}
                  </VStack>
                </TabPanel>

                {/* Popular Tab */}
                <TabPanel px={0}>
                  <SimpleGrid columns={1} spacing={2}>
                    {popularLocations.map((location, index) => (
                      <Button
                        key={index}
                        variant="ghost"
                        size="sm"
                        onClick={() => selectLocation(location)}
                        w="full"
                        h="auto"
                        py={3}
                        justifyContent="flex-start"
                        leftIcon={<FiMapPin />}
                      >
                        <Box textAlign="left" w="full">
                          <Text fontSize="sm" fontWeight="medium">
                            {location.name}
                          </Text>
                          <Text fontSize="xs" color="gray.500" noOfLines={1}>
                            {location.address}
                          </Text>
                        </Box>
                      </Button>
                    ))}
                  </SimpleGrid>
                </TabPanel>

                {/* Recent Tab */}
                <TabPanel px={0}>
                  {recentLocations.length > 0 ? (
                    <SimpleGrid columns={1} spacing={2}>
                      {recentLocations.map((location, index) => (
                        <Button
                          key={index}
                          variant="ghost"
                          size="sm"
                          onClick={() => selectLocation(location)}
                          w="full"
                          h="auto"
                          py={3}
                          justifyContent="flex-start"
                          leftIcon={<FiCrosshair />}
                        >
                          <Box textAlign="left" w="full">
                            <Text fontSize="sm" fontWeight="medium">
                              {location.name}
                            </Text>
                            <Text fontSize="xs" color="gray.500" noOfLines={1}>
                              {location.address}
                            </Text>
                          </Box>
                        </Button>
                      ))}
                    </SimpleGrid>
                  ) : (
                    <Text fontSize="sm" color="gray.500" textAlign="center" py={4}>
                      No recent locations yet
                    </Text>
                  )}
                </TabPanel>

                {/* Manual Tab */}
                <TabPanel px={0}>
                  <VStack spacing={3}>
                    <FormControl>
                      <FormLabel fontSize="sm">
                        <HStack>
                          <Icon as={FiMousePointer} />
                          <Text>Enter Coordinates</Text>
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
                        💡 Right-click on Google Maps → Copy coordinates → Paste here
                      </Text>
                    </FormControl>

                    {manualInput && !validateCoordinates(manualInput) && (
                      <Alert status="warning" size="sm">
                        <AlertIcon />
                        <AlertDescription fontSize="sm">
                          Invalid format. Use: latitude,longitude
                        </AlertDescription>
                      </Alert>
                    )}
                  </VStack>
                </TabPanel>
              </TabPanels>
            </Tabs>

            {/* Selected Location Display */}
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
                  
                  <Box w="full" p={3} bg="white" borderRadius="md">
                    <VStack spacing={2} align="stretch">
                      <Box>
                        <Text fontSize="xs" color="gray.600">Coordinates:</Text>
                        <Text fontSize="sm" fontFamily="mono" color="green.800" fontWeight="medium">
                          📍 {selectedCoordinates}
                        </Text>
                      </Box>
                      {selectedAddress && (
                        <Box>
                          <Text fontSize="xs" color="gray.600">Location:</Text>
                          <Text fontSize="sm" color="gray.700">
                            🏢 {selectedAddress}
                          </Text>
                        </Box>
                      )}
                    </VStack>
                  </Box>
                </VStack>
              </Box>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <HStack spacing={3}>
            <Button variant="ghost" onClick={handleClear} size="sm">
              Clear
            </Button>
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

export default AdvancedLocationPicker;
