import React, { useState, useEffect } from 'react';
import {
    Modal,
    ModalOverlay,
    ModalContent,
    ModalHeader,
    ModalFooter,
    ModalBody,
    ModalCloseButton,
    Button,
    VStack,
    HStack,
    Text,
    Badge,
    Table,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    Divider,
    Box,
    Textarea,
    useToast,
    FormControl,
    FormLabel,
    Tabs,
    TabList,
    TabPanels,
    Tab,
    TabPanel,
    Spinner,
} from '@chakra-ui/react';
import { PurchaseRequest } from '../../types/purchaseRequest';
import purchaseRequestService from '../../services/purchaseRequestService';
import approvalService, { ApprovalRequest, ApprovalAction } from '../../services/approvalService';
import cbsService from '../../services/cbsService';
import { PRCBSMapping } from '../../types/cbs';
import MaterialImpactCard from './MaterialImpactCard';
import ApprovalTimeline from './ApprovalTimeline';

interface PRDetailModalProps {
    isOpen: boolean;
    onClose: () => void;
    pr: PurchaseRequest | null;
    onUpdate: () => void;
}

const PRDetailModal: React.FC<PRDetailModalProps> = ({ isOpen, onClose, pr, onUpdate }) => {
    const toast = useToast();
    const [isLoading, setIsLoading] = useState(false);
    const [comments, setComments] = useState('');
    const [isRejecting, setIsRejecting] = useState(false);

    // CBS State
    const [cbsMappings, setCbsMappings] = useState<PRCBSMapping[]>([]);
    const [loadingCBS, setLoadingCBS] = useState(false);

    // Approval state
    const [approvalRequest, setApprovalRequest] = useState<ApprovalRequest | null>(null);
    const [loadingApproval, setLoadingApproval] = useState(false);
    const [currentUserRole, setCurrentUserRole] = useState<string>('');

    useEffect(() => {
        if (pr && isOpen) {
            fetchApprovalStatus();
            fetchCBSMappings();
            // Get current user role from localStorage or auth context
            const userData = localStorage.getItem('user');
            if (userData) {
                const user = JSON.parse(userData);
                setCurrentUserRole(user.role || '');
            }
        }
    }, [pr, isOpen]);

    const fetchCBSMappings = async () => {
        if (!pr) return;
        try {
            setLoadingCBS(true);
            const mappings = await cbsService.getPRCBSMappings(pr.id);
            setCbsMappings(mappings);
        } catch (error) {
            console.error('Error fetching CBS mappings:', error);
        } finally {
            setLoadingCBS(false);
        }
    };

    const fetchApprovalStatus = async () => {
        if (!pr) return;

        try {
            setLoadingApproval(true);
            // Fetch approval request for this PR
            const response = await approvalService.getApprovalRequests({
                entity_type: 'PURCHASE_REQUEST',
                page: 1,
                limit: 1,
            });

            // Find the approval request for this specific PR
            if (response.data && response.data.length > 0) {
                const prApproval = response.data.find((req: any) => req.entity_id === pr.id);
                if (prApproval) {
                    setApprovalRequest(prApproval);
                }
            }
        } catch (error) {
            console.error('Error fetching approval status:', error);
        } finally {
            setLoadingApproval(false);
        }
    };

    const canApproveCurrentStep = (): boolean => {
        if (!approvalRequest || !currentUserRole) return false;

        // Find active step
        const activeStep = approvalRequest.approval_steps?.find(step => step.is_active && step.status === 'PENDING');
        if (!activeStep) return false;

        // Check if current user's role matches the required role
        return activeStep.step.approver_role.toLowerCase() === currentUserRole.toLowerCase();
    };

    const getCurrentActiveStep = (): ApprovalAction | null => {
        if (!approvalRequest) return null;
        return approvalRequest.approval_steps?.find(step => step.is_active && step.status === 'PENDING') || null;
    };

    const handleApprove = async () => {
        if (!pr || !approvalRequest) return;

        const activeStep = getCurrentActiveStep();
        if (!activeStep) {
            toast({
                title: 'Error',
                description: 'No active approval step found',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        try {
            setIsLoading(true);
            await approvalService.approveSaleStep(approvalRequest.id, activeStep.step_id, { comments });

            toast({
                title: 'Success',
                description: `${activeStep.step.step_name} approved successfully`,
                status: 'success',
                duration: 3000,
                isClosable: true,
            });
            setComments('');
            onUpdate();
            fetchApprovalStatus(); // Refresh approval status
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.response?.data?.error || 'Failed to approve',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleReject = async () => {
        if (!comments.trim()) {
            toast({
                title: 'Error',
                description: 'Please provide a reason for rejection',
                status: 'warning',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        if (!pr || !approvalRequest) return;

        const activeStep = getCurrentActiveStep();
        if (!activeStep) {
            toast({
                title: 'Error',
                description: 'No active approval step found',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
            return;
        }

        try {
            setIsLoading(true);
            await approvalService.rejectSaleStep(approvalRequest.id, activeStep.step_id, { comments });

            toast({
                title: 'Success',
                description: 'Purchase Request rejected',
                status: 'success',
                duration: 3000,
                isClosable: true,
            });
            onUpdate();
            onClose();
            setIsRejecting(false);
            setComments('');
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.response?.data?.error || 'Failed to reject',
                status: 'error',
                duration: 3000,
                isClosable: true,
            });
        } finally {
            setIsLoading(false);
        }
    };

    if (!pr) return null;

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

    const showApprovalActions = canApproveCurrentStep() && approvalRequest?.status === 'PENDING';

    return (
        <Modal isOpen={isOpen} onClose={onClose} size="xl">
            <ModalOverlay />
            <ModalContent maxW="900px">
                <ModalHeader>
                    <HStack spacing={4}>
                        <Text>Purchase Request Details</Text>
                        <Badge colorScheme={getStatusColor(pr.status)} fontSize="md">
                            {pr.status.replace('_', ' ')}
                        </Badge>
                    </HStack>
                    <Text fontSize="sm" fontWeight="normal" color="gray.500" mt={1}>
                        {pr.code}
                    </Text>
                </ModalHeader>
                <ModalCloseButton />
                <ModalBody>
                    <Tabs variant="enclosed" colorScheme="blue">
                        <TabList>
                            <Tab>Details</Tab>
                            <Tab>CBS Breakdown</Tab>
                            <Tab>Approval Timeline</Tab>
                            <Tab>Material Impact</Tab>
                        </TabList>

                        <TabPanels>
                            {/* Details Tab */}
                            <TabPanel>
                                <VStack align="stretch" spacing={6}>
                                    {/* Header Info */}
                                    <Box p={4} bg="gray.50" borderRadius="md">
                                        <HStack justify="space-between" align="start" spacing={8} flexWrap="wrap">
                                            <VStack align="start" spacing={1}>
                                                <Text fontSize="sm" color="gray.500">Project</Text>
                                                <Text fontWeight="medium">{pr.project?.project_name}</Text>
                                            </VStack>
                                            <VStack align="start" spacing={1}>
                                                <Text fontSize="sm" color="gray.500">Requester</Text>
                                                {pr.requester?.role ? (
                                                    <Badge colorScheme="purple" variant="outline" fontSize="sm" textTransform="uppercase">
                                                        {pr.requester.role.replace('_', ' ')}
                                                    </Badge>
                                                ) : (
                                                    <Text fontWeight="medium" color="gray.400">Unknown</Text>
                                                )}
                                            </VStack>
                                            <VStack align="start" spacing={1}>
                                                <Text fontSize="sm" color="gray.500">Request Date</Text>
                                                <Text fontWeight="medium">{new Date(pr.request_date).toLocaleDateString('id-ID')}</Text>
                                            </VStack>
                                            <VStack align="start" spacing={1}>
                                                <Text fontSize="sm" color="gray.500">Required Date</Text>
                                                <Text fontWeight="medium">
                                                    {pr.required_date ? new Date(pr.required_date).toLocaleDateString('id-ID') : '-'}
                                                </Text>
                                            </VStack>
                                        </HStack>
                                    </Box>

                                    {/* Notes */}
                                    {pr.notes && (
                                        <Box>
                                            <Text fontSize="sm" color="gray.500" mb={1}>Notes</Text>
                                            <Text fontSize="sm" p={3} bg="gray.50" borderRadius="md" whiteSpace="pre-wrap">
                                                {pr.notes}
                                            </Text>
                                        </Box>
                                    )}

                                    <Divider />

                                    {/* Items Table */}
                                    <Box>
                                        <Text fontWeight="bold" mb={3}>Items</Text>
                                        <Table size="sm" variant="simple">
                                            <Thead>
                                                <Tr>
                                                    <Th>Item Name</Th>
                                                    <Th isNumeric>Qty</Th>
                                                    <Th>Unit</Th>
                                                    <Th isNumeric>Est. Price</Th>
                                                    <Th isNumeric>Total</Th>
                                                </Tr>
                                            </Thead>
                                            <Tbody>
                                                {pr.items.map((item) => (
                                                    <Tr key={item.id}>
                                                        <Td>
                                                            <Text fontWeight="medium">{item.item_name}</Text>
                                                            {item.notes && <Text fontSize="xs" color="gray.500">{item.notes}</Text>}
                                                        </Td>
                                                        <Td isNumeric>{item.quantity}</Td>
                                                        <Td>{item.unit}</Td>
                                                        <Td isNumeric>
                                                            {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(item.estimated_price)}
                                                        </Td>
                                                        <Td isNumeric fontWeight="bold">
                                                            {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(item.total_price)}
                                                        </Td>
                                                    </Tr>
                                                ))}
                                                <Tr bg="gray.50">
                                                    <Td colSpan={4} fontWeight="bold" textAlign="right">Total Amount</Td>
                                                    <Td isNumeric fontWeight="bold" fontSize="md">
                                                        {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(pr.total_amount)}
                                                    </Td>
                                                </Tr>
                                            </Tbody>
                                        </Table>
                                    </Box>
                                </VStack>
                            </TabPanel>

                            {/* CBS Breakdown Tab */}
                            <TabPanel>
                                {loadingCBS ? (
                                    <Box textAlign="center" py={8}>
                                        <Spinner />
                                        <Text mt={2} fontSize="sm" color="gray.500">Loading CBS data...</Text>
                                    </Box>
                                ) : cbsMappings.length > 0 ? (
                                    <VStack align="stretch" spacing={4}>
                                        <Box p={4} bg="purple.50" borderRadius="md" borderWidth="1px" borderColor="purple.200">
                                            <HStack>
                                                <Badge colorScheme="purple">VERIFIED</Badge>
                                                <Text fontSize="sm" color="purple.800">
                                                    This PR has been verified and mapped to the Cost Breakdown Structure.
                                                </Text>
                                            </HStack>
                                        </Box>
                                        <Table size="sm" variant="simple">
                                            <Thead>
                                                <Tr>
                                                    <Th>Item</Th>
                                                    <Th>CBS Code</Th>
                                                    <Th>CBS Name</Th>
                                                    <Th isNumeric>Allocated Amount</Th>
                                                </Tr>
                                            </Thead>
                                            <Tbody>
                                                {cbsMappings.map((mapping) => (
                                                    <Tr key={mapping.id}>
                                                        <Td>{pr?.items.find(i => i.id === mapping.pr_item_id)?.item_name || 'Unknown Item'}</Td>
                                                        <Td><Badge>{mapping.cbs_node?.code}</Badge></Td>
                                                        <Td>{mapping.cbs_node?.name}</Td>
                                                        <Td isNumeric>
                                                            {new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(mapping.allocated_amount)}
                                                        </Td>
                                                    </Tr>
                                                ))}
                                            </Tbody>
                                        </Table>
                                    </VStack>
                                ) : (
                                    <Box textAlign="center" py={8} color="gray.500">
                                        <Text>No CBS mappings found.</Text>
                                        {pr?.status === 'PENDING_VERIFICATION' && (
                                            <Text fontSize="sm" mt={2}>This PR is pending verification by Cost Control.</Text>
                                        )}
                                    </Box>
                                )}
                            </TabPanel>

                            {/* Approval Timeline Tab */}
                            <TabPanel>
                                {loadingApproval ? (
                                    <Box textAlign="center" py={8}>
                                        <Spinner />
                                        <Text mt={2} fontSize="sm" color="gray.500">Loading approval timeline...</Text>
                                    </Box>
                                ) : (
                                    <ApprovalTimeline
                                        approval_request={approvalRequest || undefined}
                                        approval_actions={approvalRequest?.approval_steps || []}
                                        isLoading={loadingApproval}
                                    />
                                )}

                                {/* Approval Actions */}
                                {showApprovalActions && (
                                    <Box mt={6} p={4} bg="blue.50" borderRadius="md" borderWidth="1px" borderColor="blue.200">
                                        <Text fontWeight="bold" mb={3} color="blue.800">
                                            🔔 Action Required - {getCurrentActiveStep()?.step.step_name}
                                        </Text>

                                        <FormControl mb={4}>
                                            <FormLabel fontSize="sm">Comments {isRejecting && <Text as="span" color="red.500">*</Text>}</FormLabel>
                                            <Textarea
                                                value={comments}
                                                onChange={(e) => setComments(e.target.value)}
                                                placeholder={isRejecting ? "Please explain why this request is being rejected..." : "Optional comments..."}
                                                size="sm"
                                            />
                                        </FormControl>

                                        {!isRejecting ? (
                                            <HStack>
                                                <Button
                                                    colorScheme="red"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => setIsRejecting(true)}
                                                >
                                                    Reject
                                                </Button>
                                                <Button
                                                    colorScheme="green"
                                                    size="sm"
                                                    onClick={handleApprove}
                                                    isLoading={isLoading}
                                                >
                                                    Approve
                                                </Button>
                                            </HStack>
                                        ) : (
                                            <HStack>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => {
                                                        setIsRejecting(false);
                                                        setComments('');
                                                    }}
                                                >
                                                    Cancel
                                                </Button>
                                                <Button
                                                    colorScheme="red"
                                                    size="sm"
                                                    onClick={handleReject}
                                                    isLoading={isLoading}
                                                >
                                                    Confirm Rejection
                                                </Button>
                                            </HStack>
                                        )}
                                    </Box>
                                )}
                            </TabPanel>

                            {/* Material Impact Tab */}
                            <TabPanel>
                                <MaterialImpactCard purchaseRequestId={pr.id} />
                            </TabPanel>
                        </TabPanels>
                    </Tabs>
                </ModalBody>

                <ModalFooter>
                    <Button variant="ghost" onClick={() => {
                        setIsRejecting(false);
                        setComments('');
                        onClose();
                    }}>
                        Close
                    </Button>
                </ModalFooter>
            </ModalContent>
        </Modal>
    );
};

export default PRDetailModal;
