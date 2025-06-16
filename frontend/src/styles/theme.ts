import { Dimensions } from 'react-native';

const { width, height } = Dimensions.get('window');

export const theme = {
  colors: {
    // Gradient colors similar to HTML demo
    primary: '#667eea',
    secondary: '#764ba2',
    
    // Status colors
    connected: '#4CAF50',
    waiting: '#FF9800',
    error: '#F44336',
    info: '#2196F3',
    
    // UI colors
    background: '#667eea',
    surface: 'rgba(255, 255, 255, 0.1)',
    surfaceVariant: 'rgba(255, 255, 255, 0.2)',
    overlay: 'rgba(0, 0, 0, 0.3)',
    
    // Text colors
    onSurface: '#FFFFFF',
    onSurfaceVariant: 'rgba(255, 255, 255, 0.8)',
    onBackground: '#FFFFFF',
    
    // Border colors
    border: 'rgba(255, 255, 255, 0.3)',
    divider: 'rgba(255, 255, 255, 0.1)',
  },
  
  spacing: {
    xs: 4,
    sm: 8,
    md: 16,
    lg: 24,
    xl: 32,
    xxl: 48,
  },
  
  borderRadius: {
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    round: 25,
  },
  
  typography: {
    h1: {
      fontSize: 32,
      fontWeight: '300' as const,
      lineHeight: 40,
    },
    h2: {
      fontSize: 24,
      fontWeight: '400' as const,
      lineHeight: 32,
    },
    h3: {
      fontSize: 20,
      fontWeight: '500' as const,
      lineHeight: 28,
    },
    body: {
      fontSize: 16,
      fontWeight: '400' as const,
      lineHeight: 24,
    },
    bodySmall: {
      fontSize: 14,
      fontWeight: '400' as const,
      lineHeight: 20,
    },
    caption: {
      fontSize: 12,
      fontWeight: '400' as const,
      lineHeight: 16,
    },
    button: {
      fontSize: 16,
      fontWeight: '500' as const,
      lineHeight: 24,
    },
  },
  
  shadows: {
    small: {
      shadowColor: '#000',
      shadowOffset: {
        width: 0,
        height: 2,
      },
      shadowOpacity: 0.1,
      shadowRadius: 3,
      elevation: 2,
    },
    medium: {
      shadowColor: '#000',
      shadowOffset: {
        width: 0,
        height: 4,
      },
      shadowOpacity: 0.15,
      shadowRadius: 6,
      elevation: 4,
    },
    large: {
      shadowColor: '#000',
      shadowOffset: {
        width: 0,
        height: 8,
      },
      shadowOpacity: 0.2,
      shadowRadius: 12,
      elevation: 8,
    },
  },
  
  dimensions: {
    screenWidth: width,
    screenHeight: height,
    buttonHeight: 48,
    inputHeight: 56,
  },
};

export type Theme = typeof theme; 