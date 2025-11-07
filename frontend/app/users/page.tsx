'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { DataTable } from '@/components/common/DataTable';
import {
  Box,
  Heading,
  Text,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner,
  Button,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  Select,
  useDisclosure,
  useToast,
  VStack,
  HStack,
  IconButton,
  Tooltip
} from '@chakra-ui/react';
import { FiEye, FiEdit, FiTrash2, FiShield } from 'react-icons/fi';
import UserPermissions from '@/components/users/UserPermissions';

interface User {
  id: number;
  username: string;
  email: string;
  full_name: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface CreateUserForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  first_name: string;
  last_name: string;
  role: string;
  department: string;
  position: string;
  phone: string;
}

// Columns will be defined dynamically in the component

const UsersPage: React.FC = () => {
  const { user, token } = useAuth();
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();
  
  // State for view modal
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [viewUser, setViewUser] = useState<User | null>(null);
  
  // State for edit modal
  const [isEditMode, setIsEditMode] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  
  // State for permissions modal
  const [isPermissionsModalOpen, setIsPermissionsModalOpen] = useState(false);
  const [permissionsUser, setPermissionsUser] = useState<User | null>(null);
  
  const [formData, setFormData] = useState<CreateUserForm>({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    first_name: '',
    last_name: '',
    role: 'employee',
    department: '',
    position: '',
    phone: '',
  });

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const res = await fetch(`/api/v1/users`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
        });

        if (!res.ok) {
          throw new Error('Failed to fetch users data');
        }

        const response = await res.json();
        // Backend returns { data: [...], pagination: {...} }
        const usersData = response.data || [];
        setUsers(usersData.map((user: any) => ({
          ...user,
          status: user.is_active ? t('common.active') : t('common.inactive'),
          created_at: new Date(user.created_at).toLocaleDateString(),
        })));
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const resetForm = () => {
    setFormData({
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      first_name: '',
      last_name: '',
      role: 'employee',
      department: '',
      position: '',
      phone: '',
    });
  };

  // handleCreateUser is now replaced by handleSaveUser above

  const handleModalClose = () => {
    resetForm();
    setIsEditMode(false);
    setSelectedUser(null);
    onClose();
  };
  
  // Handle view user
  const handleView = (user: User) => {
    setViewUser(user);
    setIsViewModalOpen(true);
  };
  
  // Handle permissions
  const handlePermissions = (user: User) => {
    setPermissionsUser(user);
    setIsPermissionsModalOpen(true);
  };
  
  // Handle edit user
  const handleEdit = (user: User) => {
    setSelectedUser(user);
    setFormData({
      username: user.username,
      email: user.email,
      password: '', // Leave empty for editing
      confirmPassword: '',
      first_name: user.full_name.split(' ')[0] || '',
      last_name: user.full_name.split(' ').slice(1).join(' ') || '',
      role: user.role,
      department: '',
      position: '',
      phone: '',
    });
    setIsEditMode(true);
    onOpen();
  };
  
  // Handle delete user
  const handleDelete = async (userId: number) => {
    if (!window.confirm(t('messages.confirmDelete'))) {
      return;
    }
    
    try {
      const response = await fetch(`/api/v1/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete user');
      }
      
      // Refresh users list
      const res = await fetch('/api/v1/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (res.ok) {
        const response = await res.json();
        const usersData = response.data || [];
        setUsers(usersData.map((user: any) => ({
          ...user,
          status: user.is_active ? 'Active' : 'Inactive',
          created_at: new Date(user.created_at).toLocaleDateString(),
        })));
      }
      
      toast({
        title: t('messages.success'),
        description: t('messages.dataDeleted'),
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      toast({
        title: t('common.error'),
        description: error.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };
  
  // Update handleCreateUser to handle both create and edit
  const handleSaveUser = async () => {
    // Validation
    if (!formData.username.trim()) {
      toast({
        title: 'Error',
        description: 'Username is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!formData.email.trim()) {
      toast({
        title: 'Error',
        description: 'Email is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    // Only validate password for new users
    if (!isEditMode) {
      if (!formData.password.trim() || formData.password.length < 6) {
        toast({
          title: 'Error',
          description: 'Password must be at least 6 characters',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
        return;
      }

      if (formData.password !== formData.confirmPassword) {
        toast({
          title: 'Error',
          description: 'Passwords do not match',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
        return;
      }
    }

    setCreating(true);

    try {
      const { confirmPassword, ...createData } = formData;
      
      // Remove password fields if editing and password is empty
      if (isEditMode && !createData.password) {
        delete createData.password;
      }
      
      const url = isEditMode 
        ? `/api/v1/users/${selectedUser?.id}`
        : '/api/v1/users';
        
      const response = await fetch(url, {
        method: isEditMode ? 'PUT' : 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(createData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Failed to ${isEditMode ? 'update' : 'create'} user`);
      }

      const newUser = await response.json();
      
      // Refresh the users list
      const res = await fetch('/api/v1/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (res.ok) {
        const response = await res.json();
        const usersData = response.data || [];
        setUsers(usersData.map((user: any) => ({
          ...user,
          status: user.is_active ? 'Active' : 'Inactive',
          created_at: new Date(user.created_at).toLocaleDateString(),
        })));
      }

      toast({
        title: t('messages.success'),
        description: isEditMode ? t('messages.dataUpdated') : t('messages.dataSaved'),
        status: 'success',
        duration: 3000,
        isClosable: true,
      });

      handleModalClose();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
<SimpleLayout allowedRoles={['admin']}>
        <Box>
          <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
          <Text ml={4}>{t('common.loading')}</Text>
        </Box>
      </SimpleLayout>
    );
  }

  // Define columns with action buttons
  const columns = [
    { header: t('users.name'), accessor: 'full_name' },
    { header: t('users.email'), accessor: 'email' },
    { header: t('users.role'), accessor: 'role' },
    { header: t('users.status'), accessor: 'status' },
    { header: t('users.created'), accessor: 'created_at' },
    {
      header: t('users.actions'),
      accessor: (user: User) => (
        <HStack spacing={2}>
          <Tooltip label="View Details" placement="top">
            <IconButton
              aria-label="View user"
              icon={<FiEye />}
              size="sm"
              variant="outline"
              colorScheme="blue"
              onClick={() => handleView(user)}
            />
          </Tooltip>
          <Tooltip label="Edit User" placement="top">
            <IconButton
              aria-label="Edit user"
              icon={<FiEdit />}
              size="sm"
              variant="outline"
              colorScheme="green"
              onClick={() => handleEdit(user)}
            />
          </Tooltip>
          <Tooltip label="Manage Permissions" placement="top">
            <IconButton
              aria-label="Manage permissions"
              icon={<FiShield />}
              size="sm"
              variant="outline"
              colorScheme="purple"
              onClick={() => handlePermissions(user)}
            />
          </Tooltip>
          <Tooltip label="Delete User" placement="top">
            <IconButton
              aria-label="Delete user"
              icon={<FiTrash2 />}
              size="sm"
              variant="outline"
              colorScheme="red"
              onClick={() => handleDelete(user.id)}
            />
          </Tooltip>
        </HStack>
      ),
    },
  ];

  return (
<SimpleLayout allowedRoles={['admin']}>
      <Box>
        <HStack justify="space-between" align="center" mb={6}>
          <Heading as="h1" size="xl">Users Management</Heading>
          <Button 
            colorScheme="blue" 
            onClick={() => {
              setIsEditMode(false);
              resetForm();
              onOpen();
            }}
            size="md"
          >
            Create Account
          </Button>
        </HStack>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle>Error:</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <Box bg="white" borderRadius="lg" overflow="hidden" boxShadow="sm">
          <DataTable 
            columns={columns} 
            data={users} 
            keyField="id"
            title="Users List"
            searchable={true}
            pagination={true}
            pageSize={10}
          />
        </Box>

        {/* Create/Edit User Modal */}
        <Modal isOpen={isOpen} onClose={handleModalClose} size="xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>{isEditMode ? 'Edit User Account' : 'Create New User Account'}</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4}>
                <HStack spacing={4} width="100%">
                  <FormControl isRequired>
                    <FormLabel>First Name</FormLabel>
                    <Input
                      name="first_name"
                      value={formData.first_name}
                      onChange={handleInputChange}
                      placeholder="Enter first name"
                    />
                  </FormControl>
                  <FormControl isRequired>
                    <FormLabel>Last Name</FormLabel>
                    <Input
                      name="last_name"
                      value={formData.last_name}
                      onChange={handleInputChange}
                      placeholder="Enter last name"
                    />
                  </FormControl>
                </HStack>

                <HStack spacing={4} width="100%">
                  <FormControl isRequired>
                    <FormLabel>Username</FormLabel>
                    <Input
                      name="username"
                      value={formData.username}
                      onChange={handleInputChange}
                      placeholder="Enter username"
                    />
                  </FormControl>
                  <FormControl isRequired>
                    <FormLabel>Email</FormLabel>
                    <Input
                      type="email"
                      name="email"
                      value={formData.email}
                      onChange={handleInputChange}
                      placeholder="Enter email address"
                    />
                  </FormControl>
                </HStack>

                <HStack spacing={4} width="100%">
                  <FormControl isRequired={!isEditMode}>
                    <FormLabel>Password {isEditMode && '(leave empty to keep current)'}</FormLabel>
                    <Input
                      type="password"
                      name="password"
                      value={formData.password}
                      onChange={handleInputChange}
                      placeholder={isEditMode ? "Leave empty to keep current password" : "Enter password (min. 6 characters)"}
                    />
                  </FormControl>
                  <FormControl isRequired={!isEditMode}>
                    <FormLabel>Confirm Password</FormLabel>
                    <Input
                      type="password"
                      name="confirmPassword"
                      value={formData.confirmPassword}
                      onChange={handleInputChange}
                      placeholder="Confirm password"
                    />
                  </FormControl>
                </HStack>

                <HStack spacing={4} width="100%">
                  <FormControl isRequired>
                    <FormLabel>Role</FormLabel>
                    <Select
                      name="role"
                      value={formData.role}
                      onChange={handleInputChange}
                    >
                      <option value="employee">Employee</option>
                      <option value="admin">Administrator</option>
                      <option value="finance">Finance</option>
                      <option value="director">Director</option>
                      <option value="inventory_manager">Inventory Manager</option>
                    </Select>
                  </FormControl>
                  <FormControl>
                    <FormLabel>Phone</FormLabel>
                    <Input
                      name="phone"
                      value={formData.phone}
                      onChange={handleInputChange}
                      placeholder="Enter phone number"
                    />
                  </FormControl>
                </HStack>

                <HStack spacing={4} width="100%">
                  <FormControl>
                    <FormLabel>Department</FormLabel>
                    <Input
                      name="department"
                      value={formData.department}
                      onChange={handleInputChange}
                      placeholder="Enter department"
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Position</FormLabel>
                    <Input
                      name="position"
                      value={formData.position}
                      onChange={handleInputChange}
                      placeholder="Enter position"
                    />
                  </FormControl>
                </HStack>
              </VStack>
            </ModalBody>

            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={handleModalClose}>
                Cancel
              </Button>
              <Button 
                colorScheme="blue" 
                onClick={handleSaveUser}
                isLoading={creating}
                loadingText={isEditMode ? "Updating..." : "Creating..."}
              >
                {isEditMode ? 'Update Account' : 'Create Account'}
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
        
        {/* View User Modal */}
        <Modal isOpen={isViewModalOpen} onClose={() => setIsViewModalOpen(false)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>User Details</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              {viewUser && (
                <VStack spacing={3} align="stretch">
                  <Box>
                    <Text fontWeight="bold" display="inline">Full Name: </Text>
                    <Text display="inline">{viewUser.full_name}</Text>
                  </Box>
                  <Box>
                    <Text fontWeight="bold" display="inline">Username: </Text>
                    <Text display="inline">{viewUser.username}</Text>
                  </Box>
                  <Box>
                    <Text fontWeight="bold" display="inline">Email: </Text>
                    <Text display="inline">{viewUser.email}</Text>
                  </Box>
                  <Box>
                    <Text fontWeight="bold" display="inline">Role: </Text>
                    <Text display="inline">{viewUser.role}</Text>
                  </Box>
                  <Box>
                    <Text fontWeight="bold" display="inline">Status: </Text>
                    <Text display="inline">{viewUser.is_active ? 'Active' : 'Inactive'}</Text>
                  </Box>
                  <Box>
                    <Text fontWeight="bold" display="inline">Created Date: </Text>
                    <Text display="inline">{new Date(viewUser.created_at).toLocaleString()}</Text>
                  </Box>
                  <Box>
                    <Text fontWeight="bold" display="inline">Last Updated: </Text>
                    <Text display="inline">{new Date(viewUser.updated_at).toLocaleString()}</Text>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button onClick={() => setIsViewModalOpen(false)}>Close</Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
        
        {/* Permissions Modal */}
        <UserPermissions
          isOpen={isPermissionsModalOpen}
          onClose={() => {
            setIsPermissionsModalOpen(false);
            setPermissionsUser(null);
          }}
          user={permissionsUser}
          token={token || ''}
        />
      </Box>
    </SimpleLayout>
  );
};

export default UsersPage;
