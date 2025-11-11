'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  Button,
  Heading,
  Card,
  CardBody,
  Grid,
  Text,
  Badge,
  HStack,
  VStack,
  Icon,
  useColorModeValue,
  Spinner,
  Center,
  useToast,
} from '@chakra-ui/react';
import { FiPlus, FiFolder, FiCalendar, FiDollarSign, FiArrowRight } from 'react-icons/fi';
import Layout from '@/components/layout/UnifiedLayout';
import projectService from '@/services/projectService';
import { Project } from '@/types/project';

// Mock data removed - now using real backend API

export default function ProjectsPage() {
  const router = useRouter();
  const toast = useToast();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const progressBgColor = useColorModeValue('gray.200', 'gray.700');

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    try {
      setLoading(true);
      const data = await projectService.getAllProjects();
      // Handle null or undefined data
      setProjects(data || []);
    } catch (error) {
      console.error('Error fetching projects:', error);
      // Show empty state when backend fails
      setProjects([]);
      toast({
        title: 'Error',
        description: 'Failed to load projects. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreateProject = () => {
    router.push('/projects/create');
  };

  const handleViewProject = (projectId: string) => {
    router.push(`/projects/${projectId}`);
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(value);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'green';
      case 'completed':
        return 'green';
      case 'on-hold':
        return 'yellow';
      case 'archived':
        return 'gray';
      default:
        return 'gray';
    }
  };

  if (loading) {
    return (
      <Layout>
      </Layout>
    );
  }

  return (
    <Layout>
      <Box p={8}>
        {/* Header */}
        <HStack justify="space-between" mb={8}>
          <VStack align="start" spacing={1}>
            <Heading size="lg" color={textColor}>
              Project Management
            </Heading>
            <Text color={subtextColor}>Manage your construction projects</Text>
          </VStack>
          <Button
            leftIcon={<FiPlus />}
            colorScheme="green"
            size="lg"
            onClick={handleCreateProject}
          >
            Create New Project
          </Button>
        </HStack>

        {/* Projects Grid */}
        {projects.length === 0 ? (
          <Center h="40vh">
            <VStack spacing={4}>
              <Icon as={FiFolder} boxSize={16} color="gray.400" />
              <Text color={subtextColor} fontSize="lg">
                No projects yet
              </Text>
              <Button leftIcon={<FiPlus />} colorScheme="green" onClick={handleCreateProject}>
                Create Your First Project
              </Button>
            </VStack>
          </Center>
        ) : (
          <Grid templateColumns="repeat(auto-fill, minmax(350px, 1fr))" gap={6}>
            {projects.map((project) => (
              <Card
                key={project.id}
                bg={bgColor}
                borderColor={borderColor}
                borderWidth="1px"
                transition="all 0.3s"
                _hover={{
                  transform: 'translateY(-4px)',
                  shadow: 'lg',
                  borderColor: 'green.400',
                }}
                cursor="pointer"
                onClick={() => handleViewProject(project.id)}
              >
                <CardBody>
                  <VStack align="stretch" spacing={4}>
                    {/* Header */}
                    <HStack justify="space-between">
                      <Badge colorScheme={getStatusColor(project.status)} fontSize="xs">
                        {project.status.toUpperCase()}
                      </Badge>
                      <Badge colorScheme="purple" fontSize="xs">
                        {project.project_type}
                      </Badge>
                    </HStack>

                    {/* Project Name */}
                    <Heading size="md" color={textColor} noOfLines={2}>
                      {project.project_name}
                    </Heading>

                    {/* Customer & Location */}
                    <VStack align="start" spacing={1}>
                      <Text fontSize="sm" color={textColor} fontWeight="medium">
                        {project.customer}
                      </Text>
                      <Text fontSize="sm" color={subtextColor}>
                        {project.city}
                      </Text>
                    </VStack>

                    {/* Progress */}
                    <Box>
                      <HStack justify="space-between" mb={2}>
                        <Text fontSize="sm" color={subtextColor}>
                          Overall Progress
                        </Text>
                        <Text fontSize="sm" fontWeight="bold" color={textColor}>
                          {project.overall_progress}%
                        </Text>
                      </HStack>
                      <Box
                        w="full"
                        h="8px"
                        bg={progressBgColor}
                        borderRadius="full"
                        overflow="hidden"
                      >
                        <Box
                          w={`${project.overall_progress}%`}
                          h="full"
                          bg="green.500"
                          transition="width 0.3s"
                        />
                      </Box>
                    </Box>

                    {/* Budget & Deadline */}
                    <HStack spacing={4}>
                      <HStack flex={1} spacing={2}>
                        <Icon as={FiDollarSign} color="green.500" />
                        <VStack align="start" spacing={0}>
                          <Text fontSize="xs" color={subtextColor}>
                            Budget
                          </Text>
                          <Text fontSize="sm" fontWeight="medium" color={textColor}>
                            {formatCurrency(project.budget)}
                          </Text>
                        </VStack>
                      </HStack>
                      <HStack flex={1} spacing={2}>
                        <Icon as={FiCalendar} color="green.500" />
                        <VStack align="start" spacing={0}>
                          <Text fontSize="xs" color={subtextColor}>
                            Deadline
                          </Text>
                          <Text fontSize="sm" fontWeight="medium" color={textColor}>
                            {formatDate(project.deadline)}
                          </Text>
                        </VStack>
                      </HStack>
                    </HStack>

                    {/* View Details Button */}
                    <Button
                      rightIcon={<FiArrowRight />}
                      variant="ghost"
                      colorScheme="green"
                      size="sm"
                      w="full"
                    >
                      View Details
                    </Button>
                  </VStack>
                </CardBody>
              </Card>
            ))}
          </Grid>
        )}
      </Box>
    </Layout>
  );
}

