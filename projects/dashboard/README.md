# Phoenix Dashboard

## Overview

Modern React-based web interface for the Phoenix Platform with real-time monitoring, experiment management, and cost analytics. Built with React 18, TypeScript, and Vite for optimal development experience.

## Architecture

The dashboard connects to Phoenix API via REST and WebSocket (both on port 8080) for real-time updates:

```
┌─────────────────────────────────┐
│      Phoenix Dashboard          │
│   (React 18 + TypeScript)       │
├─────────────────────────────────┤
│  • Experiment Wizard            │
│  • Real-time Monitoring         │
│  • Cost Analytics               │
│  • Pipeline Builder             │
└──────────┬──────────────────────┘
           │
     REST + WebSocket
           │
    ┌──────▼──────┐
    │ Phoenix API │
    │ (Port 8080) │
    └─────────────┘
```

## Development

### Prerequisites

- Node.js 18+
- npm or yarn
- Docker (optional)
- Phoenix API running on port 8080

### Setup

```bash
# Install dependencies
npm install

# Run tests
npm test

# Build for production
npm run build
```

### Running Locally

```bash
# Start development server with hot reload
npm run dev

# Dashboard will be available at:
# http://localhost:3000

# API connection defaults to:
# http://localhost:8080 (REST + WebSocket)
```

## Configuration

Configuration is managed through environment variables:

```bash
# .env file
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
VITE_ENABLE_AUTH=false
VITE_REFRESH_INTERVAL=5000
```

For production builds, use `.env.production`.

## Key Features

### 1. Experiment Management
- **Wizard Interface**: Step-by-step experiment creation
- **A/B Testing**: Compare baseline vs candidate pipelines
- **Real-time Status**: Live updates via WebSocket
- **Cost Analytics**: See savings as they happen (70% demonstrated)

### 2. Pipeline Templates
- **Adaptive Filter**: Dynamic metric filtering
- **TopK**: Keep only top metrics by value
- **Hybrid**: Combined optimization strategies

### 3. Real-time Monitoring
- **Agent Fleet View**: Monitor all connected agents
- **Live Metrics**: Cardinality and cost reduction metrics
- **WebSocket Updates**: Instant experiment progress

### 4. Visual Pipeline Builder
- **Drag-and-Drop**: Create custom pipelines visually
- **Template Library**: Pre-configured optimization strategies
- **Impact Preview**: See estimated savings before deployment

## Testing

```bash
# Run unit tests
npm test

# Run tests in watch mode
npm run test:watch

# Run with coverage
npm run test:coverage

# Run E2E tests (requires API running)
npm run test:e2e
```

## Deployment

```bash
# Build Docker image
docker build -t phoenix/dashboard .

# Run container
docker run -p 3000:80 \
  -e VITE_API_URL=http://phoenix-api:8080 \
  phoenix/dashboard

# Or use docker-compose
docker-compose up dashboard
```

## Project Structure

```
dashboard/
├── src/
│   ├── components/     # Reusable UI components
│   ├── pages/         # Route pages
│   ├── hooks/         # Custom React hooks
│   ├── services/      # API and WebSocket services
│   ├── store/         # Redux store and slices
│   └── utils/         # Helper functions
├── public/            # Static assets
└── vite.config.ts     # Vite configuration
```

## WebSocket Integration

```typescript
// Real-time updates
const ws = useWebSocket();

ws.on('experiment_update', (data) => {
  // Update experiment status
  console.log(`Savings: ${data.savings_percent}%`);
});

ws.on('agent_status', (data) => {
  // Update agent fleet view
});
```

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md)
