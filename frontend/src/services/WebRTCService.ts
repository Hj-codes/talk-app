import { mediaDevices, RTCPeerConnection, RTCSessionDescription, RTCIceCandidate } from 'react-native-webrtc';
import API_CONFIG from '../config/api';
import WebSocketService from './WebSocketService';

export type ConnectionState = 'new' | 'connecting' | 'connected' | 'disconnected' | 'failed' | 'closed';
export type ConnectionStateHandler = (state: ConnectionState) => void;
export type RemoteStreamHandler = (stream: any) => void;

class WebRTCService {
  private peerConnection: RTCPeerConnection | any | null = null;
  private localStream: any = null;
  private connectionStateHandlers: Set<ConnectionStateHandler> = new Set();
  private remoteStreamHandlers: Set<RemoteStreamHandler> = new Set();
  private isInitialized = false;
  private remoteCandidates: RTCIceCandidate[] = [];

  // Add connection state handler
  addConnectionStateHandler(handler: ConnectionStateHandler) {
    this.connectionStateHandlers.add(handler);
  }

  // Remove connection state handler
  removeConnectionStateHandler(handler: ConnectionStateHandler) {
    this.connectionStateHandlers.delete(handler);
  }

  // Add remote stream handler
  addRemoteStreamHandler(handler: RemoteStreamHandler) {
    this.remoteStreamHandlers.add(handler);
  }

  // Remove remote stream handler
  removeRemoteStreamHandler(handler: RemoteStreamHandler) {
    this.remoteStreamHandlers.delete(handler);
  }

  // Initialize WebRTC with microphone access
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    try {
      console.log('Requesting microphone access...');
      
      const mediaConstraints = {
        audio: true,
        video: false,
      };

      this.localStream = await mediaDevices.getUserMedia(mediaConstraints);
      console.log('Microphone access granted');
      this.isInitialized = true;
    } catch (error) {
      console.error('Error accessing microphone:', error);
      throw new Error('Microphone access denied');
    }
  }

  // Create peer connection with proper event handlers
  private createPeerConnection(): RTCPeerConnection {
    if (this.peerConnection) {
      this.peerConnection.close();
    }

    this.peerConnection = new RTCPeerConnection(API_CONFIG.WEBRTC_CONFIG);

    // Handle connection state changes
    this.peerConnection.addEventListener('connectionstatechange', () => {
      if (this.peerConnection) {
        const state = this.peerConnection.connectionState as ConnectionState;
        console.log(`WebRTC connection state: ${state}`);
        this.notifyConnectionStateHandlers(state);
      }
    });

    // Handle ICE connection state changes
    this.peerConnection.addEventListener('iceconnectionstatechange', () => {
      if (this.peerConnection) {
        console.log(`ICE connection state: ${this.peerConnection.iceConnectionState}`);
        
        switch (this.peerConnection.iceConnectionState) {
          case 'connected':
          case 'completed':
            console.log('Call connected successfully');
            break;
          case 'disconnected':
          case 'failed':
            console.log('Call disconnected or failed');
            break;
        }
      }
    });

    // Handle ICE candidates
    this.peerConnection.addEventListener('icecandidate', (event: any) => {
      if (event.candidate) {
        console.log('Sending ICE candidate');
        WebSocketService.sendIceCandidate({
          candidate: event.candidate.candidate,
          sdpMLineIndex: event.candidate.sdpMLineIndex,
          sdpMid: event.candidate.sdpMid,
        });
      } else {
        console.log('ICE candidate gathering finished');
      }
    });

    // Handle ICE candidate errors
    this.peerConnection.addEventListener('icecandidateerror', (event: any) => {
      console.warn('ICE candidate error:', event);
    });

    // Handle remote stream using modern 'track' event
    this.peerConnection.addEventListener('track', (event: any) => {
      console.log('Received remote track');
      if (event.streams && event.streams[0]) {
        console.log('Remote stream received');
        this.notifyRemoteStreamHandlers(event.streams[0]);
      }
    });

    // Handle signaling state changes
    this.peerConnection.addEventListener('signalingstatechange', () => {
      if (this.peerConnection) {
        console.log(`Signaling state: ${this.peerConnection.signalingState}`);
        if (this.peerConnection.signalingState === 'closed') {
          console.log('Signaling closed');
        }
      }
    });

    // Handle negotiation needed
    this.peerConnection.addEventListener('negotiationneeded', () => {
      console.log('Negotiation needed');
    });

    // Add local stream tracks to peer connection
    if (this.localStream) {
      this.localStream.getTracks().forEach((track: any) => {
        if (this.peerConnection) {
          this.peerConnection.addTrack(track, this.localStream);
          console.log(`Added ${track.kind} track to peer connection`);
        }
      });
    }

    return this.peerConnection;
  }

  // Create offer (caller role)
  async createOffer(): Promise<void> {
    if (!this.isInitialized) {
      throw new Error('WebRTC not initialized');
    }

    const pc = this.createPeerConnection();

    try {
      // Modern constraint format
      const offerOptions = {
        offerToReceiveAudio: true,
        offerToReceiveVideo: false,
        voiceActivityDetection: true
      };

      const offer = await pc.createOffer(offerOptions);
      await pc.setLocalDescription(offer);

      console.log('Created and set local offer');
      WebSocketService.sendOffer({
        type: 'offer',
        sdp: offer.sdp,
      });
    } catch (error) {
      console.error('Error creating offer:', error);
      throw error;
    }
  }

  // Handle incoming offer (callee role)
  async handleOffer(offerSdp: string): Promise<void> {
    if (!this.isInitialized) {
      throw new Error('WebRTC not initialized');
    }

    const pc = this.createPeerConnection();

    try {
      const offerDescription = new RTCSessionDescription({
        type: 'offer',
        sdp: offerSdp,
      });

      await pc.setRemoteDescription(offerDescription);
      console.log('Set remote offer');

      // Modern constraint format for answer
      const answerOptions = {
        offerToReceiveAudio: true,
        offerToReceiveVideo: false,
        voiceActivityDetection: true
      };

    //   const answer = await pc.createAnswer(answerOptions);
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);

      console.log('Created and set local answer');
      
      // Process any queued remote candidates
      this.processCandidates();

      WebSocketService.sendAnswer({
        type: 'answer',
        sdp: answer.sdp,
      });
    } catch (error) {
      console.error('Error handling offer:', error);
      throw error;
    }
  }

  // Handle incoming answer (caller role)
  async handleAnswer(answerSdp: string): Promise<void> {
    if (!this.peerConnection) {
      throw new Error('No peer connection');
    }

    try {
      const answerDescription = new RTCSessionDescription({
        type: 'answer',
        sdp: answerSdp,
      });

      await this.peerConnection.setRemoteDescription(answerDescription);
      console.log('Set remote answer');
      
      // Process any queued remote candidates
      this.processCandidates();
    } catch (error) {
      console.error('Error handling answer:', error);
      throw error;
    }
  }

  // Handle incoming ICE candidate
  async handleIceCandidate(candidateData: { 
    candidate: string; 
    sdpMLineIndex: number; 
    sdpMid: string 
  }): Promise<void> {
    const iceCandidate = new RTCIceCandidate(candidateData);
    
    if (!this.peerConnection) {
      console.warn('No peer connection, queuing ICE candidate');
      this.remoteCandidates.push(iceCandidate);
      return;
    }

    if (!this.peerConnection.remoteDescription) {
      console.warn('No remote description, queuing ICE candidate');
      this.remoteCandidates.push(iceCandidate);
      return;
    }

    try {
      await this.peerConnection.addIceCandidate(iceCandidate);
      console.log('Added ICE candidate');
    } catch (error) {
      console.error('Error adding ICE candidate:', error);
    }
  }

  // Process queued remote candidates
  private processCandidates(): void {
    if (this.remoteCandidates.length < 1 || !this.peerConnection) {
      return;
    }

    console.log(`Processing ${this.remoteCandidates.length} queued candidates`);
    
    this.remoteCandidates.forEach(async (candidate) => {
      try {
        await this.peerConnection!.addIceCandidate(candidate);
        console.log('Processed queued candidate');
      } catch (error) {
        console.error('Error processing queued candidate:', error);
      }
    });
    
    this.remoteCandidates = [];
  }

  // Close connection and cleanup
  close(): void {
    console.log('Closing WebRTC connection');

    if (this.peerConnection) {
      this.peerConnection.close();
      this.peerConnection = null;
    }

    // Clear queued candidates
    this.remoteCandidates = [];
    
    // Notify that remote stream is null
    this.notifyRemoteStreamHandlers(null);
  }

  // Stop local stream
  stopLocalStream(): void {
    if (this.localStream) {
      this.localStream.getTracks().forEach((track: any) => {
        track.stop();
      });
      this.localStream = null;
      console.log('Stopped local stream');
    }
    this.isInitialized = false;
  }

  // Get local stream
  getLocalStream(): any {
    return this.localStream;
  }

  // Get connection state
  getConnectionState(): ConnectionState {
    return (this.peerConnection?.connectionState as ConnectionState) || 'new';
  }

  // Check if connected
  isConnected(): boolean {
    return this.getConnectionState() === 'connected';
  }

  // Get ICE connection state
  getIceConnectionState(): string {
    return this.peerConnection?.iceConnectionState || 'new';
  }

  // Get signaling state
  getSignalingState(): string {
    return this.peerConnection?.signalingState || 'stable';
  }

  // Complete cleanup
  cleanup(): void {
    this.close();
    this.stopLocalStream();
    this.connectionStateHandlers.clear();
    this.remoteStreamHandlers.clear();
  }

  // Private notification methods
  private notifyConnectionStateHandlers(state: ConnectionState) {
    this.connectionStateHandlers.forEach(handler => {
      try {
        handler(state);
      } catch (error) {
        console.error('Error in connection state handler:', error);
      }
    });
  }

  private notifyRemoteStreamHandlers(stream: any) {
    this.remoteStreamHandlers.forEach(handler => {
      try {
        handler(stream);
      } catch (error) {
        console.error('Error in remote stream handler:', error);
      }
    });
  }
}

export default new WebRTCService();
