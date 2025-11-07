import React, { useState } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Box,
  Text,
  VStack,
  HStack,
  Button,
  Divider,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
  Badge,
  Flex,
  useColorModeValue,
  Grid,
  GridItem,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Tooltip,
  useToast
} from '@chakra-ui/react';
import { FiDownload, FiTrendingUp, FiDollarSign, FiPieChart } from 'react-icons/fi';
import { formatCurrency } from '../../utils/formatters';
import JournalDrilldownButton from './JournalDrilldownButton';

interface EnhancedProfitLossModalProps {
  isOpen: boolean;
  onClose: () => void;
  data: any;
  onJournalDrilldown?: (itemName: string, accountCode?: string, amount?: number) => void;
  onExport?: (format: 'pdf' | 'excel') => void;
}

const EnhancedProfitLossModal: React.FC<EnhancedProfitLossModalProps> = ({
  isOpen,
  onClose,
  data,
  onJournalDrilldown,
  onExport
}) => {
  const [activeTab, setActiveTab] = useState<'statement' | 'metrics' | 'analysis'>('statement');
  const toast = useToast();
  
  // 🔍 DEBUG: Confirm this file is loaded with tooltips
  React.useEffect(() => {
    console.log('🎉 [PL Modal v2.0 WITH TOOLTIPS] Component loaded at:', new Date().toLocaleTimeString());
    console.log('📊 Data sections:', data?.sections?.length || 0);
  }, [data]);
  
  // Color mode values
  const modalBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const sectionBg = useColorModeValue('gray.50', 'gray.700');
  const subsectionBg = useColorModeValue('blue.50', 'blue.900');
  const metricsBg = useColorModeValue('green.50', 'green.900');
  const textColor = useColorModeValue('gray.800', 'white');
  const secondaryTextColor = useColorModeValue('gray.600', 'gray.300');
  const positiveColor = useColorModeValue('green.600', 'green.300');
  const negativeColor = useColorModeValue('red.600', 'red.300');
  
  if (!data) return null;

  const handleExport = (format: 'pdf' | 'excel') => {
    if (onExport) {
      onExport(format);
    } else {
      toast({
        title: 'Export Feature',
        description: `${format.toUpperCase()} export will be implemented soon`,
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  const renderFinancialMetrics = () => {
    if (!data.financialMetrics) return null;
    
    const metrics = data.financialMetrics;
    
    return (
      <Grid templateColumns="repeat(2, 1fr)" gap={4} mb={6}>
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <HStack spacing={1}>
                  <StatLabel>Gross Profit</StatLabel>
                  <Tooltip 
                    label="📈 LABA KOTOR: Pendapatan dikurangi HPP/COGS. Margin ini menunjukkan efisiensi produksi/pembelian. Semakin tinggi, semakin baik!"
                    fontSize="sm"
                    maxW="300px"
                    hasArrow
                    placement="top"
                  >
                    <Box as="span" cursor="help" color="blue.500" fontSize="xs">ℹ️</Box>
                  </Tooltip>
                </HStack>
                <StatNumber color={metrics.grossProfit >= 0 ? positiveColor : negativeColor}>
                  {formatCurrency(metrics.grossProfit)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type={metrics.grossProfitMargin >= 0 ? 'increase' : 'decrease'} />
                  {metrics.grossProfitMargin.toFixed(1)}%
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <HStack spacing={1}>
                  <StatLabel>Operating Income</StatLabel>
                  <Tooltip 
                    label="💼 LABA OPERASIONAL: Laba kotor dikurangi biaya operasional. Menunjukkan profitabilitas dari aktivitas bisnis inti sebelum bunga dan pajak."
                    fontSize="sm"
                    maxW="300px"
                    hasArrow
                    placement="top"
                  >
                    <Box as="span" cursor="help" color="blue.500" fontSize="xs">ℹ️</Box>
                  </Tooltip>
                </HStack>
                <StatNumber color={metrics.operatingIncome >= 0 ? positiveColor : negativeColor}>
                  {formatCurrency(metrics.operatingIncome)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type={metrics.operatingMargin >= 0 ? 'increase' : 'decrease'} />
                  {metrics.operatingMargin.toFixed(1)}%
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <HStack spacing={1}>
                  <StatLabel>EBITDA</StatLabel>
                  <Tooltip 
                    label="📊 EBITDA (Earnings Before Interest, Tax, Depreciation & Amortization): Laba operasional sebelum bunga, pajak, depresiasi, dan amortisasi. Menunjukkan cash-generating ability bisnis."
                    fontSize="sm"
                    maxW="300px"
                    hasArrow
                    placement="top"
                  >
                    <Box as="span" cursor="help" color="blue.500" fontSize="xs">ℹ️</Box>
                  </Tooltip>
                </HStack>
                <StatNumber color={metrics.ebitda >= 0 ? positiveColor : negativeColor}>
                  {formatCurrency(metrics.ebitda)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type={metrics.ebitdaMargin >= 0 ? 'increase' : 'decrease'} />
                  {metrics.ebitdaMargin.toFixed(1)}%
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <HStack spacing={1}>
                  <StatLabel>Net Income</StatLabel>
                  <Tooltip 
                    label="✨ LABA BERSIH: Keuntungan final (bottom line) setelah semua biaya, pajak, bunga. Ini menunjukkan profitabilitas sesungguhnya dan berapa banyak yang bisa ditahan/dibagikan ke pemilik."
                    fontSize="sm"
                    maxW="300px"
                    hasArrow
                    placement="top"
                  >
                    <Box as="span" cursor="help" color="blue.500" fontSize="xs">ℹ️</Box>
                  </Tooltip>
                </HStack>
                <StatNumber color={metrics.netIncome >= 0 ? positiveColor : negativeColor}>
                  {formatCurrency(metrics.netIncome)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type={metrics.netIncomeMargin >= 0 ? 'increase' : 'decrease'} />
                  {metrics.netIncomeMargin.toFixed(1)}%
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
      </Grid>
    );
  };

  // ✅ Helper function untuk tooltip penjelasan istilah akuntansi
  const getSectionTooltip = (sectionName: string): string => {
    const tooltips: Record<string, string> = {
      'REVENUE': '💰 PENDAPATAN: Total uang yang diterima dari penjualan barang/jasa kepada pelanggan. Ini adalah sumber utama pemasukan perusahaan sebelum dikurangi biaya-biaya.',
      'COST OF GOODS SOLD': '📦 HARGA POKOK PENJUALAN (HPP/COGS): Biaya langsung untuk memproduksi atau membeli barang yang dijual. Termasuk biaya bahan baku, tenaga kerja langsung, dan overhead produksi.',
      'GROSS PROFIT': '📈 LABA KOTOR: Pendapatan dikurangi HPP. Menunjukkan berapa banyak keuntungan dari penjualan sebelum dikurangi biaya operasional (gaji, sewa, listrik, dll).',
      'OPERATING INCOME': '💼 LABA OPERASIONAL: Laba kotor dikurangi biaya operasional (operating expenses). Menunjukkan keuntungan dari aktivitas bisnis utama sebelum bunga dan pajak.',
      'NET INCOME': '✨ LABA BERSIH: Keuntungan final setelah semua biaya, pajak, bunga, dan pendapatan lain-lain. Ini adalah "bottom line" yang menunjukkan profitabilitas sesungguhnya.',
      'OPERATING EXPENSES': '🏢 BIAYA OPERASIONAL: Biaya untuk menjalankan bisnis sehari-hari seperti gaji karyawan, sewa kantor, listrik, marketing, administrasi. Tidak termasuk HPP.',
      'OTHER INCOME': '🎁 PENDAPATAN LAIN-LAIN: Pendapatan di luar bisnis utama, seperti bunga deposito, keuntungan penjualan aset, atau pendapatan sewa.',
      'OTHER EXPENSES': '📉 BIAYA LAIN-LAIN: Biaya di luar operasional normal, seperti bunga pinjaman, kerugian penjualan aset, atau denda.'
    };
    
    return tooltips[sectionName] || '';
  };

  const renderSection = (section: any, index: number) => {
    const isCalculated = section.isCalculated;
    const hasSubsections = section.subsections && section.subsections.length > 0;
    const tooltipText = getSectionTooltip(section.name);
    
    // 🔍 DEBUG: Log section name and tooltip availability
    console.log(`[PL Modal] Section: "${section.name}", Has Tooltip: ${!!tooltipText}`);
    
    return (
      <Card key={index} mb={4} bg={isCalculated ? metricsBg : 'transparent'}>
        <CardBody>
          <Flex justify="space-between" align="center" mb={3}>
            <HStack spacing={2}>
              <Heading size="md" color={textColor}>
                {section.name}
              </Heading>
              {/* Always render tooltip icon for testing */}
              <Tooltip 
                label={tooltipText || `Info for ${section.name}`}
                fontSize="sm"
                maxW="400px"
                hasArrow
                placement="top"
                bg="blue.600"
                color="white"
                p={3}
                borderRadius="md"
              >
                <Box 
                  as="span" 
                  cursor="help" 
                  color="blue.500" 
                  fontSize="lg"
                  _hover={{ color: "blue.600" }}
                  title={tooltipText ? "Has custom tooltip" : "No custom tooltip"}
                >
                  ℹ️
                </Box>
              </Tooltip>
            </HStack>
            <Text fontWeight="bold" fontSize="lg" color={textColor}>
              {formatCurrency(section.total)}
            </Text>
          </Flex>
          
          {hasSubsections ? (
            <Accordion allowMultiple>
              {section.subsections.map((subsection: any, subIndex: number) => (
                <AccordionItem key={subIndex} border="none">
                  <AccordionButton
                    bg={subsectionBg}
                    _hover={{ bg: useColorModeValue('blue.100', 'blue.800') }}
                    borderRadius="md"
                    mb={2}
                  >
                    <Box flex="1" textAlign="left">
                      <HStack justify="space-between">
                        <Text fontWeight="semibold">{subsection.name}</Text>
                        <Text fontWeight="bold">{formatCurrency(subsection.total)}</Text>
                      </HStack>
                    </Box>
                    <AccordionIcon />
                  </AccordionButton>
                  <AccordionPanel pb={4}>
                    <VStack spacing={2} align="stretch">
                      {subsection.items.map((item: any, itemIndex: number) => (
                        <HStack key={itemIndex} justify="space-between" pl={4}>
                          <HStack>
                            <Text fontSize="sm" color={secondaryTextColor}>
                              {item.name}
                            </Text>
                            {onJournalDrilldown && item.accountCode && (
                              <JournalDrilldownButton
                                size="xs"
                                onClick={() => onJournalDrilldown(item.name, item.accountCode, item.amount)}
                              />
                            )}
                          </HStack>
                          <Text fontSize="sm" fontWeight="medium">
                            {item.isPercentage ? `${item.amount.toFixed(1)}%` : formatCurrency(item.amount)}
                          </Text>
                        </HStack>
                      ))}
                    </VStack>
                  </AccordionPanel>
                </AccordionItem>
              ))}
            </Accordion>
          ) : (
            <VStack spacing={2} align="stretch">
              {section.items?.map((item: any, itemIndex: number) => (
                <HStack key={itemIndex} justify="space-between">
                  <HStack>
                    <Text color={secondaryTextColor}>
                      {item.name}
                    </Text>
                    {onJournalDrilldown && item.accountCode && (
                      <JournalDrilldownButton
                        size="xs"
                        onClick={() => onJournalDrilldown(item.name, item.accountCode, item.amount)}
                      />
                    )}
                  </HStack>
                  <Text fontWeight={isCalculated ? "bold" : "medium"}>
                            {item.isPercentage ? `${item.amount.toFixed(1)}%` : formatCurrency(item.amount)}
                          </Text>
                </HStack>
              ))}
            </VStack>
          )}
        </CardBody>
      </Card>
    );
  };

  const renderAnalysisTab = () => {
    if (!data.financialMetrics) {
      return (
        <Box textAlign="center" py={8}>
          <Text color={secondaryTextColor}>
            Financial analysis not available for this report format
          </Text>
        </Box>
      );
    }

    const metrics = data.financialMetrics;
    
    return (
      <VStack spacing={6} align="stretch">
        <Card>
          <CardHeader>
            <Heading size="sm">Profitability Analysis</Heading>
          </CardHeader>
          <CardBody>
            <Table size="sm">
              <Thead>
                <Tr>
                  <Th>Metric</Th>
                  <Th isNumeric>Value</Th>
                  <Th isNumeric>Percentage</Th>
                  <Th>Assessment</Th>
                </Tr>
              </Thead>
              <Tbody>
                <Tr>
                  <Td>Gross Profit</Td>
                  <Td isNumeric>{formatCurrency(metrics.grossProfit)}</Td>
                  <Td isNumeric>{metrics.grossProfitMargin.toFixed(1)}%</Td>
                  <Td>
                    <Badge colorScheme={metrics.grossProfitMargin > 20 ? 'green' : metrics.grossProfitMargin > 10 ? 'yellow' : 'red'}>
                      {metrics.grossProfitMargin > 20 ? 'Excellent' : metrics.grossProfitMargin > 10 ? 'Good' : 'Needs Improvement'}
                    </Badge>
                  </Td>
                </Tr>
                <Tr>
                  <Td>Operating Income</Td>
                  <Td isNumeric>{formatCurrency(metrics.operatingIncome)}</Td>
                  <Td isNumeric>{metrics.operatingMargin.toFixed(1)}%</Td>
                  <Td>
                    <Badge colorScheme={metrics.operatingMargin > 15 ? 'green' : metrics.operatingMargin > 5 ? 'yellow' : 'red'}>
                      {metrics.operatingMargin > 15 ? 'Strong' : metrics.operatingMargin > 5 ? 'Moderate' : 'Weak'}
                    </Badge>
                  </Td>
                </Tr>
                <Tr>
                  <Td>EBITDA</Td>
                  <Td isNumeric>{formatCurrency(metrics.ebitda)}</Td>
                  <Td isNumeric>{metrics.ebitdaMargin.toFixed(1)}%</Td>
                  <Td>
                    <Badge colorScheme={metrics.ebitdaMargin > 20 ? 'green' : metrics.ebitdaMargin > 10 ? 'yellow' : 'red'}>
                      {metrics.ebitdaMargin > 20 ? 'Excellent' : metrics.ebitdaMargin > 10 ? 'Good' : 'Poor'}
                    </Badge>
                  </Td>
                </Tr>
                <Tr>
                  <Td>Net Income</Td>
                  <Td isNumeric>{formatCurrency(metrics.netIncome)}</Td>
                  <Td isNumeric>{metrics.netIncomeMargin.toFixed(1)}%</Td>
                  <Td>
                    <Badge colorScheme={metrics.netIncomeMargin > 10 ? 'green' : metrics.netIncomeMargin > 3 ? 'yellow' : 'red'}>
                      {metrics.netIncomeMargin > 10 ? 'Profitable' : metrics.netIncomeMargin > 3 ? 'Marginal' : 'Unprofitable'}
                    </Badge>
                  </Td>
                </Tr>
              </Tbody>
            </Table>
          </CardBody>
        </Card>
        
        <Card>
          <CardHeader>
            <Heading size="sm">Key Insights</Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              <Text fontSize="sm">
                🎯 <strong>Profitability:</strong> {metrics.netIncomeMargin > 0 ? 'The company is generating positive returns' : 'The company is experiencing losses'}
              </Text>
              <Text fontSize="sm">
                📊 <strong>Operating Efficiency:</strong> {metrics.operatingMargin > 10 ? 'Strong operational performance' : 'Room for operational improvement'}
              </Text>
              <Text fontSize="sm">
                💰 <strong>Cash Generation:</strong> EBITDA margin of {metrics.ebitdaMargin.toFixed(1)}% indicates {metrics.ebitdaMargin > 15 ? 'strong' : 'moderate'} cash generation ability
              </Text>
            </VStack>
          </CardBody>
        </Card>
      </VStack>
    );
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="6xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent bg={modalBg} maxH="90vh">
        <ModalHeader borderBottom="1px" borderColor={borderColor}>
          <VStack align="stretch" spacing={2}>
            <HStack justify="space-between">
              <VStack align="start" spacing={0}>
                <Text fontSize="xl" fontWeight="bold">
                  {data.title || 'Enhanced Profit and Loss Statement'}
                </Text>
                <Text fontSize="sm" color={secondaryTextColor}>
                  {data.period}
                </Text>
                {data.company && (
                  <Text fontSize="sm" color={secondaryTextColor}>
                    {data.company.name}
                  </Text>
                )}
              </VStack>
              {data.enhanced && (
                <Badge colorScheme="blue" variant="solid">
                  Enhanced
                </Badge>
              )}
            </HStack>
            
            <HStack spacing={1}>
              <Button
                size="sm"
                variant={activeTab === 'statement' ? 'solid' : 'ghost'}
                onClick={() => setActiveTab('statement')}
                leftIcon={<FiDollarSign />}
              >
                Statement
              </Button>
              {data.enhanced && (
                <>
                  <Button
                    size="sm"
                    variant={activeTab === 'metrics' ? 'solid' : 'ghost'}
                    onClick={() => setActiveTab('metrics')}
                    leftIcon={<FiTrendingUp />}
                  >
                    Metrics
                  </Button>
                  <Button
                    size="sm"
                    variant={activeTab === 'analysis' ? 'solid' : 'ghost'}
                    onClick={() => setActiveTab('analysis')}
                    leftIcon={<FiPieChart />}
                  >
                    Analysis
                  </Button>
                </>
              )}
            </HStack>
          </VStack>
        </ModalHeader>
        <ModalCloseButton />

        <ModalBody py={6}>
          {activeTab === 'statement' && (
            <VStack spacing={4} align="stretch">
              {data.enhanced && renderFinancialMetrics()}
              {data.sections?.map(renderSection)}
            </VStack>
          )}
          
          {activeTab === 'metrics' && data.enhanced && renderFinancialMetrics()}
          
          {activeTab === 'analysis' && renderAnalysisTab()}
        </ModalBody>

        <ModalFooter borderTop="1px" borderColor={borderColor}>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiDownload />}
              size="sm"
              variant="outline"
              onClick={() => handleExport('pdf')}
            >
              Export PDF
            </Button>
            <Button
              leftIcon={<FiDownload />}
              size="sm"
              variant="outline"
              onClick={() => handleExport('excel')}
            >
              Export Excel
            </Button>
            <Button onClick={onClose} size="sm">
              Close
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default EnhancedProfitLossModal;
