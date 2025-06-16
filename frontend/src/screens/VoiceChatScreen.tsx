import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  StatusBar,
  SafeAreaView,
  ScrollView,
  Alert,
} from 'react-native';
import { Button, Card, Title, Paragraph, ActivityIndicator } from 'react-native-paper';
import LinearGradient from 'react-native-linear-gradient';
import { useVoiceChat } from '../hooks/useVoiceChat';
import { theme } from '../styles/theme';
import LogsView from '../components/LogsView';
import UserInfoCard from '../components/UserInfoCard';
import RemoteAudioView from '../components/RemoteAudioView';

const VoiceChatScreen: React.FC = () => {
  const {
    isConnected,
    isMatched,
    isVoiceConnected,
    userId,
    partnerId,
    status,
    statusType,
    remoteStream,
    connect,
    disconnect,
    findMatch,
    logs,
  } = useVoiceChat();

  const handleConnect = async () => {
    try {
      await connect();
    } catch (error) {
      Alert.alert('Connection Failed', 'Failed to connect to the voice chat service. Please check your microphone permissions and network connection.');
    }
  };

  const handleDisconnect = () => {
    Alert.alert(
      'Disconnect',
      'Are you sure you want to leave the voice chat?',
      [
        {
          text: 'Cancel',
          style: 'cancel',
        },
        {
          text: 'Leave',
          style: 'destructive',
          onPress: disconnect,
        },
      ]
    );
  };

  const getStatusColor = () => {
    switch (statusType) {
      case 'connected':
        return theme.colors.connected;
      case 'waiting':
        return theme.colors.waiting;
      case 'error':
        return theme.colors.error;
      default:
        return theme.colors.info;
    }
  };

  const renderStatusIcon = () => {
    if (statusType === 'waiting') {
      return <ActivityIndicator size="small" color={theme.colors.onSurface} style={styles.statusIcon} />;
    }
    return null;
  };

  return (
    <LinearGradient
      colors={[theme.colors.primary, theme.colors.secondary]}
      style={styles.container}
    >
      <StatusBar
        barStyle="light-content"
        backgroundColor="transparent"
        translucent
      />
      <SafeAreaView style={styles.safeArea}>
        <ScrollView contentContainerStyle={styles.scrollContent}>
          {/* Header */}
          <View style={styles.header}>
            <Title style={styles.title}>🎙️ Voice Chat</Title>
          </View>

          {/* Status Card */}
          <Card style={[styles.statusCard, { backgroundColor: getStatusColor() + '40' }]}>
            <Card.Content style={styles.statusContent}>
              {renderStatusIcon()}
              <Paragraph style={styles.statusText}>{status}</Paragraph>
            </Card.Content>
          </Card>

          {/* User Info */}
          {userId && (
            <UserInfoCard
              userId={userId}
              partnerId={partnerId}
              isMatched={isMatched}
              isVoiceConnected={isVoiceConnected}
            />
          )}

          {/* Control Buttons */}
          <View style={styles.buttonContainer}>
            <Button
              mode="contained"
              onPress={handleConnect}
              disabled={isConnected}
              style={[
                styles.button,
                isConnected && styles.buttonDisabled,
              ]}
              contentStyle={styles.buttonContent}
              labelStyle={styles.buttonLabel}
            >
              {isConnected ? 'Connected' : 'Join Chat'}
            </Button>

            <Button
              mode="contained"
              onPress={findMatch}
              disabled={!isConnected || isMatched}
              style={[
                styles.button,
                (!isConnected || isMatched) && styles.buttonDisabled,
              ]}
              contentStyle={styles.buttonContent}
              labelStyle={styles.buttonLabel}
            >
              {isMatched ? 'Matched' : 'Find Match'}
            </Button>

            <Button
              mode="contained"
              onPress={handleDisconnect}
              disabled={!isConnected}
              style={[
                styles.button,
                styles.leaveButton,
                !isConnected && styles.buttonDisabled,
              ]}
              contentStyle={styles.buttonContent}
              labelStyle={styles.buttonLabel}
            >
              Leave
            </Button>
          </View>

          {/* Remote Audio */}
          {remoteStream && (
            <RemoteAudioView stream={remoteStream} />
          )}

          {/* Logs */}
          <LogsView logs={logs} />
        </ScrollView>
      </SafeAreaView>
    </LinearGradient>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  safeArea: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    padding: theme.spacing.lg,
  },
  header: {
    alignItems: 'center',
    marginBottom: theme.spacing.xl,
    marginTop: theme.spacing.lg,
  },
  title: {
    ...theme.typography.h1,
    color: theme.colors.onBackground,
    textAlign: 'center',
  },
  statusCard: {
    marginBottom: theme.spacing.lg,
    borderRadius: theme.borderRadius.lg,
    ...theme.shadows.medium,
  },
  statusContent: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 60,
    paddingVertical: theme.spacing.md,
  },
  statusIcon: {
    marginRight: theme.spacing.sm,
  },
  statusText: {
    ...theme.typography.body,
    color: theme.colors.onSurface,
    textAlign: 'center',
    flex: 1,
  },
  buttonContainer: {
    marginBottom: theme.spacing.lg,
  },
  button: {
    marginBottom: theme.spacing.md,
    borderRadius: theme.borderRadius.round,
    backgroundColor: theme.colors.surfaceVariant,
    ...theme.shadows.small,
  },
  buttonDisabled: {
    backgroundColor: theme.colors.overlay,
  },
  leaveButton: {
    backgroundColor: theme.colors.error + '80',
  },
  buttonContent: {
    height: theme.dimensions.buttonHeight,
  },
  buttonLabel: {
    ...theme.typography.button,
    color: theme.colors.onSurface,
  },
});

export default VoiceChatScreen; 