'use client';

import React from 'react';
import {
  Box,
  Card,
  CardBody,
  HStack,
  VStack,
  Text,
  Badge,
  Progress,
  Icon,
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  useColorModeValue,
  Tooltip,
} from '@chakra-ui/react';
import {
  FiCalendar,
  FiCheckCircle,
  FiClock,
  FiMoreVertical,
  FiEdit,
  FiTrash2,
  FiAlertTriangle,
} from 'react-icons/fi';
import { Milestone } from '@/types/project';

interface MilestoneCardProps {
  milestone: Milestone;
  onEdit: (milestone: Milestone) => void;
  onDelete: (milestoneId: string) => void;
  onComplete?: (milestoneId: string) => void;
}

const MilestoneCard: React.FC<MilestoneCardProps> = ({
  milestone,
  onEdit,
  onDelete,
  onComplete,
}) => {
  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');

  // Status colors mapping
  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'pending':
        return 'gray';
      case 'in-progress':
        return 'blue';
      case 'completed':
        return 'green';
      case 'delayed':
        return 'red';
      default:
        return 'gray';
    }
  };

  // Format date
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  };

  // Calculate days until/overdue
  const getDaysInfo = () => {
    const targetDate = new Date(milestone.target_date);
    const today = new Date();
    const diffTime = targetDate.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (milestone.status === 'completed') {
      return {
        text: `Completed on ${formatDate(milestone.completion_date!)}`,
        color: 'green.500',
        icon: FiCheckCircle,
      };
    }

    if (diffDays < 0) {
      return {
        text: `${Math.abs(diffDays)} days overdue`,
        color: 'red.500',
        icon: FiAlertTriangle,
      };
    }

    if (diffDays === 0) {
      return {
        text: 'Due today',
        color: 'orange.500',
        icon: FiClock,
      };
    }

    return {
      text: `${diffDays} days remaining`,
      color: 'blue.500',
      icon: FiClock,
    };
  };

  const daysInfo = getDaysInfo();
  const progressValue = milestone.progress || 0;

  return (
    <Card
      bg={bgColor}
      borderWidth="1px"
      borderColor={borderColor}
      transition="all 0.3s"
      _hover={{
        boxShadow: 'lg',
        transform: 'translateY(-2px)',
        borderColor: `${getStatusColor(milestone.status)}.400`,
      }}
    >
      <CardBody>
        <VStack align="stretch" spacing={4}>
          {/* Header with Title and Actions */}
          <HStack justify="space-between" align="start">
            <VStack align="start" spacing={1} flex={1}>
              <Text fontSize="lg" fontWeight="bold" color={textColor}>
                {milestone.title}
              </Text>
              <Badge colorScheme={getStatusColor(milestone.status)} fontSize="xs">
                {milestone.status.toUpperCase().replace('-', ' ')}
              </Badge>
            </VStack>
            <Menu>
              <MenuButton
                as={IconButton}
                icon={<FiMoreVertical />}
                variant="ghost"
                size="sm"
                aria-label="Actions"
              />
              <MenuList>
                {milestone.status !== 'completed' && onComplete && (
                  <MenuItem
                    icon={<FiCheckCircle />}
                    onClick={() => onComplete(milestone.id)}
                    color="green.500"
                  >
                    Mark as Complete
                  </MenuItem>
                )}
                <MenuItem icon={<FiEdit />} onClick={() => onEdit(milestone)}>
                  Edit
                </MenuItem>
                <MenuItem
                  icon={<FiTrash2 />}
                  color="red.500"
                  onClick={() => onDelete(milestone.id)}
                >
                  Delete
                </MenuItem>
              </MenuList>
            </Menu>
          </HStack>

          {/* Description */}
          {milestone.description && (
            <Text fontSize="sm" color={subtextColor} noOfLines={2}>
              {milestone.description}
            </Text>
          )}

          {/* Progress Bar */}
          <Box>
            <HStack justify="space-between" mb={2}>
              <Text fontSize="xs" fontWeight="semibold" color={subtextColor}>
                Progress
              </Text>
              <Text fontSize="xs" fontWeight="bold" color={textColor}>
                {progressValue}%
              </Text>
            </HStack>
            <Progress
              value={progressValue}
              size="sm"
              colorScheme={getStatusColor(milestone.status)}
              borderRadius="full"
            />
          </Box>

          {/* Dates */}
          <VStack align="stretch" spacing={2}>
            <HStack spacing={2}>
              <Icon as={FiCalendar} color="gray.500" boxSize={4} />
              <Text fontSize="xs" color={subtextColor}>
                Target: {formatDate(milestone.target_date)}
              </Text>
            </HStack>
            <HStack spacing={2}>
              <Icon as={daysInfo.icon} color={daysInfo.color} boxSize={4} />
              <Text fontSize="xs" color={daysInfo.color} fontWeight="medium">
                {daysInfo.text}
              </Text>
            </HStack>
          </VStack>
        </VStack>
      </CardBody>
    </Card>
  );
};

export default MilestoneCard;

