'use client';

import React from 'react';
import SimpleLayout from './SimpleLayout';
import { UserRole } from '@/contexts/AuthContext';

interface LayoutProps {
  children: React.ReactNode;
  allowedRoles?: (UserRole | string)[];
  showSidebar?: boolean;
  requireAuth?: boolean;
}

// Layout component as a wrapper around SimpleLayout for backward compatibility
const Layout: React.FC<LayoutProps> = ({ 
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

export default Layout;
