'use client';

import React, { useState, useEffect } from 'react';
import {
  Input,
  InputGroup,
  InputLeftElement,
  Text,
  FormControl,
  FormLabel,
  FormErrorMessage,
} from '@chakra-ui/react';

interface CurrencyInputProps {
  value: number;
  onChange: (value: number) => void;
  placeholder?: string;
  isDisabled?: boolean;
  isInvalid?: boolean;
  size?: string;
  label?: string;
  isRequired?: boolean;
  errorMessage?: string;
  min?: number;
  max?: number;
  showLabel?: boolean;
}

const CurrencyInput: React.FC<CurrencyInputProps> = ({
  value,
  onChange,
  placeholder = "Masukkan jumlah",
  isDisabled = false,
  isInvalid = false,
  size = "md",
  label,
  isRequired = false,
  errorMessage,
  min = 0,
  max,
  showLabel = true,
}) => {
  const [displayValue, setDisplayValue] = useState<string>('');
  const [isFocused, setIsFocused] = useState(false);

  // Format number to IDR currency display
  const formatToIDR = (amount: number): string => {
    if (amount === 0) return '';
    return new Intl.NumberFormat('id-ID').format(amount);
  };

  // Parse formatted string back to number
  const parseFromIDR = (formatted: string): number => {
    const cleaned = formatted.replace(/[^\d]/g, '');
    return cleaned === '' ? 0 : parseInt(cleaned, 10);
  };

  // Update display value when value prop changes
  useEffect(() => {
    if (!isFocused) {
      setDisplayValue(formatToIDR(value));
    }
  }, [value, isFocused]);

  // Handle input change
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const inputValue = e.target.value;
    
    // Allow only numbers, dots, and commas for formatting
    const cleaned = inputValue.replace(/[^\d]/g, '');
    
    // Update display with formatting
    const formatted = cleaned === '' ? '' : new Intl.NumberFormat('id-ID').format(parseInt(cleaned, 10));
    setDisplayValue(formatted);
    
    // Convert back to number and call onChange
    const numericValue = cleaned === '' ? 0 : parseInt(cleaned, 10);
    
    // Apply min/max constraints
    let constrainedValue = numericValue;
    if (min !== undefined && numericValue < min) {
      constrainedValue = min;
    }
    if (max !== undefined && numericValue > max) {
      constrainedValue = max;
    }
    
    onChange(constrainedValue);
  };

  // Handle focus
  const handleFocus = () => {
    setIsFocused(true);
    // Show raw number when focused for easier editing
    setDisplayValue(value === 0 ? '' : value.toString());
  };

  // Handle blur
  const handleBlur = () => {
    setIsFocused(false);
    // Format back to currency display
    setDisplayValue(formatToIDR(value));
  };

  // Handle key press
  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    // Allow only numbers, backspace, delete, arrow keys
    const allowedKeys = ['Backspace', 'Delete', 'ArrowLeft', 'ArrowRight', 'Tab'];
    const isNumber = /^\d$/.test(e.key);
    
    if (!isNumber && !allowedKeys.includes(e.key)) {
      e.preventDefault();
    }
  };

  const inputElement = (
    <InputGroup size={size}>
      <InputLeftElement pointerEvents="none">
        <Text color="gray.500" fontSize={size === 'sm' ? 'xs' : 'sm'} fontWeight="medium">
          Rp
        </Text>
      </InputLeftElement>
      <Input
        value={displayValue}
        onChange={handleInputChange}
        onFocus={handleFocus}
        onBlur={handleBlur}
        onKeyDown={handleKeyPress}
        placeholder={placeholder}
        isDisabled={isDisabled}
        isInvalid={isInvalid}
        paddingLeft="40px"
        textAlign="right"
        fontFamily="mono"
      />
    </InputGroup>
  );

  if (label && showLabel) {
    return (
      <FormControl isRequired={isRequired} isInvalid={isInvalid}>
        <FormLabel fontSize="sm" fontWeight="medium">
          {label}
        </FormLabel>
        {inputElement}
        {errorMessage && (
          <FormErrorMessage fontSize="xs">
            {errorMessage}
          </FormErrorMessage>
        )}
      </FormControl>
    );
  }

  return inputElement;
};

export default CurrencyInput;
