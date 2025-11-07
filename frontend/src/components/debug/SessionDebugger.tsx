'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import api from '@/services/api';

interface SessionInfo {
  session: {
    session_id: string;
    is_active: boolean;
    created_at: string;
    expires_at: string;
    last_activity: string;
    ip_address: string;
    device_info: string;
  };
  user: {
    id: string;
    username: string;
    email: string;
    role: string;
    is_active: boolean;
  };
  timestamp: string;
}

interface ActiveSession {
  id: number;
  session_id: string;
  user_id: number;
  ip_address: string;
  device_info: string;
  created_at: string;
  expires_at: string;
  last_activity: string;
  is_active: boolean;
}

const SessionDebugger: React.FC = () => {
  const { user, token } = useAuth();
  const [sessionInfo, setSessionInfo] = useState<SessionInfo | null>(null);
  const [activeSessions, setActiveSessions] = useState<ActiveSession[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSessionInfo = async () => {
    if (!token) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const response = await api.get('/api/v1/auth/session-info');
      setSessionInfo(response.data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch session info');
    } finally {
      setLoading(false);
    }
  };

  const fetchActiveSessions = async () => {
    if (!token) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const response = await api.get('/api/v1/monitoring/sessions/active');
      setActiveSessions(response.data.data.sessions || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch active sessions');
    } finally {
      setLoading(false);
    }
  };

  const refreshData = () => {
    fetchSessionInfo();
    fetchActiveSessions();
  };

  useEffect(() => {
    if (token) {
      refreshData();
    }
  }, [token]);

  if (!user) {
    return (
      <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
        <p className="text-yellow-800">Please login to view session information</p>
      </div>
    );
  }

  return (
    <div className="p-6 bg-white rounded-lg shadow-lg">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold text-gray-800">Session Debugger</h2>
        <button
          onClick={refreshData}
          disabled={loading}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {loading ? 'Loading...' : 'Refresh'}
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Current Session Info */}
        <div className="bg-gray-50 p-4 rounded-lg">
          <h3 className="text-lg font-semibold mb-3 text-gray-700">Current Session</h3>
          {sessionInfo ? (
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="font-medium">Session ID:</span>
                <span className="font-mono text-xs">{sessionInfo.session.session_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">Status:</span>
                <span className={`px-2 py-1 rounded text-xs ${
                  sessionInfo.session.is_active 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-red-100 text-red-800'
                }`}>
                  {sessionInfo.session.is_active ? 'Active' : 'Inactive'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">Created:</span>
                <span>{new Date(sessionInfo.session.created_at).toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">Expires:</span>
                <span className={new Date(sessionInfo.session.expires_at) < new Date() ? 'text-red-600' : ''}>
                  {new Date(sessionInfo.session.expires_at).toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">Last Activity:</span>
                <span>{new Date(sessionInfo.session.last_activity).toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">IP Address:</span>
                <span>{sessionInfo.session.ip_address}</span>
              </div>
              <div className="flex justify-between">
                <span className="font-medium">Device:</span>
                <span>{sessionInfo.session.device_info}</span>
              </div>
            </div>
          ) : (
            <p className="text-gray-500">No session info available</p>
          )}
        </div>

        {/* Active Sessions */}
        <div className="bg-gray-50 p-4 rounded-lg">
          <h3 className="text-lg font-semibold mb-3 text-gray-700">
            Active Sessions ({activeSessions.length})
          </h3>
          {activeSessions.length > 0 ? (
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {activeSessions.map((session) => (
                <div key={session.id} className="bg-white p-3 rounded border text-xs">
                  <div className="flex justify-between items-start mb-2">
                    <span className="font-mono text-xs">{session.session_id}</span>
                    <span className={`px-2 py-1 rounded ${
                      session.is_active 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {session.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  <div className="text-gray-600 space-y-1">
                    <div>Created: {new Date(session.created_at).toLocaleString()}</div>
                    <div>Expires: {new Date(session.expires_at).toLocaleString()}</div>
                    <div>IP: {session.ip_address}</div>
                    <div>Device: {session.device_info}</div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-500">No active sessions found</p>
          )}
        </div>
      </div>

      {/* User Info */}
      <div className="mt-6 bg-blue-50 p-4 rounded-lg">
        <h3 className="text-lg font-semibold mb-3 text-gray-700">User Information</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="font-medium">User ID:</span> {user.id}
          </div>
          <div>
            <span className="font-medium">Email:</span> {user.email}
          </div>
          <div>
            <span className="font-medium">Role:</span> {user.role}
          </div>
          <div>
            <span className="font-medium">Token Length:</span> {token?.length || 0}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SessionDebugger;
