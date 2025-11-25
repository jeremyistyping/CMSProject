import React, { useEffect, useState } from 'react';
import {
    Box,
    Heading,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Badge,
    Alert,
    AlertIcon,
    Spinner,
    Text,
    VStack,
    Card,
    CardHeader,
    CardBody,
    useColorModeValue,
} from '@chakra-ui/react';
import purchaseRequestService from '@/services/purchaseRequestService';
import { MaterialImpact } from '@/types/purchaseRequest';

interface MaterialImpactCardProps {
    purchaseRequestId: number;
}

const MaterialImpactCard: React.FC<MaterialImpactCardProps> = ({ purchaseRequestId }) => {
    const [impacts, setImpacts] = useState<MaterialImpact[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const cardBg = useColorModeValue('white', 'gray.700');
    const borderColor = useColorModeValue('gray.200', 'gray.600');

    useEffect(() => {
        loadMaterialImpact();
    }, [purchaseRequestId]);

    const loadMaterialImpact = async () => {
        setIsLoading(true);
        setError(null);
        try {
            const data = await purchaseRequestService.getMaterialImpact(purchaseRequestId);
            setImpacts(data || []);
        } catch (err: any) {
            console.error('Failed to load material impact:', err);
            setError(err.response?.data?.error || 'Failed to load material impact data');
        } finally {
            setIsLoading(false);
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'CRITICAL':
                return 'red';
            case 'LOW':
                return 'orange';
            case 'OK':
            default:
                return 'green';
        }
    };

    if (isLoading) {
        return (
            <Card bg={cardBg} variant="outline">
                <CardBody>
                    <VStack py={4}>
                        <Spinner size="md" />
                        <Text fontSize="sm" color="gray.500">Loading material impact...</Text>
                    </VStack>
                </CardBody>
            </Card>
        );
    }

    if (error) {
        return (
            <Card bg={cardBg} variant="outline">
                <CardBody>
                    <Alert status="error" borderRadius="md">
                        <AlertIcon />
                        {error}
                    </Alert>
                </CardBody>
            </Card>
        );
    }

    if (impacts.length === 0) {
        return (
            <Card bg={cardBg} variant="outline">
                <CardBody>
                    <Alert status="info" borderRadius="md">
                        <AlertIcon />
                        No material impact data available for this purchase request.
                    </Alert>
                </CardBody>
            </Card>
        );
    }

    return (
        <Card bg={cardBg} variant="outline">
            <CardHeader pb={2}>
                <Heading size="sm">Estimated Material Impact</Heading>
                <Text fontSize="xs" color="gray.500" mt={1}>
                    How this purchase will affect project material stock
                </Text>
            </CardHeader>
            <CardBody pt={2}>
                <Box overflowX="auto">
                    <Table variant="simple" size="sm">
                        <Thead>
                            <Tr>
                                <Th>Material</Th>
                                <Th>Code</Th>
                                <Th isNumeric>Requested</Th>
                                <Th isNumeric>Current Stock</Th>
                                <Th isNumeric>Projected Stock</Th>
                                <Th>Status</Th>
                            </Tr>
                        </Thead>
                        <Tbody>
                            {impacts.map((impact, idx) => (
                                <Tr key={idx}>
                                    <Td>
                                        <Text fontWeight="medium">{impact.product_name}</Text>
                                    </Td>
                                    <Td>
                                        <Text fontSize="xs" color="gray.500">{impact.product_code}</Text>
                                    </Td>
                                    <Td isNumeric>
                                        {impact.requested_qty} {impact.unit}
                                    </Td>
                                    <Td isNumeric>
                                        {impact.current_stock} {impact.unit}
                                    </Td>
                                    <Td isNumeric fontWeight="bold">
                                        {impact.projected_stock} {impact.unit}
                                    </Td>
                                    <Td>
                                        <Badge colorScheme={getStatusColor(impact.status)}>
                                            {impact.status}
                                        </Badge>
                                    </Td>
                                </Tr>
                            ))}
                        </Tbody>
                    </Table>
                </Box>

                {impacts.some(i => i.status === 'CRITICAL' || i.status === 'LOW') && (
                    <Alert status="warning" mt={4} borderRadius="md">
                        <AlertIcon />
                        <Box>
                            <Text fontWeight="bold" fontSize="sm">Stock Alert</Text>
                            <Text fontSize="xs">
                                Some materials have low or critical stock levels. This purchase will help replenish inventory.
                            </Text>
                        </Box>
                    </Alert>
                )}
            </CardBody>
        </Card>
    );
};

export default MaterialImpactCard;
