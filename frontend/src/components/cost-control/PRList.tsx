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
import { FiMoreVertical, FiEye, FiCheck, FiX, FiEdit, FiTrash2 } from 'react-icons/fi';
import { PurchaseRequest } from '../../types/purchaseRequest';

interface PRListProps {
    purchaseRequests: PurchaseRequest[];
    onView: (pr: PurchaseRequest) => void;
    onEdit?: (pr: PurchaseRequest) => void;
    onDelete?: (pr: PurchaseRequest) => void;
    onApprove?: (pr: PurchaseRequest) => void;
    onReject?: (pr: PurchaseRequest) => void;
    onVerify?: (pr: PurchaseRequest) => void;
}

const PRList: React.FC<PRListProps> = ({ purchaseRequests, onView, onEdit, onDelete, onApprove, onReject, onVerify }) => {
    const getStatusColor = (status: string) => {
        switch (status) {
            case 'APPROVED': return 'green';
            case 'REJECTED': return 'red';
            case 'REVISION': return 'orange';
            case 'PO_CREATED': return 'blue';
            case 'VERIFIED': return 'cyan';
            case 'PENDING_VERIFICATION': return 'purple';
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
                            <Td>
                                {pr.requester?.role ? (
                                    <Badge colorScheme="purple" variant="outline" fontSize="sm" textTransform="uppercase">
                                        {pr.requester.role.replace('_', ' ')}
                                    </Badge>
                                ) : (
                                    <Text color="gray.400">Unknown</Text>
                                )}
                            </Td>
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
                                    <Menu>
                                        <MenuButton
                                            as={IconButton}
                                            aria-label="Options"
                                            icon={<FiMoreVertical />}
                                            size="sm"
                                            variant="ghost"
                                        />
                                        <MenuList>
                                            {onVerify && pr.status === 'PENDING_VERIFICATION' && (
                                                <MenuItem icon={<FiCheck />} onClick={() => onVerify(pr)} color="purple.500">
                                                    Verify & Map CBS
                                                </MenuItem>
                                            )}
                                            {onApprove && pr.status === 'VERIFIED' && (
                                                <MenuItem icon={<FiCheck />} onClick={() => onApprove(pr)} color="green.500">
                                                    Approve
                                                </MenuItem>
                                            )}
                                            {onReject && (pr.status === 'PENDING' || pr.status === 'PENDING_VERIFICATION' || pr.status === 'VERIFIED') && (
                                                <MenuItem icon={<FiX />} onClick={() => onReject(pr)} color="red.500">
                                                    Reject
                                                </MenuItem>
                                            )}
                                            {onEdit && pr.status === 'PENDING' && (
                                                <MenuItem icon={<FiEdit />} onClick={() => onEdit(pr)}>
                                                    Edit
                                                </MenuItem>
                                            )}
                                            {onDelete && pr.status === 'PENDING' && (
                                                <MenuItem icon={<FiTrash2 />} onClick={() => onDelete(pr)} color="red.500">
                                                    Delete
                                                </MenuItem>
                                            )}
                                        </MenuList>
                                    </Menu>
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
