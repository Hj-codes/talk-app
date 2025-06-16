import { useState, useEffect, useCallback } from 'react';
import WebSocketService, { WSMessage, WSMessageHandler, WSStatusHandler } from '../services/WebSocketService';
import WebRTCService, { ConnectionState, ConnectionStateHandler, RemoteStreamHandler } from '../services/WebRTCService';

export interface VoiceChatState {
  // Connection states
  isConnected: boolean;
  isMatched: boolean;
  isVoiceConnected: boolean;
  
  // User info
  userId: string | null;
  partnerId: string | null;
  
  // Status
  status: string;
  statusType: 'info' | 'waiting' | 'connected' | 'error';
  
  // Voice
  remoteStream: any;
  
  // Actions
  connect: () => Promise<void>;
  disconnect: () => void;
  findMatch: () => void;
  
  // Logs
  logs: Array<{ timestamp: string; message: string; type: string }>;
}

export const useVoiceChat = (): VoiceChatState => {
  const [isConnected, setIsConnected] = useState(false);
  const [isMatched, setIsMatched] = useState(false);
  const [isVoiceConnected, setIsVoiceConnected] = useState(false);
  const [userId, setUserId] = useState<string | null>(null);
  const [partnerId, setPartnerId] = useState<string | null>(null);
  const [status, setStatus] = useState('Click "Join" to start connecting');
  const [statusType, setStatusType] = useState<'info' | 'waiting' | 'connected' | 'error'>('info');
  const [remoteStream, setRemoteStream] = useState<any>(null);
  const [logs, setLogs] = useState<Array<{ timestamp: string; message: string; type: string }>>([]);

  // Logging function
  const addLog = useCallback((message: string, type: string = 'info') => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs(prev => [...prev, { timestamp, message, type }]);
    console.log(`[${type.toUpperCase()}] ${message}`);
  }, []);

  // Update status helper
  const updateStatus = useCallback((message: string, type: 'info' | 'waiting' | 'connected' | 'error' = 'info') => {
    setStatus(message);
    setStatusType(type);
    addLog(`Status: ${message}`);
  }, [addLog]);

  // WebSocket message handler
  const handleWSMessage: WSMessageHandler = useCallback(async (message: WSMessage) => {
    addLog(`Received: ${message.type}`, 'info');
    
    try {
      switch (message.type) {
        case 'session':
          setUserId(message.payload.user_id);
          addLog(`Session established. User ID: ${message.payload.user_id}`);
          break;
          
        case 'waiting':
          updateStatus('🔍 Looking for someone to chat with...', 'waiting');
          break;
          
        case 'match_found':
          setPartnerId(message.payload.partner_id);
          setIsMatched(true);
          addLog(`Matched with partner: ${message.payload.partner_id}`);
          updateStatus('👥 Match found! Setting up voice connection...', 'connected');
          
          // Initialize WebRTC based on role
          if (message.payload.role === 'caller') {
            addLog('Acting as caller - creating offer');
            await WebRTCService.createOffer();
          } else {
            addLog('Acting as callee - waiting for offer');
          }
          break;
          
        case 'offer':
          addLog('Received WebRTC offer');
          await WebRTCService.handleOffer(message.payload.sdp);
          break;
          
        case 'answer':
          addLog('Received WebRTC answer');
          await WebRTCService.handleAnswer(message.payload.sdp);
          break;
          
        case 'ice_candidate':
          addLog('Received ICE candidate');
          await WebRTCService.handleIceCandidate(message.payload);
          break;
          
        case 'partner_disconnected':
          addLog('Partner disconnected');
          updateStatus('Partner left the conversation', 'error');
          setIsMatched(false);
          setPartnerId(null);
          setIsVoiceConnected(false);
          setRemoteStream(null);
          WebRTCService.close();
          break;
          
        case 'error':
          let errorMessage;
          if (typeof message.payload === 'string') {
            errorMessage = message.payload;
          } else if (typeof message.payload === 'object' && message.payload.message) {
            errorMessage = message.payload.message;
          } else {
            errorMessage = 'Unknown error occurred';
          }
          addLog(`Server error: ${errorMessage}`, 'error');
          updateStatus(`Error: ${errorMessage}`, 'error');
          break;
          
        case 'ping':
          // Respond to ping
          WebSocketService.sendMessage('pong', {});
          break;
          
        default:
          addLog(`Unknown message type: ${message.type}`, 'warning');
      }
    } catch (error) {
      addLog(`Error handling message: ${error}`, 'error');
      updateStatus('Message handling error', 'error');
    }
  }, [addLog, updateStatus]);

  // WebSocket status handler
  const handleWSStatus: WSStatusHandler = useCallback((wsStatus) => {
    switch (wsStatus) {
      case 'connecting':
        updateStatus('Connecting to server...', 'waiting');
        break;
      case 'connected':
        setIsConnected(true);
        updateStatus('Connected! Click "Find Match" to find someone to talk to.', 'connected');
        break;
      case 'disconnected':
        setIsConnected(false);
        setIsMatched(false);
        updateStatus('Disconnected from server', 'error');
        cleanup();
        break;
      case 'error':
        setIsConnected(false);
        updateStatus('Connection error - check if server is running', 'error');
        break;
    }
  }, [updateStatus]);

  // WebRTC connection state handler
  const handleRTCConnectionState: ConnectionStateHandler = useCallback((state: ConnectionState) => {
    addLog(`Connection state: ${state}`);
    switch (state) {
      case 'connected':
        setIsVoiceConnected(true);
        updateStatus('🔊 Voice call connected!', 'connected');
        break;
      case 'connecting':
        updateStatus('Establishing voice connection...', 'waiting');
        break;
      case 'disconnected':
        setIsVoiceConnected(false);
        updateStatus('Voice connection lost', 'error');
        break;
      case 'failed':
        setIsVoiceConnected(false);
        updateStatus('Voice connection failed', 'error');
        addLog('WebRTC connection failed - may need TURN servers', 'error');
        break;
      case 'closed':
        setIsVoiceConnected(false);
        updateStatus('Voice connection closed', 'error');
        break;
    }
  }, [addLog, updateStatus]);

  // WebRTC remote stream handler
  const handleRemoteStream: RemoteStreamHandler = useCallback((stream) => {
    addLog('Remote stream received');
    setRemoteStream(stream);
  }, [addLog]);

  // Connect function
  const connect = useCallback(async () => {
    try {
      addLog('Requesting microphone access...');
      updateStatus('Requesting microphone access...', 'waiting');
      
      // Initialize WebRTC first to get microphone access
      await WebRTCService.initialize();
      addLog('Microphone access granted');

      // Connect to WebSocket
      await WebSocketService.connect();
    } catch (error) {
      addLog(`Error during connection: ${error}`, 'error');
      updateStatus('Connection failed', 'error');
    }
  }, [addLog, updateStatus]);

  // Disconnect function
  const disconnect = useCallback(() => {
    WebSocketService.disconnect();
    WebRTCService.cleanup();
    cleanup();
    updateStatus('Disconnected', 'error');
    addLog('Disconnected from service');
  }, [addLog, updateStatus]);

  // Find match function
  const findMatch = useCallback(() => {
    if (!isConnected) {
      addLog('Not connected to server', 'error');
      return;
    }
    
    if (isMatched) {
      addLog('Already matched with a partner', 'warning');
      return;
    }
    
    addLog('Looking for a match...');
    updateStatus('🔍 Looking for someone to chat with...', 'waiting');
    WebSocketService.findMatch();
  }, [isConnected, isMatched, addLog, updateStatus]);

  // Cleanup function
  const cleanup = useCallback(() => {
    setUserId(null);
    setPartnerId(null);
    setIsMatched(false);
    setIsVoiceConnected(false);
    setRemoteStream(null);
  }, []);

  // Setup event listeners
  useEffect(() => {
    WebSocketService.addMessageHandler(handleWSMessage);
    WebSocketService.addStatusHandler(handleWSStatus);
    WebRTCService.addConnectionStateHandler(handleRTCConnectionState);
    WebRTCService.addRemoteStreamHandler(handleRemoteStream);

    return () => {
      WebSocketService.removeMessageHandler(handleWSMessage);
      WebSocketService.removeStatusHandler(handleWSStatus);
      WebRTCService.removeConnectionStateHandler(handleRTCConnectionState);
      WebRTCService.removeRemoteStreamHandler(handleRemoteStream);
    };
  }, [handleWSMessage, handleWSStatus, handleRTCConnectionState, handleRemoteStream]);

  return {
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
  };
}; 