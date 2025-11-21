'use client';

import React from 'react';
import {
    Card,
    CardBody,
    HStack,
    VStack,
    Text,
    Badge,
    Icon,
    useColorModeValue,
} from '@chakra-ui/react';
import {
    FiCalendar,
    FiClock,
    FiUsers,
    FiBriefcase,
} from 'react-icons/fi';
import { TimelineSchedule } from '@/types/project';
import { format, parseISO } from 'date-fns';

interface TimelineScheduleCardProps {
    schedule: TimelineSchedule;
}

const TimelineScheduleCard: React.FC<TimelineScheduleCardProps> = ({ schedule }) => {
    const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
    const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
    const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
    const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');

    const getStatusColor = (status: string): string => {
        switch (status) {
            case 'not-started':
                return 'gray';
            case 'in-progress':
                return 'yellow';
            case 'completed':
                return 'green';
            default:
                return 'gray';
        }
    };

    const formatDate = (dateString: string) => {
        try {
            return format(parseISO(dateString), 'MMM dd, yyyy');
        } catch {
            return dateString;
        }
    };

    return (
        <Card
            bg={bgColor}
            borderWidth="1px"
            borderColor={borderColor}
            transition="all 0.2s"
            _hover={{
                boxShadow: 'md',
                borderColor: `${getStatusColor(schedule.status)}.400`,
            }}
        >
            <CardBody p={4}>
                <HStack justify="space-between" align="start" spacing={4}>
                    <VStack align="start" spacing={3} flex={1}>
                        {/* Header: Work Area & Status */}
                        <HStack spacing={2} wrap="wrap">
                            <Text fontSize="md" fontWeight="semibold" color={textColor}>
                                {schedule.work_area}
                            </Text>
                            <Badge colorScheme={getStatusColor(schedule.status)} fontSize="xs">
                                {schedule.status.replace('-', ' ').toUpperCase()}
                            </Badge>
                        </HStack>

                        {/* Details */}
                        <VStack align="start" spacing={1.5} w="full">
                            {/* Assigned Team */}
                            <HStack spacing={2}>
                                <Icon as={FiUsers} color="gray.500" boxSize={3.5} />
                                <Text fontSize="xs" color={subtextColor}>
                                    {schedule.assigned_team || 'Unassigned'}
                                </Text>
                            </HStack>

                            {/* Date Range */}
                            <HStack spacing={2}>
                                <Icon as={FiCalendar} color="gray.500" boxSize={3.5} />
                                <Text fontSize="xs" color={subtextColor}>
                                    {formatDate(schedule.start_date)} - {formatDate(schedule.end_date)}
                                </Text>
                            </HStack>

                            {/* Duration */}
                            <HStack spacing={2}>
                                <Icon as={FiClock} color="gray.500" boxSize={3.5} />
                                <Text fontSize="xs" color={subtextColor}>
                                    {schedule.duration ? `${schedule.duration} days` : '-'}
                                </Text>
                            </HStack>
                        </VStack>
                    </VStack>
                </HStack>
            </CardBody>
        </Card>
    );
};

export default TimelineScheduleCard;
