'use client';

import React from 'react';
import {
  Button,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  IconButton,
  Tooltip,
  Portal,
} from '@chakra-ui/react';
import { 
  FiDownload, 
  FiFilePlus, 
  FiChevronDown,
  FiFileText 
} from 'react-icons/fi';

interface ExportButtonProps {
  onExportExcel?: () => void;
  onExportPDF?: () => void;
  onExportCSV?: () => void;
  isLoading?: boolean;
  variant?: 'button' | 'icon';
  size?: 'sm' | 'md' | 'lg';
  colorScheme?: string;
  disabled?: boolean;
  showExcel?: boolean;
  showPDF?: boolean;
  showCSV?: boolean;
}

const ExportButton: React.FC<ExportButtonProps> = ({
  onExportExcel,
  onExportPDF,
  onExportCSV,
  isLoading = false,
  variant = 'button',
  size = 'md',
  colorScheme = 'gray',
  disabled = false,
  showExcel = true,
  showPDF = true,
  showCSV = false,
}) => {
  // Count available export options
  const availableOptions = [showExcel, showPDF, showCSV].filter(Boolean).length;
  
  // If only one option is available, show direct button
  if (availableOptions === 1) {
    if (showPDF && onExportPDF) {
      return variant === 'icon' ? (
        <Tooltip label="Export to PDF">
          <IconButton
            aria-label="Export to PDF"
            icon={<FiFilePlus />}
            variant="outline"
            colorScheme="red"
            size={size}
            onClick={onExportPDF}
            isLoading={isLoading}
            isDisabled={disabled}
          />
        </Tooltip>
      ) : (
        <Button
          leftIcon={<FiFilePlus />}
          colorScheme="red"
          variant="outline"
          size={size}
          onClick={onExportPDF}
          isLoading={isLoading}
          isDisabled={disabled}
        >
          Export PDF
        </Button>
      );
    }
    
    if (showExcel && onExportExcel) {
      return variant === 'icon' ? (
        <Tooltip label="Export to Excel">
          <IconButton
            aria-label="Export to Excel"
            icon={<FiDownload />}
            variant="outline"
            colorScheme={colorScheme}
            size={size}
            onClick={onExportExcel}
            isLoading={isLoading}
            isDisabled={disabled}
          />
        </Tooltip>
      ) : (
        <Button
          leftIcon={<FiDownload />}
          colorScheme={colorScheme}
          variant="outline"
          size={size}
          onClick={onExportExcel}
          isLoading={isLoading}
          isDisabled={disabled}
        >
          Export Excel
        </Button>
      );
    }
    
    if (showCSV && onExportCSV) {
      return variant === 'icon' ? (
        <Tooltip label="Export to CSV">
          <IconButton
            aria-label="Export to CSV"
            icon={<FiFileText />}
            variant="outline"
            colorScheme={colorScheme}
            size={size}
            onClick={onExportCSV}
            isLoading={isLoading}
            isDisabled={disabled}
          />
        </Tooltip>
      ) : (
        <Button
          leftIcon={<FiFileText />}
          colorScheme={colorScheme}
          variant="outline"
          size={size}
          onClick={onExportCSV}
          isLoading={isLoading}
          isDisabled={disabled}
        >
          Export CSV
        </Button>
      );
    }
  }

  // Multiple options - show dropdown menu
  return (
    <Menu placement="bottom-end" strategy="fixed" isLazy>
      {variant === 'icon' ? (
        <Tooltip label="Export Data">
          <MenuButton
            as={IconButton}
            aria-label="Export Data"
            icon={<FiDownload />}
            variant="outline"
            colorScheme={colorScheme}
            size={size}
            isLoading={isLoading}
            isDisabled={disabled}
          />
        </Tooltip>
      ) : (
        <MenuButton
          as={Button}
          leftIcon={<FiDownload />}
          rightIcon={<FiChevronDown />}
          colorScheme={colorScheme}
          variant="outline"
          size={size}
          isLoading={isLoading}
          isDisabled={disabled}
        >
          Export
        </MenuButton>
      )}
      
      <Portal>
        <MenuList zIndex={1500} minW="200px" shadow="lg">
          {showExcel && onExportExcel && (
            <MenuItem
              icon={<FiDownload />}
              onClick={onExportExcel}
            >
              Export to Excel
            </MenuItem>
          )}
          
          {showPDF && onExportPDF && (
            <MenuItem
              icon={<FiFilePlus />}
              onClick={onExportPDF}
              color="red.600"
            >
              Export to PDF
            </MenuItem>
          )}
          
          {showCSV && onExportCSV && (
            <>
              {(showExcel || showPDF) && <MenuDivider />}
              <MenuItem
                icon={<FiFileText />}
                onClick={onExportCSV}
              >
                Export to CSV
              </MenuItem>
            </>
          )}
        </MenuList>
      </Portal>
    </Menu>
  );
};

export default ExportButton;
