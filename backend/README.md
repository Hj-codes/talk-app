# Voice Chat Backend

A scalable, real-time voice chat backend built with Go, designed for random voice chat applications like Omegle.

## 🚀 Features

- **Real-time WebSocket Communication**: Low-latency signaling for WebRTC
- **Random Matchmaking**: Intelligent user pairing system
- **Room-based Architecture**: Isolated chat sessions with proper cleanup
- **Connection Management**: Heartbeat monitoring and automatic cleanup
- **Graceful Shutdown**: Proper resource cleanup on server shutdown
- **Monitoring**: Health checks and real-time statistics
- **Security**: JWT-based session management
- **Scalability**: Thread-safe operations with proper concurrency handling

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   WebRTC        │
│   (React/Expo)  │◄──►│   (Go Server)   │◄──►│   Signaling     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   User Pool     │
                       │   Management    │
                       └─────────────────┘
```

## 📦 Installation

### Prerequisites
- Go 1.22.3 or higher
- Git

### Setup
```bash
# Clone the repository
git clone <your-repo-url>
cd talk-app/backend

# Install dependencies
go mod download

# Build the application
go build -o bin/voice-chat-app .

# Run the server
./bin/voice-chat-app
```

## 🔧 Configuration

The server can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `JWT_SECRET` | `my-secret-key-change-in-production` | JWT signing secret |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `15s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `HEARTBEAT_INTERVAL` | `30s` | WebSocket heartbeat interval |
| `CLEANUP_INTERVAL` | `30s` | Connection cleanup interval |
| `CONNECTION_TIMEOUT` | `60s` | WebSocket connection timeout |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |
| `MAX_CONNECTIONS` | `1000` | Maximum concurrent connections |
| `RATE_LIMIT_PER_MINUTE` | `60` | Rate limit per minute |

### Example .env file
```bash
PORT=8080
JWT_SECRET=your-super-secret-jwt-key-here
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
MAX_CONNECTIONS=5000
```

## 🌐 API Endpoints

### WebSocket Endpoint
- **URL**: `/ws`
- **Protocol**: WebSocket
- **Description**: Main signaling endpoint for real-time communication

### HTTP Endpoints

#### Health Check
- **URL**: `/health`
- **Method**: GET
- **Response**: 
```json
{
  "status": "healthy",
  "time": "2024-01-01T12:00:00Z"
}
```

#### Statistics
- **URL**: `/stats`
- **Method**: GET
- **Response**:
```json
{
  "waiting_users": 5,
  "active_users": 10,
  "active_rooms": 5,
  "server_uptime": "2024-01-01T12:00:00Z"
}
```

## 📡 WebSocket Message Protocol

### Client → Server Messages

#### Session Initialization
Automatically sent when connection is established:
```json
{
  "type": "session",
  "payload": {
    "user_id": "uuid-here",
    "token": "jwt-token-here"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Find Match
```json
{
  "type": "find_match"
}
```

#### WebRTC Signaling
```json
{
  "type": "offer|answer|ice_candidate",
  "payload": { /* WebRTC data */ },
  "to": "partner-user-id"
}
```

#### Disconnect
```json
{
  "type": "disconnect"
}
```

#### Heartbeat Response
```json
{
  "type": "pong"
}
```

### Server → Client Messages

#### Match Found
```json
{
  "type": "match_found",
  "payload": {
    "partner_id": "partner-uuid",
    "room_id": "room-uuid",
    "role": "caller|callee"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Waiting for Match
```json
{
  "type": "waiting",
  "payload": {
    "status": "Looking for a partner..."
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Partner Disconnected
```json
{
  "type": "partner_disconnected",
  "payload": {
    "reason": "Partner left the conversation"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Heartbeat
```json
{
  "type": "ping",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 🔒 Security Considerations

### Current Implementation
- JWT-based session management
- CORS protection (configurable)
- Connection timeout handling
- Rate limiting (configurable)

### Production Recommendations
1. **Use HTTPS/WSS**: Always use secure connections in production
2. **JWT Secret**: Use a strong, randomly generated JWT secret
3. **CORS**: Specify exact allowed origins instead of `*`
4. **Rate Limiting**: Implement proper rate limiting per IP
5. **Input Validation**: Add comprehensive input validation
6. **Logging**: Implement structured logging with log levels
7. **Monitoring**: Add metrics collection (Prometheus/Grafana)

## 🚀 Deployment

### Docker Deployment
```dockerfile
FROM golang:1.22.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o voice-chat-app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/voice-chat-app .
EXPOSE 8080
CMD ["./voice-chat-app"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: voice-chat-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: voice-chat-backend
  template:
    metadata:
      labels:
        app: voice-chat-backend
    spec:
      containers:
      - name: voice-chat-backend
        image: your-registry/voice-chat-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: voice-chat-secrets
              key: jwt-secret
```

## 📊 Performance & Scalability

### Current Capabilities
- **Concurrent Connections**: 1000+ (configurable)
- **Memory Usage**: ~50MB base + ~1KB per connection
- **CPU Usage**: Low latency message routing
- **Cleanup**: Automatic cleanup of stale connections

### Scaling Strategies
1. **Horizontal Scaling**: Deploy multiple instances behind a load balancer
2. **Database Integration**: Add Redis for session persistence across instances
3. **Message Queue**: Use Redis Pub/Sub for cross-instance communication
4. **CDN**: Use WebSocket-compatible CDN for global distribution

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Benchmark tests
go test -bench=. ./...
```

## 🔧 Development

### Project Structure
```
backend/
├── handlers/          # HTTP and WebSocket handlers
├── models/           # Data models and business logic
├── utils/            # Utility functions and configuration
├── tests/            # Test files
├── bin/              # Compiled binaries
├── main.go           # Application entry point
├── go.mod            # Go module definition
└── Makefile          # Build automation
```

### Adding New Features
1. Add models in `models/` package
2. Implement handlers in `handlers/` package
3. Add utilities in `utils/` package
4. Update main.go for new routes
5. Add tests in `tests/` package

## 🤝 Integration with React-Expo Frontend

### WebSocket Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  handleMessage(message);
};

// Find match
ws.send(JSON.stringify({ type: 'find_match' }));
```

### WebRTC Integration
```javascript
// Handle match found
if (message.type === 'match_found') {
  const { partner_id, room_id, role } = message.payload;
  
  if (role === 'caller') {
    // Create offer
    const offer = await peerConnection.createOffer();
    await peerConnection.setLocalDescription(offer);
    
    ws.send(JSON.stringify({
      type: 'offer',
      payload: offer,
      to: partner_id
    }));
  }
}
```

## 📝 License

MIT License - see LICENSE file for details.

## 🆘 Support

For issues and questions:
1. Check the logs: `tail -f /var/log/voice-chat-app.log`
2. Monitor stats: `curl http://localhost:8080/stats`
3. Health check: `curl http://localhost:8080/health` 