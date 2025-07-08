import { Dimensions } from 'react-native';

const { width, height } = Dimensions.get('window');

export const theme = {
  colors: {
    // Material Design 3 Dark Theme - Professional & Classy
    primary: '#BB86FC',
    primaryVariant: '#985EFF',
    secondary: '#03DAC6',
    secondaryVariant: '#00B4A6',
    
    // Background colors - Deep, sophisticated
    background: '#0F0F0F',
    surface: '#1E1E1E',
    surfaceVariant: '#2A2A2A',
    surfaceContainerHighest: '#353535',
    overlay: 'rgba(0, 0, 0, 0.6)',
    
    // Status colors - Professional with proper contrast
    connected: '#4CAF50',
    waiting: '#FF9800',
    error: '#CF6679',
    info: '#82B1FF',
    warning: '#FFB74D',
    
    // Text colors - High contrast for readability
    onSurface: '#FFFFFF',
    onSurfaceVariant: '#B3B3B3',
    onBackground: '#FFFFFF',
    onPrimary: '#000000',
    onSecondary: '#000000',
    onError: '#000000',
    
    // Border and divider colors
    border: 'rgba(255, 255, 255, 0.12)',
    divider: 'rgba(255, 255, 255, 0.08)',
    outline: 'rgba(255, 255, 255, 0.16)',
    outlineVariant: 'rgba(255, 255, 255, 0.08)',
  },
  
  spacing: {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
    xxl: 48,
    xxxl: 64,
  },
  
  borderRadius: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24,
    round: 50,
  },
  
  typography: {
    h1: {
      fontSize: 32,
      fontWeight: '300' as const,
      lineHeight: 40,
      letterSpacing: -0.5,
    },
    h2: {
      fontSize: 28,
      fontWeight: '300' as const,
      lineHeight: 36,
      letterSpacing: -0.25,
    },
    h3: {
      fontSize: 24,
      fontWeight: '400' as const,
      lineHeight: 32,
      letterSpacing: 0,
    },
    h4: {
      fontSize: 20,
      fontWeight: '500' as const,
      lineHeight: 28,
      letterSpacing: 0,
    },
    body: {
      fontSize: 16,
      fontWeight: '400' as const,
      lineHeight: 24,
      letterSpacing: 0.15,
    },
    bodySmall: {
      fontSize: 14,
      fontWeight: '400' as const,
      lineHeight: 20,
      letterSpacing: 0.25,
    },
    caption: {
      fontSize: 12,
      fontWeight: '400' as const,
      lineHeight: 16,
      letterSpacing: 0.4,
    },
    button: {
      fontSize: 16,
      fontWeight: '500' as const,
      lineHeight: 24,
      letterSpacing: 0.1,
    },
    overline: {
      fontSize: 10,
      fontWeight: '500' as const,
      lineHeight: 16,
      letterSpacing: 1.5,
      textTransform: 'uppercase' as const,
    },
  },
  
  shadows: {
    none: {
      shadowColor: 'transparent',
      shadowOffset: { width: 0, height: 0 },
      shadowOpacity: 0,
      shadowRadius: 0,
      elevation: 0,
    },
    small: {
      shadowColor: '#000000',
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.25,
      shadowRadius: 4,
      elevation: 2,
    },
    medium: {
      shadowColor: '#000000',
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 4,
    },
    large: {
      shadowColor: '#000000',
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.35,
      shadowRadius: 16,
      elevation: 8,
    },
    xl: {
      shadowColor: '#000000',
      shadowOffset: { width: 0, height: 12 },
      shadowOpacity: 0.4,
      shadowRadius: 24,
      elevation: 12,
    },
  },
  
  dimensions: {
    screenWidth: width,
    screenHeight: height,
    buttonHeight: 56,
    inputHeight: 56,
    cardMinHeight: 120,
  },
  
  // Animation durations for consistent motion
  animation: {
    fast: 150,
    normal: 250,
    slow: 400,
  },
};

export type Theme = typeof theme; 