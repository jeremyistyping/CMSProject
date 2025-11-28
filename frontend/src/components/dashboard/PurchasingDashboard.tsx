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
    SimpleGrid,
    Spinner,
    Flex,
} from '@chakra-ui/react';
import { FiFileText, FiCheckSquare, FiArrowRight } from 'react-icons/fi';
import projectService from '@/services/projectService';
import { useRouter } from 'next/navigation';

export const PurchasingDashboard = () => {
    const [pendingCount, setPendingCount] = useState<number | null>(null);
    const [loading, setLoading] = useState(true);
    const router = useRouter();

    useEffect(() => {
        const fetchStats = async () => {
            try {
                setLoading(true);
                const data = await projectService.getPendingDailyUpdates();
                setPendingCount(data ? data.length : 0);
            } catch (error) {
                console.error('Error fetching stats:', error);
                setPendingCount(0);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    return (
        <Box>
            <Heading as="h2" size="xl" color="gray.800" mb={6}>
                Purchasing Dashboard
            </Heading>

            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={6}>
                {/* Daily Report Approval Card */}
                <Card>
                    <CardHeader pb={0}>
                        <HStack justify="space-between">
                            <Heading size="md" display="flex" alignItems="center">
                                <Icon as={FiFileText} mr={2} color="blue.500" />
                                Daily Reports
                            </Heading>
                            <Icon as={FiCheckSquare} color="gray.400" boxSize={6} />
                        </HStack>
                    </CardHeader>
                    <CardBody>
                        <VStack align="start" spacing={4}>
                            <Box>
                                <Text fontSize="sm" color="gray.500">Pending Approvals</Text>
                                {loading ? (
                                    <Spinner size="sm" color="blue.500" mt={1} />
                                ) : (
                                    <Text fontSize="3xl" fontWeight="bold" color={pendingCount && pendingCount > 0 ? "orange.500" : "green.500"}>
                                        {pendingCount}
                                    </Text>
                                )}
                            </Box>

                            <Button
                                rightIcon={<FiArrowRight />}
                                colorScheme="blue"
                                variant="outline"
                                size="sm"
                                width="full"
                                onClick={() => router.push('/daily-report-approval')}
                            >
                                Go to Approvals
                            </Button>
                        </VStack>
                    </CardBody>
                </Card>

                {/* Placeholder for future widgets */}
                <Card>
                    <CardHeader pb={0}>
                        <Heading size="md" color="gray.400">
                            Future Widget
                        </Heading>
                    </CardHeader>
                    <CardBody>
                        <Flex height="100px" align="center" justify="center" color="gray.300">
                            <Text>More stats coming soon...</Text>
                        </Flex>
                    </CardBody>
                </Card>
            </SimpleGrid>
        </Box>
    );
};
