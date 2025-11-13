'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardBody,
  VStack,
  HStack,
  Heading,
  Text,
  Grid,
  GridItem,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  useToast,
  Badge,
  Icon,
  IconButton,
  useColorModeValue,
  Spinner,
  Center,
} from '@chakra-ui/react';
import { FiFileText, FiDownload, FiTrash2, FiBarChart } from 'react-icons/fi';
import weeklyReportService, {
  WeeklyReportDTO,
  CreateWeeklyReportRequest,
} from '@/services/weeklyReportService';

interface WeeklyReportsTabProps {
  projectId: number;
  projectName: string;
}

export default function WeeklyReportsTab({ projectId, projectName }: WeeklyReportsTabProps) {
  const [reports, setReports] = useState<WeeklyReportDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const toast = useToast();

  // Form state
  const [formData, setFormData] = useState({
    week: weeklyReportService.getCurrentWeek(),
    year: new Date().getFullYear(),
    project_manager: '',
    total_work_days: 5,
    weather_delays: 0,
    team_size: 1,
    accomplishments: '',
    challenges: '',
    next_week_priorities: '',
  });

  // Get current week in ISO format (YYYY-Www)
  const getCurrentWeekString = () => {
    const now = new Date();
    const year = now.getFullYear();
    const week = weeklyReportService.getCurrentWeek();
    return `${year}-W${week.toString().padStart(2, '0')}`;
  };

  const [selectedWeek, setSelectedWeek] = useState(getCurrentWeekString());

  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const textColor = useColorModeValue('gray.800', 'white');
  const subtextColor = useColorModeValue('gray.500', 'gray.400');

  useEffect(() => {
    loadReports();
  }, [projectId]);

  const loadReports = async () => {
    setLoading(true);
    try {
      const data = await weeklyReportService.getWeeklyReports(projectId);
      setReports(data);
    } catch (error: any) {
      console.error('Failed to load reports:', error);
      toast({
        title: 'Error',
        description: 'Failed to load weekly reports',
        status: 'error',
        duration: 3000,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleWeekChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const weekString = e.target.value; // Format: YYYY-Www (e.g., 2025-W46)
    setSelectedWeek(weekString);

    if (weekString) {
      // Parse the week string
      const [yearStr, weekStr] = weekString.split('-W');
      const year = parseInt(yearStr);
      const week = parseInt(weekStr);

      setFormData((prev) => ({
        ...prev,
        week,
        year,
      }));
    }
  };

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]:
        name === 'total_work_days' ||
        name === 'weather_delays' ||
        name === 'team_size'
          ? parseInt(value) || 0
          : value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      const requestData: CreateWeeklyReportRequest = {
        ...formData,
        project_id: projectId,
      };

      await weeklyReportService.createWeeklyReport(projectId, requestData);

      toast({
        title: 'Success',
        description: 'Weekly report created successfully',
        status: 'success',
        duration: 3000,
      });

      // Reset form
      setFormData({
        ...formData,
        accomplishments: '',
        challenges: '',
        next_week_priorities: '',
      });

      // Reload reports
      loadReports();
    } catch (error: any) {
      console.error('Failed to create report:', error);
      toast({
        title: 'Error',
        description: error.response?.data?.details || 'Failed to create weekly report',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDownloadPDF = (reportId: number) => {
    const url = weeklyReportService.getPDFUrl(projectId, reportId);
    window.open(url, '_blank');
  };

  const handleDelete = async (reportId: number) => {
    if (!confirm('Are you sure you want to delete this report?')) return;

    try {
      await weeklyReportService.deleteWeeklyReport(projectId, reportId);
      toast({
        title: 'Success',
        description: 'Report deleted successfully',
        status: 'success',
        duration: 3000,
      });
      loadReports();
    } catch (error) {
      console.error('Failed to delete report:', error);
      toast({
        title: 'Error',
        description: 'Failed to delete report',
        status: 'error',
        duration: 3000,
      });
    }
  };

  return (
    <VStack align="stretch" spacing={6}>
      {/* Form Section */}
      <Card bg={bgColor} borderWidth="1px" borderColor={borderColor}>
        <CardBody>
          <Heading size="md" mb={4} color={textColor}>
            Generate Weekly Report
          </Heading>

          <form onSubmit={handleSubmit}>
            <VStack spacing={4} align="stretch">
              <Grid templateColumns="repeat(auto-fit, minmax(200px, 1fr))" gap={4}>
                <FormControl isRequired>
                  <FormLabel>Report Week</FormLabel>
                  <Input
                    type="week"
                    value={selectedWeek}
                    onChange={handleWeekChange}
                    size="md"
                  />
                </FormControl>

                <FormControl>
                  <FormLabel>Project Manager</FormLabel>
                  <Input
                    name="project_manager"
                    value={formData.project_manager}
                    onChange={handleInputChange}
                    placeholder="Manager name"
                  />
                </FormControl>
              </Grid>

              <Grid templateColumns="repeat(auto-fit, minmax(200px, 1fr))" gap={4}>
                <FormControl isRequired>
                  <FormLabel>Total Work Days</FormLabel>
                  <Input
                    type="number"
                    name="total_work_days"
                    value={formData.total_work_days}
                    onChange={handleInputChange}
                    min={0}
                  />
                </FormControl>

                <FormControl>
                  <FormLabel>Weather Delays (days)</FormLabel>
                  <Input
                    type="number"
                    name="weather_delays"
                    value={formData.weather_delays}
                    onChange={handleInputChange}
                    min={0}
                  />
                </FormControl>

                <FormControl isRequired>
                  <FormLabel>Team Size</FormLabel>
                  <Input
                    type="number"
                    name="team_size"
                    value={formData.team_size}
                    onChange={handleInputChange}
                    min={1}
                  />
                </FormControl>
              </Grid>

              <FormControl>
                <FormLabel>Major Accomplishments</FormLabel>
                <Textarea
                  name="accomplishments"
                  value={formData.accomplishments}
                  onChange={handleInputChange}
                  rows={4}
                  placeholder="List major accomplishments this week..."
                />
              </FormControl>

              <FormControl>
                <FormLabel>Challenges & Issues</FormLabel>
                <Textarea
                  name="challenges"
                  value={formData.challenges}
                  onChange={handleInputChange}
                  rows={4}
                  placeholder="Describe any challenges encountered..."
                />
              </FormControl>

              <FormControl>
                <FormLabel>Next Week's Priorities</FormLabel>
                <Textarea
                  name="next_week_priorities"
                  value={formData.next_week_priorities}
                  onChange={handleInputChange}
                  rows={4}
                  placeholder="List next week's priorities..."
                />
              </FormControl>

              <Button type="submit" colorScheme="green" isLoading={submitting} size="lg">
                Generate Report
              </Button>
            </VStack>
          </form>
        </CardBody>
      </Card>

      {/* Previous Reports Section */}
      <Card bg={bgColor} borderWidth="1px" borderColor={borderColor}>
        <CardBody>
          <HStack justify="space-between" mb={4}>
            <Heading size="md" color={textColor}>
              Previous Reports
            </Heading>
            <Button
              size="sm"
              colorScheme="blue"
              leftIcon={<FiDownload />}
              onClick={() => toast({ title: 'Export all feature - coming soon', status: 'info' })}
            >
              Export All PDF
            </Button>
          </HStack>

          {loading ? (
            <Center py={10}>
              <Spinner size="xl" color="green.500" />
            </Center>
          ) : reports.length === 0 ? (
            <Center py={10}>
              <VStack spacing={4}>
                <Icon as={FiBarChart} boxSize={16} color="gray.400" />
                <Text color={subtextColor} fontSize="lg">
                  No weekly reports yet
                </Text>
                <Text color={subtextColor} fontSize="sm">
                  Generate your first weekly report to track progress
                </Text>
              </VStack>
            </Center>
          ) : (
            <VStack spacing={4} align="stretch">
              {reports.map((report) => (
                <Card
                  key={report.id}
                  borderWidth="1px"
                  borderColor={borderColor}
                  bg={bgColor}
                  transition="all 0.2s"
                  _hover={{ shadow: 'md' }}
                >
                  <CardBody>
                    <HStack justify="space-between" mb={3}>
                      <VStack align="start" spacing={1}>
                        <Heading size="sm" color={textColor}>
                          {report.week_label}
                        </Heading>
                        <Text fontSize="xs" color={subtextColor}>
                          Generated: {new Date(report.generated_date).toLocaleDateString()}
                        </Text>
                      </VStack>
                      <HStack spacing={2}>
                        <IconButton
                          aria-label="Download PDF"
                          icon={<FiDownload />}
                          colorScheme="blue"
                          size="sm"
                          onClick={() => handleDownloadPDF(report.id)}
                        />
                        <IconButton
                          aria-label="Delete"
                          icon={<FiTrash2 />}
                          colorScheme="red"
                          size="sm"
                          onClick={() => handleDelete(report.id)}
                        />
                      </HStack>
                    </HStack>

                    <Grid templateColumns="repeat(auto-fit, minmax(150px, 1fr))" gap={3} mb={3}>
                      <Box>
                        <Text fontSize="xs" color={subtextColor}>
                          Work Days
                        </Text>
                        <Text fontSize="lg" fontWeight="bold" color={textColor}>
                          {report.total_work_days}
                        </Text>
                      </Box>
                      <Box>
                        <Text fontSize="xs" color={subtextColor}>
                          Weather Delays
                        </Text>
                        <Text fontSize="lg" fontWeight="bold" color={textColor}>
                          {report.weather_delays}
                        </Text>
                      </Box>
                      <Box>
                        <Text fontSize="xs" color={subtextColor}>
                          Team Size
                        </Text>
                        <Text fontSize="lg" fontWeight="bold" color={textColor}>
                          {report.team_size}
                        </Text>
                      </Box>
                      <Box>
                        <Text fontSize="xs" color={subtextColor}>
                          Manager
                        </Text>
                        <Text fontSize="lg" fontWeight="bold" color={textColor}>
                          {report.project_manager || 'N/A'}
                        </Text>
                      </Box>
                    </Grid>

                    {report.accomplishments && (
                      <Box borderTopWidth="1px" borderColor={borderColor} pt={3}>
                        <Text fontSize="sm" fontWeight="bold" mb={2} color={textColor}>
                          Accomplishments:
                        </Text>
                        <Text fontSize="sm" color={subtextColor} whiteSpace="pre-wrap">
                          {report.accomplishments}
                        </Text>
                      </Box>
                    )}
                  </CardBody>
                </Card>
              ))}
            </VStack>
          )}
        </CardBody>
      </Card>
    </VStack>
  );
}

