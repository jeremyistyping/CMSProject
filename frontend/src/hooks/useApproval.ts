import { useEffect, useState, useCallback } from 'react';
import approvalService, { ApprovalHistory } from '@/services/approvalService';

export function useApproval(purchaseId?: number) {
  const [history, setHistory] = useState<ApprovalHistory[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchHistory = useCallback(async () => {
    if (!purchaseId) return;
    try {
      setLoading(true);
      setError(null);
      const res = await approvalService.getApprovalHistory(purchaseId);
      setHistory(res.approval_history || []);
    } catch (e: any) {
      // Treat 404 (not found) as no history rather than an error
      if (e?.response?.status === 404) {
        setHistory([]);
        setError(null);
      } else {
        setError(e?.response?.data?.error || e?.message || 'Failed to fetch approval history');
        setHistory([]);
      }
    } finally {
      setLoading(false);
    }
  }, [purchaseId]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return { history, loading, error, refresh: fetchHistory };
}

