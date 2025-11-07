'use client';

import React from 'react';
import {
  SimpleGrid,
  Card,
  CardBody,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  HStack,
  Icon,
  Text,
  Flex,
  Box,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiShoppingBag,
  FiDollarSign,
  FiAlertCircle,
  FiBarChart2,
  FiTrendingUp
} from 'react-icons/fi';

interface SalesStats {
  total_sales: number;
  total_amount: number;
  total_outstanding: number;
  avg_order_value: number;
  total_paid?: number;
}

interface EnhancedStatsCardsProps {
  stats: SalesStats;
  formatCurrency: (amount: number) => string;
}

const EnhancedStatsCards: React.FC<EnhancedStatsCardsProps> = ({
  stats,
  formatCurrency
}) => {
  // Calculate percentage of paid vs outstanding if we have total_paid
  const paidPercentage = stats.total_paid
    ? Math.round((stats.total_paid / (stats.total_amount || 1)) * 100)
    : 0;

  // Theme-aware colors
  const cardBg = useColorModeValue('white', 'var(--bg-secondary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const textColor = useColorModeValue('gray.600', 'var(--text-secondary)');
  const primaryTextColor = useColorModeValue('gray.800', 'var(--text-primary)');

  // Background gradient colors for cards
  const gradients = {
    blue: useColorModeValue(
      'linear-gradient(135deg, var(--accent-color), #4dabf7)',
      'linear-gradient(135deg, var(--accent-color), #4dabf7)'
    ),
    green: useColorModeValue(
      'linear-gradient(135deg, var(--success-color), #51cf66)',
      'linear-gradient(135deg, var(--success-color), #51cf66)'
    ),
    orange: useColorModeValue(
      'linear-gradient(135deg, #ff9800, #ed8936)',
      'linear-gradient(135deg, #ff9800, #ed8936)'
    ),
    purple: useColorModeValue(
      'linear-gradient(135deg, #9f7aea, #b794f4)',
      'linear-gradient(135deg, #9f7aea, #b794f4)'
    ),
  };

  return (
    <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
      {/* Total Sales Card */}
      <Card
        boxShadow="xl"
        bg={gradients.blue}
        color="white"
        borderRadius="lg"
        _hover={{ transform: 'translateY(-2px)', transition: 'transform 0.3s ease' }}
      >
        <CardBody>
          <Stat>
            <HStack>
              <Icon as={FiShoppingBag} boxSize={6} />
              <Box>
                <StatLabel fontSize="sm" opacity={0.9}>
                  Total Sales
                </StatLabel>
                <StatNumber fontSize="3xl" fontWeight="bold">
                  {stats.total_sales}
                </StatNumber>
                <StatHelpText fontSize="xs" opacity={0.8}>
                  Transactions this period
                </StatHelpText>
              </Box>
            </HStack>
          </Stat>
        </CardBody>
      </Card>

      {/* Total Revenue Card */}
      <Card
        boxShadow="xl"
        bg={gradients.green}
        color="white"
        borderRadius="lg"
        _hover={{ transform: 'translateY(-2px)', transition: 'transform 0.3s ease' }}
      >
        <CardBody>
          <Stat>
            <HStack>
              <Icon as={FiDollarSign} boxSize={6} />
              <Box>
                <StatLabel fontSize="sm" opacity={0.9}>
                  Total Revenue
                </StatLabel>
                <StatNumber fontSize="3xl" fontWeight="bold">
                  {formatCurrency(stats.total_amount)}
                </StatNumber>
                <StatHelpText fontSize="xs" opacity={0.8}>
                  Gross revenue
                </StatHelpText>
              </Box>
            </HStack>
          </Stat>
        </CardBody>
      </Card>

      {/* Outstanding Card */}
      <Card
        boxShadow="xl"
        bg={gradients.orange}
        color="white"
        borderRadius="lg"
        _hover={{ transform: 'translateY(-2px)', transition: 'transform 0.3s ease' }}
      >
        <CardBody>
          <Stat>
            <HStack>
              <Icon as={FiAlertCircle} boxSize={6} />
              <Box>
                <StatLabel fontSize="sm" opacity={0.9}>
                  Outstanding
                </StatLabel>
                <StatNumber fontSize="3xl" fontWeight="bold">
                  {formatCurrency(stats.total_outstanding)}
                </StatNumber>
                <StatHelpText fontSize="xs" opacity={0.8}>
                  Unpaid invoices
                </StatHelpText>
              </Box>
            </HStack>
          </Stat>
        </CardBody>
      </Card>

      {/* Average Order Value Card */}
      <Card
        boxShadow="xl"
        bg={gradients.purple}
        color="white"
        borderRadius="lg"
        _hover={{ transform: 'translateY(-2px)', transition: 'transform 0.3s ease' }}
      >
        <CardBody>
          <Stat>
            <HStack>
              <Icon as={FiBarChart2} boxSize={6} />
              <Box>
                <StatLabel fontSize="sm" opacity={0.9}>
                  Avg Order Value
                </StatLabel>
                <StatNumber fontSize="3xl" fontWeight="bold">
                  {formatCurrency(stats.avg_order_value)}
                </StatNumber>
                <StatHelpText fontSize="xs" opacity={0.8}>
                  Per transaction
                </StatHelpText>
              </Box>
            </HStack>
          </Stat>
        </CardBody>
      </Card>
    </SimpleGrid>
  );
};

export default EnhancedStatsCards;
