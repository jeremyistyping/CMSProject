'use client';

import React from 'react';
import {
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  FormControl,
  FormLabel,
  FormHelperText,
  FormErrorMessage,
} from '@chakra-ui/react';

interface RupiahInputProps {
  value: number;
  onChange: (value: number) => void;
  label?: string;
  placeholder?: string;
  helperText?: string;
  errorMessage?: string;
  isRequired?: boolean;
  isInvalid?: boolean;
  isDisabled?: boolean;
  min?: number;
  max?: number;
  size?: 'xs' | 'sm' | 'md' | 'lg';
  allowDecimals?: boolean;
}

const RupiahInput: React.FC<RupiahInputProps> = ({
  value,
  onChange,
  label,
  placeholder = "Rp 0",
  helperText,
  errorMessage,
  isRequired = false,
  isInvalid = false,
  isDisabled = false,
  min = 0,
  max,
  size = 'md',
  allowDecimals = false,
}) => {
  // Format number as Rupiah
  const formatRupiah = (val: number | string): string => {
    if (!val || val === 0) return '';
    const numValue = typeof val === 'string' ? parseFloat(val) : val;
    if (isNaN(numValue)) return '';
    return `Rp ${numValue.toLocaleString('id-ID')}`;
  };

  // Parse Rupiah string to number
  const parseRupiah = (val: string): string => {
    return val.replace(/^Rp\s?/, '').replace(/[.,]/g, '');
  };

  const handleChange = (valueString: string, valueNumber: number) => {
    if (isNaN(valueNumber)) {
      onChange(0);
    } else {
      onChange(valueNumber);
    }
  };

  const inputComponent = (
    <NumberInput
      value={value}
      onChange={handleChange}
      min={min}
      max={max}
      precision={allowDecimals ? 2 : 0}
      format={formatRupiah}
      parse={parseRupiah}
      size={size}
      isDisabled={isDisabled}
    >
      <NumberInputField 
        placeholder={placeholder}
        fontFamily="mono"
      />
      <NumberInputStepper>
        <NumberIncrementStepper />
        <NumberDecrementStepper />
      </NumberInputStepper>
    </NumberInput>
  );

  // If no label, return just the input
  if (!label) {
    return inputComponent;
  }

  // Return with FormControl wrapper
  return (
    <FormControl isRequired={isRequired} isInvalid={isInvalid}>
      <FormLabel>{label}</FormLabel>
      {inputComponent}
      {helperText && <FormHelperText>{helperText}</FormHelperText>}
      {errorMessage && <FormErrorMessage>{errorMessage}</FormErrorMessage>}
    </FormControl>
  );
};

export default RupiahInput;
