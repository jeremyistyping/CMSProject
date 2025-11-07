import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  useToast,
  HStack,
  Badge,
  Input,
  InputGroup,
  InputLeftElement,
  Flex,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  Text,
} from '@chakra-ui/react';
import { FiEdit, FiTrash2, FiPlus, FiSearch } from 'react-icons/fi';
import ProductService, { WarehouseLocation } from '@/services/productService';
import WarehouseLocationForm from './WarehouseLocationForm';

const WarehouseLocationManagement: React.FC = () => {
  const [locations, setLocations] = useState<WarehouseLocation[]>([]);
  const [filteredLocations, setFilteredLocations] = useState<WarehouseLocation[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLocation, setSelectedLocation] = useState<WarehouseLocation | null>(null);
  const [locationToDelete, setLocationToDelete] = useState<WarehouseLocation | null>(null);
  const { isOpen: isFormOpen, onOpen: onFormOpen, onClose: onFormClose } = useDisclosure();
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const cancelRef = React.useRef<HTMLButtonElement>(null);
  const toast = useToast();

  useEffect(() => {
    fetchLocations();
  }, []);

  useEffect(() => {
    if (searchTerm) {
      const filtered = locations.filter(loc =>
        loc.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        loc.code.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (loc.address && loc.address.toLowerCase().includes(searchTerm.toLowerCase()))
      );
      setFilteredLocations(filtered);
    } else {
      setFilteredLocations(locations);
    }
  }, [searchTerm, locations]);

  const fetchLocations = async () => {
    try {
      const data = await ProductService.getWarehouseLocations();
      setLocations(data.data || []);
    } catch (error) {
      toast({
        title: 'Failed to fetch warehouse locations',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleAddClick = () => {
    setSelectedLocation(null);
    onFormOpen();
  };

  const handleEditClick = (location: WarehouseLocation) => {
    setSelectedLocation(location);
    onFormOpen();
  };

  const handleDeleteClick = (location: WarehouseLocation) => {
    setLocationToDelete(location);
    onDeleteOpen();
  };

  const confirmDelete = async () => {
    if (!locationToDelete?.id) return;

    try {
      await ProductService.deleteWarehouseLocation(locationToDelete.id);
      toast({
        title: 'Warehouse location deleted successfully',
        status: 'success',
        isClosable: true,
      });
      fetchLocations();
      onDeleteClose();
    } catch (error: any) {
      toast({
        title: 'Failed to delete warehouse location',
        description: error?.response?.data?.error || 'An error occurred',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleSaveLocation = (location: WarehouseLocation) => {
    fetchLocations();
    onFormClose();
    setSelectedLocation(null);
  };

  const handleCancelForm = () => {
    onFormClose();
    setSelectedLocation(null);
  };

  return (
    <Box>
      {/* Header with Add Button and Search */}
      <Flex justify="space-between" align="center" mb={4}>
        <Button
          leftIcon={<FiPlus />}
          colorScheme="orange"
          onClick={handleAddClick}
        >
          Add Warehouse Location
        </Button>

        <InputGroup maxW="300px">
          <InputLeftElement pointerEvents="none">
            <FiSearch color="gray.300" />
          </InputLeftElement>
          <Input
            placeholder="Search locations..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </InputGroup>
      </Flex>

      {/* Warehouse Locations Table */}
      <Box overflowX="auto">
        <Table variant="simple" size="sm">
          <Thead>
            <Tr>
              <Th>Code</Th>
              <Th>Name</Th>
              <Th>Address</Th>
              <Th>Description</Th>
              <Th>Status</Th>
              <Th>Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {filteredLocations.length === 0 ? (
              <Tr>
                <Td colSpan={6} textAlign="center" py={8}>
                  <Text color="gray.500">
                    {searchTerm ? 'No warehouse locations found matching your search' : 'No warehouse locations yet. Add your first location!'}
                  </Text>
                </Td>
              </Tr>
            ) : (
              filteredLocations.map((location) => (
                <Tr key={location.id}>
                  <Td fontWeight="medium">{location.code}</Td>
                  <Td>{location.name}</Td>
                  <Td>{location.address || '-'}</Td>
                  <Td>{location.description || '-'}</Td>
                  <Td>
                    <Badge colorScheme={location.is_active ? 'green' : 'red'}>
                      {location.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </Td>
                  <Td>
                    <HStack spacing={2}>
                      <Button
                        size="sm"
                        leftIcon={<FiEdit />}
                        colorScheme="blue"
                        variant="ghost"
                        onClick={() => handleEditClick(location)}
                      >
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        leftIcon={<FiTrash2 />}
                        colorScheme="red"
                        variant="ghost"
                        onClick={() => handleDeleteClick(location)}
                      >
                        Delete
                      </Button>
                    </HStack>
                  </Td>
                </Tr>
              ))
            )}
          </Tbody>
        </Table>
      </Box>

      {/* Add/Edit Warehouse Location Modal */}
      <Modal isOpen={isFormOpen} onClose={onFormClose} size="4xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>
            {selectedLocation ? 'Edit Warehouse Location' : 'Add Warehouse Location'}
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <WarehouseLocationForm
              location={selectedLocation || undefined}
              onSave={handleSaveLocation}
              onCancel={handleCancelForm}
            />
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        isOpen={isDeleteOpen}
        leastDestructiveRef={cancelRef}
        onClose={onDeleteClose}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Delete Warehouse Location
            </AlertDialogHeader>

            <AlertDialogBody>
              Are you sure you want to delete <strong>{locationToDelete?.name}</strong>? 
              This action cannot be undone.
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onDeleteClose}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={confirmDelete} ml={3}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </Box>
  );
};

export default WarehouseLocationManagement;
