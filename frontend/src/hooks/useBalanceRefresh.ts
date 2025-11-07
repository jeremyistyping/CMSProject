import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';

interface BalanceRefreshOptions {
  interval?: number; // Refresh interval in milliseconds
  enabled?: boolean; // Whether auto-refresh is enabled
  onRefresh?: (balances: any) => void; // Callback when balances are refreshed
}

export const useBalanceRefresh = (options: BalanceRefreshOptions = {}) => {
  const { token } = useAuth();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);

  const {
    interval = 30000, // Default 30 seconds
    enabled = true,
    onRefresh
  } = options;

  // Manual refresh function
  const refreshBalances = useCallback(async () => {
    if (!token) return;

    setIsRefreshing(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/accounts/hierarchy', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to refresh balances: ${response.statusText}`);
      }

      const data = await response.json();
      setLastRefresh(new Date());
      
      if (onRefresh) {
        onRefresh(data.data);
      }

      console.log('🔄 Balances refreshed successfully');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      console.error('❌ Failed to refresh balances:', errorMessage);
    } finally {
      setIsRefreshing(false);
    }
  }, [token, onRefresh]);

  // Auto-refresh effect
  useEffect(() => {
    if (!enabled || !token) return;

    const intervalId = setInterval(refreshBalances, interval);

    return () => {
      clearInterval(intervalId);
    };
  }, [enabled, token, interval, refreshBalances]);

  // Initial refresh on mount
  useEffect(() => {
    if (enabled && token) {
      refreshBalances();
    }
  }, [enabled, token, refreshBalances]);

  return {
    refreshBalances,
    isRefreshing,
    lastRefresh,
    error,
    clearError: () => setError(null)
  };
};
