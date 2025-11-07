'use client';

import React from 'react';
import { useTheme } from '@/contexts/SimpleThemeContext';
import { useColorMode } from '@chakra-ui/react';

interface SimpleThemeToggleProps {
  className?: string;
}

const SimpleThemeToggle: React.FC<SimpleThemeToggleProps> = ({ className = '' }) => {
  const { theme, toggleTheme, mounted } = useTheme();
  const { setColorMode } = useColorMode();

  const handleToggle = () => {
    toggleTheme();
    // Sync with Chakra UI color mode
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setColorMode(newTheme);
  };

  if (!mounted) {
    return (
      <button
        className={`btn btn-ghost ${className}`}
        disabled
      >
        🌙
      </button>
    );
  }

  return (
    <button
      onClick={handleToggle}
      className={`btn btn-ghost theme-toggle-icon ${className}`}
      title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
      aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      {theme === 'dark' ? '🌞' : '🌙'}
    </button>
  );
};

export default SimpleThemeToggle;
