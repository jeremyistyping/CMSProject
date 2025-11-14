'use client';

import React, { useEffect, useState } from 'react';
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
  Badge,
  FormControl,
  FormLabel,
  Input,
  Select,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  useToast,
} from '@chakra-ui/react';
import { BudgetVsActualReport, PortfolioBudgetVsActualReport, Project, ProjectBudgetVsActualSummary, ProgressVsCostReport, ProgressVsCostPoint } from '@/types/project';
import { reportService } from '@/services/reportService';
import projectService from '@/services/projectService';

const BudgetVsActualPage: React.FC = () => {
  const { canView, loading } = useModulePermissions('cost_control');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const boxBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const toast = useToast();

  const [projects, setProjects] = useState<Project[]>([]);
  const [report, setReport] = useState<PortfolioBudgetVsActualReport | null>(null);
  const [progressVsCost, setProgressVsCost] = useState<ProgressVsCostReport | null>(null);
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
        setProjects(Array.isArray(data) ? data : []);
      } catch (error) {
        console.error('Failed to load projects for cost control:', error);
      }
    })();
  }, []);

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount || 0);

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

    const buildPortfolioFromPerProjectReports = async (): Promise<PortfolioBudgetVsActualReport> => {
      // Tentukan proyek target (semua atau satu project)
      const targetProjects = selectedProjectId
        ? projects.filter((p) => String(p.id) === selectedProjectId)
        : projects;

      if (targetProjects.length === 0) {
        return {
          report_date: new Date().toISOString(),
          start_date: new Date(startDate).toISOString(),
          end_date: new Date(endDate).toISOString(),
          projects: [],
        };
      }

      // Fallback sederhana: gunakan field budget / actual_cost di Project
      // Catatan: fallback ini tidak menghormati filter tanggal; ini adalah lifetime view per project.
      const summaries: ProjectBudgetVsActualSummary[] = targetProjects.map((project) => {
        const budget = project.budget || 0;
        const actual = project.actual_cost || 0;
        const variance = budget - actual;
        const costProgress = budget > 0 ? (actual / budget) * 100 : 0;
        const physicalProgress = project.overall_progress || 0;
        const progressGap = costProgress - physicalProgress;

        let status: string;
        if (budget === 0 && actual === 0) {
          status = 'NO_BUDGET';
        } else if (budget > 0 && actual > budget * 1.05) {
          status = 'OVER_BUDGET';
        } else if (costProgress + 10 < physicalProgress) {
          status = 'UNDER_UTILIZED';
        } else {
          status = 'ON_TRACK';
        }

        return {
          project_id: Number(project.id),
          project_name: project.project_name,
          budget,
          actual,
          variance,
          variance_percent: budget > 0 ? (variance / budget) * 100 : 0,
          cost_progress: costProgress,
          physical_progress: physicalProgress,
          progress_gap: progressGap,
          status,
        };
      });

      return {
        report_date: new Date().toISOString(),
        start_date: new Date(startDate).toISOString(),
        end_date: new Date(endDate).toISOString(),
        projects: summaries,
      };
    };

    try {
      setLoadingReport(true);
      let data: PortfolioBudgetVsActualReport;

      try {
        // Coba endpoint portfolio bawaan backend dulu
        data = (await reportService.getPortfolioBudgetVsActual({
          start_date: startDate,
          end_date: endDate,
          project_id: selectedProjectId || undefined,
          format: 'json',
        })) as PortfolioBudgetVsActualReport;
      } catch (error: any) {
        // Jika endpoint belum tersedia (404), fallback ke agregasi per-project di frontend
        if (error instanceof Error && error.message.includes('404')) {
          console.warn('Portfolio endpoint 404, falling back to per-project aggregation');
          data = await buildPortfolioFromPerProjectReports();
        } else {
          throw error;
        }
      }

      setReport(data);

      // Jika hanya satu project dipilih, sekalian ambil Progress vs Cost time-series
      if (selectedProjectId) {
        try {
          const pvc = (await reportService.getProjectProgressVsCost({
            start_date: startDate,
            end_date: endDate,
            project_id: selectedProjectId,
            format: 'json',
          })) as ProgressVsCostReport;
          setProgressVsCost(pvc);
        } catch (err) {
          console.warn('Failed to load Progress vs Cost report:', err);
          setProgressVsCost(null);
        }
      } else {
        setProgressVsCost(null);
      }
    } catch (error: any) {
      console.error('Failed to load portfolio budget vs actual:', error);
      toast({
        title: 'Gagal memuat data',
        description: error?.message || 'Terjadi kesalahan saat mengambil data portfolio.',
        status: 'error',
        duration: 6000,
        isClosable: true,
      });
    } finally {
      setLoadingReport(false);
    }
  };

  const renderTable = () => {
    if (!report || !report.projects || report.projects.length === 0) {
      return (
        <Text fontSize="sm" color={textColor}>
          Belum ada data untuk rentang tanggal yang dipilih.
        </Text>
      );
    }

    const projectsData: ProjectBudgetVsActualSummary[] = report.projects;
    const totalBudget = projectsData.reduce((sum, p) => sum + (p.budget || 0), 0);
    const totalActual = projectsData.reduce((sum, p) => sum + (p.actual || 0), 0);

    return (
      <VStack align="stretch" spacing={4}>
        <HStack spacing={4} flexWrap="wrap">
          <Stat>
            <StatLabel>Total Projects</StatLabel>
            <StatNumber>{projectsData.length}</StatNumber>
          </Stat>
          <Stat>
            <StatLabel>Total Budget</StatLabel>
            <StatNumber>{formatCurrency(totalBudget)}</StatNumber>
          </Stat>
          <Stat>
            <StatLabel>Total Actual</StatLabel>
            <StatNumber>{formatCurrency(totalActual)}</StatNumber>
          </Stat>
          <Stat>
            <StatLabel>Portfolio Variance</StatLabel>
            <StatNumber color={totalBudget - totalActual < 0 ? 'red.500' : 'green.500'}>
              {formatCurrency(totalBudget - totalActual)}
            </StatNumber>
            <StatHelpText>Budget - Actual (semua proyek)</StatHelpText>
          </Stat>
        </HStack>

        <TableContainer>
          <Table size="sm">
            <Thead>
              <Tr>
                <Th>Project</Th>
                <Th isNumeric>Budget</Th>
                <Th isNumeric>Actual</Th>
                <Th isNumeric>Variance</Th>
                <Th isNumeric>Cost Progress</Th>
                <Th isNumeric>Physical Progress</Th>
                <Th isNumeric>Progress Gap</Th>
                <Th>Status</Th>
              </Tr>
            </Thead>
            <Tbody>
              {projectsData.map((p) => {
                const costProgress = p.cost_progress ?? 0;
                const physicalProgress = p.physical_progress ?? 0;
                const progressGap = p.progress_gap ?? 0;
                const isProblem = p.status === 'OVER_BUDGET' || p.status === 'UNDER_UTILIZED';

                let statusColorScheme: string = 'green';
                if (p.status === 'OVER_BUDGET') statusColorScheme = 'red';
                else if (p.status === 'UNDER_UTILIZED') statusColorScheme = 'orange';
                else if (p.status === 'NO_BUDGET') statusColorScheme = 'gray';

                return (
                  <Tr key={p.project_id}>
                    <Td>{p.project_name}</Td>
                    <Td isNumeric>{formatCurrency(p.budget)}</Td>
                    <Td isNumeric>{formatCurrency(p.actual)}</Td>
                    <Td isNumeric color={p.variance < 0 ? 'red.500' : 'green.500'}>
                      {formatCurrency(p.variance)}
                    </Td>
                    <Td isNumeric>{costProgress.toFixed(1)}%</Td>
                    <Td isNumeric>{physicalProgress.toFixed(1)}%</Td>
                    <Td isNumeric color={isProblem ? 'red.500' : 'gray.800'}>
                      {progressGap.toFixed(1)}%
                    </Td>
                    <Td>
                      <Badge colorScheme={statusColorScheme}>{p.status}</Badge>
                    </Td>
                  </Tr>
                );
              })}
            </Tbody>
          </Table>
        </TableContainer>
      </VStack>
    );
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
          <Heading size="lg" color={headingColor}>Budget vs Actual per Project</Heading>
          <Text fontSize="sm" color={textColor} maxW="3xl">
            Analisis perbandingan budget dan realisasi biaya per proyek, termasuk progress fisik dan indikator over/under budget.
          </Text>
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
                <Text fontSize="sm" color={textColor}>Memuat data portfolio...</Text>
              </HStack>
            ) : (
              <VStack align="stretch" spacing={8} mt={4}>
                {renderTable()}

                {/* Progress vs Cost chart placeholder - only when a single project is selected */}
                {progressVsCost && progressVsCost.points && progressVsCost.points.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={2}>
                      Progress vs Cost - {progressVsCost.project_name}
                    </Heading>
                    <Text fontSize="xs" color={textColor} mb={3}>
                      Grafik korelasi progress fisik vs biaya kumulatif (WIP: visualisasi sederhana, bisa diupgrade ke chart library).
                    </Text>
                    <TableContainer>
                      <Table size="xs">
                        <Thead>
                          <Tr>
                            <Th>Tanggal</Th>
                            <Th isNumeric>Progress Fisik</Th>
                            <Th isNumeric>Cost Progress</Th>
                            <Th isNumeric>Gap</Th>
                            <Th isNumeric>Kumulatif Actual</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {progressVsCost.points.map((pt: ProgressVsCostPoint) => (
                            <Tr key={pt.date}>
                              <Td>{new Date(pt.date).toLocaleDateString('id-ID')}</Td>
                              <Td isNumeric>{pt.physical_progress.toFixed(1)}%</Td>
                              <Td isNumeric>{pt.cost_progress.toFixed(1)}%</Td>
                              <Td isNumeric color={Math.abs(pt.progress_gap) > 10 ? 'red.500' : 'gray.700'}>
                                {pt.progress_gap.toFixed(1)}%
                              </Td>
                              <Td isNumeric>{formatCurrency(pt.cumulative_actual)}</Td>
                            </Tr>
                          ))}
                        </Tbody>
                      </Table>
                    </TableContainer>
                  </Box>
                )}
              </VStack>
            )}
          </VStack>
        </Box>
      </Box>
    </SimpleLayout>
  );
};

export default BudgetVsActualPage;

