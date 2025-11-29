import React from 'react';
import {
    Box,
    VStack,
    HStack,
    Text,
    Badge,
    Icon,
    Divider,
    Spinner,
} from '@chakra-ui/react';
import { FiCheck, FiX, FiClock, FiUser } from 'react-icons/fi';

export interface ApprovalHistoryItem {
    id: number;
    user_id: number;
    user?: {
        id: number;
        name: string;
        role: string;
    };
    action: string;
    comments: string;
    created_at: string;
}

export interface ApprovalAction {
    id: number;
    step_id: number;
    status: string;
    is_active: boolean;
    approver_id?: number;
    action_date?: string;
    comments?: string;
    step: {
        id: number;
        step_order: number;
        step_name: string;
        approver_role: string;
    };
    approver?: {
        id: number;
        name: string;
    };
}

export interface ApprovalTimelineProps {
    approval_request?: {
        id: number;
        request_code: string;
        status: string;
        priority: string;
        created_at: string;
    };
    approval_actions?: ApprovalAction[];
    approval_history?: ApprovalHistoryItem[];
    isLoading?: boolean;
}

const ApprovalTimeline: React.FC<ApprovalTimelineProps> = ({
    approval_request,
    approval_actions = [],
    approval_history = [],
    isLoading = false,
}) => {
    if (isLoading) {
        return (
            <Box p={4} textAlign="center">
                <Spinner size="sm" mr={2} />
                <Text fontSize="sm" display="inline">Loading approval timeline...</Text>
            </Box>
        );
    }

    if (!approval_request) {
        return (
            <Box p={4} bg="gray.50" borderRadius="md" textAlign="center">
                <Text fontSize="sm" color="gray.500">No approval workflow initiated yet.</Text>
            </Box>
        );
    }

    const getStepIcon = (status: string, isActive: boolean) => {
        if (status === 'APPROVED') return { icon: FiCheck, color: 'green.500', bg: 'green.50' };
        if (status === 'REJECTED') return { icon: FiX, color: 'red.500', bg: 'red.50' };
        if (isActive) return { icon: FiClock, color: 'blue.500', bg: 'blue.50' };
        return { icon: FiUser, color: 'gray.400', bg: 'gray.50' };
    };

    const getStatusBadge = (status: string) => {
        const colorMap: Record<string, string> = {
            'PENDING': 'yellow',
            'APPROVED': 'green',
            'REJECTED': 'red',
        };
        return <Badge colorScheme={colorMap[status] || 'gray'}>{status}</Badge>;
    };

    // Sort actions by step order
    const sortedActions = [...approval_actions].sort((a, b) => a.step.step_order - b.step.step_order);

    return (
        <Box>
            {/* Header */}
            <HStack justify="space-between" mb={4}>
                <Text fontWeight="bold" fontSize="lg">Approval Timeline</Text>
                <HStack>
                    <Text fontSize="sm" color="gray.500">Request Code:</Text>
                    <Badge colorScheme="purple">{approval_request.request_code}</Badge>
                    {getStatusBadge(approval_request.status)}
                </HStack>
            </HStack>

            <Divider mb={4} />

            {/* Timeline */}
            <VStack align="stretch" spacing={4}>
                {sortedActions.map((action, index) => {
                    const { icon: StepIcon, color: iconColor, bg: iconBg } = getStepIcon(action.status, action.is_active);
                    const isCompleted = action.status === 'APPROVED' || action.status === 'REJECTED';
                    const isLast = index === sortedActions.length - 1;

                    return (
                        <Box key={action.id} position="relative">
                            <HStack align="start" spacing={4}>
                                {/* Icon */}
                                <Box position="relative">
                                    <Box
                                        w="40px"
                                        h="40px"
                                        borderRadius="full"
                                        bg={iconBg}
                                        display="flex"
                                        alignItems="center"
                                        justifyContent="center"
                                        borderWidth={action.is_active ? '2px' : '0'}
                                        borderColor={action.is_active ? iconColor : 'transparent'}
                                    >
                                        <Icon as={StepIcon} color={iconColor} boxSize={5} />
                                    </Box>
                                    {/* Connecting line */}
                                    {!isLast && (
                                        <Box
                                            position="absolute"
                                            left="50%"
                                            top="40px"
                                            transform="translateX(-50%)"
                                            w="2px"
                                            h="40px"
                                            bg={isCompleted ? iconColor : 'gray.200'}
                                        />
                                    )}
                                </Box>

                                {/* Content */}
                                <Box flex="1" pb={isLast ? 0 : 6}>
                                    <HStack justify="space-between" mb={1}>
                                        <VStack align="start" spacing={0}>
                                            <HStack>
                                                <Badge colorScheme="gray" fontSize="xs">Step {action.step.step_order}</Badge>
                                                <Text fontWeight="semibold">{action.step.step_name}</Text>
                                            </HStack>
                                            <Text fontSize="xs" color="gray.500">
                                                Role: {action.step.approver_role}
                                            </Text>
                                        </VStack>
                                        {getStatusBadge(action.status)}
                                    </HStack>

                                    {/* Approver info */}
                                    {action.approver && (
                                        <Box mt={2} p={2} bg="gray.50" borderRadius="md">
                                            <HStack spacing={2} fontSize="sm">
                                                <Icon as={FiUser} color="gray.500" />
                                                <Text>
                                                    <b>{action.approver.name}</b>
                                                    {action.status === 'APPROVED' && ' approved'}
                                                    {action.status === 'REJECTED' && ' rejected'}
                                                </Text>
                                                {action.action_date && (
                                                    <Text color="gray.500">
                                                        on {new Date(action.action_date).toLocaleDateString('id-ID')}
                                                    </Text>
                                                )}
                                            </HStack>
                                            {action.comments && (
                                                <Text fontSize="sm" mt={1} color="gray.700">
                                                    "{action.comments}"
                                                </Text>
                                            )}
                                        </Box>
                                    )}

                                    {/* Active step indicator */}
                                    {action.is_active && action.status === 'PENDING' && (
                                        <Box mt={2} p={2} bg="blue.50" borderRadius="md" borderLeft="3px solid" borderColor="blue.500">
                                            <Text fontSize="sm" color="blue.700">
                                                ⏳ Awaiting approval from {action.step.approver_role}
                                            </Text>
                                        </Box>
                                    )}
                                </Box>
                            </HStack>
                        </Box>
                    );
                })}
            </VStack>

            {/* History Section */}
            {approval_history && approval_history.length > 0 && (
                <Box mt={6}>
                    <Divider mb={3} />
                    <Text fontWeight="semibold" fontSize="sm" mb={2} color="gray.600">Activity History</Text>
                    <VStack align="stretch" spacing={2}>
                        {approval_history.map((item) => (
                            <Box key={item.id} p={2} bg="gray.50" borderRadius="md" fontSize="sm">
                                <HStack justify="space-between">
                                    <HStack>
                                        <Text fontWeight="medium">{item.user?.name || 'System'}</Text>
                                        <Text color="gray.600">• {item.action}</Text>
                                    </HStack>
                                    <Text color="gray.500" fontSize="xs">
                                        {new Date(item.created_at).toLocaleString('id-ID')}
                                    </Text>
                                </HStack>
                                {item.comments && (
                                    <Text color="gray.600" mt={1}>"{item.comments}"</Text>
                                )}
                            </Box>
                        ))}
                    </VStack>
                </Box>
            )}
        </Box>
    );
};

export default ApprovalTimeline;
