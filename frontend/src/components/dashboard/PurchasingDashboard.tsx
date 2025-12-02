'use client';

import React from 'react';
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
    Flex,
} from '@chakra-ui/react';
import { FiFileText, FiCheckSquare, FiArrowRight } from 'react-icons/fi';
import { useRouter } from 'next/navigation';

export const PurchasingDashboard = () => {
    const router = useRouter();

    return (
        <Box>
            <Heading as="h2" size="xl" color="gray.800" mb={6}>
                Purchasing Dashboard
            </Heading>

            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={6}>
                {/* Purchase Request Management Card */}
                <Card>
                    <CardHeader pb={0}>
                        <HStack justify="space-between">
                            <Heading size="md" display="flex" alignItems="center">
                                <Icon as={FiCheckSquare} mr={2} color="blue.500" />
                                Purchase Requests
                            </Heading>
                            <Icon as={FiFileText} color="gray.400" boxSize={6} />
                        </HStack>
                    </CardHeader>
                    <CardBody>
                        <VStack align="start" spacing={4}>
                            <Box>
                                <Text fontSize="sm" color="gray.500">Manage Requests</Text>
                                <Text fontSize="md" color="gray.600" mt={1}>
                                    View, edit, and manage purchase requests.
                                </Text>
                            </Box>

                            <Button
                                rightIcon={<FiArrowRight />}
                                colorScheme="blue"
                                variant="outline"
                                size="sm"
                                width="full"
                                onClick={() => router.push('/cost-control/purchase-requests')}
                            >
                                Go to PR Management
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
