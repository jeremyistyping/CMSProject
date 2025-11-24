import React from 'react';
import {
    Box,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Badge,
    IconButton,
    Menu,
    MenuButton,
    MenuList,
    MenuItem,
    Text,
    HStack,
    Tooltip,
} from '@chakra-ui/react';
import { FiMoreVertical, FiEye, FiCheck, FiX, FiEdit } from 'react-icons/fi';
import { PurchaseRequest } from '../../types/purchaseRequest';

interface PRListProps {
    purchaseRequests: PurchaseRequest[];
    onView: (pr: PurchaseRequest) => void;
    onApprove?: (pr: PurchaseRequest) => void;
    onReject?: (pr: PurchaseRequest) => void;
}

const PRList: React.FC<PRListProps> = ({ purchaseRequests, onView, onApprove, onReject }) => {
    const getStatusColor = (status: string) => {
        switch (status) {
            case 'APPROVED': return 'green';
            case 'REJECTED': return 'red';
            case 'REVISION': return 'orange';
            case 'PO_CREATED': return 'blue';
            default: return 'yellow';
        }
    };

    return (
        <Box overflowX="auto">
            <Table variant="simple">
                <Thead>
                    <Tr>
                        <Th>Code</Th>
                        <Th>Date</Th>
                        <Th>Project</Th>
                        <Th>Requester</Th>
                        <Th isNumeric>Total Amount</Th>
                        <Th>Status</Th>
                        <Th>Actions</Th>
                    </Tr>
                </Thead>
                <Tbody>
                    {purchaseRequests.map((pr) => (
                        <Tr key={pr.id} _hover={{ bg: 'gray.50' }}>
                            <Td fontWeight="medium">{pr.code}</Td>
                            <Td>{new Date(pr.request_date).toLocaleDateString('id-ID')}</Td>
                            <Td>
                                <Tooltip label={pr.project?.project_name}>
                                    <Text noOfLines={1} maxW="200px">{pr.project?.project_name}</Text>
                                </Tooltip>
                            </Td>
                            <Td>{pr.requester?.name || 'Unknown'}</Td>
                            <Td isNumeric>
                                {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(pr.total_amount)}
                            </Td>
                            <Td>
                                <Badge colorScheme={getStatusColor(pr.status)}>
                                    {pr.status.replace('_', ' ')}
                                </Badge>
                            </Td>
                            <Td>
                                <HStack spacing={2}>
                                    <IconButton
                                        aria-label="View details"
                                        icon={<FiEye />}
                                        size="sm"
                                        variant="ghost"
                                        onClick={() => onView(pr)}
                                    />
                                    {pr.status === 'PENDING' && (
                                        <Menu>
                                            <MenuButton
                                                as={IconButton}
                                                aria-label="Options"
                                                icon={<FiMoreVertical />}
                                                size="sm"
                                                variant="ghost"
                                            />
                                            <MenuList>
                                                {onApprove && (
                                                    <MenuItem icon={<FiCheck />} onClick={() => onApprove(pr)}>
                                                        Approve
                                                    </MenuItem>
                                                )}
                                                {onReject && (
                                                    <MenuItem icon={<FiX />} onClick={() => onReject(pr)} color="red.500">
                                                        Reject
                                                    </MenuItem>
                                                )}
                                            </MenuList>
                                        </Menu>
                                    )}
                                </HStack>
                            </Td>
                        </Tr>
                    ))}
                    {purchaseRequests.length === 0 && (
                        <Tr>
                            <Td colSpan={7} textAlign="center" py={8} color="gray.500">
                                No Purchase Requests found
                            </Td>
                        </Tr>
                    )}
                </Tbody>
            </Table>
        </Box>
    );
};

export default PRList;
