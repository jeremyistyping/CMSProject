'use client';

import React from 'react';
import {
  FormControl,
  FormLabel,
  FormErrorMessage,
  FormHelperText,
  Tooltip,
  Icon,
  Box,
  HStack,
} from '@chakra-ui/react';
import { FiInfo } from 'react-icons/fi';

interface FormFieldWithTooltipProps {
  label: string;
  tooltip?: string;
  error?: string;
  helperText?: string;
  isRequired?: boolean;
  isInvalid?: boolean;
  children: React.ReactNode;
}

/**
 * Wrapper component untuk form field dengan tooltip informatif
 * Menampilkan icon info kecil di samping label yang bisa di-hover untuk melihat tooltip
 */
export const FormFieldWithTooltip: React.FC<FormFieldWithTooltipProps> = ({
  label,
  tooltip,
  error,
  helperText,
  isRequired = false,
  isInvalid = false,
  children,
}) => {
  return (
    <FormControl isRequired={isRequired} isInvalid={isInvalid}>
      <HStack spacing={1} mb={1}>
        <FormLabel mb={0}>{label}</FormLabel>
        {tooltip && (
          <Tooltip 
            label={tooltip} 
            fontSize="sm" 
            placement="top"
            hasArrow
            bg="gray.700"
            color="white"
            px={3}
            py={2}
            borderRadius="md"
          >
            <Box display="inline-flex" cursor="help">
              <Icon 
                as={FiInfo} 
                color="blue.500" 
                boxSize={4}
                _hover={{ color: 'blue.600' }}
              />
            </Box>
          </Tooltip>
        )}
      </HStack>
      {children}
      {helperText && !error && (
        <FormHelperText fontSize="xs" color="gray.500">
          {helperText}
        </FormHelperText>
      )}
      {error && (
        <FormErrorMessage fontSize="xs">
          {error}
        </FormErrorMessage>
      )}
    </FormControl>
  );
};

export default FormFieldWithTooltip;
