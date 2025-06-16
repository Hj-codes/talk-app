import React, { useRef, useEffect } from 'react';
import { View, StyleSheet, ScrollView } from 'react-native';
import { Card, Text } from 'react-native-paper';
import { theme } from '../styles/theme';

interface LogEntry {
  timestamp: string;
  message: string;
  type: string;
}

interface LogsViewProps {
  logs: LogEntry[];
}

const LogsView: React.FC<LogsViewProps> = ({ logs }) => {
  const scrollViewRef = useRef<ScrollView>(null);

  useEffect(() => {
    // Auto-scroll to bottom when new logs are added
    if (scrollViewRef.current && logs.length > 0) {
      scrollViewRef.current.scrollToEnd({ animated: true });
    }
  }, [logs]);

  const getLogColor = (type: string) => {
    switch (type.toLowerCase()) {
      case 'error':
        return theme.colors.error;
      case 'warning':
        return theme.colors.waiting;
      case 'info':
        return theme.colors.info;
      default:
        return theme.colors.onSurfaceVariant;
    }
  };

  if (logs.length === 0) {
    return null;
  }

  return (
    <Card style={styles.card}>
      <Card.Content>
        <Text style={styles.title}>📋 Activity Logs</Text>
        <ScrollView
          ref={scrollViewRef}
          style={styles.logsContainer}
          showsVerticalScrollIndicator={false}
        >
          {logs.map((log, index) => (
            <View key={index} style={styles.logEntry}>
              <Text style={styles.timestamp}>[{log.timestamp}]</Text>
              <Text style={[styles.message, { color: getLogColor(log.type) }]}>
                {log.message}
              </Text>
            </View>
          ))}
        </ScrollView>
      </Card.Content>
    </Card>
  );
};

const styles = StyleSheet.create({
  card: {
    backgroundColor: theme.colors.overlay,
    borderRadius: theme.borderRadius.lg,
    marginTop: theme.spacing.md,
    ...theme.shadows.small,
  },
  title: {
    ...theme.typography.h3,
    color: theme.colors.onSurface,
    marginBottom: theme.spacing.md,
  },
  logsContainer: {
    maxHeight: 200,
    backgroundColor: 'rgba(0, 0, 0, 0.2)',
    borderRadius: theme.borderRadius.sm,
    padding: theme.spacing.sm,
  },
  logEntry: {
    marginBottom: theme.spacing.xs,
    paddingBottom: theme.spacing.xs,
    borderBottomWidth: 1,
    borderBottomColor: theme.colors.divider,
  },
  timestamp: {
    ...theme.typography.caption,
    color: theme.colors.onSurfaceVariant,
    fontFamily: 'monospace',
    marginBottom: 2,
  },
  message: {
    ...theme.typography.bodySmall,
    fontFamily: 'monospace',
    lineHeight: 16,
  },
});

export default LogsView; 