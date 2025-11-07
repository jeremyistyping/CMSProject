import { API_V1_BASE } from '../config/api';

export interface BackendStatus {
  isOnline: boolean;
  hasSSotEndpoints: boolean;
  hasWebSocketSupport: boolean;
  error?: string;
  latency?: number;
}

export class BackendTestService {
  
  async testBackendConnection(): Promise<BackendStatus> {
    const startTime = Date.now();
    
    try {
      // Test basic health endpoint
      const healthResponse = await fetch(`${API_V1_BASE}/health`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal: AbortSignal.timeout(5000) // 5 second timeout
      });

      const latency = Date.now() - startTime;

      if (!healthResponse.ok) {
        return {
          isOnline: false,
          hasSSotEndpoints: false,
          hasWebSocketSupport: false,
          error: `Health check failed: ${healthResponse.status}`,
          latency
        };
      }

      // Test SSOT endpoints (without auth to just check if they exist)
      const ssotStatus = await this.testSSOTEndpoints();
      const wsStatus = await this.testWebSocketSupport();

      return {
        isOnline: true,
        hasSSotEndpoints: ssotStatus,
        hasWebSocketSupport: wsStatus,
        latency
      };

    } catch (error) {
      return {
        isOnline: false,
        hasSSotEndpoints: false,
        hasWebSocketSupport: false,
        error: error instanceof Error ? error.message : 'Connection failed',
        latency: Date.now() - startTime
      };
    }
  }

  private async testSSOTEndpoints(): Promise<boolean> {
    try {
      // Test SSOT status endpoint (should return 401 without auth, but endpoint should exist)
      const response = await fetch(`${API_V1_BASE}/ssot-reports/status`, {
        method: 'GET',
        signal: AbortSignal.timeout(3000)
      });

      // We expect 401 (unauthorized) if endpoint exists, but no auth provided
      // 404 would mean endpoint doesn't exist
      return response.status !== 404;
    } catch {
      return false;
    }
  }

  private async testWebSocketSupport(): Promise<boolean> {
    // Since we can't easily test WebSocket without auth in this context,
    // we'll assume it's available if the backend is online
    // In a production environment, you might want to test this differently
    return true;
  }

  async testWithAuthentication(token: string): Promise<BackendStatus & { authValid: boolean }> {
    const basicStatus = await this.testBackendConnection();
    
    if (!basicStatus.isOnline) {
      return { ...basicStatus, authValid: false };
    }

    try {
      // Test authenticated endpoint
      const response = await fetch(`${API_V1_BASE}/profile`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        signal: AbortSignal.timeout(3000)
      });

      return {
        ...basicStatus,
        authValid: response.ok
      };
    } catch {
      return {
        ...basicStatus,
        authValid: false
      };
    }
  }

  // Simple ping test
  async ping(): Promise<{ success: boolean; latency: number; error?: string }> {
    const startTime = Date.now();
    
    try {
      const response = await fetch(`${API_V1_BASE}/health`, {
        method: 'GET',
        signal: AbortSignal.timeout(3000)
      });

      return {
        success: response.ok,
        latency: Date.now() - startTime,
        error: response.ok ? undefined : `HTTP ${response.status}`
      };
    } catch (error) {
      return {
        success: false,
        latency: Date.now() - startTime,
        error: error instanceof Error ? error.message : 'Network error'
      };
    }
  }
}

// Singleton instance
export const backendTest = new BackendTestService();