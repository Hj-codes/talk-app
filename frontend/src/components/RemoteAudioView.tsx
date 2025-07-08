import React, { useEffect, useRef } from 'react';
import { View, StyleSheet, Animated } from 'react-native';
import { Card, Text, IconButton, Surface } from 'react-native-paper';
import { RTCView } from 'react-native-webrtc';
import { theme } from '../styles/theme';

interface RemoteAudioViewProps {
  stream: any;
}

const RemoteAudioView: React.FC<RemoteAudioViewProps> = ({ stream }) => {
  const [isPlaying, setIsPlaying] = React.useState(true);
  const pulseAnim = useRef(new Animated.Value(1)).current;

  useEffect(() => {
    // Create a subtle pulsing animation for the audio indicator when playing
    if (isPlaying) {
      const pulse = Animated.sequence([
        Animated.timing(pulseAnim, {
          toValue: 1.2,
          duration: 1000,
          useNativeDriver: true,
        }),
        Animated.timing(pulseAnim, {
          toValue: 1,
          duration: 1000,
          useNativeDriver: true,
        }),
      ]);

      const loop = Animated.loop(pulse);
      loop.start();

      return () => loop.stop();
    } else {
      pulseAnim.setValue(1);
    }
  }, [isPlaying, pulseAnim]);

  const togglePlayback = () => {
    if (stream) {
      stream.getTracks().forEach((track: any) => {
        track.enabled = !track.enabled;
      });
      setIsPlaying(!isPlaying);
    }
  };

  return (
    <Card style={styles.card} mode="elevated">
      <Card.Content style={styles.content}>
        <View style={styles.header}>
          <View style={styles.titleSection}>
            <Text style={styles.title}>Partner Audio</Text>
            <Text style={styles.subtitle}>
              {isPlaying ? 'Audio stream active' : 'Audio stream muted'}
            </Text>
          </View>
          
          <Surface style={styles.controlSurface} elevation={1}>
            <IconButton
              icon={isPlaying ? 'volume-high' : 'volume-off'}
              iconColor={isPlaying ? theme.colors.connected : theme.colors.error}
              size={28}
              onPress={togglePlayback}
              style={styles.toggleButton}
            />
          </Surface>
        </View>
        
        <View style={styles.audioIndicatorSection}>
          <Animated.View 
            style={[
              styles.audioIndicator,
              {
                transform: [{ scale: pulseAnim }],
                backgroundColor: isPlaying ? theme.colors.connected : theme.colors.error,
              }
            ]} 
          />
          <View style={styles.indicatorTextContainer}>
            <Text style={styles.indicatorMainText}>
              {isPlaying ? 'Connected' : 'Muted'}
            </Text>
            <Text style={styles.indicatorSubText}>
              {isPlaying ? 'Receiving audio from partner' : 'Audio playback disabled'}
            </Text>
          </View>
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
    marginBottom: theme.spacing.xxxl,
    backgroundColor: theme.colors.surface,
    borderRadius: theme.borderRadius.lg,
  },
  content: {
    paddingVertical: theme.spacing.xl,
    paddingHorizontal: theme.spacing.lg,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: theme.spacing.xl,
  },
  titleSection: {
    flex: 1,
  },
  title: {
    ...theme.typography.h4,
    color: theme.colors.onSurface,
    marginBottom: theme.spacing.xs,
  },
  subtitle: {
    ...theme.typography.bodySmall,
    color: theme.colors.onSurfaceVariant,
  },
  controlSurface: {
    borderRadius: theme.borderRadius.round,
    backgroundColor: theme.colors.surfaceVariant,
  },
  toggleButton: {
    margin: 0,
  },
  audioIndicatorSection: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: theme.spacing.lg,
  },
  audioIndicator: {
    width: 24,
    height: 24,
    borderRadius: 12,
    marginRight: theme.spacing.lg,
  },
  indicatorTextContainer: {
    flex: 1,
    alignItems: 'flex-start',
  },
  indicatorMainText: {
    ...theme.typography.body,
    color: theme.colors.onSurface,
    fontWeight: '600',
    marginBottom: theme.spacing.xs,
  },
  indicatorSubText: {
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