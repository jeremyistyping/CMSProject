'use client';

import React from 'react';
import {
  Box,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  useColorModeValue,
  Icon,
  Flex,
} from '@chakra-ui/react';

interface StatCardProps {
  title: string;
  stat: string;
  icon: any;
  change?: string;
  changeType?: 'increase' | 'decrease';
  color?: string;
}

export const StatCard: React.FC<StatCardProps> = ({
  title,
  stat,
  icon,
  change,
  changeType = 'increase',
  color = 'brand.500',
}) => {
  return (
    <Stat
      px={{ base: 4, md: 8 }}
      py={'5'}
      shadow={'lg'}
      border={'1px solid'}
      borderColor={useColorModeValue('gray.200', 'gray.500')}
      rounded={'lg'}
      bg={useColorModeValue('white', 'gray.700')}
      transition="all 0.2s"
      _hover={{
        shadow: 'xl',
        transform: 'translateY(-2px)'
      }}
    >
      <Flex justifyContent={'space-between'} align={'center'} gap={4}>
        <Box pl={{ base: 2, md: 4 }} minW={0}>
          <StatLabel 
            fontWeight={'medium'} 
            color={useColorModeValue('gray.600', 'gray.300')}
            noOfLines={1}
          >
            {title}
          </StatLabel>
          <StatNumber 
            fontSize={'2xl'} 
            fontWeight={'bold'}
            noOfLines={1}
            overflow="hidden"
            textOverflow="ellipsis"
            whiteSpace="nowrap"
            maxW="100%"
          >
            {stat}
          </StatNumber>
          {change && (
            <StatHelpText noOfLines={1}>
              <StatArrow type={changeType} />
              {change}
            </StatHelpText>
          )}
        </Box>
        <Box
          my={'auto'}
          color={useColorModeValue(color, 'gray.200')}
          alignContent={'center'}
          flexShrink={0}
        >
          <Icon as={icon} w={8} h={8} />
        </Box>
      </Flex>
    </Stat>
  );
};
