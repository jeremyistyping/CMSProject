import React from 'react';
import { Box, Flex, Text, Icon, Badge, useColorModeValue } from '@chakra-ui/react';
import { useRouter } from 'next/navigation';
import { IconType } from 'react-icons';
import { FiArrowRight } from 'react-icons/fi';

interface DashboardCardProps {
    title: string;
    description: string;
    icon: IconType;
    badge: string;
    href: string;
    badgeColor?: string;
}

const DashboardCard: React.FC<DashboardCardProps> = ({
    title,
    description,
    icon,
    badge,
    href,
    badgeColor = 'green'
}) => {
    const router = useRouter();
    const bgColor = useColorModeValue('white', 'gray.800');
    const borderColor = useColorModeValue('gray.200', 'gray.700');
    const hoverBg = useColorModeValue('gray.50', 'gray.700');
    const iconBg = useColorModeValue('green.50', 'green.900');
    const iconColor = useColorModeValue('green.600', 'green.400');

    return (
        <Box
            bg={bgColor}
            borderWidth="1px"
            borderColor={borderColor}
            borderRadius="lg"
            p={6}
            cursor="pointer"
            transition="all 0.2s"
            _hover={{
                bg: hoverBg,
                transform: 'translateY(-2px)',
                boxShadow: 'md'
            }}
            onClick={() => router.push(href)}
            position="relative"
        >
            {/* Icon Circle */}
            <Flex
                w="48px"
                h="48px"
                bg={iconBg}
                borderRadius="full"
                alignItems="center"
                justifyContent="center"
                mb={4}
            >
                <Icon as={icon} fontSize="24px" color={iconColor} />
            </Flex>

            {/* Title and Badge */}
            <Flex alignItems="center" mb={2} gap={2}>
                <Text fontSize="lg" fontWeight="bold" color={useColorModeValue('gray.800', 'white')}>
                    {title}
                </Text>
            </Flex>

            {/* Badge */}
            <Badge colorScheme={badgeColor} fontSize="xs" mb={3} textTransform="uppercase">
                {badge}
            </Badge>

            {/* Description */}
            <Text fontSize="sm" color={useColorModeValue('gray.600', 'gray.400')} mb={4} noOfLines={2}>
                {description}
            </Text>

            {/* Arrow Icon */}
            <Flex alignItems="center" color={iconColor} fontSize="sm" fontWeight="medium">
                <Icon as={FiArrowRight} fontSize="18px" />
            </Flex>
        </Box>
    );
};

export default DashboardCard;
