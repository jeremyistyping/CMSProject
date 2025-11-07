'use client'

import React, { useState } from 'react';
import { Box, Button, Text, VStack, Code, Alert, AlertIcon } from '@chakra-ui/react';

const SalesDataTester: React.FC = () => {
  const [testResult, setTestResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const testSalesAPI = async () => {
    setLoading(true);
    setTestResult(null);

    try {
      // First test: Check if we can reach the API
      const response = await fetch('/api/v1/sales/1');
      const data = await response.json();
      
      setTestResult({
        success: response.ok,
        status: response.status,
        data: data,
        analysis: {
          hasCustomer: !!data.customer,
          customerName: data.customer?.name || 'N/A',
          hasSalesPerson: !!data.sales_person,
          salesPersonName: data.sales_person?.name || 'N/A',
          invoiceNumber: data.invoice_number || 'Empty',
          dueDate: data.due_date || 'Empty',
          paymentTerms: data.payment_terms || 'Empty',
          isDueDateZero: data.due_date === '0001-01-01T00:00:00Z'
        }
      });
    } catch (error: any) {
      setTestResult({
        success: false,
        error: error.message,
        analysis: null
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box p={6} border="1px" borderColor="gray.300" borderRadius="md">
      <VStack spacing={4} align="stretch">
        <Text fontSize="lg" fontWeight="bold">Sales API Data Tester</Text>
        
        <Button 
          onClick={testSalesAPI} 
          isLoading={loading}
          colorScheme="blue"
          size="sm"
        >
          Test Sales API (ID: 1)
        </Button>

        {testResult && (
          <Box>
            {testResult.success ? (
              <Alert status="success">
                <AlertIcon />
                API call successful (Status: {testResult.status})
              </Alert>
            ) : (
              <Alert status="error">
                <AlertIcon />
                API call failed: {testResult.error}
              </Alert>
            )}

            {testResult.analysis && (
              <VStack spacing={2} align="stretch" mt={4}>
                <Text fontWeight="bold">Data Analysis:</Text>
                <Text fontSize="sm">Has Customer: {testResult.analysis.hasCustomer ? '✅' : '❌'}</Text>
                <Text fontSize="sm">Customer Name: {testResult.analysis.customerName}</Text>
                <Text fontSize="sm">Has Sales Person: {testResult.analysis.hasSalesPerson ? '✅' : '❌'}</Text>
                <Text fontSize="sm">Sales Person Name: {testResult.analysis.salesPersonName}</Text>
                <Text fontSize="sm">Invoice Number: {testResult.analysis.invoiceNumber}</Text>
                <Text fontSize="sm">Due Date: {testResult.analysis.dueDate} {testResult.analysis.isDueDateZero ? '(Zero date)' : ''}</Text>
                <Text fontSize="sm">Payment Terms: {testResult.analysis.paymentTerms}</Text>
              </VStack>
            )}

            <Box mt={4}>
              <Text fontWeight="bold" mb={2}>Raw API Response:</Text>
              <Code display="block" whiteSpace="pre-wrap" p={3} fontSize="xs" bg="gray.100">
                {JSON.stringify(testResult.data, null, 2)}
              </Code>
            </Box>
          </Box>
        )}
      </VStack>
    </Box>
  );
};

export default SalesDataTester;