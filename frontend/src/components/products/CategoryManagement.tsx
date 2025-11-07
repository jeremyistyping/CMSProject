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
import ProductService, { Category } from '@/services/productService';
import CategoryForm from './CategoryForm';

const CategoryManagement: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [filteredCategories, setFilteredCategories] = useState<Category[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [categoryToDelete, setCategoryToDelete] = useState<Category | null>(null);
  const { isOpen: isFormOpen, onOpen: onFormOpen, onClose: onFormClose } = useDisclosure();
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const cancelRef = React.useRef<HTMLButtonElement>(null);
  const toast = useToast();

  useEffect(() => {
    fetchCategories();
  }, []);

  useEffect(() => {
    if (searchTerm) {
      const filtered = categories.filter(cat =>
        cat.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        cat.code.toLowerCase().includes(searchTerm.toLowerCase())
      );
      setFilteredCategories(filtered);
    } else {
      setFilteredCategories(categories);
    }
  }, [searchTerm, categories]);

  const fetchCategories = async () => {
    try {
      const data = await ProductService.getCategories();
      setCategories(data.data || []);
    } catch (error) {
      toast({
        title: 'Failed to fetch categories',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleAddClick = () => {
    setSelectedCategory(null);
    onFormOpen();
  };

  const handleEditClick = (category: Category) => {
    setSelectedCategory(category);
    onFormOpen();
  };

  const handleDeleteClick = (category: Category) => {
    setCategoryToDelete(category);
    onDeleteOpen();
  };

  const confirmDelete = async () => {
    if (!categoryToDelete?.id) return;

    try {
      await ProductService.deleteCategory(categoryToDelete.id);
      toast({
        title: 'Category deleted successfully',
        status: 'success',
        isClosable: true,
      });
      fetchCategories();
      onDeleteClose();
    } catch (error: any) {
      toast({
        title: 'Failed to delete category',
        description: error?.response?.data?.error || 'An error occurred',
        status: 'error',
        isClosable: true,
      });
    }
  };

  const handleSaveCategory = (category: Category) => {
    fetchCategories();
    onFormClose();
    setSelectedCategory(null);
  };

  const handleCancelForm = () => {
    onFormClose();
    setSelectedCategory(null);
  };

  return (
    <Box>
      {/* Header with Add Button and Search */}
      <Flex justify="space-between" align="center" mb={4}>
        <Button
          leftIcon={<FiPlus />}
          colorScheme="green"
          onClick={handleAddClick}
        >
          Add Category
        </Button>

        <InputGroup maxW="300px">
          <InputLeftElement pointerEvents="none">
            <FiSearch color="gray.300" />
          </InputLeftElement>
          <Input
            placeholder="Search categories..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </InputGroup>
      </Flex>

      {/* Categories Table */}
      <Box overflowX="auto">
        <Table variant="simple" size="sm">
          <Thead>
            <Tr>
              <Th>Code</Th>
              <Th>Name</Th>
              <Th>Description</Th>
              <Th>Parent</Th>
              <Th>Status</Th>
              <Th>Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {filteredCategories.length === 0 ? (
              <Tr>
                <Td colSpan={6} textAlign="center" py={8}>
                  <Text color="gray.500">
                    {searchTerm ? 'No categories found matching your search' : 'No categories yet. Add your first category!'}
                  </Text>
                </Td>
              </Tr>
            ) : (
              filteredCategories.map((category) => (
                <Tr key={category.id}>
                  <Td fontWeight="medium">{category.code}</Td>
                  <Td>{category.name}</Td>
                  <Td>{category.description || '-'}</Td>
                  <Td>{category.parent?.name || '-'}</Td>
                  <Td>
                    <Badge colorScheme={category.is_active ? 'green' : 'red'}>
                      {category.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </Td>
                  <Td>
                    <HStack spacing={2}>
                      <Button
                        size="sm"
                        leftIcon={<FiEdit />}
                        colorScheme="blue"
                        variant="ghost"
                        onClick={() => handleEditClick(category)}
                      >
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        leftIcon={<FiTrash2 />}
                        colorScheme="red"
                        variant="ghost"
                        onClick={() => handleDeleteClick(category)}
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

      {/* Add/Edit Category Modal */}
      <Modal isOpen={isFormOpen} onClose={onFormClose} size="4xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>
            {selectedCategory ? 'Edit Category' : 'Add Category'}
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <CategoryForm
              category={selectedCategory || undefined}
              onSave={handleSaveCategory}
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
              Delete Category
            </AlertDialogHeader>

            <AlertDialogBody>
              Are you sure you want to delete <strong>{categoryToDelete?.name}</strong>? 
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

export default CategoryManagement;
