'use client';

import React from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { DailyReportApprovalList } from '@/components/projects/DailyReportApprovalList';
import { Box } from '@chakra-ui/react';

export default function DailyReportApprovalPage() {
    return (
        <SimpleLayout allowedRoles={['gm', 'admin', 'project_director', 'managing_director']}>
            <Box p={6}>
                <DailyReportApprovalList />
            </Box>
        </SimpleLayout>
    );
}
