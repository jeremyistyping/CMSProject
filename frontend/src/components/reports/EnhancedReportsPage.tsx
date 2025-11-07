'use client';

import React, { useState } from 'react';
import {
  Box,
  VStack,
  SimpleGrid,
  Heading,
  Text,
  Button,
  HStack,
  Icon,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalCloseButton,
  ModalBody,
  Card,
  CardBody,
  CardHeader
} from '@chakra-ui/react';
import { 
  FiFileText, 
  FiBarChart, 
  FiBarChart3, 
  FiTrendingUp, 
  FiDollarSign,
  FiEye
} from 'react-icons/fi';
import PDFReportExample from './PDFReportExample';
import EnhancedTrialBalanceReport from './EnhancedTrialBalanceReport';
import { EnhancedBalanceSheetReport } from './EnhancedBalanceSheetReport';

const EnhancedReportsPage: React.FC = () => {
  const [selectedReportType, setSelectedReportType] = useState<string | null>(null);
  
  const { 
    isOpen: isTrialBalanceOpen, 
    onOpen: onTrialBalanceOpen, 
    onClose: onTrialBalanceClose 
  } = useDisclosure();
  
  const { 
    isOpen: isBalanceSheetOpen, 
    onOpen: onBalanceSheetOpen, 
    onClose: onBalanceSheetClose 
  } = useDisclosure();

  const reportTypes = [
    {
      id: 'trial-balance',
      title: 'Trial Balance Report',
      description: 'Generate comprehensive trial balance with account details and real-time SSOT integration',
      icon: FiBarChart3,
      color: 'purple',
      onOpen: onTrialBalanceOpen
    },
    {
      id: 'balance-sheet',
      title: 'Balance Sheet Report', 
      description: 'Generate detailed balance sheet with assets, liabilities and equity breakdown',
      icon: FiBarChart,
      color: 'blue',
      onOpen: onBalanceSheetOpen
    },
    {
      id: 'profit-loss',
      title: 'Profit & Loss Report',
      description: 'Comprehensive income statement with revenue, expenses and financial metrics',
      icon: FiTrendingUp,
      color: 'green',
      onOpen: () => console.log('P&L Report - Coming Soon')
    },
    {
      id: 'cash-flow',
      title: 'Cash Flow Report',
      description: 'Track cash movements across operating, investing and financing activities',
      icon: FiDollarSign,
      color: 'orange',
      onOpen: () => console.log('Cash Flow Report - Coming Soon')
    }
  ];

  return (
    <Box p={6}>
      <VStack spacing={6} align="stretch">
        <Box>
          <Heading as="h1" size="xl" mb={2}>
            Financial Reports & Documents
          </Heading>
          <Text color="gray.600" fontSize="lg">
            Generate professional financial reports with real-time data and export capabilities
          </Text>
        </Box>

        {/* Financial Reports Grid */}
        <SimpleGrid columns={[1, 2, 2]} spacing={6}>
          {reportTypes.map((report) => (
            <Card key={report.id} cursor="pointer" transition="all 0.2s" _hover={{ transform: 'translateY(-2px)', boxShadow: 'lg' }}>
              <CardHeader>
                <HStack>
                  <Box p={3} bg={`${report.color}.100`} borderRadius="md">
                    <Icon as={report.icon} color={`${report.color}.600`} boxSize={6} />
                  </Box>
                  <VStack align="start" spacing={1} flex={1}>
                    <Text fontSize="lg" fontWeight="bold" color="gray.700">
                      {report.title}
                    </Text>
                    <Text fontSize="sm" color="gray.600">
                      {report.description}
                    </Text>
                  </VStack>
                </HStack>
              </CardHeader>
              <CardBody pt={0}>
                <Button
                  colorScheme={report.color}
                  leftIcon={<FiEye />}
                  onClick={report.onOpen}
                  size="sm"
                  width="full"
                >
                  Generate Report
                </Button>
              </CardBody>
            </Card>
          ))}
        </SimpleGrid>

        {/* PDF Generator Section */}
        <Box>
          <Heading as="h2" size="lg" mb={4}>
            Document Generator
          </Heading>
          <SimpleGrid columns={[1, 1, 1]} spacing={6}>
            <PDFReportExample />
          </SimpleGrid>
        </Box>

        <Box bg="blue.50" p={4} borderRadius="md" border="1px" borderColor="blue.200">
          <Text fontSize="sm" color="blue.800">
            <strong>Integration Guide:</strong> All reports are integrated with SSOT (Single Source of Truth) data system. 
            Reports automatically use your company settings (logo, name, address, etc.) from the Settings page and 
            provide real-time data with advanced export capabilities (CSV, PDF).
          </Text>
        </Box>
      </VStack>

      {/* Trial Balance Modal */}
      <Modal isOpen={isTrialBalanceOpen} onClose={onTrialBalanceClose} size="6xl">
        <ModalOverlay />
        <ModalContent maxW="90vw">
          <ModalHeader>Trial Balance Report</ModalHeader>
          <ModalCloseButton />
          <ModalBody p={0}>
            <EnhancedTrialBalanceReport onClose={onTrialBalanceClose} />
          </ModalBody>
        </ModalContent>
      </Modal>

      {/* Balance Sheet Modal */}
      <Modal isOpen={isBalanceSheetOpen} onClose={onBalanceSheetClose} size="6xl">
        <ModalOverlay />
        <ModalContent maxW="90vw">
          <ModalHeader>Balance Sheet Report</ModalHeader>
          <ModalCloseButton />
          <ModalBody p={0}>
            <EnhancedBalanceSheetReport onClose={onBalanceSheetClose} />
          </ModalBody>
        </ModalContent>
      </Modal>
    </Box>
  );
};

export default EnhancedReportsPage;