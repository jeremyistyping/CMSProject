'use client';

import React from 'react';
import {
  IconButton,
  Tooltip,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiSearch, FiEye, FiList } from 'react-icons/fi';

interface JournalDrilldownButtonProps {
  onClick: () => void;
  isDisabled?: boolean;
  variant?: 'search' | 'eye' | 'list';
  size?: 'xs' | 'sm' | 'md' | 'lg';
  label?: string;
  placement?: 'top' | 'bottom' | 'left' | 'right';
}

export const JournalDrilldownButton: React.FC<JournalDrilldownButtonProps> = ({
  onClick,
  isDisabled = false,
  variant = 'search',
  size = 'xs',
  label,
  placement = 'top',
}) => {
  const getIcon = () => {
    switch (variant) {
      case 'eye':
        return FiEye;
      case 'list':
        return FiList;
      default:
        return FiSearch;
    }
  };

  const getDefaultLabel = () => {
    switch (variant) {
      case 'eye':
        return 'View journal entries';
      case 'list':
        return 'Show journal entries';
      default:
        return 'Drill down to journal entries';
    }
  };

  const buttonBg = useColorModeValue('white', 'gray.700');
  const buttonHoverBg = useColorModeValue('blue.50', 'blue.900');
  const iconColor = useColorModeValue('blue.500', 'blue.300');

  return (
    <Tooltip 
      label={label || getDefaultLabel()} 
      fontSize="xs" 
      placement={placement}
      hasArrow
    >
      <IconButton
        aria-label={label || getDefaultLabel()}
        icon={React.createElement(getIcon())}
        size={size}
        variant="ghost"
        colorScheme="blue"
        bg={buttonBg}
        color={iconColor}
        _hover={{
          bg: buttonHoverBg,
          transform: 'scale(1.05)',
        }}
        _active={{
          transform: 'scale(0.95)',
        }}
        isDisabled={isDisabled}
        onClick={onClick}
        transition="all 0.2s"
      />
    </Tooltip>
  );
};

export default JournalDrilldownButton;
