# Load Testing Strategy for Voice Chat App

This document outlines comprehensive load testing strategies for the voice chat application to ensure it can handle production traffic and concurrent users.

## Overview

The voice chat app needs to handle:
- WebSocket connections for real-time communication
- User matchmaking and room management
- WebRTC signaling message relay
- Concurrent user sessions

## Load Testing Tools

### 1. Artillery.js (Recommended for WebSocket testing)

Install Artillery:
```bash
npm install -g artillery
```

Create `artillery-config.yml`:
```yaml
config:
  target: 'ws://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 120
      arrivalRate: 50
      name: "Ramp up load"
    - duration: 300
      arrivalRate: 100
      name: "Sustained load"
  ws:
    connect:
      timeout: 10
scenarios:
  - name: "WebSocket Connection Test"
    weight: 100
    engine: ws
    flow:
      - connect:
          url: "/ws"
      - think: 1
      - send:
          payload: '{"type": "find_match"}'
      - think: 5
      - send:
          payload: '{"type": "ping"}'
      - think: 10
```

Run the test:
```bash
artillery run artillery-config.yml
```

### 2. Vegeta (HTTP endpoints)

Install Vegeta:
```bash
go install github.com/tsenart/vegeta/attack@latest
```

Test health endpoint:
```bash
echo "GET http://localhost:8080/health" | vegeta attack -duration=30s -rate=100 | vegeta report
```

Test stats endpoint:
```bash
echo "GET http://localhost:8080/stats" | vegeta attack -duration=30s -rate=50 | vegeta report
```

### 3. Custom Go Load Test

Create `load_test.go`:
```go
package main

import (
    "fmt"
    "log"
    "net/url"
    "sync"
    "time"
    
    "github.com/gorilla/websocket"
)

func main() {
    concurrentUsers := 1000
    testDuration := 5 * time.Minute
    
    var wg sync.WaitGroup
    results := make(chan TestResult, concurrentUsers)
    
    start := time.Now()
    
    for i := 0; i < concurrentUsers; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            result := simulateUser(userID, testDuration)
            results <- result
        }(i)
        
        // Stagger connections to avoid overwhelming the server
        time.Sleep(10 * time.Millisecond)
    }
    
    wg.Wait()
    close(results)
    
    // Analyze results
    analyzeResults(results, time.Since(start))
}

type TestResult struct {
    UserID          int
    Connected       bool
    MatchFound      bool
    MessagesExchanged int
    Errors          []error
    Duration        time.Duration
}

func simulateUser(userID int, duration time.Duration) TestResult {
    result := TestResult{UserID: userID}
    start := time.Now()
    defer func() {
        result.Duration = time.Since(start)
    }()
    
    // Connect to WebSocket
    u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        result.Errors = append(result.Errors, err)
        return result
    }
    defer conn.Close()
    
    result.Connected = true
    
    // Read session message
    var sessionMsg map[string]interface{}
    if err := conn.ReadJSON(&sessionMsg); err != nil {
        result.Errors = append(result.Errors, err)
        return result
    }
    
    // Request match
    findMatchMsg := map[string]interface{}{
        "type": "find_match",
    }
    if err := conn.WriteJSON(findMatchMsg); err != nil {
        result.Errors = append(result.Errors, err)
        return result
    }
    
    // Wait for match or timeout
    timeout := time.After(30 * time.Second)
    for {
        select {
        case <-timeout:
            return result
        default:
            var msg map[string]interface{}
            if err := conn.ReadJSON(&msg); err != nil {
                result.Errors = append(result.Errors, err)
                return result
            }
            
            if msgType, ok := msg["type"].(string); ok {
                if msgType == "match_found" {
                    result.MatchFound = true
                    // Simulate some signaling messages
                    result.MessagesExchanged = simulateSignaling(conn, duration)
                    return result
                }
            }
        }
    }
}

func simulateSignaling(conn *websocket.Conn, duration time.Duration) int {
    messageCount := 0
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    timeout := time.After(duration)
    
    for {
        select {
        case <-timeout:
            return messageCount
        case <-ticker.C:
            // Send a fake ICE candidate
            iceMsg := map[string]interface{}{
                "type": "ice_candidate",
                "payload": map[string]interface{}{
                    "candidate": "fake-ice-candidate",
                },
            }
            if err := conn.WriteJSON(iceMsg); err != nil {
                return messageCount
            }
            messageCount++
        }
    }
}

func analyzeResults(results <-chan TestResult, totalDuration time.Duration) {
    var (
        totalUsers      int
        connectedUsers  int
        matchedUsers    int
        totalMessages   int
        totalErrors     int
    )
    
    for result := range results {
        totalUsers++
        if result.Connected {
            connectedUsers++
        }
        if result.MatchFound {
            matchedUsers++
        }
        totalMessages += result.MessagesExchanged
        totalErrors += len(result.Errors)
    }
    
    fmt.Printf("Load Test Results:\n")
    fmt.Printf("==================\n")
    fmt.Printf("Total Duration: %v\n", totalDuration)
    fmt.Printf("Total Users: %d\n", totalUsers)
    fmt.Printf("Connected Users: %d (%.2f%%)\n", connectedUsers, float64(connectedUsers)/float64(totalUsers)*100)
    fmt.Printf("Matched Users: %d (%.2f%%)\n", matchedUsers, float64(matchedUsers)/float64(totalUsers)*100)
    fmt.Printf("Total Messages: %d\n", totalMessages)
    fmt.Printf("Total Errors: %d\n", totalErrors)
    fmt.Printf("Connection Success Rate: %.2f%%\n", float64(connectedUsers)/float64(totalUsers)*100)
    fmt.Printf("Match Success Rate: %.2f%%\n", float64(matchedUsers)/float64(connectedUsers)*100)
}
```

## Load Testing Scenarios

### 1. Connection Load Test
- **Objective**: Test maximum concurrent WebSocket connections
- **Metrics**: Connection success rate, memory usage, CPU usage
- **Target**: 10,000 concurrent connections

### 2. Matchmaking Load Test
- **Objective**: Test user matching under load
- **Metrics**: Match success rate, match latency, queue length
- **Target**: 1,000 users requesting matches simultaneously

### 3. Signaling Load Test
- **Objective**: Test WebRTC signaling message relay
- **Metrics**: Message delivery rate, latency, dropped messages
- **Target**: 500 concurrent rooms with active signaling

### 4. Stress Test
- **Objective**: Find breaking point of the system
- **Metrics**: Error rates, response times, resource usage
- **Method**: Gradually increase load until system fails

### 5. Endurance Test
- **Objective**: Test system stability over time
- **Duration**: 24 hours
- **Load**: 70% of maximum capacity

## Monitoring During Load Tests

### Key Metrics to Monitor

1. **Application Metrics**:
   - Active WebSocket connections
   - Waiting users count
   - Active rooms count
   - Message throughput
   - Error rates

2. **System Metrics**:
   - CPU usage
   - Memory usage
   - Network I/O
   - File descriptors
   - Goroutine count

3. **Response Time Metrics**:
   - WebSocket connection time
   - Match finding time
   - Message relay latency

### Monitoring Tools

1. **Prometheus + Grafana**:
```go
// Add to your application
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "websocket_active_connections",
        Help: "Number of active WebSocket connections",
    })
    
    matchmakingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "matchmaking_duration_seconds",
        Help: "Time taken to find a match",
    })
)

func init() {
    prometheus.MustRegister(activeConnections)
    prometheus.MustRegister(matchmakingDuration)
}
```

2. **pprof for Go profiling**:
```go
import _ "net/http/pprof"

// Add to your main function
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

## Performance Targets

### Baseline Requirements
- **Concurrent Users**: 5,000
- **Match Latency**: < 2 seconds
- **Message Latency**: < 100ms
- **Uptime**: 99.9%
- **Memory Usage**: < 2GB for 5,000 users

### Stretch Goals
- **Concurrent Users**: 50,000
- **Match Latency**: < 1 second
- **Message Latency**: < 50ms
- **Uptime**: 99.99%

## Load Test Execution Plan

### Phase 1: Baseline Testing
1. Run unit and integration tests
2. Test with 100 concurrent users
3. Establish baseline metrics

### Phase 2: Gradual Load Increase
1. 500 users for 10 minutes
2. 1,000 users for 15 minutes
3. 2,500 users for 20 minutes
4. 5,000 users for 30 minutes

### Phase 3: Stress Testing
1. Increase load until system breaks
2. Identify bottlenecks
3. Optimize and repeat

### Phase 4: Endurance Testing
1. Run at 70% capacity for 24 hours
2. Monitor for memory leaks
3. Check for performance degradation

## Common Issues and Solutions

### 1. Too Many Open Files
```bash
# Increase file descriptor limits
ulimit -n 65536
```

### 2. Memory Leaks
- Monitor goroutine count
- Check for unclosed connections
- Profile memory usage with pprof

### 3. CPU Bottlenecks
- Profile CPU usage
- Optimize hot code paths
- Consider horizontal scaling

### 4. Network Saturation
- Monitor network I/O
- Optimize message sizes
- Consider message batching

## Continuous Load Testing

### CI/CD Integration
```yaml
# .github/workflows/load-test.yml
name: Load Test
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.22
      - name: Run Load Test
        run: |
          cd backend
          go run scripts/load_test.go
```

## Reporting

After each load test, generate a report including:
1. Test configuration and duration
2. Performance metrics and graphs
3. Error analysis
4. Resource utilization
5. Recommendations for optimization

This comprehensive load testing strategy ensures your voice chat application can handle production traffic and provides insights for optimization and scaling decisions. 