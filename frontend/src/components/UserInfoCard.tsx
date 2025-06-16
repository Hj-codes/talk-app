import React from 'react';
import { View, StyleSheet } from 'react-native';
import { Card, Text, Chip } from 'react-native-paper';
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

  return (
    <Card style={styles.card}>
      <Card.Content>
        <View style={styles.row}>
          <Text style={styles.label}>User ID:</Text>
          <Text style={styles.value}>{userId.substring(0, 8)}...</Text>
        </View>
        
        <View style={styles.row}>
          <Text style={styles.label}>Partner:</Text>
          <Text style={styles.value}>
            {partnerId ? `${partnerId.substring(0, 8)}...` : 'None'}
          </Text>
        </View>

        <View style={styles.statusRow}>
          <Chip
            style={[styles.statusChip, { backgroundColor: getStatusChipColor() + '40' }]}
            textStyle={[styles.statusText, { color: theme.colors.onSurface }]}
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
    marginBottom: theme.spacing.lg,
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.lg,
    ...theme.shadows.small,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: theme.spacing.sm,
  },
  statusRow: {
    alignItems: 'center',
    marginTop: theme.spacing.sm,
  },
  label: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurfaceVariant,
    fontWeight: '500',
  },
  value: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurface,
    fontFamily: 'monospace',
  },
  statusChip: {
    borderRadius: theme.borderRadius.round,
  },
  statusText: {
    ...theme.typography.caption,
    fontWeight: '500',
  },
});

export default UserInfoCard; 