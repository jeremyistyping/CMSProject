import React from 'react';
import {
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Badge,
    Text,
    Box,
    useColorModeValue,
} from '@chakra-ui/react';
import { MaterialMovement } from '@/services/materialTrackingService';
import { format } from 'date-fns';

interface MaterialMovementTableProps {
    movements: MaterialMovement[];
}

const MaterialMovementTable: React.FC<MaterialMovementTableProps> = ({ movements }) => {
    const borderColor = useColorModeValue('gray.200', 'gray.700');
    const hoverBg = useColorModeValue('gray.50', 'gray.700');

    if (movements.length === 0) {
        return (
            <Box p={4} textAlign="center">
                <Text color="gray.500">No movements recorded yet.</Text>
            </Box>
        );
    }

    return (
        <Box overflowX="auto">
            <Table variant="simple" size="sm">
                <Thead>
                    <Tr>
                        <Th>Date</Th>
                        <Th>Type</Th>
                        <Th>Product</Th>
                        <Th isNumeric>Quantity</Th>
                        <Th>Reference</Th>
                        <Th>Notes</Th>
                    </Tr>
                </Thead>
                <Tbody>
                    {movements.map((movement) => (
                        <Tr key={movement.id} _hover={{ bg: hoverBg }}>
                            <Td>{format(new Date(movement.transaction_date), 'dd MMM yyyy')}</Td>
                            <Td>
                                <Badge
                                    colorScheme={movement.type === 'IN' ? 'green' : 'red'}
                                    variant="subtle"
                                >
                                    {movement.type}
                                </Badge>
                            </Td>
                            <Td>
                                <Text fontWeight="medium">{movement.product.name}</Text>
                                <Text fontSize="xs" color="gray.500">{movement.product.code}</Text>
                            </Td>
                            <Td isNumeric>
                                {movement.quantity} {movement.product.unit}
                            </Td>
                            <Td>
                                <Badge variant="outline" fontSize="xs">
                                    {movement.reference_type} #{movement.reference_id}
                                </Badge>
                            </Td>
                            <Td maxW="200px" isTruncated title={movement.notes}>
                                {movement.notes || '-'}
                            </Td>
                        </Tr>
                    ))}
                </Tbody>
            </Table>
        </Box>
    );
};

export default MaterialMovementTable;
