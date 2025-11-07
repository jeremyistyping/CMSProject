'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import SimpleLayout from '@/components/layout/SimpleLayout';
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
import { FiUserPlus } from 'react-icons/fi';
import Link from 'next/link';

const RegisterContent = () => {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const { register, isAuthenticated } = useAuth();
  const router = useRouter();
  const toast = useToast();

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name || !email || !password) {
      setError('Please fill in all fields');
      return;
    }

    try {
      setIsSubmitting(true);
      await register(name, email, password);
      router.push('/dashboard');
    } catch (err) {
      setError('Registration failed. Please try again.');
      console.error('Registration error:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

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
            <FiUserPlus size="2rem" />
          </Flex>
          <Heading as="h1" size="lg" textAlign="center">
            Create your account
          </Heading>
          <Text color="gray.600" textAlign="center">
            Get started with your new accounting dashboard
          </Text>
        </Stack>
        
        {error && (
          <Alert status="error" mb={4} borderRadius="md">
            <AlertIcon />
            <AlertTitle mr={2}>Registration Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <form onSubmit={handleSubmit}>
          <Stack spacing={4}>
            <FormControl id="name" isRequired>
              <FormLabel>Full name</FormLabel>
              <Input
                type="text"
                placeholder="Enter your full name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isSubmitting}
              />
            </FormControl>
            
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
            
            <FormControl id="password" isRequired>
              <FormLabel>Password</FormLabel>
              <Input
                type="password"
                placeholder="Create a password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isSubmitting}
              />
            </FormControl>
            
            <Button
              type="submit"
              colorScheme="brand"
              isLoading={isSubmitting}
              loadingText="Creating account..."
              width="full"
            >
              Create account
            </Button>
          </Stack>
        </form>

        <Text mt={6} textAlign="center">
          Already have an account?{' '}
          <Button variant="link" colorScheme="brand" onClick={() => router.push('/login')}>
            Sign in
          </Button>
        </Text>
      </Box>
  </Flex>
  );
};

export default function RegisterPage() {
  return (
    <SimpleLayout showSidebar={false} requireAuth={false}>
      <RegisterContent />
    </SimpleLayout>
  );
}

