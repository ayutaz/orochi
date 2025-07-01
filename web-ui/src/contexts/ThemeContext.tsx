import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { ThemeProvider, createTheme } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'

interface ThemeContextType {
  darkMode: boolean
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

export const useTheme = () => {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}

interface ThemeProviderProps {
  children: ReactNode
}

export const CustomThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [darkMode, setDarkMode] = useState(() => {
    const saved = localStorage.getItem('darkMode')
    return saved ? JSON.parse(saved) : false
  })

  useEffect(() => {
    localStorage.setItem('darkMode', JSON.stringify(darkMode))
  }, [darkMode])

  const toggleTheme = () => {
    setDarkMode(!darkMode)
  }

  const theme = createTheme({
    palette: {
      mode: darkMode ? 'dark' : 'light',
      primary: {
        main: '#1976d2',
      },
      secondary: {
        main: '#dc004e',
      },
      background: {
        default: darkMode ? '#121212' : '#fafafa',
        paper: darkMode ? '#1e1e1e' : '#ffffff',
      },
    },
    typography: {
      fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            scrollbarColor: darkMode ? '#6b6b6b #2b2b2b' : '#959595 #f1f1f1',
            '&::-webkit-scrollbar, & *::-webkit-scrollbar': {
              width: 8,
              height: 8,
            },
            '&::-webkit-scrollbar-thumb, & *::-webkit-scrollbar-thumb': {
              borderRadius: 8,
              backgroundColor: darkMode ? '#6b6b6b' : '#959595',
              minHeight: 24,
              border: '2px solid transparent',
            },
            '&::-webkit-scrollbar-thumb:hover, & *::-webkit-scrollbar-thumb:hover': {
              backgroundColor: darkMode ? '#959595' : '#6b6b6b',
            },
            '&::-webkit-scrollbar-track, & *::-webkit-scrollbar-track': {
              backgroundColor: darkMode ? '#2b2b2b' : '#f1f1f1',
            },
          },
        },
      },
    },
  })

  return (
    <ThemeContext.Provider value={{ darkMode, toggleTheme }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeContext.Provider>
  )
}