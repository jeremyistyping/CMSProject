/**
 * Production-Ready API Type System
 * 
 * Strict TypeScript definitions to prevent API-related errors at compile time
 */

// Base API Response Types
export interface BaseApiResponse<T = any> {
  success: boolean;
  message: string;
  data: T;
  timestamp?: string;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface ApiError {
  success: false;
  error: string;
  code: string;
  details?: any;
  timestamp: string;
}

// Environment-specific configuration
export interface ApiConfig {
  baseURL: string;
  timeout: number;
  retries: number;
  headers: Record<string, string>;
  environment: 'development' | 'staging' | 'production';
}

// API Endpoint Path Types (ensures type safety for endpoint references)
export interface APIEndpointPaths {
  // Authentication
  readonly VALIDATE_TOKEN: string;
  readonly LOGIN: string;
  readonly REGISTER: string;
  readonly REFRESH: string;
  readonly PROFILE: string;
  
  // Grouped Endpoints
  readonly ACCOUNTS: {
    readonly LIST: string;
    readonly HIERARCHY: string;
    readonly BALANCE_SUMMARY: string;
    readonly VALIDATE_CODE: string;
    readonly FIX_HEADER_STATUS: string;
    readonly BY_CODE: (code: string) => string;
    readonly ADMIN_DELETE: (code: string) => string;
    readonly IMPORT: string;
    readonly EXPORT_PDF: string;
    readonly EXPORT_EXCEL: string;
    readonly CATALOG: string;
    readonly CREDIT: string;
  };
  
  readonly JOURNALS: {
    readonly LIST: string;
    readonly ACCOUNT_BALANCES: string;
    readonly REFRESH_BALANCES: string;
    readonly SUMMARY: string;
    readonly BY_ID: (id: number) => string;
  };
  
  readonly SSOT_REPORTS: {
    readonly GENERAL_LEDGER: string;
    readonly INTEGRATED: string;
    readonly JOURNAL_ANALYSIS: string;
    readonly PURCHASE_REPORT: string;
    readonly PURCHASE_VALIDATE: string;
    readonly PURCHASE_SUMMARY: string;
    readonly REFRESH: string;
    readonly SALES_SUMMARY: string;
    readonly STATUS: string;
    readonly TRIAL_BALANCE: string;
    readonly VENDOR_ANALYSIS: string;
    readonly PROFIT_LOSS: string;
    readonly BALANCE_SHEET: string;
    readonly BALANCE_SHEET_DETAILS: string;
    readonly CASH_FLOW: string;
  };
  
  // Additional required endpoints
  readonly HEALTH: string;
}

// HTTP Method Types
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS';

// Request Configuration Type
export interface ApiRequestConfig extends RequestInit {
  timeout?: number;
  retries?: number;
  baseURL?: string;
  validateStatus?: (status: number) => boolean;
}

// Service Response Types
export interface ServiceResponse<T = any> {
  data: T;
  success: boolean;
  error?: string;
  status: number;
  headers?: Headers;
}

// API Service Interface
export interface IApiService {
  get<T>(endpoint: string, config?: ApiRequestConfig): Promise<ServiceResponse<T>>;
  post<T>(endpoint: string, data?: any, config?: ApiRequestConfig): Promise<ServiceResponse<T>>;
  put<T>(endpoint: string, data?: any, config?: ApiRequestConfig): Promise<ServiceResponse<T>>;
  delete<T>(endpoint: string, config?: ApiRequestConfig): Promise<ServiceResponse<T>>;
  patch<T>(endpoint: string, data?: any, config?: ApiRequestConfig): Promise<ServiceResponse<T>>;
}

// Validation Result Types
export interface EndpointValidation {
  path: string;
  isValid: boolean;
  exists: boolean;
  reachable: boolean;
  error?: string;
}

export interface ApiValidationReport {
  timestamp: string;
  environment: string;
  totalEndpoints: number;
  validEndpoints: number;
  invalidEndpoints: number;
  unreachableEndpoints: number;
  criticalFailures: string[];
  validations: EndpointValidation[];
  summary: {
    healthy: boolean;
    score: number; // 0-100
    issues: string[];
    recommendations: string[];
  };
}

// Production Monitoring Types
export interface ApiMetrics {
  endpoint: string;
  method: HttpMethod;
  responseTime: number;
  statusCode: number;
  success: boolean;
  error?: string;
  timestamp: string;
  userAgent?: string;
  userId?: string;
}

export interface ApiHealthStatus {
  status: 'healthy' | 'degraded' | 'unhealthy';
  version: string;
  uptime: number;
  database: 'connected' | 'disconnected';
  externalServices: Record<string, 'healthy' | 'unhealthy'>;
  lastChecked: string;
  details?: {
    memoryUsage?: number;
    cpuUsage?: number;
    activeConnections?: number;
  };
}

// Error Handling Types
export interface ApiErrorContext {
  endpoint: string;
  method: HttpMethod;
  requestId?: string;
  userId?: string;
  timestamp: string;
  userAgent?: string;
  requestData?: any;
  responseData?: any;
  stackTrace?: string;
}

export class ApiValidationError extends Error {
  constructor(
    message: string,
    public readonly context: ApiErrorContext,
    public readonly validationErrors: string[]
  ) {
    super(message);
    this.name = 'ApiValidationError';
  }
}

export class ApiTimeoutError extends Error {
  constructor(
    message: string,
    public readonly timeout: number,
    public readonly context: ApiErrorContext
  ) {
    super(message);
    this.name = 'ApiTimeoutError';
  }
}

export class ApiNetworkError extends Error {
  constructor(
    message: string,
    public readonly context: ApiErrorContext,
    public readonly originalError?: Error
  ) {
    super(message);
    this.name = 'ApiNetworkError';
  }
}

// Type Guards
export function isApiResponse<T>(obj: any): obj is BaseApiResponse<T> {
  return obj && typeof obj === 'object' && 
         typeof obj.success === 'boolean' && 
         'data' in obj;
}

export function isApiError(obj: any): obj is ApiError {
  return obj && typeof obj === 'object' && 
         obj.success === false && 
         typeof obj.error === 'string';
}

// Utility Types
export type ApiEndpointKeys = keyof APIEndpointPaths;
export type FunctionEndpoint<T> = T extends (...args: any[]) => any ? T : never;
export type StringEndpoint<T> = T extends string ? T : never;

// Extract all string endpoints from API_ENDPOINTS structure
export type ExtractEndpointURLs<T> = {
  [K in keyof T]: T[K] extends string 
    ? T[K]
    : T[K] extends (...args: any[]) => string
    ? ReturnType<T[K]>
    : T[K] extends object
    ? ExtractEndpointURLs<T[K]>
    : never;
}[keyof T];

// Production Environment Variables Type Safety
export interface ProductionEnvVars {
  NODE_ENV: 'production';
  NEXT_PUBLIC_API_URL: string;
  API_TIMEOUT: string;
  API_RETRIES: string;
  MONITORING_ENABLED: string;
  LOGGING_LEVEL: 'error' | 'warn' | 'info' | 'debug';
  SENTRY_DSN?: string;
  NEW_RELIC_LICENSE_KEY?: string;
}

export interface DevelopmentEnvVars {
  NODE_ENV: 'development';
  NEXT_PUBLIC_API_URL?: string;
  API_TIMEOUT?: string;
  API_RETRIES?: string;
  MONITORING_ENABLED?: string;
  LOGGING_LEVEL?: 'error' | 'warn' | 'info' | 'debug';
}

export type EnvironmentVariables = ProductionEnvVars | DevelopmentEnvVars;

// API Client Configuration
export interface ApiClientConfig {
  baseURL: string;
  timeout: number;
  retries: number;
  retryDelay: number;
  validateStatus: (status: number) => boolean;
  transformRequest?: (data: any) => any;
  transformResponse?: (data: any) => any;
  interceptors?: {
    request?: (config: ApiRequestConfig) => ApiRequestConfig | Promise<ApiRequestConfig>;
    response?: (response: Response) => Response | Promise<Response>;
    error?: (error: Error) => Error | Promise<Error>;
  };
  cache?: {
    enabled: boolean;
    ttl: number; // Time to live in milliseconds
    maxSize: number; // Maximum cache entries
  };
}

// Default configurations
export const DEFAULT_API_CONFIG: ApiClientConfig = {
  baseURL: '',
  timeout: 10000,
  retries: 3,
  retryDelay: 1000,
  validateStatus: (status: number) => status >= 200 && status < 300,
  cache: {
    enabled: false,
    ttl: 5 * 60 * 1000, // 5 minutes
    maxSize: 100
  }
};

export const PRODUCTION_API_CONFIG: ApiClientConfig = {
  ...DEFAULT_API_CONFIG,
  timeout: 15000,
  retries: 3,
  retryDelay: 2000,
  cache: {
    enabled: true,
    ttl: 10 * 60 * 1000, // 10 minutes
    maxSize: 500
  }
};