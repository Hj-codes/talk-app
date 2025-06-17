# Deployment Guide - Render

This guide explains how to deploy the voice chat backend to Render.

## Prerequisites

1. GitHub repository with your code
2. Render account (free tier available)
3. Docker installed locally for testing

## Deployment Steps

### 1. Prepare Your Repository

Ensure your repository has the following files:
- `Dockerfile` (optimized for Render)
- `render.yaml` (deployment configuration)
- `.dockerignore` (build optimization)

### 2. Create Render Service

1. Go to [Render Dashboard](https://dashboard.render.com/)
2. Click "New" → "Web Service"
3. Connect your GitHub repository
4. Configure the service:
   - **Name**: `voice-chat-backend` (or your preferred name)
   - **Environment**: Docker
   - **Region**: Choose closest to your users
   - **Plan**: Starter (free tier) or higher
   - **Dockerfile Path**: `./backend/Dockerfile`
   - **Docker Context**: `./backend`

### 3. Environment Variables

Set these environment variables in Render dashboard:

**Required:**
- `ENVIRONMENT`: `production`
- `JWT_SECRET`: Generate a secure secret (32+ characters)
- `ALLOWED_ORIGINS`: Your frontend URLs (comma-separated)

**Optional (with defaults):**
- `LOG_LEVEL`: `info`
- `MAX_CONNECTIONS`: `1000`
- `HTTP_RATE_LIMIT_PER_MINUTE`: `60`
- `WS_RATE_LIMIT_PER_MINUTE`: `100`

### 4. Update Frontend Configuration

After deployment, update your frontend `api.ts`:

```typescript
const PRODUCTION_BASE_URL = 'https://your-service-name.onrender.com';
const PRODUCTION_WS_URL = 'wss://your-service-name.onrender.com/ws';
```

### 5. Configure CORS

Update the `ALLOWED_ORIGINS` environment variable to include:
- Your React Native development URLs
- Production frontend URLs
- Expo development URLs if using Expo

Example:
```
https://your-frontend.com,http://localhost:19006,exp://localhost:19000
```

## Health Checks

The service includes automatic health checks at `/health` endpoint.

## Monitoring

- **Logs**: Available in Render dashboard
- **Metrics**: Use `/stats` endpoint for connection statistics
- **Health**: Use `/health` endpoint for service status

## Free Tier Limitations

Render's free tier has these limitations:
- Service spins down after 15 minutes of inactivity
- 750 hours per month
- Slower cold start times

For production apps, consider upgrading to a paid plan.

## Troubleshooting

### Common Issues:

1. **Service Won't Start**
   - Check logs in Render dashboard
   - Verify Dockerfile builds locally
   - Ensure all required environment variables are set

2. **WebSocket Connection Fails**
   - Verify WSS protocol (not WS) for HTTPS sites
   - Check CORS configuration
   - Ensure frontend URLs are in ALLOWED_ORIGINS

3. **JWT Errors**
   - Ensure JWT_SECRET is set and secure
   - Check secret length (minimum 8 characters)

4. **Rate Limiting Issues**
   - Adjust rate limit environment variables
   - Monitor connection counts via `/stats`

### Testing Deployment

1. **Health Check**: `curl https://your-service.onrender.com/health`
2. **Stats Check**: `curl https://your-service.onrender.com/stats`
3. **WebSocket**: Test with frontend app

## Security Considerations

- Always use HTTPS in production
- Set secure JWT_SECRET (never commit to repository)
- Configure proper CORS origins
- Monitor rate limiting and connection limits
- Use environment variables for all sensitive data

## Scaling

For high traffic applications:
- Upgrade to higher Render plans
- Consider Redis integration for session management
- Implement horizontal scaling strategies
- Monitor connection limits and performance metrics 