import { useState, useEffect, useRef } from 'react';
import api from '@/services/api';
import { API_ENDPOINTS } from '@/config/api';

// Define the structure of the analytics data for Project Management System
export interface ProjectSummary {
  id: number;
  name: string;
  status: string;
  progress: number;
  budget: number;
  created_at: string;
}

export interface PurchaseRequestSummary {
  id: number;
  pr_number: string;
  project_name: string;
  total_amount: number;
  status: string;
  created_at: string;
}

export interface MonthlyData {
  month: string;
  value: number;
}

export interface DashboardAnalytics {
  // Project statistics
  totalProjects: number;
  activeProjects: number;
  completedProjects: number;
  
  // Purchase Request statistics
  totalPurchaseRequests: number;
  pendingApprovals: number;
  
  // Budget statistics
  totalBudget: number;
  totalSpent: number;
  
  // Monthly data for charts
  monthlyProjects: MonthlyData[];
  monthlyPRs: MonthlyData[];
  
  // Recent data
  recentProjects: ProjectSummary[];
  recentPurchaseRequests: PurchaseRequestSummary[];
}

// Cache for analytics data with timestamp
interface CacheEntry {
  data: DashboardAnalytics;
  timestamp: number;
}

const analyticsCache: { current: CacheEntry | null } = { current: null };
const CACHE_DURATION = 5000; // 5 seconds cache for faster updates

export const useDashboardAnalytics = (user: any, token: string | null) => {
  const [analytics, setAnalytics] = useState<DashboardAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const fetchingRef = useRef(false);

  useEffect(() => {
    if (!user || !token) {
      setLoading(false);
      return;
    }

    // Check if user has required role
    const userRoleNormalized = user.role?.toString().toLowerCase();
    if (!['admin', 'project_director', 'gm', 'managing_director', 'cost_control'].includes(userRoleNormalized)) {
      setLoading(false);
      return;
    }

    const fetchAnalytics = async () => {
      // Prevent duplicate API calls
      if (fetchingRef.current) {
        console.log('⚠️ Dashboard analytics fetch already in progress, skipping duplicate call');
        return;
      }

      // Check cache first
      if (analyticsCache.current) {
        const now = Date.now();
        const cacheAge = now - analyticsCache.current.timestamp;
        if (cacheAge < CACHE_DURATION) {
          console.log('📦 Using cached dashboard analytics data (age:', Math.round(cacheAge / 1000), 'seconds)');
          setAnalytics(analyticsCache.current.data);
          setError(null);
          setLoading(false);
          return;
        }
      }

      try {
        fetchingRef.current = true;
        console.log('🔍 Dashboard Debug Info:');
        console.log('User:', user);
        console.log('Token length:', token?.length);
        console.log('User role:', user.role);
        
        console.log('🌐 Making API request to', API_ENDPOINTS.DASHBOARD_ANALYTICS);
        const response = await api.get(API_ENDPOINTS.DASHBOARD_ANALYTICS);
        
        console.log('✅ Dashboard analytics response received:', response.data);
        
        // Extract analytics from nested response { success: true, data: {...} }
        const analyticsData = response.data.data || response.data;
        console.log('📊 Analytics data:', analyticsData);
        
        // Update cache
        analyticsCache.current = {
          data: analyticsData,
          timestamp: Date.now()
        };
        
        setAnalytics(analyticsData);
        setError(null);
      } catch (err: any) {
        console.error('❌ Failed to fetch dashboard analytics:', err);
        
        // Handle authentication errors
        if (err.response?.status === 401) {
          console.error('🚫 Authentication failed - token might be expired or invalid');
          setError('Session expired. Please login again.');
          return;
        }
        
        // Handle authorization errors  
        if (err.response?.status === 403) {
          console.error('🚫 User not authorized to view dashboard analytics');
          setError(`You are not authorized to view dashboard analytics. Current role: ${user.role}`);
        } else {
          const errorMessage = err.response?.data?.error || err.message || 'Failed to load dashboard data';
          setError(errorMessage);
        }
      } finally {
        setLoading(false);
        fetchingRef.current = false;
      }
    };

    fetchAnalytics();
  }, [user, token]);

  return { analytics, loading, error };
};
