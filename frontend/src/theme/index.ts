import { extendTheme, type ThemeConfig } from '@chakra-ui/react';

const config: ThemeConfig = {
  initialColorMode: 'light',
  useSystemColorMode: true,
  disableTransitionOnChange: false,
};

const theme = extendTheme({
  config,
  colors: {
    brand: {
      50: '#ECFDF5',
      100: '#D1FAE5', 
      200: '#A7F3D0',
      300: '#6EE7B7',
      400: '#34D399',
      500: '#10B981', // Primary green color
      600: '#059669',
      700: '#047857',
      800: '#065F46',
      900: '#064E3B',
    },
  },
  fonts: {
    heading: 'Inter, system-ui, sans-serif',
    body: 'Inter, system-ui, sans-serif',
  },
  styles: {
    global: (props) => ({
      body: {
        bg: props.colorMode === 'dark' ? 'var(--bg-primary)' : 'var(--bg-secondary)',
        color: props.colorMode === 'dark' ? 'var(--text-primary)' : 'var(--text-primary)',
        transition: 'background-color 0.3s ease, color 0.3s ease',
      },
      '*::placeholder': {
        color: props.colorMode === 'dark' ? 'var(--text-secondary)' : 'var(--text-secondary)',
      },
      '*, *::before, &::after': {
        borderColor: props.colorMode === 'dark' ? 'var(--border-color)' : 'var(--border-color)',
      },
    }),
  },
  components: {
    Button: {
      defaultProps: {
        colorScheme: 'brand',
      },
      variants: {
        solid: {
          fontWeight: 'medium',
          _hover: {
            transform: 'translateY(-1px)',
            boxShadow: 'lg',
          },
        },
      },
    },
    Card: {
      baseStyle: (props) => ({
        container: {
          borderRadius: 'lg',
          overflow: 'hidden',
          border: '1px',
          borderColor: props.colorMode === 'dark' ? 'var(--border-color)' : 'gray.200',
          bg: props.colorMode === 'dark' ? 'var(--bg-secondary)' : 'white',
          boxShadow: props.colorMode === 'dark' ? 'var(--shadow)' : 'sm',
          transition: 'all 0.3s ease',
        },
      }),
    },
    Table: {
      variants: {
        simple: (props) => ({
          th: {
            fontWeight: '500',
            fontSize: 'sm',
            color: props.colorMode === 'dark' ? 'var(--text-primary)' : 'gray.600',
            textTransform: 'none',
            borderColor: props.colorMode === 'dark' ? 'var(--border-color)' : 'gray.200',
            bg: props.colorMode === 'dark' ? 'var(--bg-tertiary)' : 'gray.50',
          },
          td: {
            borderColor: props.colorMode === 'dark' ? 'var(--border-color)' : 'gray.200',
            fontSize: 'sm',
            color: props.colorMode === 'dark' ? 'var(--text-primary)' : 'gray.800',
          },
          tbody: {
            tr: {
              _hover: {
                bg: props.colorMode === 'dark' ? 'var(--bg-tertiary)' : 'gray.50',
              },
            },
          },
        }),
      },
    },
    Heading: {
      baseStyle: {
        fontWeight: '600',
        letterSpacing: 'tight',
      },
    },
    Text: {
      baseStyle: {
        lineHeight: 'base',
      },
    },
    Alert: {
      baseStyle: {
        container: {
          bg: 'white',
          color: 'gray.800',
          borderRadius: 'md',
        },
      },
      variants: {
        solid: {
          container: {
            bg: 'white !important',
            color: 'gray.800 !important',
          },
        },
        'left-accent': {
          container: {
            bg: 'white !important',
            color: 'gray.800 !important',
          },
        },
      },
    },
  },
});

export default theme;
