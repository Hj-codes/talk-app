/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { StatusBar } from 'react-native';
import { Provider as PaperProvider, MD3LightTheme } from 'react-native-paper';
import VoiceChatScreen from './src/screens/VoiceChatScreen';
import { theme } from './src/styles/theme';

// Create a custom theme for React Native Paper
const paperTheme = {
  ...MD3LightTheme,
  colors: {
    ...MD3LightTheme.colors,
    primary: theme.colors.primary,
    secondary: theme.colors.secondary,
    surface: theme.colors.surface,
    background: theme.colors.background,
    onSurface: theme.colors.onSurface,
    onBackground: theme.colors.onBackground,
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
