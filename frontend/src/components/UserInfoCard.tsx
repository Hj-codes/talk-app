import React from 'react';
import { View, StyleSheet } from 'react-native';
import { Card, Text, Chip, Divider } from 'react-native-paper';
import { theme } from '../styles/theme';

interface UserInfoCardProps {
  userId: string;
  partnerId: string | null;
  isMatched: boolean;
  isVoiceConnected: boolean;
}

const UserInfoCard: React.FC<UserInfoCardProps> = ({
  userId,
  partnerId,
  isMatched,
  isVoiceConnected,
}) => {
  const getStatusChipColor = () => {
    if (isVoiceConnected) return theme.colors.connected;
    if (isMatched) return theme.colors.waiting;
    return theme.colors.info;
  };

  const getStatusText = () => {
    if (isVoiceConnected) return 'Voice Connected';
    if (isMatched) return 'Setting up voice...';
    return 'Connected to server';
  };

  const getStatusIcon = () => {
    if (isVoiceConnected) return '🔊';
    if (isMatched) return '⚙️';
    return '📡';
  };

  return (
    <Card style={styles.card} mode="elevated">
      <Card.Content style={styles.content}>
        <Text style={styles.cardTitle}>Session Information</Text>
        
        <View style={styles.infoSection}>
          <View style={styles.row}>
            <Text style={styles.label}>Your ID</Text>
            <Text style={styles.value}>{userId.substring(0, 8)}...</Text>
          </View>
          
          <Divider style={styles.divider} />
          
          <View style={styles.row}>
            <Text style={styles.label}>Partner</Text>
            <Text style={styles.value}>
              {partnerId ? `${partnerId.substring(0, 8)}...` : 'Waiting...'}
            </Text>
          </View>
        </View>

        <View style={styles.statusSection}>
          <Chip
            style={[styles.statusChip, { backgroundColor: getStatusChipColor() + '20' }]}
            textStyle={[styles.statusText, { color: getStatusChipColor() }]}
            icon={() => <Text style={styles.statusIcon}>{getStatusIcon()}</Text>}
            mode="flat"
          >
            {getStatusText()}
          </Chip>
        </View>
      </Card.Content>
    </Card>
  );
};

const styles = StyleSheet.create({
  card: {
    marginBottom: theme.spacing.xxxl,
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.lg,
  },
  content: {
    paddingVertical: theme.spacing.xl,
    paddingHorizontal: theme.spacing.lg,
  },
  cardTitle: {
    ...theme.typography.h4,
    color: theme.colors.onSurface,
    marginBottom: theme.spacing.lg,
    textAlign: 'center',
  },
  infoSection: {
    marginBottom: theme.spacing.xl,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: theme.spacing.md,
  },
  divider: {
    backgroundColor: theme.colors.divider,
    height: 1,
  },
  label: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurfaceVariant,
    fontWeight: '500',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  value: {
    ...theme.typography.body,
    color: theme.colors.onSurface,
    fontFamily: 'monospace',
    fontWeight: '600',
  },
  statusSection: {
    alignItems: 'center',
  },
  statusChip: {
    borderRadius: theme.borderRadius.round,
    paddingHorizontal: theme.spacing.md,
  },
  statusText: {
    ...theme.typography.bodySmall,
    fontWeight: '600',
    marginLeft: theme.spacing.xs,
  },
  statusIcon: {
    fontSize: 14,
  },
});

export default UserInfoCard; 