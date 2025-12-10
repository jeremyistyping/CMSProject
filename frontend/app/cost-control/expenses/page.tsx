'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useModulePermissions } from '@/hooks/usePermissions';
import {
    Box, Heading, Text, VStack, HStack, Spinner, Alert, AlertIcon,
    useColorModeValue, Select, FormControl, FormLabel, Button,
} from '@chakra-ui/react';
import { FiGrid, FiBarChart2 } from 'react-icons/fi';
import ExpenseTransactionList from '@/components/cost-control/ExpenseTransactionList';
import projectService from '@/services/projectService';

interface Project {
    id: number;
    project_name: string;
    city: string;
}

export default function ExpensesPage() {
    const { canView, canCreate, loading: permLoading } = useModulePermissions('cost_control');
    const [projects, setProjects] = useState<Project[]>([]);
    const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
    const [loading, setLoading] = useState(true);
    const router = useRouter();

    const headingColor = useColorModeValue('gray.800', 'gray.100');
    const textColor = useColorModeValue('gray.600', 'gray.300');
    const boxBg = useColorModeValue('white', 'gray.800');

    useEffect(() => {
        loadProjects();
    }, []);

    const loadProjects = async () => {
        try {
            const data = await projectService.getActiveProjects();
            const projectList = Array.isArray(data) ? data : [];
            setProjects(projectList);
            if (projectList.length > 0) {
                setSelectedProjectId(projectList[0].id);
            }
        } catch (error) {
            console.error('Failed to load projects:', error);
        } finally {
            setLoading(false);
        }
    };

    if (permLoading || loading) {
        return (
            <SimpleLayout>
                <Box display="flex" alignItems="center" justifyContent="center" minH="60vh">
                    <HStack spacing={3}>
                        <Spinner />
                        <Text>Loading...</Text>
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
                            <Text fontSize="sm">
                                Anda tidak memiliki akses ke modul Cost Control. Silakan hubungi administrator.
                            </Text>
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
                            <Heading size="lg" color={headingColor}>
                                Expense Transactions
                            </Heading>
                            <Text fontSize="sm" color={textColor} maxW="3xl" mt={2}>
                                Kelola transaksi biaya proyek termasuk biaya tenaga kerja, material, dan operasional.
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
                                leftIcon={<FiBarChart2 />}
                                colorScheme="purple"
                                variant="outline"
                                onClick={() => router.push('/cost-control/budget-vs-actual')}
                                size="sm"
                            >
                                Budget Report
                            </Button>
                        </HStack>
                    </HStack>
                </VStack>

                <Box bg={boxBg} borderWidth="1px" borderRadius="lg" p={6}>
                    <VStack align="stretch" spacing={6}>
                        {/* Project Selector */}
                        <FormControl maxW="400px">
                            <FormLabel>Pilih Proyek</FormLabel>
                            <Select
                                value={selectedProjectId || ''}
                                onChange={(e) => setSelectedProjectId(parseInt(e.target.value))}
                                placeholder="Pilih proyek"
                            >
                                {projects.map((project) => (
                                    <option key={project.id} value={project.id}>
                                        {project.project_name} - {project.city}
                                    </option>
                                ))}
                            </Select>
                        </FormControl>

                        {/* Expense List */}
                        {selectedProjectId ? (
                            <ExpenseTransactionList projectId={selectedProjectId} />
                        ) : (
                            <Box textAlign="center" py={10}>
                                <Text color="gray.500">Pilih proyek untuk melihat transaksi</Text>
                            </Box>
                        )}
                    </VStack>
                </Box>
            </Box>
        </SimpleLayout>
    );
}
