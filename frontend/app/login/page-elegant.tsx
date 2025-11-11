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
  VStack,
  HStack,
  InputGroup,
  InputRightElement,
  IconButton,
  Alert,
  AlertIcon,
  Container,
  useColorMode,
  Image,
  Divider,
  Link
} from '@chakra-ui/react';
import { FiLogIn, FiEye, FiEyeOff, FiMail, FiLock, FiArrowRight } from 'react-icons/fi';

const LoginContent = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  
  const { login, isAuthenticated } = useAuth();
  const router = useRouter();
  const toast = useToast();
  const { colorMode } = useColorMode();
  
  useEffect(() => {
    if (isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    if (!email || !password) {
      setError('Please enter both email and password');
      return;
    }
    
    try {
      setIsSubmitting(true);
      await login(email, password);
      
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem('lastLogoutTime');
      }
      
      toast({
        title: "Login Successful",
        description: "Welcome back!",
        status: "success",
        duration: 3000,
        isClosable: true,
        position: 'top-right',
      });
      
      router.push('/dashboard');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Invalid email or password';
      setError(errorMessage);
      
      toast({
        title: "Login Failed",
        description: errorMessage,
        status: "error",
        duration: 5000,
        isClosable: true,
        position: 'top-right',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Box 
      minH="100vh"
      position="relative"
      overflow="hidden"
      bg="linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
    >
      {/* Animated background elements */}
      <Box
        position="absolute"
        top="0"
        left="0"
        right="0"
        bottom="0"
        opacity="0.1"
        bgImage="repeating-linear-gradient(45deg, rgba(255,255,255,0.03) 0px, rgba(255,255,255,0.03) 2px, transparent 2px, transparent 4px)"
      />
      
      <Container maxW="6xl" h="100vh" display="flex" alignItems="center" py={8}>
        <Flex 
          width="100%" 
          direction={{ base: 'column', md: 'row' }}
          align="stretch"
          gap={0}
          borderRadius="3xl"
          overflow="hidden"
          boxShadow="2xl"
          bg="white"
          _dark={{ bg: 'gray.900' }}
        >
          {/* Left Panel - Branding */}
          <Box 
            flex="1"
            bg="linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
            p={{ base: 8, md: 12 }}
            display="flex"
            flexDirection="column"
            justifyContent="center"
            position="relative"
            overflow="hidden"
          >
            {/* Decorative circles */}
            <Box
              position="absolute"
              top="-50px"
              right="-50px"
              width="200px"
              height="200px"
              borderRadius="full"
              bg="whiteAlpha.200"
            />
            <Box
              position="absolute"
              bottom="-30px"
              left="-30px"
              width="150px"
              height="150px"
              borderRadius="full"
              bg="whiteAlpha.100"
            />
            
            <VStack align="flex-start" spacing={6} position="relative" zIndex={1}>
              {/* Logo */}
              <Box
                w={16}
                h={16}
                bg="whiteAlpha.300"
                backdropFilter="blur(10px)"
                borderRadius="2xl"
                display="flex"
                alignItems="center"
                justifyContent="center"
                color="white"
                fontSize="2xl"
                fontWeight="bold"
                boxShadow="lg"
              >
                U
              </Box>
              
              <VStack align="flex-start" spacing={3}>
                <Heading 
                  as="h1" 
                  size="2xl" 
                  color="white"
                  fontWeight="extrabold"
                  letterSpacing="tight"
                >
                  Welcome to Unipro
                </Heading>
                <Text 
                  fontSize="xl" 
                  color="whiteAlpha.900"
                  fontWeight="medium"
                  lineHeight="tall"
                >
                  Cost Control Management System
                </Text>
                <Text 
                  fontSize="md" 
                  color="whiteAlpha.800"
                  lineHeight="relaxed"
                  maxW="md"
                >
                  Streamline your project cost management with precision and efficiency. 
                  Track, analyze, and optimize your expenses in real-time.
                </Text>
              </VStack>
              
              {/* Features list */}
              <VStack align="flex-start" spacing={3} pt={4}>
                {[
                  'Real-time budget tracking',
                  'Advanced cost analytics',
                  'Multi-project management',
                  'Automated approval workflows'
                ].map((feature, idx) => (
                  <HStack key={idx} spacing={3}>
                    <Box
                      w={6}
                      h={6}
                      borderRadius="full"
                      bg="whiteAlpha.300"
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      color="white"
                      fontSize="sm"
                    >
                      ✓
                    </Box>
                    <Text color="white" fontSize="sm" fontWeight="medium">
                      {feature}
                    </Text>
                  </HStack>
                ))}
              </VStack>
            </VStack>
          </Box>
          
          {/* Right Panel - Login Form */}
          <Box 
            flex="1"
            p={{ base: 8, md: 12 }}
            display="flex"
            flexDirection="column"
            justifyContent="center"
            bg="white"
            _dark={{ bg: 'gray.900' }}
          >
            <VStack spacing={8} w="full" maxW="md" mx="auto">
              {/* Header */}
              <VStack spacing={2} w="full">
                <Heading 
                  as="h2" 
                  size="xl" 
                  fontWeight="bold"
                  bgGradient="linear(to-r, purple.600, pink.500)"
                  bgClip="text"
                >
                  Sign In
                </Heading>
                <Text color="gray.600" _dark={{ color: 'gray.400' }}>
                  Enter your credentials to continue
                </Text>
              </VStack>
              
              {/* Error Alert */}
              {error && (
                <Alert 
                  status="error" 
                  borderRadius="xl"
                  boxShadow="sm"
                >
                  <AlertIcon />
                  <Text fontSize="sm">{error}</Text>
                </Alert>
              )}
              
              {/* Login Form */}
              <form onSubmit={handleSubmit} style={{ width: '100%' }}>
                <VStack spacing={5} w="full">
                  <FormControl isRequired>
                    <FormLabel 
                      fontSize="sm"
                      fontWeight="semibold"
                      color="gray.700"
                      _dark={{ color: 'gray.300' }}
                    >
                      Email Address
                    </FormLabel>
                    <InputGroup size="lg">
                      <Input
                        type="email"
                        placeholder="you@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={isSubmitting}
                        borderRadius="xl"
                        borderWidth="2px"
                        borderColor="gray.200"
                        _dark={{ borderColor: 'gray.700', bg: 'gray.800' }}
                        _hover={{
                          borderColor: 'purple.400'
                        }}
                        _focus={{
                          borderColor: 'purple.500',
                          boxShadow: '0 0 0 1px var(--chakra-colors-purple-500)'
                        }}
                      />
                      <InputRightElement>
                        <FiMail color="gray" />
                      </InputRightElement>
                    </InputGroup>
                  </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel 
                      fontSize="sm"
                      fontWeight="semibold"
                      color="gray.700"
                      _dark={{ color: 'gray.300' }}
                    >
                      Password
                    </FormLabel>
                    <InputGroup size="lg">
                      <Input
                        type={showPassword ? 'text' : 'password'}
                        placeholder="••••••••"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        disabled={isSubmitting}
                        borderRadius="xl"
                        borderWidth="2px"
                        borderColor="gray.200"
                        _dark={{ borderColor: 'gray.700', bg: 'gray.800' }}
                        _hover={{
                          borderColor: 'purple.400'
                        }}
                        _focus={{
                          borderColor: 'purple.500',
                          boxShadow: '0 0 0 1px var(--chakra-colors-purple-500)'
                        }}
                      />
                      <InputRightElement>
                        <IconButton
                          aria-label={showPassword ? 'Hide password' : 'Show password'}
                          icon={showPassword ? <FiEyeOff /> : <FiEye />}
                          onClick={() => setShowPassword(!showPassword)}
                          variant="ghost"
                          size="sm"
                        />
                      </InputRightElement>
                    </InputGroup>
                  </FormControl>
                  
                  <HStack justify="flex-end" w="full">
                    <Link 
                      fontSize="sm" 
                      color="purple.600"
                      fontWeight="medium"
                      _hover={{ textDecoration: 'underline' }}
                    >
                      Forgot password?
                    </Link>
                  </HStack>
                  
                  <Button
                    type="submit"
                    size="lg"
                    w="full"
                    bgGradient="linear(to-r, purple.600, pink.500)"
                    color="white"
                    isLoading={isSubmitting}
                    loadingText="Signing in..."
                    borderRadius="xl"
                    fontWeight="bold"
                    rightIcon={<FiArrowRight />}
                    _hover={{
                      bgGradient: 'linear(to-r, purple.700, pink.600)',
                      transform: 'translateY(-2px)',
                      boxShadow: 'lg'
                    }}
                    _active={{
                      transform: 'translateY(0)'
                    }}
                    transition="all 0.2s"
                  >
                    Sign In
                  </Button>
                </VStack>
              </form>
              
              {/* Divider */}
              <HStack w="full">
                <Divider />
                <Text fontSize="sm" color="gray.500" whiteSpace="nowrap" px={2}>
                  or
                </Text>
                <Divider />
              </HStack>
              
              {/* Sign up link */}
              <Text fontSize="sm" color="gray.600" _dark={{ color: 'gray.400' }}>
                Don't have an account?{' '}
                <Link 
                  color="purple.600" 
                  fontWeight="semibold"
                  onClick={() => router.push('/register')}
                  _hover={{ textDecoration: 'underline' }}
                >
                  Sign up for free
                </Link>
              </Text>
            </VStack>
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};

export default function LoginPage() {
  return (
    <SimpleLayout showSidebar={false} requireAuth={false}>
      <LoginContent />
    </SimpleLayout>
  );
}

