import React, { useEffect, useState } from 'react';
import {
  Box,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  SimpleGrid,
  Spinner,
  Text,
  useColorModeValue,
  HStack,
  Icon,
} from '@chakra-ui/react';
import { FiDollarSign, FiTrendingUp, FiTrendingDown } from 'react-icons/fi';
import cbsService from '@/services/cbsService';

interface CBSBudgetSummaryProps {
  projectId: number;
}

interface BudgetSummary {
  project_id: number;
  total_budget: number;
  total_actual: number;
  total_variance: number;
  node_count: number;
}

const CBSBudgetSummary: React.FC<CBSBudgetSummaryProps> = ({ projectId }) => {
  const [summary, setSummary] = useState<BudgetSummary | null>(null);
  const [loading, setLoading] = useState(true);

  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');

  useEffect(() => {
    loadSummary();
  }, [projectId]);

  const loadSummary = async () => {
    try {
      setLoading(true);
      const data = await cbsService.getProjectBudgetSummary(projectId);
      setSummary(data);
    } catch (error) {
      console.error('Failed to load budget summary:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const getVarianceColor = (variance: number) => {
    if (variance > 0) return 'green.500';
    if (variance < 0) return 'red.500';
    return 'gray.500';
  };

  const getVarianceIcon = (variance: number) => {
    if (variance > 0) return FiTrendingUp;
    if (variance < 0) return FiTrendingDown;
    return FiDollarSign;
  };

  if (loading) {
    return (
      <Box
        bg={bgColor}
        p={6}
        borderRadius="lg"
        borderWidth="1px"
        borderColor={borderColor}
      >
        <HStack justify="center" spacing={3}>
          <Spinner size="sm" />
          <Text fontSize="sm">Loading budget summary...</Text>
        </HStack>
      </Box>
    );
  }

  if (!summary) {
    return null;
  }

  const variancePercentage = summary.total_budget > 0
    ? ((summary.total_variance / summary.total_budget) * 100).toFixed(1)
    : '0';

  return (
    <Box
      bg={bgColor}
      p={6}
      borderRadius="lg"
      borderWidth="1px"
      borderColor={borderColor}
      shadow="sm"
    >
      <Text fontSize="lg" fontWeight="bold" mb={4}>
        Budget Summary
      </Text>
      <SimpleGrid columns={{ base: 1, md: 4 }} spacing={4}>
        <Stat>
          <StatLabel fontSize="sm">Total Budget (CBS)</StatLabel>
          <StatNumber fontSize="xl">{formatCurrency(summary.total_budget)}</StatNumber>
          <StatHelpText fontSize="xs">{summary.node_count} CBS nodes</StatHelpText>
        </Stat>

        <Stat>
          <StatLabel fontSize="sm">Total Actual</StatLabel>
          <StatNumber fontSize="xl">{formatCurrency(summary.total_actual)}</StatNumber>
          <StatHelpText fontSize="xs">From PR allocations</StatHelpText>
        </Stat>

        <Stat>
          <StatLabel fontSize="sm">Variance</StatLabel>
          <StatNumber fontSize="xl" color={getVarianceColor(summary.total_variance)}>
            <HStack spacing={1}>
              <Icon as={getVarianceIcon(summary.total_variance)} />
              <Text>{formatCurrency(Math.abs(summary.total_variance))}</Text>
            </HStack>
          </StatNumber>
          <StatHelpText fontSize="xs">
            <StatArrow type={summary.total_variance >= 0 ? 'increase' : 'decrease'} />
            {variancePercentage}%
          </StatHelpText>
        </Stat>

        <Stat>
          <StatLabel fontSize="sm">Budget Utilization</StatLabel>
          <StatNumber fontSize="xl">
            {summary.total_budget > 0
              ? ((summary.total_actual / summary.total_budget) * 100).toFixed(1)
              : '0'}%
          </StatNumber>
          <StatHelpText fontSize="xs">
            {summary.total_variance >= 0 ? 'Under budget' : 'Over budget'}
          </StatHelpText>
        </Stat>
      </SimpleGrid>
    </Box>
  );
};

export default CBSBudgetSummary;
