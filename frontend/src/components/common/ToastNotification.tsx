import React from 'react';
import { useToast, UseToastOptions } from '@chakra-ui/react';
import { CheckCircle, XCircle, AlertTriangle, Info } from 'lucide-react';

export interface ToastNotificationOptions extends UseToastOptions {
  type?: 'success' | 'error' | 'warning' | 'info';
  title: string;
  description: string;
  actionButton?: {
    text: string;
    onClick: () => void;
  };
}

export const useEnhancedToast = () => {
  const toast = useToast();

  const showToast = (options: ToastNotificationOptions) => {
    const {
      type = 'info',
      title,
      description,
      actionButton,
      duration = 5000,
      isClosable = true,
      position = 'top-right',
      ...otherOptions
    } = options;

    const getIcon = () => {
      switch (type) {
        case 'success':
          return '✅';
        case 'error':
          return '❌';
        case 'warning':
          return '⚠️';
        case 'info':
        default:
          return 'ℹ️';
      }
    };

    const getStatus = () => {
      switch (type) {
        case 'success':
          return 'success' as const;
        case 'error':
          return 'error' as const;
        case 'warning':
          return 'warning' as const;
        case 'info':
        default:
          return 'info' as const;
      }
    };

    toast({
      title: `${getIcon()} ${title}`,
      description: description,
      status: getStatus(),
      duration,
      isClosable,
      position,
      variant: 'left-accent',
      ...otherOptions,
    });
  };

  const success = (title: string, description: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'success',
      title,
      description,
      duration: 3000,
      ...options,
    });
  };

  const error = (title: string, description: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'error',
      title,
      description,
      duration: 7000,
      ...options,
    });
  };

  const warning = (title: string, description: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'warning',
      title,
      description,
      duration: 5000,
      ...options,
    });
  };

  const info = (title: string, description: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'info',
      title,
      description,
      duration: 4000,
      ...options,
    });
  };

  // Business logic specific toasts
  const validationError = (description: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'warning',
      title: 'Validation Error',
      description,
      duration: 6000,
      ...options,
    });
  };

  const networkError = (options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'error',
      title: 'Connection Error',
      description: 'Unable to connect to the server. Please check your internet connection and try again.',
      duration: 8000,
      ...options,
    });
  };

  const serverError = (options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'error',
      title: 'Server Error',
      description: 'The server encountered an error. Please try again in a few moments.',
      duration: 6000,
      ...options,
    });
  };

  const permissionError = (action: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'warning',
      title: 'Access Denied',
      description: `You don't have permission to ${action}. Contact your administrator for access.`,
      duration: 5000,
      ...options,
    });
  };

  const saveSuccess = (resourceName: string, isUpdate = false, options?: Partial<ToastNotificationOptions>) => {
    const action = isUpdate ? 'updated' : 'created';
    showToast({
      type: 'success',
      title: 'Success',
      description: `${resourceName} has been ${action} successfully.`,
      duration: 3000,
      ...options,
    });
  };

  const deleteSuccess = (resourceName: string, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'success',
      title: 'Deleted',
      description: `${resourceName} has been deleted successfully.`,
      duration: 3000,
      ...options,
    });
  };

  const insufficientStock = (productName: string, available: number, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'warning',
      title: 'Insufficient Stock',
      description: `Only ${available} units of "${productName}" are available. Please reduce the quantity or add more stock.`,
      duration: 6000,
      ...options,
    });
  };

  const creditLimitExceeded = (availableCredit: number, options?: Partial<ToastNotificationOptions>) => {
    showToast({
      type: 'warning',
      title: 'Credit Limit Exceeded',
      description: `Available credit limit: ${availableCredit}. Consider reducing the order amount or request a credit limit increase.`,
      duration: 6000,
      ...options,
    });
  };

  return {
    showToast,
    success,
    error,
    warning,
    info,
    validationError,
    networkError,
    serverError,
    permissionError,
    saveSuccess,
    deleteSuccess,
    insufficientStock,
    creditLimitExceeded,
  };
};

export default useEnhancedToast;