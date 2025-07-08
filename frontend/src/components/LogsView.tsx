import React, { useRef, useEffect } from 'react';
import { View, StyleSheet, ScrollView } from 'react-native';
import { Card, Text, Surface } from 'react-native-paper';
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
        return theme.colors.warning;
      case 'info':
        return theme.colors.info;
      default:
        return theme.colors.onSurfaceVariant;
    }
  };

  const getLogIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'error':
        return '❌';
      case 'warning':
        return '⚠️';
      case 'info':
        return 'ℹ️';
      default:
        return '📝';
    }
  };

  if (logs.length === 0) {
    return null;
  }

  return (
    <Card style={styles.card} mode="elevated">
      <Card.Content style={styles.content}>
        <View style={styles.header}>
          <Text style={styles.title}>Activity Logs</Text>
          <Text style={styles.subtitle}>{logs.length} entries</Text>
        </View>
        
        <Surface style={styles.logsContainer} elevation={0}>
          <ScrollView
            ref={scrollViewRef}
            style={styles.scrollView}
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
          >
            {logs.map((log, index) => (
              <View key={index} style={styles.logEntry}>
                <View style={styles.logHeader}>
                  <Text style={styles.logIcon}>{getLogIcon(log.type)}</Text>
                  <Text style={styles.timestamp}>{log.timestamp}</Text>
                  <View style={[styles.typeBadge, { backgroundColor: getLogColor(log.type) + '20' }]}>
                    <Text style={[styles.typeText, { color: getLogColor(log.type) }]}>
                      {log.type.toUpperCase()}
                    </Text>
                  </View>
                </View>
                <Text style={[styles.message, { color: theme.colors.onSurface }]}>
                  {log.message}
                </Text>
              </View>
            ))}
          </ScrollView>
        </Surface>
      </Card.Content>
    </Card>
  );
};

const styles = StyleSheet.create({
  card: {
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.lg,
    marginBottom: theme.spacing.xl,
  },
  content: {
    paddingVertical: theme.spacing.xl,
    paddingHorizontal: theme.spacing.lg,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: theme.spacing.lg,
  },
  title: {
    ...theme.typography.h4,
    color: theme.colors.onSurface,
  },
  subtitle: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurfaceVariant,
    fontWeight: '500',
  },
  logsContainer: {
    maxHeight: 250,
    backgroundColor: theme.colors.surfaceVariant,
    borderRadius: theme.borderRadius.md,
    borderWidth: 1,
    borderColor: theme.colors.outline,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: theme.spacing.md,
  },
  logEntry: {
    marginBottom: theme.spacing.md,
    paddingBottom: theme.spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: theme.colors.divider,
  },
  logHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: theme.spacing.xs,
  },
  logIcon: {
    fontSize: 12,
    marginRight: theme.spacing.sm,
  },
  timestamp: {
    ...theme.typography.caption,
    color: theme.colors.onSurfaceVariant,
    fontFamily: 'monospace',
    flex: 1,
    fontWeight: '500',
  },
  typeBadge: {
    paddingHorizontal: theme.spacing.sm,
    paddingVertical: theme.spacing.xs,
    borderRadius: theme.borderRadius.xs,
  },
  typeText: {
    ...theme.typography.caption,
    fontWeight: '600',
    fontSize: 10,
  },
  message: {
    ...theme.typography.bodySmall,
    fontFamily: 'monospace',
    lineHeight: 18,
    paddingLeft: theme.spacing.lg,
  },
});

export default LogsView; 