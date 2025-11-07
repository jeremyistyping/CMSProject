'use client';

import React from 'react';
import {
  Box,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Button,
  VStack,
  Text,
  Code,
  Divider,
} from '@chakra-ui/react';

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error; resetError: () => void }>;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({
      error,
      errorInfo,
    });
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback;
        return <FallbackComponent error={this.state.error!} resetError={this.handleReset} />;
      }

      return (
        <Box p={6}>
          <Alert status="error" flexDirection="column" alignItems="flex-start">
            <AlertIcon boxSize={6} mr={0} mb={4} />
            <AlertTitle fontSize="lg" mb={2}>
              Oops! Something went wrong
            </AlertTitle>
            <AlertDescription maxWidth="none" mb={4}>
              <Text mb={2}>
                We encountered an unexpected error. Please try refreshing the page or contact support if the problem persists.
              </Text>
              <VStack align="stretch" spacing={3}>
                <Button 
                  colorScheme="red" 
                  variant="outline" 
                  onClick={this.handleReset}
                  size="sm"
                >
                  Try Again
                </Button>
                <Button 
                  variant="ghost" 
                  onClick={() => window.location.reload()}
                  size="sm"
                >
                  Refresh Page
                </Button>
              </VStack>
            </AlertDescription>
            
            {process.env.NODE_ENV === 'development' && this.state.error && (
              <Box mt={4} w="100%">
                <Divider mb={4} />
                <Text fontSize="sm" fontWeight="bold" mb={2}>
                  Developer Information:
                </Text>
                <Code 
                  display="block" 
                  p={4} 
                  borderRadius="md" 
                  bg="gray.50" 
                  overflowX="auto"
                  fontSize="xs"
                  whiteSpace="pre-wrap"
                >
                  {this.state.error.toString()}
                  {this.state.errorInfo?.componentStack}
                </Code>
              </Box>
            )}
          </Alert>
        </Box>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;

// Functional component version for simple error display
export const ErrorFallback: React.FC<{ 
  error: Error; 
  resetError: () => void;
  title?: string;
  description?: string;
}> = ({ error, resetError, title = "Something went wrong", description }) => (
  <Box p={6}>
    <Alert status="error" flexDirection="column" alignItems="center" textAlign="center">
      <AlertIcon boxSize={10} mr={0} mb={4} />
      <AlertTitle fontSize="xl" mb={2}>
        {title}
      </AlertTitle>
      <AlertDescription maxWidth="md" mb={4}>
        {description || "An unexpected error occurred. Please try again or contact support if the problem persists."}
      </AlertDescription>
      <VStack spacing={2}>
        <Button colorScheme="red" onClick={resetError}>
          Try Again
        </Button>
        <Button variant="ghost" size="sm" onClick={() => window.location.reload()}>
          Refresh Page
        </Button>
      </VStack>
    </Alert>
  </Box>
);
