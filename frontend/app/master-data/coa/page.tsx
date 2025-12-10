'use client';

import { Box, Heading, Text, VStack, useColorModeValue } from '@chakra-ui/react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import COAList from '@/components/master-data/COAList';

// Chart of Accounts management page
export default function COAPage() {
    const headingColor = useColorModeValue('gray.800', 'gray.100');
    const textColor = useColorModeValue('gray.600', 'gray.300');
    const boxBg = useColorModeValue('white', 'gray.800');
    const borderColor = useColorModeValue('gray.200', 'gray.700');

    return (
        <SimpleLayout>
            <Box>
                <VStack align="start" spacing={4} mb={6}>
                    <Heading size="lg" color={headingColor}>Chart of Accounts</Heading>
                    <Text fontSize="sm" color={textColor} maxW="3xl">
                        Kelola Chart of Accounts untuk kategorisasi biaya proyek konstruksi
                    </Text>
                </VStack>

                <Box bg={boxBg} borderWidth="1px" borderColor={borderColor} borderRadius="lg" p={6}>
                    <COAList />
                </Box>
            </Box>
        </SimpleLayout>
    );
}
