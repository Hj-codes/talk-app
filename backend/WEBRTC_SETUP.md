# WebRTC Setup Guide

This guide will help you set up a complete WebRTC voice chat system with your Go backend and React Native frontend.

## Backend Setup

### 1. WebRTC Signaling Server

The Go backend now includes a complete WebRTC signaling server with the following features:

- **SDP Offer/Answer Exchange**: Handles WebRTC session descriptions
- **ICE Candidate Exchange**: Manages network connectivity candidates
- **Call State Management**: Tracks call states (idle, ringing, answered, ended)
- **STUN/TURN Server Configuration**: Provides ICE servers for NAT traversal
- **Message Validation**: Validates SDP and ICE candidate formats
- **Error Handling**: Comprehensive error handling for failed connections

### 2. Message Types

The signaling server supports these message types:

#### Client → Server
```json
{
  "type": "find_match",
  "timestamp": "2024-01-01T00:00:00Z"
}

{
  "type": "offer",
  "payload": {
    "type": "offer",
    "sdp": "v=0\r\no=..."
  },
  "to": "partner_user_id"
}

{
  "type": "answer", 
  "payload": {
    "type": "answer",
    "sdp": "v=0\r\no=..."
  },
  "to": "partner_user_id"
}

{
  "type": "ice_candidate",
  "payload": {
    "candidate": "candidate:...",
    "sdpMLineIndex": 0,
    "sdpMid": "0"
  },
  "to": "partner_user_id"
}

{
  "type": "call_start",
  "timestamp": "2024-01-01T00:00:00Z"
}

{
  "type": "call_accept",
  "timestamp": "2024-01-01T00:00:00Z"
}

{
  "type": "call_reject",
  "timestamp": "2024-01-01T00:00:00Z"
}

{
  "type": "call_end",
  "timestamp": "2024-01-01T00:00:00Z"
}

{
  "type": "get_ice_servers",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### Server → Client
```json
{
  "type": "session",
  "payload": {
    "user_id": "uuid",
    "token": "jwt_token"
  }
}

{
  "type": "match_found",
  "payload": {
    "partner_id": "partner_uuid",
    "room_id": "room_id",
    "role": "caller" // or "callee"
  }
}

{
  "type": "call_incoming",
  "from": "caller_id",
  "payload": {
    "caller_id": "uuid",
    "room_id": "room_id"
  }
}

{
  "type": "ice_servers",
  "payload": {
    "iceServers": [
      {
        "urls": ["stun:stun.l.google.com:19302"]
      },
      {
        "urls": ["turn:your-turn-server.com:3478"],
        "username": "user",
        "credential": "pass"
      }
    ]
  }
}
```

## STUN/TURN Server Setup

### Using Coturn (Recommended)

1. **Install Coturn**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install coturn
   
   # CentOS/RHEL
   sudo yum install coturn
   ```

2. **Configure Coturn** (`/etc/turnserver.conf`):
   ```conf
   # Basic configuration
   listening-port=3478
   tls-listening-port=5349
   
   # External IP (replace with your server's IP)
   external-ip=YOUR_SERVER_IP
   
   # Relay configuration  
   min-port=10000
   max-port=20000
   
   # Authentication
   use-auth-secret
   static-auth-secret=YOUR_SECRET_KEY
   
   # Database (optional)
   userdb=/var/lib/turn/turndb
   
   # Logging
   log-file=/var/log/turnserver.log
   verbose
   
   # Security
   fingerprint
   no-multicast-peers
   no-cli
   no-loopback-peers
   no-tcp-relay
   
   # Certificates (for TLS)
   cert=/path/to/cert.pem
   pkey=/path/to/private.pem
   ```

3. **Start Coturn**:
   ```bash
   sudo systemctl enable coturn
   sudo systemctl start coturn
   ```

4. **Configure Environment Variables**:
   ```bash
   # .env file
   TURN_SERVER_1_URL=turn:your-server.com:3478
   TURN_SERVER_1_USERNAME=your-username
   TURN_SERVER_1_CREDENTIAL=your-password
   
   # Multiple TURN servers
   TURN_SERVER_2_URL=turns:your-server.com:5349
   TURN_SERVER_2_USERNAME=your-username
   TURN_SERVER_2_CREDENTIAL=your-password
   ```

### Using Public STUN Servers (Development Only)

For development, you can use public STUN servers (already configured):
- `stun:stun.l.google.com:19302`
- `stun:stun1.l.google.com:19302`
- `stun:stun2.l.google.com:19302`

**Note**: Public STUN servers don't provide TURN functionality, which is needed for users behind symmetric NATs.

## React Native Frontend Integration

### 1. Install Dependencies

```bash
npm install react-native-webrtc
# For Expo managed workflow
expo install react-native-webrtc
```

### 2. Basic WebRTC Setup

```javascript
// WebRTCService.js
import {
  RTCPeerConnection,
  RTCSessionDescription,
  RTCIceCandidate,
  mediaDevices
} from 'react-native-webrtc';

class WebRTCService {
  constructor() {
    this.pc = null;
    this.localStream = null;
    this.ws = null;
    this.configuration = {
      iceServers: [], // Will be fetched from server
    };
  }

  async init() {
    // Fetch ICE servers from backend
    const response = await fetch('http://your-server:8080/ice-servers');
    const { iceServers } = await response.json();
    this.configuration.iceServers = iceServers;

    // Initialize WebSocket connection
    this.ws = new WebSocket('ws://your-server:8080/ws');
    this.setupWebSocketHandlers();

    // Get user media
    await this.getUserMedia();
  }

  async getUserMedia() {
    try {
      const stream = await mediaDevices.getUserMedia({
        audio: true,
        video: false, // Voice only
      });
      this.localStream = stream;
    } catch (error) {
      console.error('Error accessing microphone:', error);
    }
  }

  setupWebSocketHandlers() {
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleSignalingMessage(message);
    };
  }

  async handleSignalingMessage(message) {
    switch (message.type) {
      case 'match_found':
        await this.handleMatchFound(message);
        break;
      case 'offer':
        await this.handleOffer(message);
        break;
      case 'answer':
        await this.handleAnswer(message);
        break;
      case 'ice_candidate':
        await this.handleIceCandidate(message);
        break;
      case 'call_incoming':
        this.handleIncomingCall(message);
        break;
    }
  }

  async createPeerConnection() {
    this.pc = new RTCPeerConnection(this.configuration);
    
    // Add local stream
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => {
        this.pc.addTrack(track, this.localStream);
      });
    }

    // Handle ICE candidates
    this.pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.sendMessage({
          type: 'ice_candidate',
          payload: {
            candidate: event.candidate.candidate,
            sdpMLineIndex: event.candidate.sdpMLineIndex,
            sdpMid: event.candidate.sdpMid,
          },
          to: this.partnerId,
        });
      }
    };

    // Handle remote stream
    this.pc.ontrack = (event) => {
      console.log('Received remote stream');
      // Handle remote audio stream
    };
  }

  async handleMatchFound(message) {
    this.partnerId = message.payload.partner_id;
    await this.createPeerConnection();

    if (message.payload.role === 'caller') {
      await this.createOffer();
    }
  }

  async createOffer() {
    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);

    this.sendMessage({
      type: 'offer',
      payload: {
        type: 'offer',
        sdp: offer.sdp,
      },
      to: this.partnerId,
    });
  }

  async handleOffer(message) {
    if (!this.pc) {
      await this.createPeerConnection();
    }

    const offer = new RTCSessionDescription({
      type: 'offer',
      sdp: message.payload.sdp,
    });

    await this.pc.setRemoteDescription(offer);
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);

    this.sendMessage({
      type: 'answer',
      payload: {
        type: 'answer',
        sdp: answer.sdp,
      },
      to: message.from,
    });
  }

  async handleAnswer(message) {
    const answer = new RTCSessionDescription({
      type: 'answer',
      sdp: message.payload.sdp,
    });

    await this.pc.setRemoteDescription(answer);
  }

  async handleIceCandidate(message) {
    const candidate = new RTCIceCandidate({
      candidate: message.payload.candidate,
      sdpMLineIndex: message.payload.sdpMLineIndex,
      sdpMid: message.payload.sdpMid,
    });

    await this.pc.addIceCandidate(candidate);
  }

  sendMessage(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        ...message,
        timestamp: new Date().toISOString(),
      }));
    }
  }

  findMatch() {
    this.sendMessage({ type: 'find_match' });
  }

  endCall() {
    if (this.pc) {
      this.pc.close();
      this.pc = null;
    }
    this.sendMessage({ type: 'call_end' });
  }
}

export default new WebRTCService();
```

### 3. React Native Component Example

```javascript
// VoiceChatScreen.js
import React, { useEffect, useState } from 'react';
import { View, Button, Text, Alert } from 'react-native';
import WebRTCService from './WebRTCService';

export default function VoiceChatScreen() {
  const [status, setStatus] = useState('disconnected');
  const [isMatching, setIsMatching] = useState(false);

  useEffect(() => {
    WebRTCService.init();
    
    // Add event listeners for status updates
    // You'll need to implement these in WebRTCService
    
    return () => {
      WebRTCService.endCall();
    };
  }, []);

  const findMatch = () => {
    setIsMatching(true);
    WebRTCService.findMatch();
  };

  const endCall = () => {
    WebRTCService.endCall();
    setStatus('disconnected');
    setIsMatching(false);
  };

  return (
    <View style={{ flex: 1, justifyContent: 'center', padding: 20 }}>
      <Text>Status: {status}</Text>
      
      {status === 'disconnected' && (
        <Button
          title={isMatching ? "Finding match..." : "Find Someone to Talk"}
          onPress={findMatch}
          disabled={isMatching}
        />
      )}
      
      {status === 'connected' && (
        <Button title="End Call" onPress={endCall} />
      )}
    </View>
  );
}
```

## Production Deployment

### 1. Environment Variables

```bash
# Server configuration
PORT=8080
JWT_SECRET=your-super-secret-jwt-key

# WebRTC configuration
STUN_SERVERS=stun:stun.l.google.com:19302,stun:stun1.l.google.com:19302

# TURN servers
TURN_SERVER_1_URL=turn:your-turn-server.com:3478
TURN_SERVER_1_USERNAME=username
TURN_SERVER_1_CREDENTIAL=password

TURN_SERVER_2_URL=turns:your-turn-server.com:5349
TURN_SERVER_2_USERNAME=username
TURN_SERVER_2_CREDENTIAL=password
```

### 2. Docker Deployment

The project includes a `Dockerfile` and `docker-compose.yml`. To deploy with TURN server:

```yaml
# docker-compose.yml
version: '3.8'
services:
  voice-chat-backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      - TURN_SERVER_1_URL=turn:coturn:3478
      - TURN_SERVER_1_USERNAME=user
      - TURN_SERVER_1_CREDENTIAL=pass
    depends_on:
      - coturn

  coturn:
    image: coturn/coturn:latest
    ports:
      - "3478:3478"
      - "3478:3478/udp"
      - "5349:5349"
      - "10000-20000:10000-20000/udp"
    volumes:
      - ./coturn/turnserver.conf:/etc/coturn/turnserver.conf
    command: ["-c", "/etc/coturn/turnserver.conf"]
```

### 3. SSL/TLS Setup

For production, you need HTTPS for WebRTC:

```go
// In main.go, replace ListenAndServe with:
log.Println("Voice chat server starting on :8443 (HTTPS)")
if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
    log.Fatalf("Server failed to start: %v", err)
}
```

## Testing

### 1. Backend Testing

```bash
# Run existing tests
go test ./...

# Test WebSocket connection
curl -i -N -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" \
     -H "Sec-WebSocket-Key: test" \
     http://localhost:8080/ws
```

### 2. Frontend Testing

```javascript
// Test ICE servers endpoint
fetch('http://localhost:8080/ice-servers')
  .then(r => r.json())
  .then(data => console.log('ICE Servers:', data));
```

## Architecture Overview

```
┌─────────────────┐     WebSocket     ┌─────────────────┐
│  React Native   │ ◄────────────────► │   Go Backend    │
│     Client      │   (Signaling)     │  (Signaling     │
│                 │                   │   Server)       │
└─────────────────┘                   └─────────────────┘
         │                                     │
         │ WebRTC P2P                         │
         │ (Voice Data)                       │
         │                                     │
         │              ┌─────────────────┐   │
         └──────────────►│  STUN/TURN     │◄──┘
                        │   Server        │
                        └─────────────────┘
```

This setup provides a complete, production-ready WebRTC voice chat system with proper NAT traversal, error handling, and scalable architecture. 