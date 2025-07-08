/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { StatusBar } from 'react-native';
import { Provider as PaperProvider, MD3DarkTheme } from 'react-native-paper';
import VoiceChatScreen from './src/screens/VoiceChatScreen';
import { theme } from './src/styles/theme';

// Create a custom dark theme for React Native Paper
const paperTheme = {
  ...MD3DarkTheme,
  colors: {
    ...MD3DarkTheme.colors,
    primary: theme.colors.primary,
    secondary: theme.colors.secondary,
    surface: theme.colors.surface,
    surfaceVariant: theme.colors.surfaceVariant,
    background: theme.colors.background,
    onSurface: theme.colors.onSurface,
    onSurfaceVariant: theme.colors.onSurfaceVariant,
    onBackground: theme.colors.onBackground,
    outline: theme.colors.outline,
    outlineVariant: theme.colors.outlineVariant,
  },
};

function App() {
  return (
    <PaperProvider theme={paperTheme}>
      <StatusBar
        barStyle="light-content"
        backgroundColor="transparent"
        translucent
      />
      <VoiceChatScreen />
    </PaperProvider>
  );
}

export default App;
