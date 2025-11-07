'use client';

import React from 'react';
import SimpleLayout from './SimpleLayout';
import { UserRole } from '@/contexts/AuthContext';

interface UnifiedLayoutProps {
  children: React.ReactNode;
  allowedRoles?: (UserRole | string)[];
  showSidebar?: boolean;
  requireAuth?: boolean;
}

// UnifiedLayout component - another wrapper around SimpleLayout
const UnifiedLayout: React.FC<UnifiedLayoutProps> = ({ 
  children, 
  allowedRoles = [], 
  showSidebar = true,
  requireAuth = true 
}) => {
  return (
    <SimpleLayout 
      allowedRoles={allowedRoles}
      showSidebar={showSidebar}
      requireAuth={requireAuth}
    >
      {children}
    </SimpleLayout>
  );
};

export default UnifiedLayout;
