import React, { useState } from 'react';
import {
  Box,
  Button,
  VStack,
  HStack,
  Heading,
  Text,
  Code,
  Divider,
  useColorModeValue,
} from '@chakra-ui/react';
import ErrorAlert, { ErrorDetail } from '../common/ErrorAlert';
import { useEnhancedToast } from '../common/ToastNotification';

const ErrorAlertDemo: React.FC = () => {
  const [showValidationError, setShowValidationError] = useState(false);
  const [showBusinessError, setShowBusinessError] = useState(false);
  const [showNetworkError, setShowNetworkError] = useState(false);
  const enhancedToast = useEnhancedToast();

  const bg = useColorModeValue('gray.50', 'gray.900');
  const cardBg = useColorModeValue('white', 'gray.800');

  // Example validation errors (like insufficient stock)
  const validationErrors: ErrorDetail[] = [
    {
      field: 'items[0].quantity',
      message: 'Insufficient stock available',
      context: {
        available_stock: 5,
        product_name: 'Laptop Dell XPS 13',
        requested_quantity: 10
      },
      suggestion: 'Only 5 units of "Laptop Dell XPS 13" are available. Please reduce the quantity or add more stock first.'
    },
    {
      field: 'customer_id',
      message: 'Customer is required',
      suggestion: 'Please select a customer from the dropdown list.'
    },
    {
      field: 'items[1].unit_price',
      message: 'Unit price cannot be negative',
      suggestion: 'Please enter a positive price for this item.'
    }
  ];

  // Example business rule errors
  const businessErrors: ErrorDetail[] = [
    {
      field: 'customer_id',
      message: 'Order would exceed customer credit limit',
      context: {
        credit_limit: 10000,
        current_outstanding: 8500,
        order_amount: 2000,
        available_credit: 1500
      },
      suggestion: 'Maximum available credit: 1500. Consider requesting a credit limit increase or reducing the order amount.'
    }
  ];

  // Example network errors
  const networkErrors: ErrorDetail[] = [
    {
      message: 'Unable to connect to the server',
      suggestion: 'Check your internet connection and try again.'
    }
  ];

  const handleToastDemo = () => {
    enhancedToast.insufficientStock('Product ABC', 5);
  };

  const handleSuccessDemo = () => {
    enhancedToast.saveSuccess('Sale', false);
  };

  const handleNetworkDemo = () => {
    enhancedToast.networkError();
  };

  return (
    <Box p={6} bg={bg} minH="100vh">
      <VStack spacing={6} maxW="4xl" mx="auto">
        <Box textAlign="center">
          <Heading size="lg" mb={2}>
            🚨 Error Alert System Demo
          </Heading>
          <Text color="gray.600">
            Interactive demo of the new user-friendly error alert system
          </Text>
        </Box>

        {/* Validation Error Example */}
        <Box bg={cardBg} p={6} borderRadius="lg" shadow="md" w="full">
          <VStack spacing={4} align="stretch">
            <Heading size="md">📋 Validation Errors</Heading>
            <Text fontSize="sm" color="gray.600">
              Example: Insufficient stock, missing required fields, invalid values
            </Text>
            
            {showValidationError && (
              <ErrorAlert
                title="Sale Validation Failed"
                message="Please correct the following issues:"
                errors={validationErrors}
                type="error"
                onClose={() => setShowValidationError(false)}
              />
            )}
            
            <HStack>
              <Button
                colorScheme="red"
                variant="outline"
                onClick={() => setShowValidationError(true)}
                isDisabled={showValidationError}
              >
                Show Validation Error
              </Button>
              {showValidationError && (
                <Button
                  size="sm"
                  onClick={() => setShowValidationError(false)}
                >
                  Hide
                </Button>
              )}
            </HStack>
            
            <Code fontSize="xs" p={2} borderRadius="md">
              {`// Backend Response
{
  status: 400,
  error: "validation failed: Insufficient stock available; Customer is required"
}`}
            </Code>
          </VStack>
        </Box>

        {/* Business Rule Error Example */}
        <Box bg={cardBg} p={6} borderRadius="lg" shadow="md" w="full">
          <VStack spacing={4} align="stretch">
            <Heading size="md">⚠️ Business Rule Violations</Heading>
            <Text fontSize="sm" color="gray.600">
              Example: Credit limit exceeded, account restrictions, business logic violations
            </Text>
            
            {showBusinessError && (
              <ErrorAlert
                title="Business Rule Violation"
                message="This transaction violates business rules:"
                errors={businessErrors}
                type="warning"
                onClose={() => setShowBusinessError(false)}
              />
            )}
            
            <HStack>
              <Button
                colorScheme="orange"
                variant="outline"
                onClick={() => setShowBusinessError(true)}
                isDisabled={showBusinessError}
              >
                Show Business Error
              </Button>
              {showBusinessError && (
                <Button
                  size="sm"
                  onClick={() => setShowBusinessError(false)}
                >
                  Hide
                </Button>
              )}
            </HStack>
            
            <Code fontSize="xs" p={2} borderRadius="md">
              {`// Backend Response
{
  status: 422,
  error: "Order would exceed customer credit limit"
}`}
            </Code>
          </VStack>
        </Box>

        {/* Network Error Example */}
        <Box bg={cardBg} p={6} borderRadius="lg" shadow="md" w="full">
          <VStack spacing={4} align="stretch">
            <Heading size="md">🌐 Network/Server Errors</Heading>
            <Text fontSize="sm" color="gray.600">
              Example: Connection failures, server errors, timeout issues
            </Text>
            
            {showNetworkError && (
              <ErrorAlert
                title="Connection Error"
                message="Unable to connect to the server"
                errors={networkErrors}
                type="error"
                onClose={() => setShowNetworkError(false)}
              />
            )}
            
            <HStack>
              <Button
                colorScheme="red"
                variant="outline"
                onClick={() => setShowNetworkError(true)}
                isDisabled={showNetworkError}
              >
                Show Network Error
              </Button>
              {showNetworkError && (
                <Button
                  size="sm"
                  onClick={() => setShowNetworkError(false)}
                >
                  Hide
                </Button>
              )}
            </HStack>
            
            <Code fontSize="xs" p={2} borderRadius="md">
              {`// No Response (Network Error)
{
  code: "NETWORK_ERROR",
  message: "Network request failed"
}`}
            </Code>
          </VStack>
        </Box>

        <Divider />

        {/* Toast Notifications Demo */}
        <Box bg={cardBg} p={6} borderRadius="lg" shadow="md" w="full">
          <VStack spacing={4} align="stretch">
            <Heading size="md">🍞 Toast Notifications</Heading>
            <Text fontSize="sm" color="gray.600">
              For simple feedback messages and non-critical errors
            </Text>
            
            <HStack spacing={4} flexWrap="wrap">
              <Button colorScheme="red" onClick={handleToastDemo}>
                Insufficient Stock Toast
              </Button>
              <Button colorScheme="green" onClick={handleSuccessDemo}>
                Success Toast
              </Button>
              <Button colorScheme="orange" onClick={handleNetworkDemo}>
                Network Error Toast
              </Button>
            </HStack>
            
            <Code fontSize="xs" p={2} borderRadius="md">
              {`// Usage
const enhancedToast = useEnhancedToast();
enhancedToast.insufficientStock('Product Name', 5);
enhancedToast.saveSuccess('Sale', false);
enhancedToast.networkError();`}
            </Code>
          </VStack>
        </Box>

        {/* Integration Info */}
        <Box bg={cardBg} p={6} borderRadius="lg" shadow="md" w="full">
          <VStack spacing={4} align="stretch">
            <Heading size="md">🔧 Integration</Heading>
            <Text fontSize="sm" color="gray.600">
              The new error system automatically parses backend errors and shows appropriate UI:
            </Text>
            <VStack spacing={2} align="stretch" fontSize="sm">
              <Text>• <strong>Validation errors (400)</strong> → ErrorAlert with detailed field-level errors</Text>
              <Text>• <strong>Business rule violations (422)</strong> → ErrorAlert with contextual suggestions</Text>
              <Text>• <strong>Server errors (500+)</strong> → Toast notification with retry suggestion</Text>
              <Text>• <strong>Network errors</strong> → Toast notification with connection advice</Text>
              <Text>• <strong>Success operations</strong> → Toast notification with confirmation</Text>
            </VStack>
            
            <Code fontSize="xs" p={2} borderRadius="md" whiteSpace="pre-wrap">
{`// In SalesForm.tsx
try {
  await salesService.createSale(data);
  enhancedToast.saveSuccess('Sale', false);
} catch (error) {
  const parsedError = ErrorHandler.parseValidationError(error);
  if (parsedError.type === 'validation') {
    setValidationError(parsedError); // Show ErrorAlert
  } else {
    enhancedToast.error('Error', parsedError.message); // Show toast
  }
}`}
            </Code>
          </VStack>
        </Box>
      </VStack>
    </Box>
  );
};

export default ErrorAlertDemo;