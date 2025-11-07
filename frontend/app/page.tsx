'use client';

import React from 'react';
import { redirect } from 'next/navigation';

// Simple redirect page - let Next.js handle the redirect
export default function Home() {
  // Check if we're on client side and redirect accordingly
  React.useEffect(() => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    if (token) {
      window.location.replace('/dashboard');
    } else {
      window.location.replace('/login');
    }
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <h1 className="text-3xl font-bold mb-4">Accounting Application</h1>
        <p className="text-gray-600 mb-6">Please wait while we redirect you...</p>
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
      </div>
    </div>
  );
}
