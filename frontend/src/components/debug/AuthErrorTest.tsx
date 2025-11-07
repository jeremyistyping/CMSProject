'use client';

import React, { useState } from 'react';
import {
  Box,
  Button,
  VStack,
  HStack,
  Text,
  Alert,
  AlertIcon,
  useToast,
  Badge,
  Code,
} from '@chakra-ui/react';
import api from '@/services/api';
import { useAuth } from '@/contexts/AuthContext';
import { handleApiError, isAuthError } from '@/utils/authErrorHandler';

const AuthErrorTest: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [lastError, setLastError] = useState<any>(null);
  const { user, token } = useAuth();
  const toast = useToast();

  const testValidCall = async () => {
    setLoading(true);
    setLastError(null);
    
    try {
      const response = await api.get('/permissions/me');
      console.log('Valid call success:', response.data);
      
      toast({
        title: 'Success',
        description: 'Valid API call completed successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      console.error('Valid call failed:', error);
      setLastError(error);
      
      const errorResult = handleApiError(error, 'AuthErrorTest.testValidCall');
      toast({
        title: 'Error',
        description: errorResult.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const testInvalidToken = async () => {
    setLoading(true);
    setLastError(null);
    
    try {
      // Simulate invalid token by making request to non-existent endpoint
      // or by corrupting the token temporarily
      if (typeof window !== 'undefined') {
        const originalToken = localStorage.getItem('token');
        localStorage.setItem('token', 'invalid-token-123');
        
        try {
          await api.get('/permissions/me');
        } finally {
          // Restore original token
          if (originalToken) {
            localStorage.setItem('token', originalToken);
          }
        }
      }
    } catch (error: any) {
      console.error('Invalid token test:', error);
      setLastError(error);
      
      toast({
        title: 'Test Result',
        description: `Auth error detected: ${isAuthError(error) ? 'YES' : 'NO'}`,
        status: isAuthError(error) ? 'info' : 'warning',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const testPermissionError = async () => {
    setLoading(true);
    setLastError(null);
    
    try {
      // Try to access admin-only endpoint (this should return 403)
      await api.get('/admin/test-endpoint');
    } catch (error: any) {
      console.error('Permission test:', error);
      setLastError(error);
      
      const errorResult = handleApiError(error, 'AuthErrorTest.testPermissionError');
      toast({
        title: 'Test Result',
        description: errorResult.message,
        status: error?.response?.status === 403 ? 'info' : 'warning',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const clearLocalStorage = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
      localStorage.removeItem('lastLogoutTime');
      
      toast({
        title: 'Cleared',
        description: 'LocalStorage auth data cleared',
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }
  };
  
  const simulateLogout = () => {
    if (typeof window !== 'undefined') {
      // Simulate the logout process
      localStorage.removeItem('token');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
      localStorage.setItem('lastLogoutTime', Date.now().toString());
      
      toast({
        title: 'Simulated Logout',
        description: 'Auth data cleared with logout timestamp set',
        status: 'warning',
        duration: 5000,
        isClosable: true,
      });
    }
  };
  
  const checkLogoutTimestamp = () => {
    if (typeof window !== 'undefined') {
      const lastLogoutTime = localStorage.getItem('lastLogoutTime');
      if (lastLogoutTime) {
        const timeSinceLogout = Date.now() - parseInt(lastLogoutTime);
        const minutesSince = Math.floor(timeSinceLogout / (1000 * 60));
        
        toast({
          title: 'Logout Timestamp Check',
          description: `Last logout: ${minutesSince} minutes ago`,
          status: 'info',
          duration: 5000,
          isClosable: true,
        });
      } else {
        toast({
          title: 'No Logout Timestamp',
          description: 'No recent logout timestamp found',
          status: 'info',
          duration: 3000,
          isClosable: true,
        });
      }
    }
  };

  return (
    <Box p={6} maxW="800px" mx="auto">
      <VStack spacing={6} align="stretch">
        <Box>
          <Text fontSize="2xl" fontWeight="bold" mb={4}>
            Authentication Error Handling Test
          </Text>
          <Text color="gray.600">
            This component helps test the authentication error handling system.
          </Text>
        </Box>

        <Alert status="info">
          <AlertIcon />
          <VStack align="start" spacing={1} flex={1}>
            <Text fontWeight="bold">Current Auth Status:</Text>
            <HStack>
              <Text>User:</Text>
              <Badge colorScheme={user ? 'green' : 'red'}>
                {user ? `${user.name} (${user.role})` : 'Not logged in'}
              </Badge>
            </HStack>
            <HStack>
              <Text>Token:</Text>
              <Badge colorScheme={token ? 'green' : 'red'}>
                {token ? `${token.substring(0, 20)}...` : 'None'}
              </Badge>
            </HStack>
            <HStack>
              <Text>Logout Timestamp:</Text>
              <Badge colorScheme={typeof window !== 'undefined' && localStorage.getItem('lastLogoutTime') ? 'orange' : 'gray'}>
                {typeof window !== 'undefined' && localStorage.getItem('lastLogoutTime') 
                  ? new Date(parseInt(localStorage.getItem('lastLogoutTime')!)).toLocaleTimeString()
                  : 'None'
                }
              </Badge>
            </HStack>
          </VStack>
        </Alert>

        <VStack spacing={4}>
          <Text fontWeight="bold">Test Scenarios:</Text>
          
          <HStack spacing={4} width="100%">
            <Button
              onClick={testValidCall}
              isLoading={loading}
              colorScheme="green"
              flex={1}
            >
              Test Valid API Call
            </Button>
            
            <Button
              onClick={testInvalidToken}
              isLoading={loading}
              colorScheme="orange"
              flex={1}
            >
              Test Invalid Token (401)
            </Button>
            
            <Button
              onClick={testPermissionError}
              isLoading={loading}
              colorScheme="red"
              flex={1}
            >
              Test Permission Error (403)
            </Button>
          </HStack>
          
            <HStack spacing={4} width="100%">
              <Button
                onClick={simulateLogout}
                colorScheme="orange"
                variant="outline"
                size="sm"
                flex={1}
              >
                Simulate Logout
              </Button>
              
              <Button
                onClick={checkLogoutTimestamp}
                colorScheme="purple"
                variant="outline"
                size="sm"
                flex={1}
              >
                Check Logout Timestamp
              </Button>
            </HStack>
            
            <Button
              onClick={clearLocalStorage}
              colorScheme="gray"
              variant="outline"
              size="sm"
            >
              Clear All LocalStorage Data
            </Button>
        </VStack>

        {lastError && (
          <Alert status="warning">
            <AlertIcon />
            <VStack align="start" spacing={2} flex={1}>
              <Text fontWeight="bold">Last Error Details:</Text>
              <Text fontSize="sm">Status: {lastError?.response?.status || 'N/A'}</Text>
              <Text fontSize="sm">Message: {lastError?.message || 'N/A'}</Text>
              <Text fontSize="sm">Is Auth Error: {isAuthError(lastError) ? 'Yes' : 'No'}</Text>
              <Code fontSize="xs" p={2} borderRadius="md" maxH="200px" overflow="auto" w="100%">
                {JSON.stringify(lastError, null, 2)}
              </Code>
            </VStack>
          </Alert>
        )}
      </VStack>
    </Box>
  );
};

export default AuthErrorTest;
