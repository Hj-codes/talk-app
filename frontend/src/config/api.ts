// Configuration for different environments
const getEnvironmentConfig = () => {
  // For development, you can use your local server or ngrok
  const DEVELOPMENT_BASE_URL = 'https://talk-app-xfbw.onrender.com';
  const DEVELOPMENT_WS_URL = 'wss://talk-app-xfbw.onrender.com/ws';
  
  // For production, use your Render deployment URL
  // Replace 'your-app-name' with your actual Render service name
  const PRODUCTION_BASE_URL = 'https://talk-app-xfbw.onrender.com';
  const PRODUCTION_WS_URL = 'wss://talk-app-xfbw.onrender.com/ws';
  
  if (__DEV__) {
    return {
      BASE_URL: DEVELOPMENT_BASE_URL,
      WS_URL: DEVELOPMENT_WS_URL,
    };
  }
  
  return {
    BASE_URL: PRODUCTION_BASE_URL,
    WS_URL: PRODUCTION_WS_URL,
  };
};

const environmentConfig = getEnvironmentConfig();

export const API_CONFIG = {
    // Backend server configuration
    BASE_URL: environmentConfig.BASE_URL,
    WS_URL: environmentConfig.WS_URL,

    // API endpoints
    ENDPOINTS: {
      HEALTH: '/health',
      STATS: '/stats',
      ICE_SERVERS: '/ice-servers',
      WEBSOCKET: '/ws',
    },
    
    // WebRTC configuration - will be fetched from backend
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