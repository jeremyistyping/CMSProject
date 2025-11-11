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
  InputGroup,
  InputRightElement,
  IconButton,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  VStack,
  HStack,
  Divider,
  useColorMode,
  useColorModeValue,
  Card,
  CardBody,
  Image,
  Container,
  SimpleGrid,
  ScaleFade,
  Badge,
  Center
} from '@chakra-ui/react';
import { FiLogIn, FiEye, FiEyeOff, FiShield, FiLock, FiMail, FiUsers, FiTrendingUp, FiPieChart } from 'react-icons/fi';

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
  
  // Black themed design - Fixed colors (no light mode)
  const bgGradient = 'linear(to-br, gray.900, black, gray.900)';
  const cardBg = 'gray.900';
  const headingColor = 'white';
  const textColor = 'gray.300';
  const borderColor = 'gray.700';
  const inputBg = 'gray.800';
  const featureCardBg = 'gray.800';
  const iconColor = 'green.400';
  const accentColor = 'green.500';
  const inputHoverBorderColor = 'green.500';
  const inputPlaceholderColor = 'gray.500';
  const buttonHoverBg = 'green.600';
  const decorativeElementBg1 = 'rgba(16, 185, 129, 0.1)'; // green with opacity
  const decorativeElementBg2 = 'rgba(139, 92, 246, 0.1)'; // purple with opacity
  const passwordToggleHoverBg = 'gray.700';
  // Error alert colors
  const errorAlertBg = 'red.900';
  const errorAlertBorderColor = 'red.700';
  
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
      
      // Clear logout timestamp on successful login
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem('lastLogoutTime');
      }
      
      toast({
        render: () => (
          <Box
            color={colorMode === 'dark' ? 'gray.800' : 'white'}
            p={4}
            bg={colorMode === 'dark' ? 'white' : 'green.500'}
            borderRadius='md'
            boxShadow='lg'
          >
            <HStack spacing={3}>
              <Box as='span' fontSize='xl' color={colorMode === 'dark' ? 'green.500' : 'white'}>✓</Box>
              <VStack align='start' spacing={0}>
                <Text fontWeight='bold' fontSize='sm'>Login Successful</Text>
                <Text fontSize='sm'>Welcome back!</Text>
              </VStack>
            </HStack>
          </Box>
        ),
        duration: 3000,
        isClosable: true,
        position: 'bottom',
      });
      
      router.push('/dashboard');
    } catch (err) {
      // Use the actual error message from the caught exception
      const errorMessage = err instanceof Error ? err.message : 'Invalid email or password';
      setError(errorMessage);
      
      toast({
        render: () => (
          <Box
            color={colorMode === 'dark' ? 'gray.800' : 'white'}
            p={4}
            bg={colorMode === 'dark' ? 'white' : 'red.500'}
            borderRadius='md'
            boxShadow='lg'
            borderLeft={colorMode === 'dark' ? '4px solid' : 'none'}
            borderLeftColor={colorMode === 'dark' ? 'red.500' : 'transparent'}
          >
            <HStack spacing={3}>
              <Box as='span' fontSize='xl' color={colorMode === 'dark' ? 'red.500' : 'white'}>✕</Box>
              <VStack align='start' spacing={0}>
                <Text fontWeight='bold' fontSize='sm'>Login Failed</Text>
                <Text fontSize='sm'>{errorMessage}</Text>
              </VStack>
            </HStack>
          </Box>
        ),
        duration: 5000,
        isClosable: true,
        position: 'bottom',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Box 
      minH="100vh"
      bgGradient={bgGradient}
      position="relative"
      overflow="hidden"
    >
      {/* Background decorative elements */}
      <Box
        position="absolute"
        top="-50px"
        right="-50px"
        w="200px"
        h="200px"
        bg={decorativeElementBg1}
        borderRadius="full"
        opacity={0.3}
      />
      <Box
        position="absolute"
        bottom="-100px"
        left="-100px"
        w="300px"
        h="300px"
        bg={decorativeElementBg2}
        borderRadius="full"
        opacity={0.2}
      />
      
      <Container maxW="7xl" py={8}>
        <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={8} alignItems="center" minH="80vh">
          {/* Left side - Branding and Features */}
          <Box order={{ base: 1, lg: 1 }}>
          <ScaleFade initialScale={0.9} in={true}>
            <VStack spacing={8} align="stretch">
              {/* Company Branding */}
              <Box textAlign={{ base: 'center', lg: 'left' }}>
                <HStack spacing={4} justify={{ base: 'center', lg: 'flex-start' }} mb={6}>
                  <Box
                    w={12}
                    h={12}
                    bg={accentColor}
                    borderRadius="xl"
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    color="white"
                    boxShadow="lg"
                  >
                    <FiPieChart size={24} />
                  </Box>
                  <VStack spacing={0} align={{ base: 'center', lg: 'flex-start' }}>
                    <Heading 
                      size="xl" 
                      color={headingColor}
                      fontWeight="bold"
                      letterSpacing="tight"
                    >
                      Unipro
                    </Heading>
                    <Text color={textColor} fontSize="sm" fontWeight="medium">
                      Cost Control Management System
                    </Text>
                  </VStack>
                </HStack>
                
                <Text 
                  fontSize="lg" 
                  color={textColor} 
                  mb={8}
                  lineHeight="tall"
                >
                  Streamline your project cost management with precision. 
                  Our comprehensive cost control system provides powerful 
                  tools for tracking, analyzing, and optimizing project expenses.
                </Text>
              </Box>
              
              {/* Feature highlights */}
              <VStack spacing={4} align="stretch">
                <Heading size="md" color={headingColor} mb={2}>
                  Key Features
                </Heading>
                
                <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                  <Card 
                    bg={featureCardBg} 
                    shadow="sm" 
                    borderWidth="1px" 
                    borderColor={borderColor}
                    _hover={{ transform: 'translateY(-2px)', shadow: 'md' }}
                    transition="all 0.2s"
                  >
                    <CardBody p={4}>
                      <HStack spacing={3}>
                        <Center 
                          w={10} 
                          h={10} 
                          bg="gray.700"
                          borderRadius="lg"
                        >
                          <FiShield color={iconColor} size={20} />
                        </Center>
                        <VStack align="start" spacing={0}>
                          <Text fontWeight="semibold" color={headingColor} fontSize="sm">
                            Budget Control
                          </Text>
                          <Text color={textColor} fontSize="xs">
                            Real-time budget tracking
                          </Text>
                        </VStack>
                      </HStack>
                    </CardBody>
                  </Card>
                  
                  <Card 
                    bg={featureCardBg} 
                    shadow="sm" 
                    borderWidth="1px" 
                    borderColor={borderColor}
                    _hover={{ transform: 'translateY(-2px)', shadow: 'md' }}
                    transition="all 0.2s"
                  >
                    <CardBody p={4}>
                      <HStack spacing={3}>
                        <Center 
                          w={10} 
                          h={10} 
                          bg="gray.700"
                          borderRadius="lg"
                        >
                          <FiTrendingUp color={iconColor} size={20} />
                        </Center>
                        <VStack align="start" spacing={0}>
                          <Text fontWeight="semibold" color={headingColor} fontSize="sm">
                            Cost Analytics
                          </Text>
                          <Text color={textColor} fontSize="xs">
                            Track project expenses
                          </Text>
                        </VStack>
                      </HStack>
                    </CardBody>
                  </Card>
                  
                  <Card 
                    bg={featureCardBg} 
                    shadow="sm" 
                    borderWidth="1px" 
                    borderColor={borderColor}
                    _hover={{ transform: 'translateY(-2px)', shadow: 'md' }}
                    transition="all 0.2s"
                  >
                    <CardBody p={4}>
                      <HStack spacing={3}>
                        <Center 
                          w={10} 
                          h={10} 
                          bg="gray.700"
                          borderRadius="lg"
                        >
                          <FiUsers color={iconColor} size={20} />
                        </Center>
                        <VStack align="start" spacing={0}>
                          <Text fontWeight="semibold" color={headingColor} fontSize="sm">
                            Project Management
                          </Text>
                          <Text color={textColor} fontSize="xs">
                            Multi-project tracking
                          </Text>
                        </VStack>
                      </HStack>
                    </CardBody>
                  </Card>
                  
                  <Card 
                    bg={featureCardBg} 
                    shadow="sm" 
                    borderWidth="1px" 
                    borderColor={borderColor}
                    _hover={{ transform: 'translateY(-2px)', shadow: 'md' }}
                    transition="all 0.2s"
                  >
                    <CardBody p={4}>
                      <HStack spacing={3}>
                        <Center 
                          w={10} 
                          h={10} 
                          bg="gray.700"
                          borderRadius="lg"
                        >
                          <FiLock color={iconColor} size={20} />
                        </Center>
                        <VStack align="start" spacing={0}>
                          <Text fontWeight="semibold" color={headingColor} fontSize="sm">
                            Cost Approval
                          </Text>
                          <Text color={textColor} fontSize="xs">
                            Automated cost control
                          </Text>
                        </VStack>
                      </HStack>
                    </CardBody>
                  </Card>
                </SimpleGrid>
                
                {/* Trust badges */}
                <Box pt={6}>
                  <HStack spacing={4} justify={{ base: 'center', lg: 'flex-start' }} wrap="wrap">
                    <Badge 
                      colorScheme="green" 
                      px={3} 
                      py={1} 
                      borderRadius="full"
                      fontSize="xs"
                      fontWeight="semibold"
                      bg="green.900"
                      color="green.200"
                    >
                      ✓ Secure
                    </Badge>
                    <Badge 
                      colorScheme="green" 
                      px={3} 
                      py={1} 
                      borderRadius="full"
                      fontSize="xs"
                      fontWeight="semibold"
                      bg="green.900"
                      color="green.200"
                    >
                      ✓ Fast
                    </Badge>
                    <Badge 
                      colorScheme="green" 
                      px={3} 
                      py={1} 
                      borderRadius="full"
                      fontSize="xs"
                      fontWeight="semibold"
                      bg="green.900"
                      color="green.200"
                    >
                      ✓ Reliable
                    </Badge>
                  </HStack>
                </Box>
              </VStack>
            </VStack>
          </ScaleFade>
          </Box>
          
          {/* Right side - Login Form */}
          <Flex align="center" justify="center" py={8} order={{ base: 2, lg: 2 }}>
            <ScaleFade initialScale={0.9} in={true}>
              <Card 
                maxW="md"
                w="full"
                bg={cardBg}
                shadow="2xl"
                borderWidth="1px"
                borderColor={borderColor}
                borderRadius="2xl"
                overflow="hidden"
              >
                <CardBody p={8}>
                  {/* Header */}
                  <VStack spacing={6} mb={8}>
                    <Box
                      w={20}
                      h={20}
                      bg={accentColor}
                      borderRadius="2xl"
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      color="white"
                      boxShadow="lg"
                      position="relative"
                      _before={{
                        content: '""',
                        position: 'absolute',
                        inset: '-2px',
                        bg: 'linear-gradient(45deg, transparent 30%, rgba(255,255,255,0.3) 50%, transparent 70%)',
                        borderRadius: '2xl',
                        opacity: 0,
                        transition: 'opacity 0.3s',
                        zIndex: -1
                      }}
                      _hover={{
                        transform: 'scale(1.05)',
                        transition: 'all 0.2s ease'
                      }}
                    >
                      <FiLogIn size="2.5rem" />
                    </Box>
                    
                    <VStack spacing={2}>
                      <Heading 
                        as="h1" 
                        size="xl" 
                        textAlign="center" 
                        color={headingColor}
                        fontWeight="bold"
                        letterSpacing="tight"
                      >
                        Welcome Back
                      </Heading>
                      <Text 
                        color={textColor} 
                        textAlign="center" 
                        fontSize="md"
                        fontWeight="medium"
                      >
                        Sign in to access your cost control dashboard
                      </Text>
                    </VStack>
                  </VStack>
                  
                  {/* Error Alert */}
                  {error && (
                    <Alert 
                      status="error" 
                      mb={6} 
                      borderRadius="xl"
                      bg={errorAlertBg}
                      borderWidth="1px"
                      borderColor={errorAlertBorderColor}
                    >
                      <AlertIcon />
                      <Box>
                        <AlertTitle fontSize="sm">Login Failed!</AlertTitle>
                        <AlertDescription fontSize="sm">{error}</AlertDescription>
                      </Box>
                    </Alert>
                  )}
                  
                  {/* Login Form */}
                  <form onSubmit={handleSubmit}>
                    <VStack spacing={6}>
                      <FormControl id="email" isRequired>
                        <FormLabel 
                          color={headingColor}
                          fontSize="sm"
                          fontWeight="semibold"
                          mb={2}
                        >
                          Email address *
                        </FormLabel>
                        <InputGroup>
                          <Input
                            type="email"
                            placeholder="your-email@example.com"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            disabled={isSubmitting}
                            bg={inputBg}
                            color="white"
                            borderWidth="2px"
                            borderColor={borderColor}
                            borderRadius="xl"
                            py={6}
                            fontSize="md"
                            _hover={{
                              borderColor: inputHoverBorderColor
                            }}
                            _focus={{
                              borderColor: accentColor,
                              boxShadow: `0 0 0 1px ${accentColor}`,
                              bg: inputBg,
                              color: 'white'
                            }}
                            _placeholder={{
                              color: inputPlaceholderColor
                            }}
                          />
                          <InputRightElement top={2} right={2}>
                            <Center 
                              w={8} 
                              h={8} 
                              bg={passwordToggleHoverBg}
                              borderRadius="lg"
                            >
                              <FiMail color={iconColor} size={16} />
                            </Center>
                          </InputRightElement>
                        </InputGroup>
                      </FormControl>
                      
                      <FormControl id="password" isRequired>
                        <FormLabel 
                          color={headingColor}
                          fontSize="sm"
                          fontWeight="semibold"
                          mb={2}
                        >
                          Password *
                        </FormLabel>
                        <InputGroup>
                          <Input
                            type={showPassword ? 'text' : 'password'}
                            placeholder="Enter your password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            disabled={isSubmitting}
                            bg={inputBg}
                            color="white"
                            borderWidth="2px"
                            borderColor={borderColor}
                            borderRadius="xl"
                            py={6}
                            fontSize="md"
                            _hover={{
                              borderColor: inputHoverBorderColor
                            }}
                            _focus={{
                              borderColor: accentColor,
                              boxShadow: `0 0 0 1px ${accentColor}`,
                              bg: inputBg,
                              color: 'white'
                            }}
                            _placeholder={{
                              color: inputPlaceholderColor
                            }}
                          />
                          <InputRightElement top={2} right={2}>
                            <IconButton
                              variant="ghost"
                              size="sm"
                              aria-label={showPassword ? 'Hide password' : 'Show password'}
                              icon={showPassword ? <FiEyeOff /> : <FiEye />}
                              onClick={() => setShowPassword(!showPassword)}
                              borderRadius="lg"
                              _hover={{
                                bg: passwordToggleHoverBg
                              }}
                            />
                          </InputRightElement>
                        </InputGroup>
                      </FormControl>
                      
                      <Button
                        type="submit"
                        bg={accentColor}
                        color="white"
                        isLoading={isSubmitting}
                        loadingText="Signing in..."
                        width="full"
                        size="lg"
                        borderRadius="xl"
                        fontWeight="bold"
                        fontSize="md"
                        py={6}
                        leftIcon={<FiLogIn />}
                        _hover={{
                          bg: buttonHoverBg,
                          transform: 'translateY(-1px)',
                          boxShadow: 'lg'
                        }}
                        _active={{
                          transform: 'translateY(0)'
                        }}
                        transition="all 0.2s ease"
                        boxShadow="md"
                      >
                        Sign In
                      </Button>
                    </VStack>
                  </form>
                  
                  {/* Divider */}
                  <HStack my={8}>
                    <Divider borderColor={borderColor} />
                    <Text 
                      px={3} 
                      color={textColor} 
                      fontSize="sm" 
                      fontWeight="medium"
                      whiteSpace="nowrap"
                    >
                      Need help?
                    </Text>
                    <Divider borderColor={borderColor} />
                  </HStack>
                  
                  {/* Support Text */}
                  <VStack spacing={3}>
                    <Text 
                      textAlign="center" 
                      color={textColor} 
                      fontSize="sm"
                    >
                      Don't have an account?{' '}
                      <Button 
                        variant="link" 
                        color={accentColor} 
                        fontWeight="semibold"
                        fontSize="sm"
                        p={0}
                        h="auto"
                        minW="auto"
                        onClick={() => router.push('/register')}
                        _hover={{
                          textDecoration: 'underline'
                        }}
                      >
                        Sign up here
                      </Button>
                    </Text>
                    
                    <Text 
                      textAlign="center" 
                      color={textColor} 
                      fontSize="xs"
                      maxW="sm"
                      lineHeight="relaxed"
                    >
                      By signing in, you agree to our terms of service and privacy policy.
                    </Text>
                  </VStack>
                </CardBody>
              </Card>
            </ScaleFade>
          </Flex>
        </SimpleGrid>
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
