'use client';

import React from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Badge,
  Icon,
  Tooltip,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiCheckCircle,
  FiClock,
  FiAlertCircle,
  FiCircle,
} from 'react-icons/fi';
import { useLanguage } from '@/contexts/LanguageContext';
import { format } from 'date-fns';

interface Milestone {
  id: number;
  name: string;
  description: string;
  target_date: string;
  actual_completion_date?: string;
  status: string;
  order_number: number;
  weight_percentage: number;
  is_overdue?: boolean;
  days_until_target?: number;
  days_delayed?: number;
}

interface MilestoneTimelineProps {
  milestones: Milestone[];
  onMilestoneClick?: (milestone: Milestone) => void;
}

const MilestoneTimeline: React.FC<MilestoneTimelineProps> = ({
  milestones,
  onMilestoneClick,
}) => {
  const { t } = useLanguage();
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const lineColor = useColorModeValue('gray.300', 'gray.600');

  // Sort milestones by order_number
  const sortedMilestones = [...milestones].sort((a, b) => a.order_number - b.order_number);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return FiCheckCircle;
      case 'in-progress':
        return FiClock;
      case 'delayed':
        return FiAlertCircle;
      default:
        return FiCircle;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'green';
      case 'in-progress':
        return 'blue';
      case 'delayed':
        return 'red';
      default:
        return 'gray';
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'completed':
        return t('milestones.statusCompleted');
      case 'in-progress':
        return t('milestones.statusInProgress');
      case 'delayed':
        return t('milestones.statusDelayed');
      default:
        return t('milestones.statusPending');
    }
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    try {
      return format(new Date(dateString), 'dd MMM yyyy');
    } catch {
      return dateString;
    }
  };

  const getDaysInfo = (milestone: Milestone) => {
    if (milestone.status === 'completed') {
      if (milestone.days_delayed && milestone.days_delayed > 0) {
        return {
          text: `${milestone.days_delayed} ${t('milestones.daysLate')}`,
          color: 'red.500',
        };
      }
      return {
        text: t('milestones.onTime'),
        color: 'green.500',
      };
    }

    if (milestone.is_overdue) {
      return {
        text: `${Math.abs(milestone.days_until_target || 0)} ${t('milestones.daysOverdue')}`,
        color: 'red.500',
      };
    }

    if (milestone.days_until_target !== undefined) {
      return {
        text: `${milestone.days_until_target} ${t('milestones.daysRemaining')}`,
        color: 'blue.500',
      };
    }

    return null;
  };

  if (!sortedMilestones || sortedMilestones.length === 0) {
    return (
      <Box p={8} textAlign="center" color="gray.500">
        <Text>{t('milestones.noMilestones')}</Text>
      </Box>
    );
  }

  return (
    <VStack align="stretch" spacing={0} position="relative" pl={8}>
      {/* Vertical line */}
      <Box
        position="absolute"
        left="15px"
        top="20px"
        bottom="20px"
        width="2px"
        bg={lineColor}
        zIndex={0}
      />

      {sortedMilestones.map((milestone, index) => {
        const StatusIconComponent = getStatusIcon(milestone.status);
        const statusColor = getStatusColor(milestone.status);
        const daysInfo = getDaysInfo(milestone);
        const isLast = index === sortedMilestones.length - 1;

        return (
          <Box key={milestone.id} position="relative" pb={isLast ? 0 : 6}>
            {/* Timeline dot/icon */}
            <Box
              position="absolute"
              left="-23px"
              top="10px"
              bg={bgColor}
              border="2px solid"
              borderColor={`${statusColor}.500`}
              borderRadius="full"
              p={1}
              zIndex={1}
            >
              <Icon
                as={StatusIconComponent}
                color={`${statusColor}.500`}
                boxSize={5}
              />
            </Box>

            {/* Milestone card */}
            <Box
              bg={bgColor}
              border="1px solid"
              borderColor={borderColor}
              borderRadius="md"
              p={4}
              cursor={onMilestoneClick ? 'pointer' : 'default'}
              onClick={() => onMilestoneClick?.(milestone)}
              _hover={onMilestoneClick ? { bg: hoverBg, transform: 'translateX(2px)' } : {}}
              transition="all 0.2s"
            >
              <VStack align="stretch" spacing={2}>
                {/* Header */}
                <HStack justify="space-between" align="start">
                  <VStack align="start" spacing={1} flex={1}>
                    <HStack>
                      <Text fontWeight="bold" fontSize="md">
                        {milestone.order_number}. {milestone.name}
                      </Text>
                      <Badge colorScheme={statusColor} fontSize="xs">
                        {getStatusLabel(milestone.status)}
                      </Badge>
                    </HStack>
                    {milestone.description && (
                      <Text fontSize="sm" color="gray.600" noOfLines={2}>
                        {milestone.description}
                      </Text>
                    )}
                  </VStack>

                  <Tooltip label={`${milestone.weight_percentage}% of total progress`}>
                    <Badge colorScheme="purple" fontSize="sm" px={2}>
                      {milestone.weight_percentage}%
                    </Badge>
                  </Tooltip>
                </HStack>

                {/* Dates & Status Info */}
                <HStack spacing={4} fontSize="sm" color="gray.600">
                  <HStack>
                    <Icon as={FiClock} />
                    <Text>
                      {t('milestones.target')}: {formatDate(milestone.target_date)}
                    </Text>
                  </HStack>

                  {milestone.actual_completion_date && (
                    <HStack>
                      <Icon as={FiCheckCircle} />
                      <Text>
                        {t('milestones.completed')}: {formatDate(milestone.actual_completion_date)}
                      </Text>
                    </HStack>
                  )}

                  {daysInfo && (
                    <Text fontWeight="semibold" color={daysInfo.color}>
                      {daysInfo.text}
                    </Text>
                  )}
                </HStack>
              </VStack>
            </Box>
          </Box>
        );
      })}
    </VStack>
  );
};

export default MilestoneTimeline;

