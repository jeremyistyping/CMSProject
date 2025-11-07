'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import api from '@/services/api';
import { API_ENDPOINTS } from '@/config/api';
import Layout from '@/components/layout/Layout';
import GroupedTable from '@/components/common/GroupedTable';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useToast,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  Switch,
  VStack,
  HStack,
} from '@chakra-ui/react';
import { FiPlus, FiEdit, FiTrash2, FiEye } from 'react-icons/fi';

// Define the Contact type
interface Contact {
  id: number;
  code?: string;
  name: string;
  type: 'CUSTOMER' | 'VENDOR' | 'EMPLOYEE';
  category?: string;
  email: string;
  phone: string;
  mobile?: string;
  fax?: string;
  website?: string;
  tax_number?: string;
  credit_limit?: number;
  payment_terms?: number;
  is_active: boolean;
  pic_name?: string;        // Person In Charge (for Customer/Vendor)
  external_id?: string;     // Employee ID, Vendor ID, Customer ID
  address?: string;         // Simple address field
  notes?: string;
  created_at: string;
  updated_at: string;
  addresses?: ContactAddress[];
}

interface ContactAddress {
  id: number;
  contact_id: number;
  type: 'BILLING' | 'SHIPPING' | 'MAILING';
  address1: string;
  address2?: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
  is_default: boolean;
}

const ContactsPage = () => {
  const { token, user } = useAuth();
  const { t } = useTranslation();
  const canEdit = user?.role?.toLowerCase() === 'admin' || user?.role?.toLowerCase() === 'finance' || user?.role?.toLowerCase() === 'inventory_manager';
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedContact, setSelectedContact] = useState<Partial<Contact> | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState<Partial<Contact>>({
    name: '',
    type: 'CUSTOMER',
    email: '',
    phone: '',
    mobile: '',
    notes: '',
    pic_name: '',
    external_id: '',
    address: '',
    is_active: true
  });
  // Fetch contacts from API
  const fetchContacts = async () => {
    try {
      const response = await api.get(API_ENDPOINTS.CONTACTS);
      // Backend returns direct array, not wrapped in data field
      setContacts(Array.isArray(response.data) ? response.data : response.data.data || []);
    } catch (err: any) {
      setError('Failed to fetch contacts. Please try again.');
      console.error('Error fetching contacts:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Load contacts on component mount
  useEffect(() => {
    if (token) {
      fetchContacts();
    }
  }, [token]);

  // Handle form submission for create/update
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    
    try {
      let response;
      if (formData.id) {
        response = await api.put(`${API_ENDPOINTS.CONTACTS}/${formData.id}`, formData);
      } else {
        response = await api.post(API_ENDPOINTS.CONTACTS, formData);
      }
      
      // Refresh contacts list
      fetchContacts();
      
      // Show success message
      toast({
        title: formData.id ? 'Contact Updated' : 'Contact Created',
        description: `Contact has been ${formData.id ? 'updated' : 'created'} successfully.`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
      // Close modal and reset form
      setIsModalOpen(false);
      setSelectedContact(null);
      setFormData({
        name: '',
        type: 'CUSTOMER',
        email: '',
        phone: '',
        mobile: '',
        notes: '',
        pic_name: '',
        external_id: '',
        address: '',
        is_active: true
      });
    } catch (err) {
      setError(`Error ${formData.id ? 'updating' : 'creating'} contact. Please try again.`);
      console.error('Error submitting contact:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle contact deletion
  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this contact?')) {
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      await api.delete(`${API_ENDPOINTS.CONTACTS}/${id}`);
      
      // Refresh contacts list after successful deletion
      fetchContacts();
      
      // Show success message
      toast({
        title: 'Contact Deleted',
        description: 'Contact has been deleted successfully.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (err) {
      setError('Error deleting contact. Please try again.');
      console.error('Error deleting contact:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Open modal for creating a new contact
  const handleCreate = () => {
    setSelectedContact(null);
    setFormData({
      name: '',
      type: 'CUSTOMER',
      email: '',
      phone: '',
      mobile: '',
      notes: '',
      pic_name: '',
      external_id: '',
      address: '',
      is_active: true
    });
    setIsModalOpen(true);
  };

  // Open modal for editing an existing contact
  const handleEdit = (contact: Contact) => {
    setSelectedContact(contact);
    setFormData(contact);
    setIsModalOpen(true);
  };

  // Handle form input changes
  const handleInputChange = (field: keyof Contact, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  // Table columns definition (removed Type column since we're grouping by type)
  // Dynamic columns based on contact type
  const getColumnsForType = (contactType?: string) => {
    const baseColumns = [
      { 
        header: 'Name', 
        accessor: 'name',
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold' },
        cellStyle: { padding: '12px 8px', fontSize: '14px' }
      },
      { 
        header: 'External ID', 
        accessor: (contact: Contact) => contact.external_id || '-',
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold', whiteSpace: 'nowrap' },
        cellStyle: { padding: '12px 8px', fontSize: '14px', whiteSpace: 'nowrap' }
      },
    ];
    
    // Only add PIC Name column for Customer and Vendor groups, not for Employee
    if (contactType !== 'EMPLOYEE') {
      baseColumns.push({
        header: 'PIC Name', 
        accessor: (contact: Contact) => contact.pic_name || '-',
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold', whiteSpace: 'nowrap' },
        cellStyle: { padding: '12px 8px', fontSize: '14px', whiteSpace: 'nowrap' }
      });
    }
    
    // Add remaining columns
    baseColumns.push(
      { 
        header: 'Email', 
        accessor: 'email',
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold' },
        cellStyle: { padding: '12px 8px', fontSize: '14px', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }
      },
      { 
        header: 'Phone', 
        accessor: 'phone',
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold', whiteSpace: 'nowrap' },
        cellStyle: { padding: '12px 8px', fontSize: '14px', whiteSpace: 'nowrap' }
      },
      { 
        header: 'Address', 
        accessor: (contact: Contact) => {
          if (contact.address) {
            // Truncate long address for table display
            return contact.address.length > 50 
              ? contact.address.substring(0, 50) + '...' 
              : contact.address;
          }
          return '-';
        },
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold' },
        cellStyle: { padding: '12px 8px', fontSize: '14px', maxWidth: '250px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }
      },
      { 
        header: 'Status', 
        accessor: (contact: Contact) => (contact.is_active ? 'Active' : 'Inactive'),
        headerStyle: { padding: '12px 8px', fontSize: '14px', fontWeight: 'semibold', whiteSpace: 'nowrap' },
        cellStyle: { padding: '12px 8px', fontSize: '14px', whiteSpace: 'nowrap' }
      }
    );
    
    return baseColumns;
  };
  
  // Default columns (for backward compatibility)
  const columns = getColumnsForType();

  const toast = useToast();

  // New state for view modal
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [viewContact, setViewContact] = useState<Contact | null>(null);

  // Handler to open view modal
  const handleView = (contact: Contact) => {
    setViewContact(contact);
    setIsViewModalOpen(true);
  };

  // Action buttons for each row
  const renderActions = (contact: Contact) => (
    <>
      <Button
        size="xs"
        variant="outline"
        leftIcon={<FiEye />}
        onClick={() => handleView(contact)}
        colorScheme="blue"
        minW="auto"
        px={2}
      >
        View
      </Button>
      {canEdit && (
        <>
          <Button
            size="xs"
            variant="outline"
            leftIcon={<FiEdit />}
            onClick={() => handleEdit(contact)}
            minW="auto"
            px={2}
          >
            Edit
          </Button>
          <Button
            size="xs"
            colorScheme="red"
            variant="outline"
            leftIcon={<FiTrash2 />}
            onClick={() => handleDelete(contact.id)}
            minW="auto"
            px={2}
          >
            Delete
          </Button>
        </>
      )}
    </>
  );

  return (
<Layout allowedRoles={['admin', 'finance', 'inventory_manager', 'employee', 'director']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Contact Master</Heading>
          {canEdit && (
            <Button
              colorScheme="brand"
              leftIcon={<FiPlus />}
              onClick={handleCreate}
            >
              {t('contacts.addContact')}
            </Button>
          )}
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <GroupedTable<Contact>
          columns={getColumnsForType}
          data={contacts}
          keyField="id"
          groupBy="type"
          title="Contacts"
          actions={renderActions}
          isLoading={isLoading}
          groupLabels={{
            VENDOR: 'Vendors',
            CUSTOMER: 'Customers', 
            EMPLOYEE: 'Employees'
          }}
        />
        
        <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <form onSubmit={handleSubmit}>
              <ModalHeader>
                {selectedContact?.id ? t('contacts.editContact') : t('contacts.createContact')}
              </ModalHeader>
              <ModalCloseButton />
              <ModalBody>
                <VStack spacing={4}>
                  <FormControl isRequired>
                    <FormLabel>{t('common.name')}</FormLabel>
                    <Input
                      value={formData.name || ''}
                      onChange={(e) => handleInputChange('name', e.target.value)}
                      placeholder="Enter contact name"
                    />
                  </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel>{t('common.type')}</FormLabel>
                    <Select
                      value={formData.type || 'CUSTOMER'}
                      onChange={(e) => handleInputChange('type', e.target.value as 'CUSTOMER' | 'VENDOR' | 'EMPLOYEE')}
                    >
                      <option value="CUSTOMER">{t('contacts.customer')}</option>
                      <option value="VENDOR">{t('contacts.vendor')}</option>
                      <option value="EMPLOYEE">{t('contacts.employee')}</option>
                    </Select>
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel>
                      {formData.type === 'CUSTOMER' ? 'Customer ID' : 
                       formData.type === 'VENDOR' ? 'Vendor ID' : 
                       formData.type === 'EMPLOYEE' ? 'Employee ID' : 'External ID'}
                    </FormLabel>
                    <Input
                      value={formData.external_id || ''}
                      onChange={(e) => handleInputChange('external_id', e.target.value)}
                      placeholder={`Enter ${formData.type?.toLowerCase() || 'external'} ID`}
                    />
                  </FormControl>
                  
                  {/* PIC Name - only show for Customer/Vendor */}
                  {(formData.type === 'CUSTOMER' || formData.type === 'VENDOR') && (
                    <FormControl>
                      <FormLabel>{t('contacts.picName')}</FormLabel>
                      <Input
                        value={formData.pic_name || ''}
                        onChange={(e) => handleInputChange('pic_name', e.target.value)}
                        placeholder="Enter person in charge name"
                      />
                    </FormControl>
                  )}
                  
                  <FormControl isRequired>
                    <FormLabel>{t('contacts.email')}</FormLabel>
                    <Input
                      type="email"
                      value={formData.email || ''}
                      onChange={(e) => handleInputChange('email', e.target.value)}
                      placeholder="Enter email address"
                    />
                  </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel>{t('contacts.phone')}</FormLabel>
                    <Input
                      value={formData.phone || ''}
                      onChange={(e) => handleInputChange('phone', e.target.value)}
                      placeholder="Enter phone number"
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel>{t('contacts.mobile')}</FormLabel>
                    <Input
                      value={formData.mobile || ''}
                      onChange={(e) => handleInputChange('mobile', e.target.value)}
                      placeholder="Enter mobile number"
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel>{t('contacts.address')}</FormLabel>
                    <Textarea
                      value={formData.address || ''}
                      onChange={(e) => handleInputChange('address', e.target.value)}
                      placeholder="Enter complete address"
                      rows={3}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel>{t('common.notes')}</FormLabel>
                    <Textarea
                      value={formData.notes || ''}
                      onChange={(e) => handleInputChange('notes', e.target.value)}
                      placeholder="Enter notes"
                      rows={3}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <HStack>
                      <FormLabel mb={0}>{t('common.active')}</FormLabel>
                      <Switch
                        isChecked={formData.is_active !== false}
                        onChange={(e) => handleInputChange('is_active', e.target.checked)}
                      />
                    </HStack>
                  </FormControl>
                </VStack>
              </ModalBody>
              <ModalFooter>
                <Button variant="ghost" mr={3} onClick={() => setIsModalOpen(false)}>
                  {t('common.cancel')}
                </Button>
                <Button
                  colorScheme="brand"
                  type="submit"
                  isLoading={isSubmitting}
                  loadingText={selectedContact?.id ? t('common.updating') : t('common.creating')}
                >
                  {selectedContact?.id ? t('common.update') : t('common.create')}
                </Button>
              </ModalFooter>
            </form>
          </ModalContent>
        </Modal>
        
        {/* View Modal */}
        <Modal isOpen={isViewModalOpen} onClose={() => setIsViewModalOpen(false)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>{t('contacts.contactDetails')}</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              {viewContact && (
                <VStack spacing={3} align="stretch">
                  <Box>
                    <strong>Name:</strong> {viewContact.name}
                  </Box>
                  <Box>
                    <strong>Type:</strong> {viewContact.type}
                  </Box>
                  <Box>
                    <strong>Code:</strong> {viewContact.code || '-'}
                  </Box>
                  <Box>
                    <strong>External ID:</strong> {viewContact.external_id || '-'}
                  </Box>
                  {(viewContact.type === 'CUSTOMER' || viewContact.type === 'VENDOR') && (
                    <Box>
                      <strong>PIC Name:</strong> {viewContact.pic_name || '-'}
                    </Box>
                  )}
                  <Box>
                    <strong>Email:</strong> {viewContact.email}
                  </Box>
                  <Box>
                    <strong>Phone:</strong> {viewContact.phone}
                  </Box>
                  <Box>
                    <strong>Mobile:</strong> {viewContact.mobile || '-'}
                  </Box>
                  <Box>
                    <strong>Address:</strong> {viewContact.address || '-'}
                  </Box>
                  <Box>
                    <strong>Status:</strong> {viewContact.is_active ? 'Active' : 'Inactive'}
                  </Box>
                  <Box>
                    <strong>Notes:</strong> {viewContact.notes || '-'}
                  </Box>
                  <Box>
                    <strong>Created:</strong> {new Date(viewContact.created_at).toLocaleDateString()}
                  </Box>
                  <Box>
                    <strong>Updated:</strong> {new Date(viewContact.updated_at).toLocaleDateString()}
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button onClick={() => setIsViewModalOpen(false)}>Close</Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </Box>
    </Layout>
  );
};

export default ContactsPage;
