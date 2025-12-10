'use client';

import { useState } from 'react';
import {
    Box, VStack, HStack, Text, Table, Thead, Tbody, Tr, Th, Td, Badge, Heading,
    Stat, StatLabel, StatNumber, StatHelpText, StatArrow, Divider, useColorModeValue,
    Accordion, AccordionItem, AccordionButton, AccordionPanel, AccordionIcon,
} from '@chakra-ui/react';
import { BudgetReportResponse, BudgetCategoryReport, WorkPackageSummary } from '@/services/expenseTransactionService';

interface BudgetReportViewerProps {
    report: BudgetReportResponse;
}

export default function BudgetReportViewer({ report }: BudgetReportViewerProps) {
    const bgColor = useColorModeValue('white', 'gray.800');
    const borderColor = useColorModeValue('gray.200', 'gray.700');

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0,
        }).format(amount);
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('id-ID', {
            day: '2-digit',
            month: 'short',
            year: 'numeric',
        });
    };

    const getVarianceColor = (variance: number) => {
        if (variance > 0) return 'green.500';
        if (variance < 0) return 'red.500';
        return 'gray.500';
    };

    const renderCategoryReport = (title: string, categoryReport?: BudgetCategoryReport, showWorkPackage: boolean = false) => {
        if (!categoryReport || categoryReport.transactions.length === 0) {
            return null;
        }

        return (
            <Box borderWidth="1px" borderRadius="lg" p={6} bg={bgColor}>
                <VStack align="stretch" spacing={4}>
                    <Heading size="md" color="blue.600">{title}</Heading>

                    {/* Summary Stats */}
                    <HStack spacing={6} flexWrap="wrap">
                        <Stat>
                            <StatLabel>Budget Estimation</StatLabel>
                            <StatNumber fontSize="lg">{formatCurrency(categoryReport.budget_estimation)}</StatNumber>
                        </Stat>
                        <Stat>
                            <StatLabel>Actual</StatLabel>
                            <StatNumber fontSize="lg">{formatCurrency(categoryReport.actual)}</StatNumber>
                        </Stat>
                        <Stat>
                            <StatLabel>Variance</StatLabel>
                            <StatNumber fontSize="lg" color={getVarianceColor(categoryReport.variance)}>
                                {formatCurrency(Math.abs(categoryReport.variance))}
                            </StatNumber>
                            <StatHelpText>
                                <StatArrow type={categoryReport.variance >= 0 ? 'increase' : 'decrease'} />
                                {categoryReport.budget_estimation > 0
                                    ? ((categoryReport.variance / categoryReport.budget_estimation) * 100).toFixed(1)
                                    : 0}%
                            </StatHelpText>
                        </Stat>
                    </HStack>

                    <Divider />

                    {/* Work Package Breakdown (for Operational Budget) */}
                    {showWorkPackage && categoryReport.by_work_package && categoryReport.by_work_package.length > 0 && (
                        <Accordion allowMultiple>
                            {categoryReport.by_work_package.map((wp, idx) => (
                                <AccordionItem key={idx}>
                                    <AccordionButton>
                                        <Box flex="1" textAlign="left">
                                            <HStack justify="space-between" w="full">
                                                <Text fontWeight="bold">{wp.work_package}</Text>
                                                <HStack spacing={4}>
                                                    <Badge colorScheme="blue">
                                                        Budget: {formatCurrency(wp.budget_estimation)}
                                                    </Badge>
                                                    <Badge colorScheme="green">
                                                        Actual: {formatCurrency(wp.actual)}
                                                    </Badge>
                                                    <Badge colorScheme={wp.variance >= 0 ? 'green' : 'red'}>
                                                        Variance: {formatCurrency(Math.abs(wp.variance))}
                                                    </Badge>
                                                </HStack>
                                            </HStack>
                                        </Box>
                                        <AccordionIcon />
                                    </AccordionButton>
                                    <AccordionPanel pb={4}>
                                        {renderTransactionTable(wp.transactions)}
                                    </AccordionPanel>
                                </AccordionItem>
                            ))}
                        </Accordion>
                    )}

                    {/* All Transactions */}
                    {!showWorkPackage && renderTransactionTable(categoryReport.transactions)}
                </VStack>
            </Box>
        );
    };

    const renderTransactionTable = (transactions: any[]) => {
        if (transactions.length === 0) {
            return <Text color="gray.500">Tidak ada transaksi</Text>;
        }

        return (
            <Box overflowX="auto">
                <Table size="sm" variant="simple">
                    <Thead>
                        <Tr>
                            <Th>Tanggal</Th>
                            <Th>Deskripsi</Th>
                            <Th>COA</Th>
                            <Th isNumeric>Qty</Th>
                            <Th>Unit</Th>
                            <Th isNumeric>Total</Th>
                            <Th>Ref</Th>
                        </Tr>
                    </Thead>
                    <Tbody>
                        {transactions.map((tx, idx) => (
                            <Tr key={idx}>
                                <Td>{formatDate(tx.date)}</Td>
                                <Td maxW="300px" isTruncated title={tx.description}>
                                    {tx.description}
                                </Td>
                                <Td fontSize="xs">
                                    <Text fontWeight="bold">{tx.coa_code}</Text>
                                    <Text color="gray.600">{tx.coa_name}</Text>
                                </Td>
                                <Td isNumeric>{tx.quantity}</Td>
                                <Td>{tx.unit}</Td>
                                <Td isNumeric fontWeight="bold">
                                    {formatCurrency(tx.total_price)}
                                </Td>
                                <Td fontSize="xs">{tx.reference_no || '-'}</Td>
                            </Tr>
                        ))}
                    </Tbody>
                </Table>
            </Box>
        );
    };

    return (
        <VStack align="stretch" spacing={6}>
            {/* Report Header */}
            <Box borderWidth="1px" borderRadius="lg" p={6} bg={bgColor}>
                <VStack align="stretch" spacing={2}>
                    <Heading size="lg">{report.project_name}</Heading>
                    <Text color="gray.600">
                        Periode: {formatDate(report.start_date)} - {formatDate(report.end_date)}
                    </Text>
                    <Text fontSize="sm" color="gray.500">
                        Report Date: {formatDate(report.report_date)}
                    </Text>
                </VStack>
            </Box>

            {/* Labour Budget */}
            {renderCategoryReport('LABOUR BUDGET', report.labour_budget, false)}

            {/* Operational Budget */}
            {renderCategoryReport('OPERASIONAL BUDGET', report.operasional_budget, true)}

            {/* Other Budget */}
            {renderCategoryReport('BIAYA OPERASIONAL LAINNYA', report.other_budget, false)}

            {/* Grand Total */}
            <Box borderWidth="2px" borderRadius="lg" p={6} bg="blue.50" borderColor="blue.300">
                <HStack justify="space-between" flexWrap="wrap" spacing={6}>
                    <Stat>
                        <StatLabel fontWeight="bold">Total Budget Estimation</StatLabel>
                        <StatNumber fontSize="2xl" color="blue.600">
                            {formatCurrency(
                                (report.labour_budget?.budget_estimation || 0) +
                                (report.operasional_budget?.budget_estimation || 0) +
                                (report.other_budget?.budget_estimation || 0)
                            )}
                        </StatNumber>
                    </Stat>
                    <Stat>
                        <StatLabel fontWeight="bold">Total Actual</StatLabel>
                        <StatNumber fontSize="2xl" color="green.600">
                            {formatCurrency(
                                (report.labour_budget?.actual || 0) +
                                (report.operasional_budget?.actual || 0) +
                                (report.other_budget?.actual || 0)
                            )}
                        </StatNumber>
                    </Stat>
                    <Stat>
                        <StatLabel fontWeight="bold">Total Variance</StatLabel>
                        <StatNumber
                            fontSize="2xl"
                            color={getVarianceColor(
                                (report.labour_budget?.variance || 0) +
                                (report.operasional_budget?.variance || 0) +
                                (report.other_budget?.variance || 0)
                            )}
                        >
                            {formatCurrency(
                                Math.abs(
                                    (report.labour_budget?.variance || 0) +
                                    (report.operasional_budget?.variance || 0) +
                                    (report.other_budget?.variance || 0)
                                )
                            )}
                        </StatNumber>
                    </Stat>
                </HStack>
            </Box>
        </VStack>
    );
}
