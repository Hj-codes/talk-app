export const API_CONFIG = {
    // Backend server configuration
    // BASE_URL: __DEV__ ? 'http://10.0.2.2:8080' : 'http://localhost:8080', // Android emulator uses 10.0.2.2
    // WS_URL: __DEV__ ? 'ws://10.0.2.2:8080/ws' : 'ws://localhost:8080/ws',
    BASE_URL: __DEV__ ? 'https://2dd3-2405-201-5c00-a0c4-d96b-2a58-ae83-23e6.ngrok-free.app' : 'https://2dd3-2405-201-5c00-a0c4-d96b-2a58-ae83-23e6.ngrok-free.app',
    WS_URL: __DEV__ ? 'wss://2dd3-2405-201-5c00-a0c4-d96b-2a58-ae83-23e6.ngrok-free.app/ws' : 'wss://2dd3-2405-201-5c00-a0c4-d96b-2a58-ae83-23e6.ngrok-free.app/ws',

    // API endpoints
    ENDPOINTS: {
      HEALTH: '/health',
      STATS: '/stats',
      ICE_SERVERS: '/ice-servers',
      WEBSOCKET: '/ws',
    },
    
    // WebRTC configuration matching the HTML demo
    WEBRTC_CONFIG: {
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'stun:stun1.l.google.com:19302' },
      ],
    },
    
    // Connection settings
    CONNECTION_TIMEOUT: 10000,
    RECONNECT_DELAY: 3000,
    MAX_RECONNECT_ATTEMPTS: 5,
  };
  
  export default API_CONFIG; 