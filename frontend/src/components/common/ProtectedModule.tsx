import React from 'react';
import {
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Flex,
  Spinner,
  Text,
  VStack,
} from '@chakra-ui/react';
import { useModulePermissions } from '@/hooks/usePermissions';
import SimpleLayout from '@/components/layout/SimpleLayout';

interface ProtectedModuleProps {
  module: string;
  children: React.ReactNode;
  fallbackRoles?: string[]; // Fallback roles if permission API fails
  requirePermission?: 'view' | 'create' | 'edit' | 'delete' | 'approve' | 'export';
}

const ProtectedModule: React.FC<ProtectedModuleProps> = ({
  module,
  children,
  fallbackRoles = [],
  requirePermission = 'view'
}) => {
  const {
    canView,
    canCreate,
    canEdit,
    canDelete,
    canApprove,
    canExport,
    loading,
    error
  } = useModulePermissions(module);

  // Show loading while checking permissions
  if (loading) {
    return (
      <SimpleLayout>
        <Flex justify="center" align="center" minH="60vh">
          <VStack spacing={4}>
            <Spinner size="xl" color="blue.500" thickness="4px" />
            <Text>Checking permissions...</Text>
          </VStack>
        </Flex>
      </SimpleLayout>
    );
  }

  // Check required permission
  let hasRequiredPermission = false;
  switch (requirePermission) {
    case 'view':
      hasRequiredPermission = canView;
      break;
    case 'create':
      hasRequiredPermission = canCreate;
      break;
    case 'edit':
      hasRequiredPermission = canEdit;
      break;
    case 'delete':
      hasRequiredPermission = canDelete;
      break;
    case 'approve':
      hasRequiredPermission = canApprove;
      break;
    case 'export':
      hasRequiredPermission = canExport;
      break;
    default:
      hasRequiredPermission = canView;
  }

  // If user doesn't have required permission, show access denied
  if (!hasRequiredPermission) {
    return (
      <SimpleLayout>
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          <AlertTitle mr={2}>Access Denied!</AlertTitle>
          <AlertDescription>
            You don't have permission to {requirePermission} this module ({module}). 
            Please contact your administrator.
          </AlertDescription>
        </Alert>
      </SimpleLayout>
    );
  }

  // If permission check succeeds, render children
  return <>{children}</>;
};

export default ProtectedModule;
