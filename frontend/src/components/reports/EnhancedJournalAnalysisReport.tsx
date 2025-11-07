import React, { useState, useEffect } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Heading,
  Card,
  CardBody,
  CardHeader,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Progress,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Flex,
  Icon,
  Tooltip,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
  Circle,
  useColorModeValue,
  Spinner,
} from '@chakra-ui/react';
import {
  FaChartPie,
  FaCheckCircle,
  FaExclamationTriangle,
  FaInfoCircle,
  FaCalendarAlt,
  FaBuilding,
  FaFileInvoice,
  FaShoppingCart,
  FaCreditCard,
  FaUniversity,
  FaCogs,
  FaEdit
} from 'react-icons/fa';
import { ssotJournalAnalysisService, SSOTJournalAnalysisData, SSOTJournalAnalysisParams } from '../../services/ssotJournalAnalysisService';

interface EnhancedJournalAnalysisReportProps {
  params: SSOTJournalAnalysisParams;
}

const EnhancedJournalAnalysisReport: React.FC<EnhancedJournalAnalysisReportProps> = ({ params }) => {
  const [data, setData] = useState<SSOTJournalAnalysisData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const headingColor = useColorModeValue('gray.800', 'white');

  useEffect(() => {
    loadJournalAnalysis();
  }, [params]);

  const loadJournalAnalysis = async () => {
    try {
      setLoading(true);
      setError(null);
      const result = await ssotJournalAnalysisService.generateSSOTJournalAnalysis(params);
      setData(result);
    } catch (err: any) {
      setError(err.message || 'Failed to load journal analysis');
    } finally {
      setLoading(false);
    }
  };

  const getSourceTypeIcon = (sourceType: string) => {
    switch (sourceType.toLowerCase()) {
      case 'sales transaction':
        return FaShoppingCart;
      case 'purchase transaction':
        return FaFileInvoice;
      case 'payment transaction':
        return FaCreditCard;
      case 'cash & bank transaction':
        return FaUniversity;
      case 'manual journal entry':
        return FaEdit;
      default:
        return FaCogs;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'high':
        return 'red';
      case 'medium':
        return 'yellow';
      case 'low':
        return 'blue';
      default:
        return 'gray';
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const getQualityGrade = (score: number) => {
    if (score >= 90) return { grade: 'A', color: 'green' };
    if (score >= 80) return { grade: 'B', color: 'blue' };
    if (score >= 70) return { grade: 'C', color: 'yellow' };
    if (score >= 60) return { grade: 'D', color: 'orange' };
    return { grade: 'F', color: 'red' };
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="400px">
        <VStack>
          <Spinner size="xl" color="blue.500" />
          <Text>Loading journal analysis...</Text>
        </VStack>
      </Box>
    );
  }

  if (error) {
    return (
      <Alert status="error">
        <AlertIcon />
        <AlertTitle>Error!</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (!data) {
    return (
      <Alert status="info">
        <AlertIcon />
        <AlertTitle>No Data</AlertTitle>
        <AlertDescription>No journal analysis data available.</AlertDescription>
      </Alert>
    );
  }

  return (
    <VStack spacing={6} align="stretch" width="100%">
      {/* Header Section */}
      <Card bg={cardBg} borderColor={borderColor}>
        <CardHeader>
          <HStack justify="space-between" align="center">
            <VStack align="start" spacing={1}>
              <Heading size="lg" color={headingColor}>
                Journal Entry Analysis Report
              </Heading>
              <Text color={textColor}>
                Period: {params.start_date} to {params.end_date}
              </Text>
              {data.company && (
                <Text fontSize="sm" color={textColor}>
                  {data.company.name}
                </Text>
              )}
            </VStack>
            <Icon as={FaChartPie} w={8} h={8} color="blue.500" />
          </HStack>
        </CardHeader>
      </Card>

      {/* Summary Statistics */}
      <SimpleGrid columns={[2, 3, 5]} spacing={4}>
        <Card bg={cardBg} borderColor={borderColor}>
          <CardBody>
            <Stat>
              <StatLabel>Total Entries</StatLabel>
              <StatNumber>{data.total_entries.toLocaleString()}</StatNumber>
              <StatHelpText>All journal entries</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
        
        <Card bg={cardBg} borderColor={borderColor}>
          <CardBody>
            <Stat>
              <StatLabel>Posted Entries</StatLabel>
              <StatNumber color="green.500">{data.posted_entries.toLocaleString()}</StatNumber>
              <StatHelpText>
                {data.total_entries > 0 
                  ? `${Math.round((data.posted_entries / data.total_entries) * 100)}%`
                  : '0%'
                } of total
              </StatHelpText>
            </Stat>
          </CardBody>
        </Card>

        <Card bg={cardBg} borderColor={borderColor}>
          <CardBody>
            <Stat>
              <StatLabel>Draft Entries</StatLabel>
              <StatNumber color="yellow.500">{data.draft_entries.toLocaleString()}</StatNumber>
              <StatHelpText>Pending approval</StatHelpText>
            </Stat>
          </CardBody>
        </Card>

        <Card bg={cardBg} borderColor={borderColor}>
          <CardBody>
            <Stat>
              <StatLabel>Reversed Entries</StatLabel>
              <StatNumber color="red.500">{data.reversed_entries.toLocaleString()}</StatNumber>
              <StatHelpText>Cancelled entries</StatHelpText>
            </Stat>
          </CardBody>
        </Card>

        <Card bg={cardBg} borderColor={borderColor}>
          <CardBody>
            <Stat>
              <StatLabel>Total Amount</StatLabel>
              <StatNumber fontSize="md">{formatCurrency(data.total_amount)}</StatNumber>
              <StatHelpText>Posted entries only</StatHelpText>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Entry Type Distribution */}
      <Card bg={cardBg} borderColor={borderColor}>
        <CardHeader>
          <Heading size="md" color={headingColor}>Entry Type Distribution</Heading>
        </CardHeader>
        <CardBody>
          {data.entries_by_type && data.entries_by_type.length > 0 ? (
            <VStack spacing={4} align="stretch">
              {data.entries_by_type.map((entry, index) => (
                <Box key={index}>
                  <HStack justify="space-between" mb={2}>
                    <HStack>
                      <Icon as={getSourceTypeIcon(entry.source_type)} color="blue.500" />
                      <Text fontWeight="medium">{entry.source_type}</Text>
                      <Badge colorScheme="blue" variant="subtle">
                        {entry.count} entries
                      </Badge>
                    </HStack>
                    <VStack align="end" spacing={0}>
                      <Text fontWeight="bold">{formatCurrency(entry.total_amount)}</Text>
                      <Text fontSize="sm" color={textColor}>
                        {entry.percentage.toFixed(1)}% of total
                      </Text>
                    </VStack>
                  </HStack>
                  <Progress 
                    value={entry.percentage} 
                    colorScheme="blue" 
                    size="sm" 
                    borderRadius="md"
                    bg="gray.100"
                  />
                </Box>
              ))}
              
              {/* Explanation for the distribution */}
              <Alert status="info" borderRadius="md">
                <AlertIcon />
                <Box>
                  <AlertTitle fontSize="sm">Understanding the Distribution</AlertTitle>
                  <AlertDescription fontSize="sm">
                    This section shows how your journal entries are distributed across different transaction types. 
                    {data.entries_by_type.length === 1 && data.entries_by_type[0].source_type.includes('Purchase') && 
                      " It appears you have only Purchase transactions in this period. Sales and other transaction types may not be posting to the journal system properly."
                    }
                    {data.entries_by_type.length > 1 && 
                      " Multiple transaction types indicate a healthy mix of business activities being recorded."
                    }
                  </AlertDescription>
                </Box>
              </Alert>
            </VStack>
          ) : (
            <Alert status="warning">
              <AlertIcon />
              <AlertTitle>No Entry Type Data</AlertTitle>
              <AlertDescription>
                No journal entries found for the selected period, or entries are not properly categorized.
              </AlertDescription>
            </Alert>
          )}
        </CardBody>
      </Card>

      {/* Detailed Analysis Sections */}
      <Accordion allowToggle>
        {/* Account Breakdown */}
        <AccordionItem border="1px" borderColor={borderColor} borderRadius="md">
          <AccordionButton bg={cardBg}>
            <Box as="span" flex="1" textAlign="left">
              <HStack>
                <Icon as={FaBuilding} color="green.500" />
                <Text fontWeight="medium">Account Breakdown</Text>
                <Badge colorScheme="green">{data.entries_by_account?.length || 0} accounts</Badge>
              </HStack>
            </Box>
            <AccordionIcon />
          </AccordionButton>
          <AccordionPanel pb={4} bg={cardBg}>
            {data.entries_by_account && data.entries_by_account.length > 0 ? (
              <TableContainer>
                <Table variant="simple" size="sm">
                  <Thead>
                    <Tr>
                      <Th>Account Code</Th>
                      <Th>Account Name</Th>
                      <Th>Entries</Th>
                      <Th isNumeric>Total Debit</Th>
                      <Th isNumeric>Total Credit</Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {data.entries_by_account.map((account, index) => (
                      <Tr key={index}>
                        <Td fontWeight="medium">{account.account_code}</Td>
                        <Td>{account.account_name}</Td>
                        <Td>
                          <Badge colorScheme="blue" variant="subtle">
                            {account.count}
                          </Badge>
                        </Td>
                        <Td isNumeric>{formatCurrency(account.total_debit)}</Td>
                        <Td isNumeric>{formatCurrency(account.total_credit)}</Td>
                      </Tr>
                    ))}
                  </Tbody>
                </Table>
              </TableContainer>
            ) : (
              <Alert status="info">
                <AlertIcon />
                <AlertTitle>No Account Data</AlertTitle>
                <AlertDescription>
                  Account-wise breakdown is not available for this period.
                </AlertDescription>
              </Alert>
            )}
          </AccordionPanel>
        </AccordionItem>

        {/* Period Breakdown */}
        <AccordionItem border="1px" borderColor={borderColor} borderRadius="md">
          <AccordionButton bg={cardBg}>
            <Box as="span" flex="1" textAlign="left">
              <HStack>
                <Icon as={FaCalendarAlt} color="purple.500" />
                <Text fontWeight="medium">Period Breakdown</Text>
                <Badge colorScheme="purple">{data.entries_by_period?.length || 0} periods</Badge>
              </HStack>
            </Box>
            <AccordionIcon />
          </AccordionButton>
          <AccordionPanel pb={4} bg={cardBg}>
            {data.entries_by_period && data.entries_by_period.length > 0 ? (
              <VStack spacing={3} align="stretch">
                {data.entries_by_period.map((period, index) => (
                  <HStack key={index} justify="space-between" p={3} bg="gray.50" borderRadius="md">
                    <VStack align="start" spacing={0}>
                      <Text fontWeight="medium">{period.period}</Text>
                      <Text fontSize="sm" color={textColor}>
                        {period.count} entries
                      </Text>
                    </VStack>
                    <Text fontWeight="bold">{formatCurrency(period.total_amount)}</Text>
                  </HStack>
                ))}
              </VStack>
            ) : (
              <Alert status="info">
                <AlertIcon />
                <AlertTitle>No Period Data</AlertTitle>
                <AlertDescription>
                  Period-wise breakdown is not available for this period.
                </AlertDescription>
              </Alert>
            )}
          </AccordionPanel>
        </AccordionItem>

        {/* Compliance Check */}
        <AccordionItem border="1px" borderColor={borderColor} borderRadius="md">
          <AccordionButton bg={cardBg}>
            <Box as="span" flex="1" textAlign="left">
              <HStack>
                <Icon as={FaCheckCircle} color="green.500" />
                <Text fontWeight="medium">Compliance Assessment</Text>
                <Badge colorScheme="green">
                  {Math.round(data.compliance_check?.compliance_score || 0)}% compliant
                </Badge>
              </HStack>
            </Box>
            <AccordionIcon />
          </AccordionButton>
          <AccordionPanel pb={4} bg={cardBg}>
            {data.compliance_check ? (
              <VStack spacing={4} align="stretch">
                <SimpleGrid columns={3} spacing={4}>
                  <Stat>
                    <StatLabel>Total Checks</StatLabel>
                    <StatNumber>{data.compliance_check.total_checks}</StatNumber>
                  </Stat>
                  <Stat>
                    <StatLabel>Passed</StatLabel>
                    <StatNumber color="green.500">{data.compliance_check.passed_checks}</StatNumber>
                  </Stat>
                  <Stat>
                    <StatLabel>Failed</StatLabel>
                    <StatNumber color="red.500">{data.compliance_check.failed_checks}</StatNumber>
                  </Stat>
                </SimpleGrid>
                
                <Progress 
                  value={data.compliance_check.compliance_score} 
                  colorScheme="green" 
                  size="lg"
                  borderRadius="md"
                />

                {data.compliance_check.issues && data.compliance_check.issues.length > 0 && (
                  <VStack spacing={2} align="stretch">
                    <Text fontWeight="medium">Issues Found:</Text>
                    {data.compliance_check.issues.map((issue, index) => (
                      <Alert key={index} status="warning" size="sm">
                        <AlertIcon />
                        <Box>
                          <AlertTitle fontSize="sm">{issue.type}</AlertTitle>
                          <AlertDescription fontSize="sm">{issue.description}</AlertDescription>
                        </Box>
                        <Badge colorScheme={getSeverityColor(issue.severity)} ml={2}>
                          {issue.severity}
                        </Badge>
                      </Alert>
                    ))}
                  </VStack>
                )}
              </VStack>
            ) : (
              <Alert status="info">
                <AlertIcon />
                <AlertTitle>No Compliance Data</AlertTitle>
                <AlertDescription>Compliance assessment is not available.</AlertDescription>
              </Alert>
            )}
          </AccordionPanel>
        </AccordionItem>

        {/* Data Quality Metrics */}
        <AccordionItem border="1px" borderColor={borderColor} borderRadius="md">
          <AccordionButton bg={cardBg}>
            <Box as="span" flex="1" textAlign="left">
              <HStack>
                <Icon as={FaInfoCircle} color="blue.500" />
                <Text fontWeight="medium">Data Quality Assessment</Text>
                {data.data_quality_metrics && (
                  <Badge colorScheme={getQualityGrade(data.data_quality_metrics.overall_score).color}>
                    Grade {getQualityGrade(data.data_quality_metrics.overall_score).grade}
                  </Badge>
                )}
              </HStack>
            </Box>
            <AccordionIcon />
          </AccordionButton>
          <AccordionPanel pb={4} bg={cardBg}>
            {data.data_quality_metrics ? (
              <VStack spacing={4} align="stretch">
                <SimpleGrid columns={4} spacing={4}>
                  <Stat>
                    <StatLabel>Overall Score</StatLabel>
                    <StatNumber>{Math.round(data.data_quality_metrics.overall_score)}%</StatNumber>
                  </Stat>
                  <Stat>
                    <StatLabel>Completeness</StatLabel>
                    <StatNumber>{Math.round(data.data_quality_metrics.completeness_score)}%</StatNumber>
                  </Stat>
                  <Stat>
                    <StatLabel>Accuracy</StatLabel>
                    <StatNumber>{Math.round(data.data_quality_metrics.accuracy_score)}%</StatNumber>
                  </Stat>
                  <Stat>
                    <StatLabel>Consistency</StatLabel>
                    <StatNumber>{Math.round(data.data_quality_metrics.consistency_score)}%</StatNumber>
                  </Stat>
                </SimpleGrid>

                {data.data_quality_metrics.issues && data.data_quality_metrics.issues.length > 0 && (
                  <VStack spacing={2} align="stretch">
                    <Text fontWeight="medium">Quality Issues:</Text>
                    {data.data_quality_metrics.issues.map((issue, index) => (
                      <Alert key={index} status="warning" size="sm">
                        <AlertIcon />
                        <Box>
                          <AlertTitle fontSize="sm">{issue.type}</AlertTitle>
                          <AlertDescription fontSize="sm">
                            {issue.description} ({issue.count} entries affected)
                          </AlertDescription>
                        </Box>
                        <Badge colorScheme={getSeverityColor(issue.severity)} ml={2}>
                          {issue.severity}
                        </Badge>
                      </Alert>
                    ))}
                  </VStack>
                )}
              </VStack>
            ) : (
              <Alert status="info">
                <AlertIcon />
                <AlertTitle>No Quality Data</AlertTitle>
                <AlertDescription>Data quality assessment is not available.</AlertDescription>
              </Alert>
            )}
          </AccordionPanel>
        </AccordionItem>
      </Accordion>

      {/* Footer */}
      <Box textAlign="center" py={4}>
        <Text fontSize="sm" color={textColor}>
          Report generated on {new Date(data.generated_at).toLocaleString('id-ID')}
        </Text>
      </Box>
    </VStack>
  );
};

export default EnhancedJournalAnalysisReport;