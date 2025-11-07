/**
 * Utility functions for consistent authentication token management across the application
 * This ensures all parts of the app use the same token retrieval strategy
 */

/**
 * Retrieves the authentication token from storage using multiple fallback methods
 * 
 * Priority order:
 * 1. localStorage with 'token' key (primary - used by AuthContext)
 * 2. Alternative storage keys as fallback
 * 3. Cookies as last resort
 * 
 * If token is found with alternative key, it's stored with the correct key for future use
 * 
 * @param throwOnError - Whether to throw an error if no token is found (default: true)
 * @returns The authentication token or null if not found and throwOnError is false
 */
export function getAuthToken(throwOnError: boolean = true): string | null {
  let token = null;
  
  if (typeof window !== 'undefined') {
    // Primary method: get from localStorage with 'token' key (AuthContext stores here)
    token = localStorage.getItem('token');
    
    // Fallback to alternative keys if needed
    if (!token) {
      token = localStorage.getItem('authToken') || 
             sessionStorage.getItem('token') || 
             sessionStorage.getItem('authToken');
             
      // If token found with alternative key, store it with the correct key for future use
      if (token) {
        console.log('Token found with alternative key, storing with correct key');
        localStorage.setItem('token', token);
      }
    }
    
    // Last resort: try cookies
    if (!token) {
      const cookies = document.cookie.split(';');
      for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'token' || name === 'authToken' || name === 'access_token') {
          token = value;
          // Store in localStorage for future use
          if (token) {
            localStorage.setItem('token', token);
          }
          break;
        }
      }
    }
  }
  
  if (!token && throwOnError) {
    throw new Error('Authentication token not found. Please login first.');
  }
  
  return token;
}

/**
 * Creates authorization headers with Bearer token for API requests
 * 
 * @param includeContentType - Whether to include Content-Type header (default: true)
 * @returns Headers object with Authorization and optionally Content-Type
 */
export function getAuthHeaders(includeContentType: boolean = true): Record<string, string> {
  const token = getAuthToken(true); // Will throw if no token found
  
  const headers: Record<string, string> = {
    'Authorization': `Bearer ${token}`,
  };
  
  if (includeContentType) {
    headers['Content-Type'] = 'application/json';
  }
  
  return headers;
}

/**
 * Checks if user is authenticated by verifying token existence
 * 
 * @returns True if token exists, false otherwise
 */
export function isAuthenticated(): boolean {
  try {
    const token = getAuthToken(false);
    return !!token;
  } catch {
    return false;
  }
}

/**
 * Clears all authentication tokens from storage
 * This is useful for logout operations or when tokens become invalid
 */
export function clearAuthTokens(): void {
  if (typeof window !== 'undefined') {
    // Clear all possible token storage locations
    localStorage.removeItem('token');
    localStorage.removeItem('authToken');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('authToken');
    
    // Clear related auth data
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
    localStorage.removeItem('authData');
    localStorage.removeItem('userData');
    
    // Clear cookies by setting them to expire
    const cookieNames = ['token', 'authToken', 'access_token'];
    cookieNames.forEach(name => {
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
    });
  }
}