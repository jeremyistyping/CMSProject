import { useContext, useEffect } from 'react';
import { AuthContext } from '@/contexts/AuthContext';

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
    // Services now use shared api.ts with interceptors,
    // so they don't need separate unauthorized handlers
    
    // Cleanup function (optional)
    return () => {
      // Could clear handlers here if needed
    };
  }, [handleUnauthorized]);
};
