import { createTheme } from '@mui/material/styles';
import type { ThemeOptions } from '@mui/material/styles';

// Aviasales-inspired color palette
const aviPalette = {
  primary: {
    main: '#0071ce', // Blue similar to Aviasales primary
    light: '#3392e3',
    dark: '#005199',
    contrastText: '#ffffff',
  },
  secondary: {
    main: '#f39c12', // Orange accent color
    light: '#f5b041',
    dark: '#d35400',
    contrastText: '#ffffff',
  },
  accent: {
    main: '#e74c3c', // Red accent for highlights
    light: '#ec7063',
    dark: '#c0392b',
  },
  neutral: {
    main: '#34495e',
    light: '#5d6d7e',
    dark: '#2c3e50',
    contrastText: '#ffffff',
  },
  background: {
    light: '#f8f9fa',
    main: '#ffffff',
  }
};

// Create the extended theme
const themeOptions: ThemeOptions = {
  palette: {
    primary: {
      main: aviPalette.primary.main,
      light: aviPalette.primary.light,
      dark: aviPalette.primary.dark,
      contrastText: aviPalette.primary.contrastText,
    },
    secondary: {
      main: aviPalette.secondary.main,
      light: aviPalette.secondary.light,
      dark: aviPalette.secondary.dark,
      contrastText: aviPalette.secondary.contrastText,
    },
    background: {
      default: aviPalette.background.light,
      paper: aviPalette.background.main,
    },
    text: {
      primary: aviPalette.neutral.dark,
      secondary: '#7f8c8d',
    },
    error: {
      main: aviPalette.accent.main,
      light: aviPalette.accent.light,
      dark: aviPalette.accent.dark,
    },
  },
  typography: {
    fontFamily: [
      '"SF Pro Display"', 
      '"Helvetica Neue"',
      'Arial',
      'sans-serif',
      '"Apple Color Emoji"',
      '"Segoe UI Emoji"',
      '"Segoe UI Symbol"',
    ].join(','),
    h1: {
      fontSize: '2.5rem',
      fontWeight: 600,
      lineHeight: 1.2,
      letterSpacing: '-0.02em',
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 600,
      lineHeight: 1.3,
    },
    h3: {
      fontSize: '1.75rem',
      fontWeight: 600,
      lineHeight: 1.3,
    },
    h4: {
      fontSize: '1.5rem',
      fontWeight: 500,
      lineHeight: 1.4,
    },
    h5: {
      fontSize: '1.25rem',
      fontWeight: 500,
      lineHeight: 1.4,
    },
    h6: {
      fontSize: '1.125rem',
      fontWeight: 500,
      lineHeight: 1.4,
    },
    subtitle1: {
      fontSize: '1rem',
      fontWeight: 400,
      lineHeight: 1.5,
    },
    body1: {
      fontSize: '0.9375rem',
      lineHeight: 1.6,
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.5,
    },
    button: {
      textTransform: 'none',
      fontWeight: 500,
      fontSize: '1rem',
    },
  },
  shape: {
    borderRadius: 8,
  },
  spacing: 8,
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 24,
          padding: '10px 24px',
          fontWeight: 500,
          boxShadow: 'none',
          '&:hover': {
            boxShadow: 'none',
          },
        },
        contained: {
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          '&:hover': {
            boxShadow: '0 4px 8px rgba(0,0,0,0.15)',
          },
        },
        sizeLarge: {
          padding: '14px 32px',
          fontSize: '1.125rem',
        },
        sizeSmall: {
          padding: '6px 16px',
          fontSize: '0.8125rem',
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
          borderRadius: '12px',
          border: '1px solid rgba(0,0,0,0.05)',
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          background: '#ffffff',
          borderRadius: '8px',
          '& .MuiOutlinedInput-root': {
            borderRadius: '8px',
            '& fieldset': {
              borderColor: '#e0e0e0',
              borderWidth: 1,
            },
            '&:hover fieldset': {
              borderColor: aviPalette.primary.main,
            },
            '&.Mui-focused fieldset': {
              borderColor: aviPalette.primary.main,
              borderWidth: 1,
            },
          },
        },
      },
    },
    MuiInputBase: {
      styleOverrides: {
        input: {
          fontSize: '1rem',
          padding: '12px 16px',
        },
      },
    },
    MuiAutocomplete: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            padding: 0,
            '& input': {
              padding: '12px 16px',
            },
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none', // Remove default MUI gradient
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          boxShadow: '0 2px 10px rgba(0,0,0,0.05)',
          backgroundColor: '#ffffff',
          color: aviPalette.neutral.main,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: '6px',
        },
        colorPrimary: {
          backgroundColor: `${aviPalette.primary.light}20`, // With opacity
          color: aviPalette.primary.dark,
        },
      },
    },
    MuiFormControl: {
      styleOverrides: {
        root: {
          '& .MuiInputBase-input': {
            height: '1.5em',
          },
        },
      },
    },
  },
};

const theme = createTheme(themeOptions);

export default theme;