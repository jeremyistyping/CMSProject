'use client';

import React from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTheme } from '@/contexts/SimpleThemeContext';
import SimpleThemeToggle from '@/components/common/SimpleThemeToggle';
import UnifiedNotifications from '@/components/notification/UnifiedNotifications';
import {
  Flex,
  Text,
  IconButton,
  HStack,
  VStack,
  Avatar,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  Button,
  useColorMode,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiMenu, FiChevronDown, FiLogOut } from 'react-icons/fi';

interface SimpleNavbarProps {
  onMenuClick?: () => void;
  isMenuOpen?: boolean;
}

const SimpleNavbar: React.FC<SimpleNavbarProps> = ({ onMenuClick, isMenuOpen = false }) => {
  const { user, logout } = useAuth();
  const { theme } = useTheme();
  const { colorMode, setColorMode } = useColorMode();
  
  // Dynamic colors based on theme
  const bgColor = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subtextColor = useColorModeValue('gray.500', 'var(--text-secondary)');

  // Sync Chakra UI color mode with our custom theme
  React.useEffect(() => {
    if (colorMode !== theme) {
      setColorMode(theme);
    }
  }, [theme, colorMode, setColorMode]);

  return (
    <Flex
      h="16"
      alignItems="center"
      justifyContent="space-between"
      px={4}
      bg={bgColor}
      borderBottom="1px"
      borderColor={borderColor}
      position="sticky"
      top={0}
      zIndex={1000}
      className="navbar"
      transition="all 0.3s ease"
    >
      {/* Mobile menu button */}
      {onMenuClick && (
        <IconButton
          display={{ base: 'flex', md: 'none' }}
          onClick={onMenuClick}
          variant="outline"
          aria-label={isMenuOpen ? 'Close menu' : 'Open menu'}
          icon={<FiMenu />}
        />
      )}

      {/* Title - visible on mobile when no sidebar */}
      <Text fontSize="lg" fontWeight="semibold" color={textColor} display={{ base: onMenuClick ? 'none' : 'block', md: 'none' }}>
        Accounting System
      </Text>

      {/* Spacer for desktop */}
      <div style={{ flex: 1 }} />

      {/* Right side */}
      <HStack spacing={4}>
        {/* Notifications */}
        <UnifiedNotifications />
        
        {/* Theme Toggle */}
        <SimpleThemeToggle />

        {/* User menu */}
        {user && (
          <Menu>
            <MenuButton as={Button} variant="ghost" rightIcon={<FiChevronDown />}>
              <HStack spacing={2}>
                <Avatar size="sm" name={user.name} bg="brand.500" />
                <VStack spacing={0} align="start" display={{ base: 'none', md: 'flex' }}>
                  <Text fontSize="sm" fontWeight="medium" color={textColor}>
                    {user.name}
                  </Text>
                  <Text fontSize="xs" color={subtextColor}>
                    {user.role}
                  </Text>
                </VStack>
              </HStack>
            </MenuButton>
            <MenuList>
              <MenuItem icon={<FiLogOut />} onClick={logout}>
                Sign out
              </MenuItem>
            </MenuList>
          </Menu>
        )}
      </HStack>
    </Flex>
  );
};

export default SimpleNavbar;
