'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import {
  Box,
  Button,
  Card,
  CardBody,
  Heading,
  HStack,
  VStack,
  Text,
  Badge,
  Icon,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Grid,
  GridItem,
  useColorModeValue,
  Spinner,
  Center,
  useToast,
  IconButton,
} from '@chakra-ui/react';
import {
  FiArrowLeft,
  FiArchive,
  FiEdit,
  FiDollarSign,
  FiCalendar,
  FiBarChart,
  FiFileText,
  FiTarget,
  FiClock,
  FiDatabase,
} from 'react-icons/fi';
import Layout from '@/components/layout/UnifiedLayout';
import projectService from '@/services/projectService';
import { Project } from '@/types/project';
import UpdateProgressModal from '@/components/projects/UpdateProgressModal';
import DailyUpdatesTab from '@/components/projects/DailyUpdatesTab';
import MilestonesTab from '@/components/projects/MilestonesTab';
import WeeklyReportsTab from '@/components/projects/WeeklyReportsTab';

// Mock data untuk demo
const MOCK_PROJECT: Project = {
  id: '1',
  project_name: 'Downtown Restaurant Kitchen Renovation',
  project_description: 'Complete kitchen renovation including new equipment and utilities',
  customer: 'Downtown Bistro LLC',
  city: 'Jakarta Pusat',
  address: 'Jl. Sudirman No. 123, Jakarta Pusat',
  project_type: 'Renovation',
  budget: 500000000,
  deadline: '2025-12-31',
  overall_progress: 45,
  foundation_progress: 100,
  utilities_progress: 80,
  interior_progress: 30,
  equipment_progress: 10,
  status: 'active',
  created_at: '2025-01-15',
  updated_at: '2025-11-10',
};

export default function ProjectDetailPage() {
  const router = useRouter();
  const params = useParams();
  const toast = useToast();
  const projectId = params?.id as string;

  const [project, setProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState(0);
  const [isProgressModalOpen, setIsProgressModalOpen] = useState(false);

  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');
  const tabBgColor = useColorModeValue('green.500', 'green.400');

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
    } catch (error) {
      console.error('Error fetching project:', error);
      // Use mock data when backend is not ready
      setProject(MOCK_PROJECT);
      toast({
        title: 'Demo Mode',
        description: 'Showing demo project data.',
        status: 'info',
        duration: 3000,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    router.push('/projects');
  };

  const handleEdit = () => {
    router.push(`/projects/${projectId}/edit`);
  };

  const handleArchive = async () => {
    try {
      await projectService.archiveProject(projectId);
      toast({
        title: 'Success',
        description: 'Project archived successfully',
        status: 'success',
        duration: 3000,
      });
      router.push('/projects');
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to archive project',
        status: 'error',
        duration: 3000,
      });
    }
  };

  const handleUpdateProgress = () => {
    setIsProgressModalOpen(true);
  };

  const handleProgressModalClose = () => {
    setIsProgressModalOpen(false);
  };

  const handleProgressUpdateSuccess = () => {
    // Refresh project data after progress update
    fetchProject();
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
      month: 'long',
      day: 'numeric',
    });
  };

  if (loading) {
    return (
      <Layout>
        <Center h="50vh">
          <Spinner size="xl" color="green.500" />
        </Center>
      </Layout>
    );
  }

  if (!project) {
    return (
      <Layout>
        <Center h="50vh">
          <VStack spacing={4}>
            <Text fontSize="xl" color={textColor}>
              Project not found
            </Text>
            <Button onClick={handleBack}>Back to Projects</Button>
          </VStack>
        </Center>
      </Layout>
    );
  }

  return (
    <Layout>
      <Box p={8}>
        {/* Back Button */}
        <Button leftIcon={<FiArrowLeft />} variant="ghost" onClick={handleBack} mb={6}>
          Back to Projects
        </Button>

        {/* Project Header */}
        <Card bg={bgColor} borderColor={borderColor} borderWidth="1px" mb={6}>
          <CardBody>
            <VStack align="stretch" spacing={4}>
              {/* Title & Actions */}
              <HStack justify="space-between">
                <VStack align="start" spacing={2}>
                  <HStack spacing={3}>
                    <Icon as={FiBarChart} boxSize={8} color="green.500" />
                    <Heading size="xl" color={textColor}>
                      {project.project_name}
                    </Heading>
                  </HStack>
                  <HStack spacing={2}>
                    <Text color={textColor} fontWeight="medium">
                      {project.customer}
                    </Text>
                    <Text color={subtextColor}>•</Text>
                    <Text color={subtextColor}>{project.city}</Text>
                    <Badge colorScheme="purple" ml={2}>
                      {project.project_type}
                    </Badge>
                  </HStack>
                  <Text color={subtextColor} fontSize="sm">
                    {project.address}
                  </Text>
                </VStack>
                <HStack spacing={2}>
                  <Button
                    leftIcon={<FiArchive />}
                    colorScheme="yellow"
                    variant="outline"
                    onClick={handleArchive}
                  >
                    Archive Project
                  </Button>
                  <Button leftIcon={<FiEdit />} colorScheme="green" onClick={handleEdit}>
                    Edit Project
                  </Button>
                </HStack>
              </HStack>

              {/* Budget & Deadline */}
              <HStack spacing={8} pt={4}>
                <HStack spacing={2}>
                  <Icon as={FiDollarSign} color="green.500" boxSize={5} />
                  <VStack align="start" spacing={0}>
                    <Text fontSize="xs" color={subtextColor}>
                      Budget
                    </Text>
                    <Text fontSize="lg" fontWeight="bold" color={textColor}>
                      {formatCurrency(project.budget)}
                    </Text>
                  </VStack>
                </HStack>
                <HStack spacing={2}>
                  <Icon as={FiCalendar} color="green.500" boxSize={5} />
                  <VStack align="start" spacing={0}>
                    <Text fontSize="xs" color={subtextColor}>
                      Deadline
                    </Text>
                    <Text fontSize="lg" fontWeight="bold" color={textColor}>
                      {formatDate(project.deadline)}
                    </Text>
                  </VStack>
                </HStack>
              </HStack>
            </VStack>
          </CardBody>
        </Card>

        {/* Tabs */}
        <Card bg={bgColor} borderColor={borderColor} borderWidth="1px">
          <Tabs index={activeTab} onChange={setActiveTab} isLazy>
            <TabList
              borderBottomWidth="2px"
              borderColor={borderColor}
              overflowX="auto"
              css={{
                '&::-webkit-scrollbar': { display: 'none' },
                msOverflowStyle: 'none',
                scrollbarWidth: 'none',
              }}
            >
              <Tab
                _selected={{ color: 'white', bg: tabBgColor }}
                fontWeight="medium"
                minW="150px"
              >
                <Icon as={FiBarChart} mr={2} />
                Dashboard
              </Tab>
              <Tab
                _selected={{ color: 'white', bg: tabBgColor }}
                fontWeight="medium"
                minW="150px"
              >
                <Icon as={FiFileText} mr={2} />
                Daily Updates
              </Tab>
              <Tab
                _selected={{ color: 'white', bg: tabBgColor }}
                fontWeight="medium"
                minW="150px"
              >
                <Icon as={FiTarget} mr={2} />
                Milestones
              </Tab>
              <Tab
                _selected={{ color: 'white', bg: tabBgColor }}
                fontWeight="medium"
                minW="150px"
              >
                <Icon as={FiFileText} mr={2} />
                Weekly Reports
              </Tab>
              <Tab
                _selected={{ color: 'white', bg: tabBgColor }}
                fontWeight="medium"
                minW="180px"
              >
                <Icon as={FiClock} mr={2} />
                Timeline Schedule
              </Tab>
              <Tab
                _selected={{ color: 'white', bg: tabBgColor }}
                fontWeight="medium"
                minW="150px"
              >
                <Icon as={FiDatabase} mr={2} />
                Technical Data
              </Tab>
            </TabList>

            <TabPanels>
              {/* Dashboard Tab */}
              <TabPanel p={6}>
                <VStack align="stretch" spacing={6}>
                  {/* Progress Section */}
                  <Box>
                    <HStack justify="space-between" mb={4}>
                      <Heading size="md" color={textColor}>
                        Project Progress
                      </Heading>
                      <Button size="sm" colorScheme="green" onClick={handleUpdateProgress}>
                        Update Progress
                      </Button>
                    </HStack>

                    <Grid templateColumns="repeat(auto-fit, minmax(200px, 1fr))" gap={4}>
                      {/* Overall Progress */}
                      <Card borderWidth="1px" borderColor={borderColor}>
                        <CardBody>
                          <VStack align="stretch" spacing={3}>
                            <HStack justify="space-between">
                              <Icon as={FiBarChart} color="green.500" boxSize={6} />
                              <Text fontSize="2xl" fontWeight="bold" color={textColor}>
                                {project.overall_progress}%
                              </Text>
                            </HStack>
                            <Text fontSize="sm" fontWeight="medium" color={textColor}>
                              Overall Progress
                            </Text>
                            <Text fontSize="xs" color={subtextColor}>
                              Complete
                            </Text>
                          </VStack>
                        </CardBody>
                      </Card>

                      {/* Foundation */}
                      <Card borderWidth="1px" borderColor={borderColor}>
                        <CardBody>
                          <VStack align="stretch" spacing={3}>
                            <HStack justify="space-between">
                              <Icon as={FiDatabase} color="orange.500" boxSize={6} />
                              <Text fontSize="2xl" fontWeight="bold" color={textColor}>
                                {project.foundation_progress}%
                              </Text>
                            </HStack>
                            <Text fontSize="sm" fontWeight="medium" color={textColor}>
                              Foundation & Structure
                            </Text>
                            <Text fontSize="xs" color={subtextColor}>
                              Complete
                            </Text>
                          </VStack>
                        </CardBody>
                      </Card>

                      {/* Utilities */}
                      <Card borderWidth="1px" borderColor={borderColor}>
                        <CardBody>
                          <VStack align="stretch" spacing={3}>
                            <HStack justify="space-between">
                              <Icon as={FiTarget} color="purple.500" boxSize={6} />
                              <Text fontSize="2xl" fontWeight="bold" color={textColor}>
                                {project.utilities_progress}%
                              </Text>
                            </HStack>
                            <Text fontSize="sm" fontWeight="medium" color={textColor}>
                              Utilities Installation
                            </Text>
                            <Text fontSize="xs" color={subtextColor}>
                              Complete
                            </Text>
                          </VStack>
                        </CardBody>
                      </Card>

                      {/* Interior */}
                      <Card borderWidth="1px" borderColor={borderColor}>
                        <CardBody>
                          <VStack align="stretch" spacing={3}>
                            <HStack justify="space-between">
                              <Icon as={FiFileText} color="pink.500" boxSize={6} />
                              <Text fontSize="2xl" fontWeight="bold" color={textColor}>
                                {project.interior_progress}%
                              </Text>
                            </HStack>
                            <Text fontSize="sm" fontWeight="medium" color={textColor}>
                              Interior & Finishes
                            </Text>
                            <Text fontSize="xs" color={subtextColor}>
                              Complete
                            </Text>
                          </VStack>
                        </CardBody>
                      </Card>

                      {/* Equipment */}
                      <Card borderWidth="1px" borderColor={borderColor}>
                        <CardBody>
                          <VStack align="stretch" spacing={3}>
                            <HStack justify="space-between">
                              <Icon as={FiClock} color="green.500" boxSize={6} />
                              <Text fontSize="2xl" fontWeight="bold" color={textColor}>
                                {project.equipment_progress}%
                              </Text>
                            </HStack>
                            <Text fontSize="sm" fontWeight="medium" color={textColor}>
                              Kitchen Equipment
                            </Text>
                            <Text fontSize="xs" color={subtextColor}>
                              Complete
                            </Text>
                          </VStack>
                        </CardBody>
                      </Card>
                    </Grid>
                  </Box>

                  {/* Project Milestone */}
                  <Box>
                    <Heading size="md" color={textColor} mb={4}>
                      Project Milestone
                    </Heading>
                    <Center h="200px" borderWidth="1px" borderColor={borderColor} borderRadius="md">
                      <Text color={subtextColor}>No milestones yet</Text>
                    </Center>
                  </Box>

                  {/* Timeline Schedule */}
                  <Box>
                    <Heading size="md" color={textColor} mb={4}>
                      Timeline Schedule
                    </Heading>
                    <Center h="300px" borderWidth="1px" borderColor={borderColor} borderRadius="md">
                      <VStack spacing={4}>
                        <Icon as={FiClock} boxSize={12} color="gray.400" />
                        <Text color={subtextColor}>No Timeline Data</Text>
                        <Text color={subtextColor} fontSize="sm" textAlign="center">
                          Add work areas with start and end dates to see the project timeline
                        </Text>
                      </VStack>
                    </Center>
                  </Box>
                </VStack>
              </TabPanel>

              {/* Daily Updates Tab */}
              <TabPanel p={6}>
                <DailyUpdatesTab projectId={projectId} project={project || undefined} />
              </TabPanel>

              {/* Milestones Tab */}
              <TabPanel p={6}>
                <MilestonesTab projectId={Number(projectId)} />
              </TabPanel>

              {/* Weekly Reports Tab */}
              <TabPanel p={6}>
                <WeeklyReportsTab projectId={Number(projectId)} projectName={project?.project_name || ''} />
              </TabPanel>

              {/* Timeline Schedule Tab */}
              <TabPanel p={6}>
                <Center h="400px">
                  <VStack spacing={4}>
                    <Icon as={FiClock} boxSize={16} color="gray.400" />
                    <Text color={subtextColor} fontSize="lg">
                      Timeline Schedule
                    </Text>
                    <Text color={subtextColor} fontSize="sm">
                      Feature coming soon
                    </Text>
                  </VStack>
                </Center>
              </TabPanel>

              {/* Technical Data Tab */}
              <TabPanel p={6}>
                <Center h="400px">
                  <VStack spacing={4}>
                    <Icon as={FiDatabase} boxSize={16} color="gray.400" />
                    <Text color={subtextColor} fontSize="lg">
                      Technical Data
                    </Text>
                    <Text color={subtextColor} fontSize="sm">
                      Feature coming soon
                    </Text>
                  </VStack>
                </Center>
              </TabPanel>
            </TabPanels>
          </Tabs>
        </Card>

        {/* Update Progress Modal */}
        {project && (
          <UpdateProgressModal
            isOpen={isProgressModalOpen}
            onClose={handleProgressModalClose}
            project={project}
            onSuccess={handleProgressUpdateSuccess}
          />
        )}
      </Box>
    </Layout>
  );
}

