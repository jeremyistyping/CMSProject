'use client';

import React from 'react';
import SessionDebugger from '@/components/debug/SessionDebugger';

const DebugSessionPage: React.FC = () => {
  return (
    <div className="min-h-screen bg-gray-100 py-8">
      <div className="max-w-6xl mx-auto px-4">
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900">Session Debugger</h1>
          <p className="text-gray-600 mt-2">
            Debug session information and authentication status
          </p>
        </div>
        
        <SessionDebugger />
      </div>
    </div>
  );
};

export default DebugSessionPage;
