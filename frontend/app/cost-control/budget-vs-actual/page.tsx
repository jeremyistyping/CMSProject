'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useModulePermissions } from '@/hooks/usePermissions';
import {
  Box,
  Heading,
  Text,
  VStack,
  HStack,
  Spinner,
  Alert,
  AlertIcon,
  useColorModeValue,
  FormControl,
  FormLabel,
  Input,
  Select,
  Button,
  useToast,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react';
import { FiGrid, FiDollarSign, FiDownload } from 'react-icons/fi';
import projectService from '@/services/projectService';
import { expenseTransactionService, BudgetReportResponse } from '@/services/expenseTransactionService';
import BudgetReportViewer from '@/components/cost-control/BudgetReportViewer';

interface ProjectItem {
  id: number;
  project_name: string;
  city: string;
}

const BudgetVsActualPage: React.FC = () => {
  const { canView, loading } = useModulePermissions('cost_control');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const boxBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const toast = useToast();
  const router = useRouter();

  const [projects, setProjects] = useState<ProjectItem[]>([]);
  const [budgetReport, setBudgetReport] = useState<BudgetReportResponse | null>(null);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [selectedProjectId, setSelectedProjectId] = useState('');
  const [loadingReport, setLoadingReport] = useState(false);

  useEffect(() => {
    // default: current month
    const now = new Date();
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
    setStartDate(firstDay.toISOString().split('T')[0]);
    setEndDate(now.toISOString().split('T')[0]);

    // load projects for filter
    (async () => {
      try {
        const data = await projectService.getActiveProjects();
        const projectList = Array.isArray(data) ? data : [];
        const mappedProjects: ProjectItem[] = projectList.map((p: any) => ({
          id: typeof p.id === 'string' ? parseInt(p.id) : p.id,
          project_name: p.project_name,
          city: p.city
        }));
        setProjects(mappedProjects);
        if (mappedProjects.length > 0) {
          setSelectedProjectId(String(mappedProjects[0].id));
        }
      } catch (error) {
        console.error('Failed to load projects for cost control:', error);
      }
    })();
  }, []);

  const handleGenerate = async () => {
    if (!startDate || !endDate) {
      toast({
        title: 'Tanggal belum lengkap',
        description: 'Pilih start date dan end date terlebih dahulu.',
        status: 'warning',
        duration: 4000,
        isClosable: true,
      });
      return;
    }

    if (!selectedProjectId) {
      toast({
        title: 'Proyek belum dipilih',
        description: 'Pilih proyek terlebih dahulu.',
        status: 'warning',
        duration: 4000,
        isClosable: true,
      });
      return;
    }

    try {
      setLoadingReport(true);
      const report = await expenseTransactionService.getBudgetReport(
        parseInt(selectedProjectId),
        startDate,
        endDate
      );
      setBudgetReport(report);
    } catch (error: any) {
      console.error('Failed to load budget report:', error);
      toast({
        title: 'Gagal memuat data',
        description: error?.response?.data?.error || error?.message || 'Terjadi kesalahan saat mengambil data report.',
        status: 'error',
        duration: 6000,
        isClosable: true,
      });
    } finally {
      setLoadingReport(false);
    }
  };

  const handleExportPDF = async () => {
    if (!startDate || !endDate || !selectedProjectId) {
      toast({
        title: 'Data belum lengkap',
        description: 'Pilih proyek dan periode terlebih dahulu.',
        status: 'warning',
        duration: 4000,
        isClosable: true,
      });
      return;
    }

    try {
      const url = await expenseTransactionService.exportBudgetReportPDF(
        parseInt(selectedProjectId),
        startDate,
        endDate
      );
      
      // Open PDF in new tab
      window.open(url, '_blank');
      
      toast({
        title: 'Export berhasil',
        description: 'PDF report telah dibuka di tab baru.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      console.error('Failed to export PDF:', error);
      toast({
        title: 'Gagal export PDF',
        description: error?.response?.data?.error || error?.message || 'Terjadi kesalahan saat export PDF.',
        status: 'error',
        duration: 6000,
        isClosable: true,
      });
    }
  };



  if (loading) {
    return (
      <SimpleLayout>
        <Box display="flex" alignItems="center" justifyContent="center" minH="60vh">
          <HStack spacing={3}>
            <Spinner />
            <Text>Checking permissions...</Text>
          </HStack>
        </Box>
      </SimpleLayout>
    );
  }

  if (!canView) {
    return (
      <SimpleLayout>
        <Box maxW="xl">
          <Alert status="error" borderRadius="md">
            <AlertIcon />
            <Box>
              <Heading size="sm" mb={1}>Access Denied</Heading>
              <Text fontSize="sm">Anda tidak memiliki akses ke modul Cost Control. Silakan hubungi administrator.</Text>
            </Box>
          </Alert>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout>
      <Box>
        <VStack align="start" spacing={4} mb={6}>
          <HStack justify="space-between" w="full">
            <Box>
              <Heading size="lg" color={headingColor}>Budget vs Actual per Project</Heading>
              <Text fontSize="sm" color={textColor} maxW="3xl" mt={2}>
                Analisis perbandingan budget dan realisasi biaya per proyek, termasuk progress fisik dan indikator over/under budget.
              </Text>
            </Box>
            <HStack spacing={2}>
              <Button
                leftIcon={<FiGrid />}
                colorScheme="blue"
                variant="outline"
                onClick={() => router.push('/cost-control/cbs')}
                size="sm"
              >
                View CBS
              </Button>
              <Button
                leftIcon={<FiDollarSign />}
                colorScheme="green"
                variant="outline"
                onClick={() => router.push('/cost-control/expenses')}
                size="sm"
              >
                View Expenses
              </Button>
            </HStack>
          </HStack>
        </VStack>

        <Box
          bg={boxBg}
          borderWidth="1px"
          borderColor={borderColor}
          borderRadius="lg"
          p={6}
        >
          <VStack align="stretch" spacing={4}>
            {/* Filter bar */}
            <HStack spacing={4} flexWrap="wrap" align="flex-end">
              <FormControl maxW={{ base: '100%', md: '200px' }}>
                <FormLabel fontSize="sm">Start Date</FormLabel>
                <Input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  size="sm"
                />
              </FormControl>
              <FormControl maxW={{ base: '100%', md: '200px' }}>
                <FormLabel fontSize="sm">End Date</FormLabel>
                <Input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  size="sm"
                />
              </FormControl>
              <FormControl maxW={{ base: '100%', md: '260px' }}>
                <FormLabel fontSize="sm">Project (Optional)</FormLabel>
                <Select
                  placeholder="All Projects"
                  value={selectedProjectId}
                  onChange={(e) => setSelectedProjectId(e.target.value)}
                  size="sm"
                >
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.project_name} - {p.city}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl maxW={{ base: '100%', md: '160px' }}>
                <Button
                  colorScheme="green"
                  onClick={handleGenerate}
                  isLoading={loadingReport}
                  width="full"
                  size="sm"
                >
                  Generate
                </Button>
              </FormControl>
            </HStack>

            {/* Content */}
            {loadingReport ? (
              <HStack spacing={3} mt={4}>
                <Spinner size="sm" />
                <Text fontSize="sm" color={textColor}>Memuat data report...</Text>
              </HStack>
            ) : budgetReport ? (
              <VStack align="stretch" spacing={8} mt={4}>
                {/* Export Button */}
                <HStack justify="flex-end">
                  <Button
                    leftIcon={<FiDownload />}
                    colorScheme="blue"
                    onClick={handleExportPDF}
                    size="sm"
                  >
                    Export PDF
                  </Button>
                </HStack>
                
                <BudgetReportViewer report={budgetReport} />
              </VStack>
            ) : (
              <Box textAlign="center" py={10} mt={4}>
                <Text color={textColor}>
                  Pilih proyek dan periode, kemudian klik Generate untuk melihat report.
                </Text>
              </Box>
            )}
          </VStack>
        </Box>
      </Box>
    </SimpleLayout>
  );
};

export default BudgetVsActualPage;

