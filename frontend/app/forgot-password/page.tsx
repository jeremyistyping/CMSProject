'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  Flex,
  FormControl,
  FormLabel,
  Input,
  Button,
  Text,
  Heading,
  useToast,
  Stack,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
} from '@chakra-ui/react';
import { FiKey, FiCheckCircle, FiArrowLeft } from 'react-icons/fi';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const router = useRouter();
  const toast = useToast();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email) {
      setError('Please enter your email address');
      return;
    }

    try {
      setIsSubmitting(true);
      
      // TODO: Implement forgot password API call
      const response = await fetch('/api/v1/auth/forgot-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email }),
      });

      if (response.ok) {
        setSuccess(true);
      } else {
        const data = await response.json();
        setError(data.message || 'Failed to send reset email');
      }
    } catch (err) {
      setError('Something went wrong. Please try again.');
      console.error('Forgot password error:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (success) {
    return (
      <Flex minH="100vh" align="center" justify="center" bg="gray.50">
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
              bg="green.500" 
              borderRadius="full"
            >
              <FiCheckCircle size="2rem" />
            </Flex>
            <Heading as="h1" size="lg" textAlign="center">
              Check your email
            </Heading>
            <Text color="gray.600" textAlign="center">
              We&apos;ve sent a password reset link to your email address
            </Text>
          </Stack>
          
          <Text textAlign="center" mt={4}>
            We&apos;ve sent a password reset link to: <strong>{email}</strong>
            <br />
            Click the link in the email to reset your password. If you don&apos;t see the email, check your spam folder.
          </Text>
          
          <Button 
            onClick={() => setSuccess(false)} 
            colorScheme="green"
            variant="outline"
            width="full"
            mt={4}
          >
            Try another email
          </Button>
          
          <Button 
            variant="link"
            mt={4}
            colorScheme="gray"
            onClick={() => router.push('/login')}
          >
            Back to sign in
          </Button>
        </Box>
      </Flex>
    );
  }

  return (
    <Flex minH="100vh" align="center" justify="center" bg="gray.50">
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
            bg="brand.500" 
            borderRadius="full"
          >
            <FiKey size="2rem" />
          </Flex>
          <Heading as="h1" size="lg" textAlign="center">
            Reset your password
          </Heading>
          <Text color="gray.600" textAlign="center">
            Enter your email address and we&apos;ll send you a reset link
          </Text>
        </Stack>
        
        {error && (
          <Alert status="error" mb={4} borderRadius="md">
            <AlertIcon />
            <AlertTitle mr={2}>Reset Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <form onSubmit={handleSubmit}>
          <Stack spacing={4}>
            <FormControl id="email" isRequired>
              <FormLabel>Email address</FormLabel>
              <Input
                type="email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isSubmitting}
              />
            </FormControl>
            
            <Button
              type="submit"
              colorScheme="brand"
              isLoading={isSubmitting}
              loadingText="Sending reset link..."
              width="full"
            >
              Send reset link
            </Button>
          </Stack>
        </form>

        <Button 
          variant="link"
          mt={4}
          leftIcon={<FiArrowLeft />}
          onClick={() => router.push('/login')}
        >
          Back to sign in
        </Button>
      </Box>
    </Flex>
  );
}
