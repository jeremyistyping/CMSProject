'use client';

import React, { useState } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Input,
  Select,
  VStack,
  HStack,
  Grid,
  GridItem,
  Switch,
  useToast,
  Divider,
  Text,
  Card,
  CardBody,
  CardHeader,
  IconButton,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiPlus, FiTrash2, FiEdit } from 'react-icons/fi';
import { ContactAddress } from '@/types/contact';

interface ContactAddressFormProps {
  addresses: ContactAddress[];
  onAddressChange: (addresses: ContactAddress[]) => void;
}

export default function ContactAddressForm({ 
  addresses = [], 
  onAddressChange 
}: ContactAddressFormProps) {
  const toast = useToast();
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [newAddress, setNewAddress] = useState<Partial<ContactAddress>>({
    type: 'BILLING',
    address1: '',
    address2: '',
    city: '',
    state: '',
    postal_code: '',
    country: 'Indonesia',
    is_default: false,
  });

  const textColor = useColorModeValue('gray.600', 'var(--text-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');

  const handleAddAddress = () => {
    if (!newAddress.address1 || !newAddress.city) {
      toast({
        title: 'Validation Error',
        description: 'Address and city are required',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    const addressToAdd: ContactAddress = {
      id: Date.now(), // Temporary ID for new addresses
      contact_id: 0, // Will be set by backend
      type: newAddress.type as 'BILLING' | 'SHIPPING' | 'MAILING',
      address1: newAddress.address1!,
      address2: newAddress.address2 || '',
      city: newAddress.city!,
      state: newAddress.state || '',
      postal_code: newAddress.postal_code || '',
      country: newAddress.country || 'Indonesia',
      is_default: newAddress.is_default || false,
    };

    const updatedAddresses = [...addresses, addressToAdd];
    onAddressChange(updatedAddresses);

    // Reset form
    setNewAddress({
      type: 'BILLING',
      address1: '',
      address2: '',
      city: '',
      state: '',
      postal_code: '',
      country: 'Indonesia',
      is_default: false,
    });
  };

  const handleRemoveAddress = (index: number) => {
    const updatedAddresses = addresses.filter((_, i) => i !== index);
    onAddressChange(updatedAddresses);
  };

  const handleUpdateAddress = (index: number, updatedAddress: ContactAddress) => {
    const updatedAddresses = addresses.map((addr, i) => 
      i === index ? updatedAddress : addr
    );
    onAddressChange(updatedAddresses);
    setEditingIndex(null);
  };

  const getAddressTypeLabel = (type: string) => {
    switch (type) {
      case 'BILLING': return 'Billing';
      case 'SHIPPING': return 'Shipping';
      case 'MAILING': return 'Mailing';
      default: return type;
    }
  };

  const getAddressTypeColor = (type: string) => {
    switch (type) {
      case 'BILLING': return 'blue';
      case 'SHIPPING': return 'green';
      case 'MAILING': return 'orange';
      default: return 'gray';
    }
  };

  return (
    <VStack spacing={4} align="stretch">
      <Text fontSize="lg" fontWeight="semibold">Contact Addresses</Text>

      {/* Existing Addresses */}
      {addresses.map((address, index) => (
        <Card key={index} variant="outline">
          <CardHeader pb={2}>
            <HStack justify="space-between">
              <HStack>
                <Text fontSize="sm" fontWeight="medium" color={`${getAddressTypeColor(address.type)}.500`}>
                  {getAddressTypeLabel(address.type)}
                </Text>
                {address.is_default && (
                  <Text fontSize="xs" color="green.500" fontWeight="medium">
                    DEFAULT
                  </Text>
                )}
              </HStack>
              <HStack spacing={1}>
                <IconButton
                  aria-label="Edit address"
                  icon={<FiEdit />}
                  size="sm"
                  variant="ghost"
                  onClick={() => setEditingIndex(index)}
                />
                <IconButton
                  aria-label="Remove address"
                  icon={<FiTrash2 />}
                  size="sm"
                  variant="ghost"
                  colorScheme="red"
                  onClick={() => handleRemoveAddress(index)}
                />
              </HStack>
            </HStack>
          </CardHeader>
          <CardBody pt={0}>
            {editingIndex === index ? (
              // Edit mode
              <VStack spacing={3} align="stretch">
                <Grid templateColumns={{ base: '1fr', md: '1fr 1fr' }} gap={3}>
                  <GridItem>
                    <FormControl>
                      <FormLabel fontSize="sm">Address Type</FormLabel>
                      <Select 
                        size="sm"
                        value={address.type}
                        onChange={(e) => handleUpdateAddress(index, { 
                          ...address, 
                          type: e.target.value as 'BILLING' | 'SHIPPING' | 'MAILING' 
                        })}
                      >
                        <option value="BILLING">Billing</option>
                        <option value="SHIPPING">Shipping</option>
                        <option value="MAILING">Mailing</option>
                      </Select>
                    </FormControl>
                  </GridItem>
                  <GridItem>
                    <FormControl display="flex" alignItems="center">
                      <FormLabel fontSize="sm" mb="0">Default Address</FormLabel>
                      <Switch
                        size="sm"
                        isChecked={address.is_default}
                        onChange={(e) => handleUpdateAddress(index, { 
                          ...address, 
                          is_default: e.target.checked 
                        })}
                      />
                    </FormControl>
                  </GridItem>
                </Grid>
                
                <FormControl>
                  <FormLabel fontSize="sm">Address Line 1</FormLabel>
                  <Input 
                    size="sm"
                    value={address.address1}
                    onChange={(e) => handleUpdateAddress(index, { 
                      ...address, 
                      address1: e.target.value 
                    })}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm">Address Line 2</FormLabel>
                  <Input 
                    size="sm"
                    value={address.address2 || ''}
                    onChange={(e) => handleUpdateAddress(index, { 
                      ...address, 
                      address2: e.target.value 
                    })}
                  />
                </FormControl>
                
                <Grid templateColumns={{ base: '1fr', md: '1fr 1fr 1fr' }} gap={3}>
                  <GridItem>
                    <FormControl>
                      <FormLabel fontSize="sm">City</FormLabel>
                      <Input 
                        size="sm"
                        value={address.city}
                        onChange={(e) => handleUpdateAddress(index, { 
                          ...address, 
                          city: e.target.value 
                        })}
                      />
                    </FormControl>
                  </GridItem>
                  <GridItem>
                    <FormControl>
                      <FormLabel fontSize="sm">State/Province</FormLabel>
                      <Input 
                        size="sm"
                        value={address.state || ''}
                        onChange={(e) => handleUpdateAddress(index, { 
                          ...address, 
                          state: e.target.value 
                        })}
                      />
                    </FormControl>
                  </GridItem>
                  <GridItem>
                    <FormControl>
                      <FormLabel fontSize="sm">Postal Code</FormLabel>
                      <Input 
                        size="sm"
                        value={address.postal_code || ''}
                        onChange={(e) => handleUpdateAddress(index, { 
                          ...address, 
                          postal_code: e.target.value 
                        })}
                      />
                    </FormControl>
                  </GridItem>
                </Grid>
                
                <FormControl>
                  <FormLabel fontSize="sm">Country</FormLabel>
                  <Input 
                    size="sm"
                    value={address.country}
                    onChange={(e) => handleUpdateAddress(index, { 
                      ...address, 
                      country: e.target.value 
                    })}
                  />
                </FormControl>
                
                <HStack>
                  <Button size="sm" colorScheme="blue" onClick={() => setEditingIndex(null)}>
                    Save
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => setEditingIndex(null)}>
                    Cancel
                  </Button>
                </HStack>
              </VStack>
            ) : (
              // View mode
              <VStack align="start" spacing={1}>
                <Text fontSize="sm" fontWeight="medium">{address.address1}</Text>
                {address.address2 && (
                  <Text fontSize="sm" color={textColor}>{address.address2}</Text>
                )}
                <Text fontSize="sm" color={textColor}>
                  {address.city}{address.state ? `, ${address.state}` : ''} {address.postal_code}
                </Text>
                <Text fontSize="sm" color={textColor}>{address.country}</Text>
              </VStack>
            )}
          </CardBody>
        </Card>
      ))}

      {/* Add New Address Form */}
      <Card variant="outline" borderStyle="dashed">
        <CardHeader pb={2}>
          <Text fontSize="md" fontWeight="medium">Add New Address</Text>
        </CardHeader>
        <CardBody pt={0}>
          <VStack spacing={3} align="stretch">
            <Grid templateColumns={{ base: '1fr', md: '1fr 1fr' }} gap={3}>
              <GridItem>
                <FormControl>
                  <FormLabel fontSize="sm">Address Type</FormLabel>
                  <Select 
                    size="sm"
                    value={newAddress.type || 'BILLING'}
                    onChange={(e) => setNewAddress({ 
                      ...newAddress, 
                      type: e.target.value as 'BILLING' | 'SHIPPING' | 'MAILING' 
                    })}
                  >
                    <option value="BILLING">Billing</option>
                    <option value="SHIPPING">Shipping</option>
                    <option value="MAILING">Mailing</option>
                  </Select>
                </FormControl>
              </GridItem>
              <GridItem>
                <FormControl display="flex" alignItems="center">
                  <FormLabel fontSize="sm" mb="0">Set as Default</FormLabel>
                  <Switch
                    size="sm"
                    isChecked={newAddress.is_default || false}
                    onChange={(e) => setNewAddress({ 
                      ...newAddress, 
                      is_default: e.target.checked 
                    })}
                  />
                </FormControl>
              </GridItem>
            </Grid>
            
            <FormControl isRequired>
              <FormLabel fontSize="sm">Address Line 1</FormLabel>
              <Input 
                size="sm"
                value={newAddress.address1 || ''}
                onChange={(e) => setNewAddress({ 
                  ...newAddress, 
                  address1: e.target.value 
                })}
                placeholder="Enter street address"
              />
            </FormControl>
            
            <FormControl>
              <FormLabel fontSize="sm">Address Line 2</FormLabel>
              <Input 
                size="sm"
                value={newAddress.address2 || ''}
                onChange={(e) => setNewAddress({ 
                  ...newAddress, 
                  address2: e.target.value 
                })}
                placeholder="Apartment, suite, etc."
              />
            </FormControl>
            
            <Grid templateColumns={{ base: '1fr', md: '1fr 1fr 1fr' }} gap={3}>
              <GridItem>
                <FormControl isRequired>
                  <FormLabel fontSize="sm">City</FormLabel>
                  <Input 
                    size="sm"
                    value={newAddress.city || ''}
                    onChange={(e) => setNewAddress({ 
                      ...newAddress, 
                      city: e.target.value 
                    })}
                    placeholder="Enter city"
                  />
                </FormControl>
              </GridItem>
              <GridItem>
                <FormControl>
                  <FormLabel fontSize="sm">State/Province</FormLabel>
                  <Input 
                    size="sm"
                    value={newAddress.state || ''}
                    onChange={(e) => setNewAddress({ 
                      ...newAddress, 
                      state: e.target.value 
                    })}
                    placeholder="Enter state"
                  />
                </FormControl>
              </GridItem>
              <GridItem>
                <FormControl>
                  <FormLabel fontSize="sm">Postal Code</FormLabel>
                  <Input 
                    size="sm"
                    value={newAddress.postal_code || ''}
                    onChange={(e) => setNewAddress({ 
                      ...newAddress, 
                      postal_code: e.target.value 
                    })}
                    placeholder="Enter postal code"
                  />
                </FormControl>
              </GridItem>
            </Grid>
            
            <FormControl>
              <FormLabel fontSize="sm">Country</FormLabel>
              <Input 
                size="sm"
                value={newAddress.country || 'Indonesia'}
                onChange={(e) => setNewAddress({ 
                  ...newAddress, 
                  country: e.target.value 
                })}
                placeholder="Enter country"
              />
            </FormControl>
            
            <Button
              leftIcon={<FiPlus />}
              size="sm"
              colorScheme="blue"
              onClick={handleAddAddress}
            >
              Add Address
            </Button>
          </VStack>
        </CardBody>
      </Card>
    </VStack>
  );
}
