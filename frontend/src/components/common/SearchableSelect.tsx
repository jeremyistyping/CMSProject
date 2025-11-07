import React, { useState, useRef, useEffect } from 'react';
import {
  Box,
  Input,
  List,
  ListItem,
  Text,
  InputGroup,
  InputRightElement,
  IconButton,
  useDisclosure,
  Collapse,
  Spinner,
  Badge,
  HStack,
} from '@chakra-ui/react';
import { FiChevronDown, FiChevronUp, FiX } from 'react-icons/fi';

interface SearchableSelectOption {
  id: number | string;
  code?: string;
  name: string;
  active?: boolean;
}

interface SearchableSelectProps {
  options: SearchableSelectOption[];
  value?: string | number;
  onChange: (value: string | number, option?: SearchableSelectOption) => void;
  placeholder?: string;
  isLoading?: boolean;
  isDisabled?: boolean;
  displayFormat?: (option: SearchableSelectOption) => string;
  filterFunction?: (option: SearchableSelectOption, searchTerm: string) => boolean;
  allowClear?: boolean;
}

const SearchableSelect: React.FC<SearchableSelectProps> = ({
  options,
  value,
  onChange,
  placeholder = "Select an option...",
  isLoading = false,
  isDisabled = false,
  displayFormat = (option) => option.code ? `${option.code} - ${option.name}` : option.name,
  filterFunction = (option, searchTerm) => {
    const term = searchTerm.toLowerCase();
    return (
      option.name.toLowerCase().includes(term) ||
      (option.code && option.code.toLowerCase().includes(term))
    );
  },
  allowClear = true,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedOption, setSelectedOption] = useState<SearchableSelectOption | null>(null);
  const { isOpen, onOpen, onClose } = useDisclosure();
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Find the selected option based on value
  useEffect(() => {
    if (value) {
      const found = options.find(option => option.id.toString() === value.toString());
      if (found) {
        setSelectedOption(found);
        setSearchTerm(''); // Clear search when an option is selected
      }
    } else {
      setSelectedOption(null);
      setSearchTerm('');
    }
  }, [value, options]);

  // Filter options based on search term
  const filteredOptions = searchTerm
    ? options.filter(option => filterFunction(option, searchTerm))
    : options;

  // Handle option selection
  const handleOptionSelect = (option: SearchableSelectOption) => {
    setSelectedOption(option);
    setSearchTerm('');
    onChange(option.id, option);
    onClose();
  };

  // Handle input change (search)
  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newSearchTerm = event.target.value;
    setSearchTerm(newSearchTerm);
    
    if (!isOpen && newSearchTerm) {
      onOpen();
    }
  };

  // Handle clear selection
  const handleClear = () => {
    setSelectedOption(null);
    setSearchTerm('');
    onChange('');
    onClose();
    inputRef.current?.focus();
  };

  // Handle input focus
  const handleInputFocus = () => {
    if (!isDisabled) {
      onOpen();
    }
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        onClose();
        setSearchTerm(''); // Clear search when closing
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [onClose]);

  // Smooth mouse-wheel scrolling when dropdown is open
  useEffect(() => {
    const el = dropdownRef.current;
    if (!isOpen || !el) return;

    const onWheel = (e: WheelEvent) => {
      // ignore zoom or horizontal scroll gestures
      if (e.ctrlKey || Math.abs(e.deltaX) > Math.abs(e.deltaY)) return;
      if (el.scrollHeight <= el.clientHeight) return;
      e.preventDefault();
      el.scrollBy({ top: e.deltaY, left: 0, behavior: 'smooth' });
    };

    el.addEventListener('wheel', onWheel, { passive: false });
    return () => {
      el.removeEventListener('wheel', onWheel as EventListener);
    };
  }, [isOpen]);

  // Ensure active/selected option is visible when opening
  useEffect(() => {
    if (!isOpen || !dropdownRef.current || !selectedOption) return;
    const el = dropdownRef.current.querySelector(
      `[data-option-id="${selectedOption.id}"]`
    ) as HTMLElement | null;
    el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
  }, [isOpen, selectedOption]);

  // Display value in input
  const getInputValue = () => {
    if (searchTerm) {
      return searchTerm; // Show search term when searching
    }
    if (selectedOption) {
      return displayFormat(selectedOption); // Show selected option when not searching
    }
    return ''; // Empty when nothing selected
  };

  return (
    <Box ref={containerRef} position="relative">
      <InputGroup>
        <Input
          ref={inputRef}
          value={getInputValue()}
          onChange={handleInputChange}
          onFocus={handleInputFocus}
          placeholder={selectedOption ? '' : placeholder}
          isDisabled={isDisabled}
          bg={selectedOption ? 'green.50' : 'white'}
          borderColor={selectedOption ? 'green.300' : 'gray.200'}
        />
        <InputRightElement width="4.5rem">
          <HStack spacing={1}>
            {allowClear && selectedOption && (
              <IconButton
                aria-label="Clear selection"
                icon={<FiX />}
                size="xs"
                variant="ghost"
                onClick={handleClear}
                isDisabled={isDisabled}
              />
            )}
            <IconButton
              aria-label="Toggle dropdown"
              icon={isOpen ? <FiChevronUp /> : <FiChevronDown />}
              size="xs"
              variant="ghost"
              onClick={isOpen ? onClose : onOpen}
              isDisabled={isDisabled}
            />
          </HStack>
        </InputRightElement>
      </InputGroup>

      <Collapse in={isOpen && !isDisabled}>
        <Box
          ref={dropdownRef}
          position="absolute"
          top="100%"
          left={0}
          right={0}
          zIndex={1000}
          bg="white"
          border="1px"
          borderColor="gray.200"
          borderRadius="md"
          boxShadow="lg"
          maxHeight="200px"
          overflowY="auto"
          mt={1}
          // Smooth programmatic scroll and better overscroll behaviour
          sx={{ scrollBehavior: 'smooth', overscrollBehavior: 'contain' }}
        >
          {isLoading ? (
            <Box p={4} textAlign="center">
              <Spinner size="sm" />
              <Text ml={2} fontSize="sm">Loading...</Text>
            </Box>
          ) : filteredOptions.length > 0 ? (
            <List>
              {filteredOptions.map((option) => (
                <ListItem
                  key={option.id}
                  data-option-id={option.id}
                  p={3}
                  cursor="pointer"
                  _hover={{ bg: 'gray.50' }}
                  bg={selectedOption?.id === option.id ? 'blue.50' : 'white'}
                  borderBottom="1px"
                  borderColor="gray.100"
                  onClick={() => handleOptionSelect(option)}
                >
                  <HStack justify="space-between">
                    <Text fontSize="sm">
                      {displayFormat(option)}
                    </Text>
                    {option.active === false && (
                      <Badge colorScheme="red" size="sm">
                        Inactive
                      </Badge>
                    )}
                  </HStack>
                </ListItem>
              ))}
            </List>
          ) : (
            <Box p={4} textAlign="center">
              <Text fontSize="sm" color="gray.500">
                {searchTerm ? 'No results found' : 'No options available'}
              </Text>
            </Box>
          )}
        </Box>
      </Collapse>
    </Box>
  );
};

export default SearchableSelect;
