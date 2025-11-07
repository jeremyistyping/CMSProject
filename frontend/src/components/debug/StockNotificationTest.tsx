'use client';

import React, { useState } from 'react';
import {
  Box,
  Button,
  Text,
  VStack,
  HStack,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Input,
  FormControl,
  FormLabel,
  Badge,
} from '@chakra-ui/react';
import axios from 'axios';
import { API_BASE_URL } from '@/config/api';

interface Product {
  id: number;
  name: string;
  code: string;
  stock: number;
  min_stock: number;
  max_stock: number;
  reorder_level: number;
}

const StockNotificationTest: React.FC = () => {
  const [testProduct, setTestProduct] = useState<Product>({
    id: 1,
    name: 'Aqua Test',
    code: 'AQUA001',
    stock: 10,
    min_stock: 5,
    max_stock: 20,
    reorder_level: 5,
  });
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string>('');

  const simulateStockUpdate = async (newStock: number) => {
    setLoading(true);
    setResult('');
    
    try {
      const token = localStorage.getItem('token');
      
      // Update stock via API (example)
      // This should trigger notification if stock <= min_stock
      const response = await axios.post(
        `${API_BASE_URL}/products/${testProduct.id}/adjust-stock`,
        {
          new_stock: newStock,
          reason: 'Test stock notification',
          type: 'adjustment'
        },
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      );

      setTestProduct(prev => ({ ...prev, stock: newStock }));
      
      if (newStock === 0) {
        setResult(`✅ Stock = 0: Should trigger OUT_OF_STOCK notification`);
      } else if (newStock <= testProduct.min_stock) {
        setResult(`⚠️ Stock (${newStock}) <= Min Stock (${testProduct.min_stock}): Should trigger LOW_STOCK notification`);
      } else {
        setResult(`✅ Stock (${newStock}) > Min Stock (${testProduct.min_stock}): No notification needed`);
      }
      
    } catch (error: any) {
      console.error('Stock update error:', error);
      setResult(`❌ Error: ${error.response?.data?.error || error.message}`);
    } finally {
      setLoading(false);
    }
  };

  const getStockStatus = () => {
    if (testProduct.stock === 0) {
      return { status: 'Out of Stock', color: 'red' };
    } else if (testProduct.stock <= testProduct.min_stock) {
      return { status: 'Low Stock', color: 'orange' };
    } else if (testProduct.stock <= testProduct.reorder_level) {
      return { status: 'Reorder Level', color: 'yellow' };
    }
    return { status: 'Normal', color: 'green' };
  };

  const stockStatus = getStockStatus();

  return (
    <Box p={6} maxW="600px" mx="auto">
      <VStack spacing={6} align="stretch">
        <Box>
          <Text fontSize="xl" fontWeight="bold" mb={4}>
            🧪 Stock Notification Logic Test
          </Text>
          
          <Alert status="info" mb={4}>
            <AlertIcon />
            <Box>
              <AlertTitle>Test Logic:</AlertTitle>
              <AlertDescription>
                • Stock = 0 → OUT_OF_STOCK notification<br/>
                • Stock ≤ Min Stock → LOW_STOCK notification<br/>
                • Stock &gt; Min Stock → No notification
              </AlertDescription>
            </Box>
          </Alert>
        </Box>

        {/* Current Product Status */}
        <Box p={4} border="1px" borderColor="gray.200" borderRadius="md">
          <Text fontWeight="bold" mb={2}>Current Product: {testProduct.name}</Text>
          <VStack align="start" spacing={2}>
            <HStack>
              <Text>Current Stock:</Text>
              <Badge colorScheme={stockStatus.color}>{testProduct.stock}</Badge>
              <Text>({stockStatus.status})</Text>
            </HStack>
            <HStack>
              <Text>Min Stock:</Text>
              <Badge>{testProduct.min_stock}</Badge>
            </HStack>
            <HStack>
              <Text>Max Stock:</Text>
              <Badge>{testProduct.max_stock}</Badge>
            </HStack>
            <HStack>
              <Text>Reorder Level:</Text>
              <Badge>{testProduct.reorder_level}</Badge>
            </HStack>
          </VStack>
        </Box>

        {/* Test Scenarios */}
        <Box>
          <Text fontWeight="bold" mb={3}>Test Scenarios:</Text>
          <VStack spacing={3}>
            <HStack spacing={3} w="full">
              <Button
                size="sm"
                colorScheme="red"
                onClick={() => simulateStockUpdate(0)}
                isLoading={loading}
              >
                Set Stock = 0
              </Button>
              <Text fontSize="sm">Should trigger OUT_OF_STOCK</Text>
            </HStack>
            
            <HStack spacing={3} w="full">
              <Button
                size="sm"
                colorScheme="orange"
                onClick={() => simulateStockUpdate(3)}
                isLoading={loading}
              >
                Set Stock = 3
              </Button>
              <Text fontSize="sm">Should trigger LOW_STOCK (3 &lt; 5)</Text>
            </HStack>
            
            <HStack spacing={3} w="full">
              <Button
                size="sm"
                colorScheme="yellow"
                onClick={() => simulateStockUpdate(5)}
                isLoading={loading}
              >
                Set Stock = 5
              </Button>
              <Text fontSize="sm">Should trigger LOW_STOCK (5 = 5)</Text>
            </HStack>
            
            <HStack spacing={3} w="full">
              <Button
                size="sm"
                colorScheme="green"
                onClick={() => simulateStockUpdate(10)}
                isLoading={loading}
              >
                Set Stock = 10
              </Button>
              <Text fontSize="sm">No notification (10 &gt; 5)</Text>
            </HStack>
          </VStack>
        </Box>

        {/* Custom Stock Input */}
        <Box>
          <FormControl>
            <FormLabel>Custom Stock Value:</FormLabel>
            <HStack>
              <Input
                type="number"
                placeholder="Enter stock value"
                onChange={(e) => {
                  const value = parseInt(e.target.value);
                  if (!isNaN(value)) {
                    simulateStockUpdate(value);
                  }
                }}
              />
            </HStack>
          </FormControl>
        </Box>

        {/* Result */}
        {result && (
          <Alert 
            status={result.includes('❌') ? 'error' : result.includes('⚠️') ? 'warning' : 'success'}
          >
            <AlertIcon />
            <Text>{result}</Text>
          </Alert>
        )}

        {/* Expected Behavior */}
        <Box p={4} bg="gray.50" borderRadius="md">
          <Text fontWeight="bold" mb={2}>Expected Notification Behavior:</Text>
          <VStack align="start" spacing={1} fontSize="sm">
            <Text>✅ Current Stock (10) &gt; Min Stock (5) = No notification</Text>
            <Text>⚠️ If Stock ≤ 5 = LOW_STOCK notification</Text>
            <Text>🚨 If Stock = 0 = OUT_OF_STOCK notification</Text>
          </VStack>
        </Box>
      </VStack>
    </Box>
  );
};

export default StockNotificationTest;
