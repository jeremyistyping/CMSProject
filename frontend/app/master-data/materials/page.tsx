'use client';

import { Box, Heading, Text, VStack, useColorModeValue } from '@chakra-ui/react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import MaterialList from '@/components/master-data/MaterialList';

// Material management page
export default function MaterialsPage() {
    const headingColor = useColorModeValue('gray.800', 'gray.100');
    const textColor = useColorModeValue('gray.600', 'gray.300');
    const boxBg = useColorModeValue('white', 'gray.800');
    const borderColor = useColorModeValue('gray.200', 'gray.700');

    return (
        <SimpleLayout>
            <Box>
                <VStack align="start" spacing={4} mb={6}>
                    <Heading size="lg" color={headingColor}>Master Material</Heading>
                    <Text fontSize="sm" color={textColor} maxW="3xl">
                        Kelola data master material untuk purchase request dan tracking inventory
                    </Text>
                </VStack>

                <Box bg={boxBg} borderWidth="1px" borderColor={borderColor} borderRadius="lg" p={6}>
                    <MaterialList />
                </Box>
            </Box>
        </SimpleLayout>
    );
}
