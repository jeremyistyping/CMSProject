'use client';

import React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import {
  Box,
  Flex,
  Button,
  Text,
  Heading,
  Stack,
  useToast,
} from '@chakra-ui/react';
import { FiXCircle, FiLayout, FiLogOut, FiRefreshCw } from 'react-icons/fi';

export default function UnauthorizedPage() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const toast = useToast();

  const handleClearAuthAndRelogin = () => {
    // Clear all auth data
    logout();
    
    // Clear localStorage completely
    if (typeof window !== 'undefined') {
      localStorage.clear();
    }
    
    toast({
      title: 'Auth Data Cleared',
      description: 'Please login again with your credentials.',
      status: 'info',
      duration: 3000,
      isClosable: true,
    });
    
    // Redirect to login
    router.push('/login');
  };

  const handleLogout = () => {
    logout();
    toast({
      title: 'Logged Out',
      description: 'You have been logged out successfully.',
      status: 'success',
      duration: 3000,
      isClosable: true,
    });
    router.push('/login');
  };

  return (
    <Flex minH="100vh" align="center" justify="center" bg="red.50">
      <Box 
        maxW="md"
        w="full"
        bg="white"
        boxShadow="lg"
        borderRadius="lg"
        p={8}
        mx={4}
      >
        <Stack align="center" spacing={4} mb={8}>
          <Flex 
            w={16} 
            h={16} 
            align="center" 
            justify="center" 
            color="white" 
            bg="red.500" 
            borderRadius="full"
          >
            <FiXCircle size="2rem" />
          </Flex>
          <Heading as="h1" size="lg" textAlign="center" color="red.600">
            Access Denied
          </Heading>
          <Text color="gray.600" textAlign="center">
            You do not have permission to view this page
          </Text>
        </Stack>
        
        <Box textAlign="center" mb={6}>
          <Heading as="h2" size="md" mb={4}>Unauthorized Access</Heading>
          <Text mb={4} color="gray.600">
            Sorry, you are not authorized to access this resource.
          </Text>
          
          <Text fontSize="sm" color="gray.700" mb={4}>
            The page you are trying to access has restricted permissions.
            {user && (
              <Text as="span"> Your current role is <Text as="strong">{user.role}</Text>.</Text>
            )}
          </Text>
          
          <Text fontSize="sm" color="gray.600" mb={6}>
            If you believe this is an error, please contact your system administrator to request access.
          </Text>
        </Box>
        
        <Stack spacing={3} w="full">
          <Button 
            onClick={handleClearAuthAndRelogin}
            colorScheme="orange" 
            leftIcon={<FiRefreshCw />}
            width="full"
          >
            Clear Auth Data & Login Again
          </Button>
          
          <Button 
            onClick={handleLogout}
            colorScheme="red" 
            variant="outline"
            leftIcon={<FiLogOut />}
            width="full"
          >
            Logout
          </Button>
          
          <Button 
            as={Link} 
            href="/dashboard" 
            colorScheme="blue" 
            variant="ghost"
            leftIcon={<FiLayout />}
            width="full"
          >
            Try Dashboard Again
          </Button>
        </Stack>
      </Box>
    </Flex>
  );
}
