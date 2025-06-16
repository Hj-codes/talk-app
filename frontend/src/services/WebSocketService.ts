import API_CONFIG from '../config/api';

// Message types matching the Go backend
export interface WSMessage {
  type: 'session' | 'waiting' | 'match_found' | 'offer' | 'answer' | 'ice_candidate' | 'partner_disconnected' | 'error' | 'ping' | 'pong' | 'find_match' | 'disconnect';
  payload: any;
  timestamp: string;
}

export interface SessionPayload {
  user_id: string;
}

export interface MatchFoundPayload {
  partner_id: string;
  role: 'caller' | 'callee';
}

export interface WebRTCPayload {
  type: 'offer' | 'answer';
  sdp: string;
}

export interface IceCandidatePayload {
  candidate: string;
  sdpMLineIndex: number;
  sdpMid: string;
}

export type WSMessageHandler = (message: WSMessage) => void;
export type WSStatusHandler = (status: 'connecting' | 'connected' | 'disconnected' | 'error') => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private messageHandlers: Set<WSMessageHandler> = new Set();
  private statusHandlers: Set<WSStatusHandler> = new Set();
  private reconnectAttempts = 0;
  private isConnecting = false;
  private isDisconnecting = false;

  // Add message handler
  addMessageHandler(handler: WSMessageHandler) {
    this.messageHandlers.add(handler);
  }

  // Remove message handler
  removeMessageHandler(handler: WSMessageHandler) {
    this.messageHandlers.delete(handler);
  }

  // Add status handler
  addStatusHandler(handler: WSStatusHandler) {
    this.statusHandlers.add(handler);
  }

  // Remove status handler
  removeStatusHandler(handler: WSStatusHandler) {
    this.statusHandlers.delete(handler);
  }

  // Connect to WebSocket
  connect(): Promise<void> {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return Promise.resolve();
    }

    this.isConnecting = true;
    this.notifyStatusHandlers('connecting');

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(API_CONFIG.WS_URL);

        this.ws.onopen = () => {
          console.log('WebSocket connected to', API_CONFIG.WS_URL);
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.notifyStatusHandlers('connected');
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            console.log('WebSocket message received:', message.type);
            this.notifyMessageHandlers(message);
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason);
          this.isConnecting = false;
          this.ws = null;
          this.notifyStatusHandlers('disconnected');

          // Auto-reconnect if not manually disconnecting
          if (!this.isDisconnecting && this.reconnectAttempts < API_CONFIG.MAX_RECONNECT_ATTEMPTS) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.isConnecting = false;
          this.notifyStatusHandlers('error');
          reject(error);
        };

        // Connection timeout
        setTimeout(() => {
          if (this.isConnecting) {
            this.isConnecting = false;
            this.ws?.close();
            reject(new Error('Connection timeout'));
          }
        }, API_CONFIG.CONNECTION_TIMEOUT);

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  // Disconnect from WebSocket
  disconnect() {
    this.isDisconnecting = true;
    this.reconnectAttempts = API_CONFIG.MAX_RECONNECT_ATTEMPTS; // Prevent reconnection
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    
    this.notifyStatusHandlers('disconnected');
  }

  // Send message
  sendMessage(type: WSMessage['type'], payload: any = {}) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return false;
    }

    const message: WSMessage = {
      type,
      payload,
      timestamp: new Date().toISOString(),
    };

    try {
      this.ws.send(JSON.stringify(message));
      console.log('WebSocket message sent:', type);
      return true;
    } catch (error) {
      console.error('Error sending WebSocket message:', error);
      return false;
    }
  }

  // Get connection status
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // Helper methods
  findMatch() {
    return this.sendMessage('find_match', {});
  }

  sendOffer(offer: WebRTCPayload) {
    return this.sendMessage('offer', offer);
  }

  sendAnswer(answer: WebRTCPayload) {
    return this.sendMessage('answer', answer);
  }

  sendIceCandidate(candidate: IceCandidatePayload) {
    return this.sendMessage('ice_candidate', candidate);
  }

  sendDisconnect() {
    return this.sendMessage('disconnect', {});
  }

  // Private methods
  private scheduleReconnect() {
    this.reconnectAttempts++;
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${API_CONFIG.MAX_RECONNECT_ATTEMPTS}`);
    
    setTimeout(() => {
      if (!this.isDisconnecting && this.reconnectAttempts <= API_CONFIG.MAX_RECONNECT_ATTEMPTS) {
        this.connect().catch(console.error);
      }
    }, API_CONFIG.RECONNECT_DELAY);
  }

  private notifyMessageHandlers(message: WSMessage) {
    this.messageHandlers.forEach(handler => {
      try {
        handler(message);
      } catch (error) {
        console.error('Error in message handler:', error);
      }
    });
  }

  private notifyStatusHandlers(status: Parameters<WSStatusHandler>[0]) {
    this.statusHandlers.forEach(handler => {
      try {
        handler(status);
      } catch (error) {
        console.error('Error in status handler:', error);
      }
    });
  }
}

// Export singleton instance
export default new WebSocketService(); 