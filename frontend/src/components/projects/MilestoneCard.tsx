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
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiCalendar,
  FiCheckCircle,
  FiClock,
  FiMoreVertical,
  FiEdit,
  FiTrash2,
  FiAlertTriangle,
  FiUsers,
  FiBriefcase,
} from 'react-icons/fi';

interface Milestone {
  id: number;
  title: string;
  description?: string;
  work_area?: string;
  priority: string;
  assigned_team?: string;
  target_date: string;
  status: string;
  completion_date?: string;
}

interface MilestoneCardProps {
  milestone: Milestone;
  onEdit: (milestone: Milestone) => void;
  onDelete: (milestoneId: number) => void;
  onComplete?: (milestoneId: number) => void;
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

  // Priority colors mapping
  const getPriorityColor = (priority: string): string => {
    switch (priority?.toLowerCase()) {
      case 'high':
        return 'red';
      case 'medium':
        return 'yellow';
      case 'low':
        return 'green';
      default:
        return 'gray';
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
        borderColor: `${getStatusColor(milestone.status)}.400`,
      }}
    >
      <CardBody>
        <HStack justify="space-between" align="start" spacing={4}>
          {/* Left Section - Main Info */}
          <VStack align="start" spacing={3} flex={1}>
            {/* Title and Status */}
            <HStack spacing={2} wrap="wrap">
              <Text fontSize="md" fontWeight="semibold" color={textColor}>
                {milestone.title}
              </Text>
              <Badge colorScheme={getStatusColor(milestone.status)} fontSize="xs">
                {milestone.status.toUpperCase().replace('-', ' ')}
              </Badge>
              {milestone.priority && (
                <Badge colorScheme={getPriorityColor(milestone.priority)} fontSize="xs">
                  {milestone.priority.toUpperCase()}
                </Badge>
              )}
            </HStack>

            {/* Description */}
            {milestone.description && (
              <Text fontSize="sm" color={subtextColor} noOfLines={2}>
                {milestone.description}
              </Text>
            )}

            {/* Additional Info */}
            <VStack align="start" spacing={1.5} w="full">
              {/* Work Area */}
              {milestone.work_area && (
                <HStack spacing={2}>
                  <Icon as={FiBriefcase} color="gray.500" boxSize={3.5} />
                  <Text fontSize="xs" color={subtextColor}>
                    {milestone.work_area}
                  </Text>
                </HStack>
              )}

              {/* Assigned Team */}
              {milestone.assigned_team && (
                <HStack spacing={2}>
                  <Icon as={FiUsers} color="gray.500" boxSize={3.5} />
                  <Text fontSize="xs" color={subtextColor}>
                    {milestone.assigned_team}
                  </Text>
                </HStack>
              )}

              {/* Target Date */}
              <HStack spacing={2}>
                <Icon as={FiCalendar} color="gray.500" boxSize={3.5} />
                <Text fontSize="xs" color={subtextColor}>
                  Target: {formatDate(milestone.target_date)}
                </Text>
              </HStack>

              {/* Days Info */}
              <HStack spacing={2}>
                <Icon as={daysInfo.icon} color={daysInfo.color} boxSize={3.5} />
                <Text fontSize="xs" color={daysInfo.color} fontWeight="medium">
                  {daysInfo.text}
                </Text>
              </HStack>
            </VStack>
          </VStack>

          {/* Right Section - Actions */}
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
      </CardBody>
    </Card>
  );
};

export default MilestoneCard;

