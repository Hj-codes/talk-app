import React, { useEffect, useRef } from 'react';
import { View, StyleSheet } from 'react-native';
import { Card, Text, IconButton } from 'react-native-paper';
import { RTCView } from 'react-native-webrtc';
import { theme } from '../styles/theme';

interface RemoteAudioViewProps {
  stream: any;
}

const RemoteAudioView: React.FC<RemoteAudioViewProps> = ({ stream }) => {
  const [isPlaying, setIsPlaying] = React.useState(true);

  const togglePlayback = () => {
    if (stream) {
      stream.getTracks().forEach((track: any) => {
        track.enabled = !track.enabled;
      });
      setIsPlaying(!isPlaying);
    }
  };

  // Note: For audio-only streams, RTCView is not needed and won't display anything
  // The audio will automatically play through the device speakers
  // This component provides UI controls for the audio stream

  return (
    <Card style={styles.card}>
      <Card.Content style={styles.content}>
        <View style={styles.header}>
          <Text style={styles.title}>🔊 Partner Audio</Text>
          <IconButton
            icon={isPlaying ? 'volume-high' : 'volume-off'}
            iconColor={theme.colors.onSurface}
            size={24}
            onPress={togglePlayback}
            style={styles.toggleButton}
          />
        </View>
        
        <View style={styles.audioIndicator}>
          <View style={[styles.indicator, { backgroundColor: isPlaying ? theme.colors.connected : theme.colors.error }]} />
          <Text style={styles.indicatorText}>
            {isPlaying ? 'Audio Playing' : 'Audio Muted'}
          </Text>
        </View>

        {/* Hidden RTCView for audio stream - required by react-native-webrtc */}
        {stream && (
          <RTCView
            streamURL={stream.toURL()}
            style={styles.hiddenView}
            objectFit="cover"
            mirror={false}
          />
        )}
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
  content: {
    alignItems: 'center',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
    marginBottom: theme.spacing.md,
  },
  title: {
    ...theme.typography.h3,
    color: theme.colors.onSurface,
  },
  toggleButton: {
    backgroundColor: theme.colors.surfaceVariant,
  },
  audioIndicator: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  indicator: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: theme.spacing.sm,
  },
  indicatorText: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurfaceVariant,
  },
  hiddenView: {
    width: 0,
    height: 0,
    opacity: 0,
  },
});

export default RemoteAudioView; 