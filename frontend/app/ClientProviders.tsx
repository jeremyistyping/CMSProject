'use client';

import React, { useEffect, useState } from 'react';
import { ChakraProvider, ColorModeScript } from '@chakra-ui/react';
import { ThemeProvider } from '@/contexts/SimpleThemeContext';
import { AuthProvider } from '@/contexts/AuthContext';
import { LanguageProvider } from '@/contexts/LanguageContext';
import theme from '@/theme';

interface ClientProvidersProps {
  children: React.ReactNode;
}

export default function ClientProviders({ children }: ClientProvidersProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    // Return null during SSR to avoid hydration mismatches
    // The actual content will render after hydration
    return null;
  }

  return (
    <ChakraProvider 
      theme={theme}
      toastOptions={{
        defaultOptions: {
          position: 'top-right',
          duration: 3000,
          isClosable: true,
          variant: 'solid',
        },
      }}
    >
      <ThemeProvider>
        <LanguageProvider>
          <AuthProvider>
            {children}
          </AuthProvider>
        </LanguageProvider>
      </ThemeProvider>
    </ChakraProvider>
  );
}
