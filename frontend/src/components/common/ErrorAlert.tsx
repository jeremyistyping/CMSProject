import React from 'react';
import { AlertTriangle, XCircle, Info, CheckCircle, X } from 'lucide-react';

export interface ErrorDetail {
  field?: string;
  message: string;
  suggestion?: string;
  context?: Record<string, any>;
}

export interface ErrorAlertProps {
  title?: string;
  message?: string;
  errors?: ErrorDetail[];
  type?: 'error' | 'warning' | 'info' | 'success';
  onClose?: () => void;
  dismissible?: boolean;
  showIcon?: boolean;
  className?: string;
}

const ErrorAlert: React.FC<ErrorAlertProps> = ({
  title,
  message,
  errors = [],
  type = 'error',
  onClose,
  dismissible = true,
  showIcon = true,
  className = '',
}) => {
  const getIcon = () => {
    switch (type) {
      case 'error':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'warning':
        return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
      case 'info':
        return <Info className="h-5 w-5 text-blue-500" />;
      case 'success':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      default:
        return <XCircle className="h-5 w-5 text-red-500" />;
    }
  };

  const getColorClasses = () => {
    switch (type) {
      case 'error':
        return {
          container: 'bg-red-50 border-red-200',
          title: 'text-red-800',
          message: 'text-red-700',
          detail: 'text-red-600',
          suggestion: 'text-red-600 bg-red-100',
          context: 'text-red-500',
        };
      case 'warning':
        return {
          container: 'bg-yellow-50 border-yellow-200',
          title: 'text-yellow-800',
          message: 'text-yellow-700',
          detail: 'text-yellow-600',
          suggestion: 'text-yellow-600 bg-yellow-100',
          context: 'text-yellow-500',
        };
      case 'info':
        return {
          container: 'bg-blue-50 border-blue-200',
          title: 'text-blue-800',
          message: 'text-blue-700',
          detail: 'text-blue-600',
          suggestion: 'text-blue-600 bg-blue-100',
          context: 'text-blue-500',
        };
      case 'success':
        return {
          container: 'bg-green-50 border-green-200',
          title: 'text-green-800',
          message: 'text-green-700',
          detail: 'text-green-600',
          suggestion: 'text-green-600 bg-green-100',
          context: 'text-green-500',
        };
      default:
        return {
          container: 'bg-red-50 border-red-200',
          title: 'text-red-800',
          message: 'text-red-700',
          detail: 'text-red-600',
          suggestion: 'text-red-600 bg-red-100',
          context: 'text-red-500',
        };
    }
  };

  const colors = getColorClasses();

  const formatFieldName = (field: string) => {
    return field
      .replace(/[_-]/g, ' ')
      .replace(/\b\w/g, (l) => l.toUpperCase())
      .replace(/Id/g, 'ID');
  };

  const renderContext = (context: Record<string, any>) => {
    return Object.entries(context).map(([key, value]) => {
      if (key === 'available_stock' || key === 'product_name') {
        return (
          <div key={key} className="text-sm">
            <span className="font-medium">{formatFieldName(key)}: </span>
            <span>{value}</span>
          </div>
        );
      }
      return null;
    }).filter(Boolean);
  };

  return (
    <div className={`rounded-lg border p-4 ${colors.container} ${className}`}>
      <div className="flex">
        <div className="flex-shrink-0">
          {showIcon && getIcon()}
        </div>
        <div className="ml-3 flex-1">
          {title && (
            <h3 className={`text-sm font-medium ${colors.title}`}>
              {title}
            </h3>
          )}
          {message && (
            <div className={`text-sm ${title ? 'mt-1' : ''} ${colors.message}`}>
              {message}
            </div>
          )}
          
          {errors.length > 0 && (
            <div className={`text-sm ${(title || message) ? 'mt-2' : ''}`}>
              <ul className="space-y-2">
                {errors.map((error, index) => (
                  <li key={index} className="space-y-1">
                    <div className={colors.detail}>
                      {error.field && (
                        <span className="font-medium">
                          {formatFieldName(error.field)}: 
                        </span>
                      )}
                      <span className={error.field ? 'ml-1' : ''}>
                        {error.message}
                      </span>
                    </div>
                    
                    {error.context && Object.keys(error.context).length > 0 && (
                      <div className={`ml-4 space-y-1 ${colors.context}`}>
                        {renderContext(error.context)}
                      </div>
                    )}
                    
                    {error.suggestion && (
                      <div className={`ml-4 p-2 rounded ${colors.suggestion} text-sm`}>
                        <strong>💡 Suggestion: </strong>
                        {error.suggestion}
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
        
        {dismissible && onClose && (
          <div className="ml-auto pl-3">
            <div className="-mx-1.5 -my-1.5">
              <button
                type="button"
                className={`inline-flex rounded-md p-1.5 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600 ${colors.detail}`}
                onClick={onClose}
              >
                <span className="sr-only">Dismiss</span>
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ErrorAlert;