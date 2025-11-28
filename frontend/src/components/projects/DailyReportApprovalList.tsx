'use client';

import React, { useEffect, useState } from 'react';
import {
    Box,
    Heading,
    Text,
    Card,
    CardHeader,
    CardBody,
    Button,
    VStack,
    HStack,
    Icon,
    Badge,
    Spinner,
    useDisclosure,
    useToast,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Flex,
} from '@chakra-ui/react';
import { FiCheckCircle, FiFileText, FiRefreshCw } from 'react-icons/fi';
import projectService from '@/services/projectService';
import { DailyUpdate } from '@/types/project';
import DailyUpdateViewModal from '@/components/projects/DailyUpdateViewModal';

export const DailyReportApprovalList = () => {
    const [updates, setUpdates] = useState<DailyUpdate[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedUpdate, setSelectedUpdate] = useState<DailyUpdate | null>(null);
    const { isOpen, onOpen, onClose } = useDisclosure();
    const toast = useToast();

    const fetchUpdates = async () => {
        try {
            setLoading(true);
            const data = await projectService.getPendingDailyUpdates();
            setUpdates(data || []);
        } catch (error) {
            console.error('Error fetching pending updates:', error);
            toast({
                title: 'Error',
                description: 'Failed to fetch pending daily updates',
                status: 'error',
                duration: 3000,
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUpdates();
    }, []);

    const handleViewUpdate = (update: DailyUpdate) => {
        setSelectedUpdate(update);
        onOpen();
    };

    const handleUpdateProcessed = () => {
        fetchUpdates();
    };

    return (
        <Box>
            <HStack justify="space-between" mb={6}>
                <Heading as="h2" size="lg" color="gray.800">
                    Daily Report Approvals
                </Heading>
                <Button
                    leftIcon={<FiRefreshCw />}
                    onClick={fetchUpdates}
                    isLoading={loading}
                    variant="outline"
                    size="sm"
                >
                    Refresh
                </Button>
            </HStack>

            <Card>
                <CardBody>
                    {loading ? (
                        <Flex justify="center" align="center" minH="200px">
                            <VStack spacing={4}>
                                <Spinner size="xl" color="blue.500" thickness="4px" />
                                <Text>Loading pending reports...</Text>
                            </VStack>
                        </Flex>
                    ) : updates.length === 0 ? (
                        <Flex justify="center" align="center" minH="200px" direction="column">
                            <Icon as={FiCheckCircle} boxSize={12} color="green.500" mb={4} />
                            <Text fontSize="lg" fontWeight="medium">All caught up!</Text>
                            <Text color="gray.500">No pending daily reports to approve.</Text>
                        </Flex>
                    ) : (
                        <Box overflowX="auto">
                            <Table variant="simple">
                                <Thead>
                                    <Tr>
                                        <Th>Project</Th>
                                        <Th>Date</Th>
                                        <Th>Submitted By</Th>
                                        <Th>Status</Th>
                                        <Th>Action</Th>
                                    </Tr>
                                </Thead>
                                <Tbody>
                                    {updates.map((update) => (
                                        <Tr key={update.id}>
                                            <Td fontWeight="medium">{update.project?.project_name || 'Unknown Project'}</Td>
                                            <Td>{new Date(update.date).toLocaleDateString('id-ID', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</Td>
                                            <Td>
                                                {update.created_by}
                                                <Badge ml={2} colorScheme="green" fontSize="xs">Employee</Badge>
                                            </Td>
                                            <Td>
                                                <Badge colorScheme="yellow">Pending</Badge>
                                            </Td>
                                            <Td>
                                                <Button
                                                    size="sm"
                                                    colorScheme="blue"
                                                    onClick={() => handleViewUpdate(update)}
                                                >
                                                    Review
                                                </Button>
                                            </Td>
                                        </Tr>
                                    ))}
                                </Tbody>
                            </Table>
                        </Box>
                    )}
                </CardBody>
            </Card>

            {selectedUpdate && (
                <DailyUpdateViewModal
                    isOpen={isOpen}
                    onClose={onClose}
                    dailyUpdate={selectedUpdate}
                    onStatusChange={handleUpdateProcessed}
                />
            )}
        </Box>
    );
};
