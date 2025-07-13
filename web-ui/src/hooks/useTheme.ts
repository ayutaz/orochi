import { useState, useEffect, useMemo } from 'react';
import { createTheme } from '@mui/material/styles';

export const useTheme = () => {
  const [isDarkMode, setIsDarkMode] = useState(() => {
    const saved = localStorage.getItem('darkMode');
    return saved ? JSON.parse(saved) : true;
  });

  useEffect(() => {
    localStorage.setItem('darkMode', JSON.stringify(isDarkMode));
  }, [isDarkMode]);

  const toggleTheme = () => {
    setIsDarkMode(!isDarkMode);
  };

  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode: isDarkMode ? 'dark' : 'light',
          primary: {
            main: '#4CAF50',
          },
          secondary: {
            main: '#ff9800',
          },
          error: {
            main: '#f44336',
          },
          background: {
            default: isDarkMode ? '#1a1a1a' : '#fafafa',
            paper: isDarkMode ? '#2d2d2d' : '#ffffff',
          },
        },
        shape: {
          borderRadius: 8,
        },
        components: {
          MuiButton: {
            styleOverrides: {
              root: {
                textTransform: 'none',
              },
            },
          },
        },
      }),
    [isDarkMode]
  );

  return { theme, isDarkMode, toggleTheme };
};
