import React from 'react';
import {
  View,
  StyleSheet,
  StatusBar,
  SafeAreaView,
  ScrollView,
  Alert,
} from 'react-native';
import { 
  Button, 
  Card, 
  Text, 
  ActivityIndicator, 
  Surface, 
  IconButton,
  Chip,
  Divider,
  Avatar 
} from 'react-native-paper';
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

  const getStatusIcon = () => {
    switch (statusType) {
      case 'connected':
        return '🟢';
      case 'waiting':
        return '🟡';
      case 'error':
        return '🔴';
      default:
        return '🔵';
    }
  };

  const renderConnectionStatusCard = () => (
    <Card style={styles.connectionCard} mode="elevated">
      <Card.Content style={styles.connectionContent}>
        <View style={styles.connectionHeader}>
          <View style={styles.statusIndicatorContainer}>
            <View style={[styles.statusDot, { backgroundColor: getStatusColor() }]} />
            <Text style={styles.connectionTitle}>Connection Status</Text>
          </View>
          {statusType === 'waiting' && (
            <ActivityIndicator size="small" color={getStatusColor()} />
          )}
        </View>
        
        <Text style={[styles.connectionStatus, { color: getStatusColor() }]}>
          {status}
        </Text>
        
        <View style={styles.statusChipsContainer}>
          <Chip 
            style={[styles.statusChip, { backgroundColor: isConnected ? theme.colors.connected + '20' : theme.colors.error + '20' }]}
            textStyle={[styles.statusChipText, { color: isConnected ? theme.colors.connected : theme.colors.error }]}
            icon={() => <Text style={styles.statusChipIcon}>{isConnected ? '📡' : '❌'}</Text>}
            compact
          >
            {isConnected ? 'Server Connected' : 'Disconnected'}
          </Chip>
          
          {isMatched && (
            <Chip 
              style={[styles.statusChip, { backgroundColor: theme.colors.secondary + '20' }]}
              textStyle={[styles.statusChipText, { color: theme.colors.secondary }]}
              icon={() => <Text style={styles.statusChipIcon}>👥</Text>}
              compact
            >
              Partner Found
            </Chip>
          )}
          
          {isVoiceConnected && (
            <Chip 
              style={[styles.statusChip, { backgroundColor: theme.colors.primary + '20' }]}
              textStyle={[styles.statusChipText, { color: theme.colors.primary }]}
              icon={() => <Text style={styles.statusChipIcon}>🔊</Text>}
              compact
            >
              Voice Active
            </Chip>
          )}
        </View>
      </Card.Content>
    </Card>
  );

  const renderUserSessionCard = () => {
    if (!userId) return null;
    
    return (
      <Card style={styles.sessionCard} mode="elevated">
        <Card.Content style={styles.sessionContent}>
          <View style={styles.sessionHeader}>
            <Text style={styles.sessionTitle}>Current Session</Text>
            <IconButton 
              icon="information-outline" 
              size={20} 
              iconColor={theme.colors.onSurfaceVariant}
              style={styles.infoButton}
            />
          </View>
          
          <View style={styles.userInfoRow}>
            <Avatar.Text 
              size={40} 
              label={userId.substring(0, 2).toUpperCase()} 
              style={styles.userAvatar}
              labelStyle={styles.avatarLabel}
            />
            <View style={styles.userDetails}>
              <Text style={styles.userLabel}>Your ID</Text>
              <Text style={styles.userValue}>{userId.substring(0, 12)}...</Text>
            </View>
          </View>
          
          <Divider style={styles.sessionDivider} />
          
          <View style={styles.userInfoRow}>
            <Avatar.Text 
              size={40} 
              label={partnerId ? partnerId.substring(0, 2).toUpperCase() : '?'} 
              style={[styles.userAvatar, { backgroundColor: partnerId ? theme.colors.secondary : theme.colors.surfaceVariant }]}
              labelStyle={styles.avatarLabel}
            />
            <View style={styles.userDetails}>
              <Text style={styles.userLabel}>Partner</Text>
              <Text style={styles.userValue}>
                {partnerId ? `${partnerId.substring(0, 12)}...` : 'Waiting for match...'}
              </Text>
            </View>
          </View>
        </Card.Content>
      </Card>
    );
  };

  const renderControlPanel = () => (
    <Card style={styles.controlCard} mode="elevated">
      <Card.Content style={styles.controlContent}>
        <Text style={styles.controlTitle}>Voice Chat Controls</Text>
        
        <View style={styles.primaryButtonContainer}>
          <Button
            mode="contained"
            onPress={handleConnect}
            disabled={isConnected}
            style={[
              styles.primaryButton,
              isConnected && styles.buttonDisabled,
            ]}
            contentStyle={styles.primaryButtonContent}
            labelStyle={styles.primaryButtonLabel}
            icon={isConnected ? 'check-circle' : 'microphone'}
          >
            {isConnected ? 'Connected to Server' : 'Join Voice Chat'}
          </Button>
        </View>
        
        <View style={styles.secondaryButtonsContainer}>
          <Button
            mode="contained-tonal"
            onPress={findMatch}
            disabled={!isConnected || isMatched}
            style={[
              styles.secondaryButton,
              (!isConnected || isMatched) && styles.buttonDisabled,
            ]}
            contentStyle={styles.secondaryButtonContent}
            labelStyle={styles.secondaryButtonLabel}
            icon={isMatched ? 'account-check' : 'account-search'}
          >
            {isMatched ? 'Partner Matched' : 'Find Partner'}
          </Button>
          
          <Button
            mode="outlined"
            onPress={handleDisconnect}
            disabled={!isConnected}
            style={[
              styles.leaveButton,
              !isConnected && styles.buttonDisabled,
            ]}
            contentStyle={styles.secondaryButtonContent}
            labelStyle={styles.leaveButtonLabel}
            icon="exit-to-app"
          >
            Leave Chat
          </Button>
        </View>
      </Card.Content>
    </Card>
  );

  return (
    <View style={styles.container}>
      <StatusBar
        barStyle="light-content"
        backgroundColor={theme.colors.background}
        translucent={false}
      />
      <SafeAreaView style={styles.safeArea}>
        <ScrollView 
          contentContainerStyle={styles.scrollContent} 
          showsVerticalScrollIndicator={false}
          bounces={false}
        >
          {/* Modern Header */}
          <View style={styles.headerContainer}>
            <Text style={styles.appTitle}>Voice Chat</Text>
            <Text style={styles.appSubtitle}>Connect with people worldwide through voice</Text>
          </View>

          {/* Connection Status */}
          {renderConnectionStatusCard()}

          {/* User Session Info */}
          {renderUserSessionCard()}

          {/* Control Panel */}
          {renderControlPanel()}

          {/* Remote Audio */}
          {remoteStream && (
            <RemoteAudioView stream={remoteStream} />
          )}

          {/* Activity Logs */}
          {logs.length > 0 && (
            <LogsView logs={logs} />
          )}
        </ScrollView>
      </SafeAreaView>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: theme.colors.background,
  },
  safeArea: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    paddingHorizontal: theme.spacing.lg,
    paddingTop: theme.spacing.md,
    paddingBottom: theme.spacing.xxxl,
  },
  
  // Modern Header Styles
  headerContainer: {
    alignItems: 'center',
    paddingVertical: theme.spacing.xxxl,
    marginBottom: theme.spacing.lg,
  },
  appTitle: {
    ...theme.typography.h1,
    color: theme.colors.onBackground,
    fontWeight: '700',
    textAlign: 'center',
    marginBottom: theme.spacing.sm,
  },
  appSubtitle: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurfaceVariant,
    textAlign: 'center',
    opacity: 0.9,
    lineHeight: 20,
  },
  
  // Connection Status Card Styles
  connectionCard: {
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.xl,
    marginBottom: theme.spacing.xl,
    ...theme.shadows.medium,
  },
  connectionContent: {
    paddingVertical: theme.spacing.xl,
    paddingHorizontal: theme.spacing.lg,
  },
  connectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: theme.spacing.lg,
  },
  statusIndicatorContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: theme.spacing.sm,
  },
  connectionTitle: {
    ...theme.typography.h4,
    color: theme.colors.onSurface,
    fontWeight: '600',
  },
  connectionStatus: {
    ...theme.typography.body,
    fontWeight: '500',
    marginBottom: theme.spacing.lg,
    lineHeight: 24,
  },
  statusChipsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: theme.spacing.sm,
  },
  statusChip: {
    borderRadius: theme.borderRadius.md,
  },
  statusChipText: {
    ...theme.typography.caption,
    fontWeight: '600',
    fontSize: 11,
  },
  statusChipIcon: {
    fontSize: 10,
  },
  
  // Session Card Styles
  sessionCard: {
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.xl,
    marginBottom: theme.spacing.xl,
    ...theme.shadows.medium,
  },
  sessionContent: {
    paddingVertical: theme.spacing.xl,
    paddingHorizontal: theme.spacing.lg,
  },
  sessionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: theme.spacing.lg,
  },
  sessionTitle: {
    ...theme.typography.h4,
    color: theme.colors.onSurface,
    fontWeight: '600',
  },
  infoButton: {
    margin: 0,
  },
  userInfoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: theme.spacing.md,
  },
  userAvatar: {
    backgroundColor: theme.colors.primary,
    marginRight: theme.spacing.lg,
  },
  avatarLabel: {
    color: theme.colors.onPrimary,
    fontWeight: '700',
    fontSize: 14,
  },
  userDetails: {
    flex: 1,
  },
  userLabel: {
    ...theme.typography.caption,
    color: theme.colors.onSurfaceVariant,
    fontWeight: '500',
    textTransform: 'uppercase',
    letterSpacing: 0.8,
    marginBottom: theme.spacing.xs,
  },
  userValue: {
    ...theme.typography.body,
    color: theme.colors.onSurface,
    fontFamily: 'monospace',
    fontWeight: '600',
  },
  sessionDivider: {
    backgroundColor: theme.colors.divider,
    marginVertical: theme.spacing.sm,
  },
  
  // Control Panel Styles
  controlCard: {
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.xl,
    marginBottom: theme.spacing.xl,
    ...theme.shadows.medium,
  },
  controlContent: {
    paddingVertical: theme.spacing.xl,
    paddingHorizontal: theme.spacing.lg,
  },
  controlTitle: {
    ...theme.typography.h4,
    color: theme.colors.onSurface,
    fontWeight: '600',
    textAlign: 'center',
    marginBottom: theme.spacing.xl,
  },
  primaryButtonContainer: {
    marginBottom: theme.spacing.lg,
  },
  primaryButton: {
    backgroundColor: theme.colors.primary,
    borderRadius: theme.borderRadius.lg,
    ...theme.shadows.small,
  },
  primaryButtonContent: {
    height: theme.dimensions.buttonHeight,
    paddingHorizontal: theme.spacing.lg,
  },
  primaryButtonLabel: {
    ...theme.typography.button,
    color: theme.colors.onPrimary,
    fontWeight: '700',
    fontSize: 16,
  },
  secondaryButtonsContainer: {
    gap: theme.spacing.md,
  },
  secondaryButton: {
    backgroundColor: theme.colors.secondary + '20',
    borderRadius: theme.borderRadius.lg,
  },
  secondaryButtonContent: {
    height: 48,
    paddingHorizontal: theme.spacing.md,
  },
  secondaryButtonLabel: {
    ...theme.typography.button,
    color: theme.colors.secondary,
    fontWeight: '600',
    fontSize: 14,
  },
  leaveButton: {
    borderColor: theme.colors.error,
    borderWidth: 1.5,
    borderRadius: theme.borderRadius.lg,
    backgroundColor: 'transparent',
  },
  leaveButtonLabel: {
    color: theme.colors.error,
    fontWeight: '600',
    fontSize: 14,
  },
  buttonDisabled: {
    backgroundColor: theme.colors.surfaceVariant,
    opacity: 0.5,
  },
});

export default VoiceChatScreen; 