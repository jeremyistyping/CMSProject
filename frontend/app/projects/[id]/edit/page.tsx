'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
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
  Spinner,
  Center,
} from '@chakra-ui/react';
import { FiArrowLeft, FiSave } from 'react-icons/fi';
import Layout from '@/components/layout/UnifiedLayout';
import projectService from '@/services/projectService';
import { ProjectFormData, Project } from '@/types/project';

export default function EditProjectPage() {
  const router = useRouter();
  const params = useParams();
  const toast = useToast();
  const projectId = params?.id as string;

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [project, setProject] = useState<Project | null>(null);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const inputBgColor = useColorModeValue('gray.50', 'var(--bg-primary)');

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

  // Fetch project data on mount
  useEffect(() => {
    if (projectId) {
      fetchProject();
    }
  }, [projectId]);

  const fetchProject = async () => {
    try {
      setLoading(true);
      const data = await projectService.getProjectById(projectId);
      setProject(data);

      // Pre-fill form with existing data
      // Convert deadline to YYYY-MM-DD format for input[type="date"]
      const deadlineDate = data.deadline ? new Date(data.deadline).toISOString().split('T')[0] : '';

      setFormData({
        project_name: data.project_name || '',
        project_description: data.project_description || '',
        customer: data.customer || '',
        city: data.city || '',
        address: data.address || '',
        project_type: data.project_type || 'New Build',
        budget: data.budget || 0,
        deadline: deadlineDate,
        overall_progress: data.overall_progress || 0,
        foundation_progress: data.foundation_progress || 0,
        utilities_progress: data.utilities_progress || 0,
        interior_progress: data.interior_progress || 0,
        equipment_progress: data.equipment_progress || 0,
      });
    } catch (error) {
      console.error('Error fetching project:', error);
      toast({
        title: 'Error',
        description: 'Failed to load project data. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      // Redirect back to projects list if fetch fails
      router.push('/projects');
    } finally {
      setLoading(false);
    }
  };

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

    // Validation
    if (!formData.project_name.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Project name is required',
        status: 'warning',
        duration: 3000,
      });
      return;
    }

    if (!formData.customer.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Customer is required',
        status: 'warning',
        duration: 3000,
      });
      return;
    }

    try {
      setSubmitting(true);

      // Format deadline to ISO 8601 for backend
      const formattedData = {
        ...formData,
        deadline: formData.deadline ? new Date(formData.deadline).toISOString() : new Date().toISOString(),
      };

      // Call backend API to update project
      const updatedProject = await projectService.updateProject(projectId, formattedData);

      toast({
        title: 'Success',
        description: 'Project updated successfully',
        status: 'success',
        duration: 3000,
      });

      // Redirect to project detail page
      router.push(`/projects/${projectId}`);
    } catch (error) {
      console.error('Error updating project:', error);

      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to update project. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancel = () => {
    // Go back to project detail page
    router.push(`/projects/${projectId}`);
  };

  const formatBudget = (value: number) => {
    if (value === 0) return '';
    return new Intl.NumberFormat('id-ID').format(value);
  };

  const formatToMillion = (value: number) => {
    return (value / 1000000).toFixed(1);
  };

  // Show loading spinner while fetching
  if (loading) {
    return (
      <Layout>
        <Center h="50vh">
          <VStack spacing={4}>
            <Spinner size="xl" color="blue.500" />
            <Text color={textColor}>Loading project data...</Text>
          </VStack>
        </Center>
      </Layout>
    );
  }

  // Show error if project not found
  if (!project) {
    return (
      <Layout>
        <Center h="50vh">
          <VStack spacing={4}>
            <Text fontSize="xl" color={textColor}>
              Project not found
            </Text>
            <Button onClick={() => router.push('/projects')}>Back to Projects</Button>
          </VStack>
        </Center>
      </Layout>
    );
  }

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
          Back to Project Detail
        </Button>

        {/* Form Card */}
        <Card bg={bgColor} borderColor={borderColor} borderWidth="1px">
          <CardBody p={8}>
            <VStack align="stretch" spacing={4} mb={6}>
              <Heading size="lg" color={textColor}>
                Edit Project
              </Heading>
              <Text color={subtextColor} fontSize="sm">
                Update project information and progress. All changes will be saved to the database.
              </Text>
            </VStack>

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
                    bg={inputBgColor}
                    borderColor={borderColor}
                    isDisabled={submitting}
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
                    bg={inputBgColor}
                    borderColor={borderColor}
                    isDisabled={submitting}
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
                        bg={inputBgColor}
                        borderColor={borderColor}
                        isDisabled={submitting}
                      />
                    </FormControl>
                  </GridItem>
                  <GridItem>
                    <FormControl isRequired>
                      <FormLabel color={textColor}>City</FormLabel>
                      <Input
                        name="city"
                        placeholder="e.g. Jakarta"
                        value={formData.city}
                        onChange={handleInputChange}
                        bg={inputBgColor}
                        borderColor={borderColor}
                        isDisabled={submitting}
                      />
                    </FormControl>
                  </GridItem>
                </Grid>

                {/* Address */}
                <FormControl isRequired>
                  <FormLabel color={textColor}>Address</FormLabel>
                  <Input
                    name="address"
                    placeholder="e.g. Jl. Sudirman No. 123, Jakarta Pusat"
                    value={formData.address}
                    onChange={handleInputChange}
                    bg={inputBgColor}
                    borderColor={borderColor}
                    isDisabled={submitting}
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
                        bg={inputBgColor}
                        borderColor={borderColor}
                        isDisabled={submitting}
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
                        placeholder="e.g. 500000000"
                        value={formData.budget || ''}
                        onChange={handleInputChange}
                        bg={inputBgColor}
                        borderColor={borderColor}
                        isDisabled={submitting}
                      />
                      {formData.budget > 0 && (
                        <Text fontSize="xs" color={subtextColor} mt={1}>
                          Rp {formatBudget(formData.budget)} ({formatToMillion(formData.budget)} juta)
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
                    bg={inputBgColor}
                    borderColor={borderColor}
                    isDisabled={submitting}
                  />
                </FormControl>

                {/* Progress Percentages */}
                <Box>
                  <Heading size="sm" mb={4} color={textColor}>
                    Project Progress (%)
                  </Heading>
                  <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                    <GridItem colSpan={2}>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">
                          Overall Progress (%)
                        </FormLabel>
                        <Input
                          name="overall_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.overall_progress}
                          onChange={handleInputChange}
                          bg={inputBgColor}
                          borderColor={borderColor}
                          isDisabled={submitting}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">
                          Foundation (%)
                        </FormLabel>
                        <Input
                          name="foundation_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.foundation_progress}
                          onChange={handleInputChange}
                          bg={inputBgColor}
                          borderColor={borderColor}
                          isDisabled={submitting}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">
                          Utilities (%)
                        </FormLabel>
                        <Input
                          name="utilities_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.utilities_progress}
                          onChange={handleInputChange}
                          bg={inputBgColor}
                          borderColor={borderColor}
                          isDisabled={submitting}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">
                          Interior (%)
                        </FormLabel>
                        <Input
                          name="interior_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.interior_progress}
                          onChange={handleInputChange}
                          bg={inputBgColor}
                          borderColor={borderColor}
                          isDisabled={submitting}
                        />
                      </FormControl>
                    </GridItem>
                    <GridItem>
                      <FormControl>
                        <FormLabel color={textColor} fontSize="sm">
                          Equipment (%)
                        </FormLabel>
                        <Input
                          name="equipment_progress"
                          type="number"
                          min="0"
                          max="100"
                          value={formData.equipment_progress}
                          onChange={handleInputChange}
                          bg={inputBgColor}
                          borderColor={borderColor}
                          isDisabled={submitting}
                        />
                      </FormControl>
                    </GridItem>
                  </Grid>
                </Box>

                {/* Action Buttons */}
                <HStack spacing={4} justify="flex-end" pt={4} borderTopWidth="1px" borderColor={borderColor}>
                  <Button variant="ghost" onClick={handleCancel} isDisabled={submitting}>
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    colorScheme="blue"
                    leftIcon={<FiSave />}
                    isLoading={submitting}
                    loadingText="Updating..."
                  >
                    Update Project
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

