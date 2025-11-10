'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  Button,
  Card,
  CardBody,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  Select,
  Grid,
  GridItem,
  Heading,
  HStack,
  VStack,
  Text,
  useColorModeValue,
  useToast,
  Icon,
} from '@chakra-ui/react';
import { FiArrowLeft } from 'react-icons/fi';
import Layout from '@/components/layout/UnifiedLayout';
import projectService from '@/services/projectService';
import { ProjectFormData } from '@/types/project';

export default function CreateProjectPage() {
  const router = useRouter();
  const toast = useToast();
  const [loading, setLoading] = useState(false);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');

  const [formData, setFormData] = useState<ProjectFormData>({
    project_name: '',
    project_description: '',
    customer: '',
    city: '',
    address: '',
    project_type: 'New Build',
    budget: 0,
    deadline: '',
    overall_progress: 0,
    foundation_progress: 0,
    utilities_progress: 0,
    interior_progress: 0,
    equipment_progress: 0,
  });

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: name.includes('progress') || name === 'budget' ? Number(value) : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      setLoading(true);
      const project = await projectService.createProject(formData);
      
      // Check if backend actually returned valid data
      if (!project || !project.id) {
        throw new Error('Invalid response from server');
      }
      
      toast({
        title: 'Success',
        description: 'Project created successfully',
        status: 'success',
        duration: 3000,
      });
      
      // Redirect to project detail page
      router.push(`/projects/${project.id}`);
    } catch (error) {
      console.error('Error creating project:', error);
      
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to create project. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    router.push('/projects');
  };

  const formatBudget = (value: number) => {
    if (value === 0) return '';
    return new Intl.NumberFormat('id-ID').format(value);
  };

  const formatToBillion = (value: number) => {
    return (value / 1000000000).toFixed(1);
  };

  return (
    <Layout>
      <Box p={8} maxW="900px" mx="auto">
        {/* Back Button */}
        <Button
          leftIcon={<FiArrowLeft />}
          variant="ghost"
          onClick={handleCancel}
          mb={6}
          color={textColor}
        >
          Back to Projects
        </Button>

        {/* Form Card */}
        <Card bg={bgColor} borderColor={borderColor} borderWidth="1px">
          <CardBody p={8}>
            <Heading size="lg" mb={6} color={textColor}>
              Create New Project
            </Heading>

            <form onSubmit={handleSubmit}>
              <VStack spacing={6} align="stretch">
                {/* Project Name */}
                <FormControl isRequired>
                  <FormLabel color={textColor}>Project Name</FormLabel>
                  <Input
                    name="project_name"
                    placeholder="e.g. Downtown Restaurant Kitchen"
                    value={formData.project_name}
                    onChange={handleInputChange}
                    bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                    borderColor={borderColor}
                  />
                </FormControl>

                {/* Project Description */}
                <FormControl isRequired>
                  <FormLabel color={textColor}>Project Description</FormLabel>
                  <Textarea
                    name="project_description"
                    placeholder="Describe the project scope, goals, and key features..."
                    value={formData.project_description}
                    onChange={handleInputChange}
                    rows={4}
                    bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                    borderColor={borderColor}
                  />
                </FormControl>

                {/* Customer & City */}
                <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                  <GridItem>
                    <FormControl isRequired>
                      <FormLabel color={textColor}>Customer</FormLabel>
                      <Input
                        name="customer"
                        placeholder="e.g. Downtown Bistro LLC"
                        value={formData.customer}
                        onChange={handleInputChange}
                        bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                        borderColor={borderColor}
                      />
                    </FormControl>
                  </GridItem>
                  <GridItem>
                    <FormControl isRequired>
                      <FormLabel color={textColor}>City</FormLabel>
                      <Input
                        name="city"
                        placeholder="e.g. Seattle"
                        value={formData.city}
                        onChange={handleInputChange}
                        bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                        borderColor={borderColor}
                      />
                    </FormControl>
                  </GridItem>
                </Grid>

                {/* Address */}
                <FormControl isRequired>
                  <FormLabel color={textColor}>Address</FormLabel>
                  <Input
                    name="address"
                    placeholder="e.g. 1234 Pike Street, Seattle, WA 98101"
                    value={formData.address}
                    onChange={handleInputChange}
                    bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                    borderColor={borderColor}
                  />
                </FormControl>

                {/* Project Type & Budget */}
                <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                  <GridItem>
                    <FormControl isRequired>
                      <FormLabel color={textColor}>Project Type</FormLabel>
                      <Select
                        name="project_type"
                        value={formData.project_type}
                        onChange={handleInputChange}
                        bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                        borderColor={borderColor}
                      >
                        <option value="New Build">New Build</option>
                        <option value="Renovation">Renovation</option>
                        <option value="Expansion">Expansion</option>
                        <option value="Maintenance">Maintenance</option>
                      </Select>
                    </FormControl>
                  </GridItem>
                  <GridItem>
                    <FormControl isRequired>
                      <FormLabel color={textColor}>Budget (IDR)</FormLabel>
                      <Input
                        name="budget"
                        type="number"
                        placeholder="Rp 500.000.000 (500 juta rupiah)"
                        value={formData.budget || ''}
                        onChange={handleInputChange}
                        bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                        borderColor={borderColor}
                      />
                      {formData.budget > 0 && (
                        <Text fontSize="xs" color={subtextColor} mt={1}>
                          Minimal: Rp {formatBudget(formData.budget)} ({formatToBillion(formData.budget)} juta rupiah)
                        </Text>
                      )}
                    </FormControl>
                  </GridItem>
                </Grid>

                {/* Deadline */}
                <FormControl isRequired>
                  <FormLabel color={textColor}>Deadline</FormLabel>
                  <Input
                    name="deadline"
                    type="date"
                    value={formData.deadline}
                    onChange={handleInputChange}
                    bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                    borderColor={borderColor}
                  />
                </FormControl>

                {/* Progress Percentages */}
                <Box>
                  <Heading size="sm" mb={4} color={textColor}>
                    Initial Progress (%)
                  </Heading>
                  <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                    <GridItem colSpan={2}>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">Overall Progress (%)</FormLabel>
                        <Input
                          name="overall_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.overall_progress}
                          onChange={handleInputChange}
                          bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                          borderColor={borderColor}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">Foundation (%)</FormLabel>
                        <Input
                          name="foundation_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.foundation_progress}
                          onChange={handleInputChange}
                          bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                          borderColor={borderColor}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">Utilities (%)</FormLabel>
                        <Input
                          name="utilities_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.utilities_progress}
                          onChange={handleInputChange}
                          bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                          borderColor={borderColor}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">Interior (%)</FormLabel>
                        <Input
                          name="interior_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.interior_progress}
                          onChange={handleInputChange}
                          bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                          borderColor={borderColor}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">Equipment (%)</FormLabel>
                        <Input
                          name="equipment_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.equipment_progress}
                          onChange={handleInputChange}
                          bg={useColorModeValue('gray.50', 'var(--bg-primary)')}
                          borderColor={borderColor}
                        />
                      </FormControl>
                    </GridItem>
                  </Grid>
                </Box>

                {/* Action Buttons */}
                <HStack spacing={4} justify="flex-end" pt={4}>
                  <Button variant="ghost" onClick={handleCancel} isDisabled={loading}>
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    colorScheme="blue"
                    isLoading={loading}
                    loadingText="Creating..."
                  >
                    Create Project
                  </Button>
                </HStack>
              </VStack>
            </form>
          </CardBody>
        </Card>
      </Box>
    </Layout>
  );
}

