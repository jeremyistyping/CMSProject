import { useContext, useEffect } from 'react';
import { AuthContext } from '@/contexts/AuthContext';
import { accountService } from '@/services/accountService';
import { contactService } from '@/services/contactService';
import salesService from '@/services/salesService';

/**
 * Custom hook to setup unauthorized error handling for API services
 * This should be used in the main app or layout components
 */
export const useAuthService = () => {
  const authContext = useContext(AuthContext);
  
  // Add null check for context
  if (!authContext) {
    console.warn('useAuthService: AuthContext not found');
    return;
  }
  
  const { handleUnauthorized } = authContext;

  useEffect(() => {
    // Set up the unauthorized handler for all services
    if (typeof accountService.setUnauthorizedHandler === 'function') {
      accountService.setUnauthorizedHandler(handleUnauthorized);
    }
    if (typeof contactService.setUnauthorizedHandler === 'function') {
      contactService.setUnauthorizedHandler(handleUnauthorized);
    }
    
    // Note: salesService uses the shared api.ts with interceptors,
    // so it doesn't need a separate unauthorized handler
    
    // Cleanup function (optional - services are singletons, so this might not be necessary)
    return () => {
      // Could clear handlers here if needed
    };
  }, [handleUnauthorized]);
};
