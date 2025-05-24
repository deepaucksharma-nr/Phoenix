# Phoenix Dashboard Technical Specification

## Overview

The Phoenix Dashboard is a React-based single-page application (SPA) that provides the primary user interface for the Phoenix Process Metrics Optimization Platform. It enables visual pipeline creation, experiment management, and comprehensive metrics analysis.

## Service Identity

- **Service Name**: phoenix-dashboard
- **Type**: Frontend Web Application
- **Repository Path**: `/dashboard/`
- **Port**: 3000 (development), 80/443 (production)
- **Technology Stack**: React 18, TypeScript, Material-UI, React Flow

## Architecture

### Component Architecture

```
dashboard/
├── src/
│   ├── components/           # Reusable UI components
│   │   ├── common/          # Generic components
│   │   ├── ExperimentBuilder/  # Pipeline visual builder
│   │   ├── ExperimentList/    # Experiment management
│   │   ├── MetricsDashboard/  # Metrics visualization
│   │   └── Layout/            # App layout components
│   ├── services/            # API client services
│   ├── hooks/               # Custom React hooks
│   ├── store/               # Redux store
│   ├── types/               # TypeScript definitions
│   ├── utils/               # Utility functions
│   └── pages/               # Route-based pages
├── public/                  # Static assets
└── tests/                   # Test files
```

### State Management

```typescript
// Redux Store Structure
interface RootState {
  auth: AuthState;
  experiments: ExperimentsState;
  pipelines: PipelinesState;
  metrics: MetricsState;
  ui: UIState;
}

interface ExperimentsState {
  list: Experiment[];
  current: Experiment | null;
  loading: boolean;
  error: string | null;
  filters: ExperimentFilters;
}
```

## Core Components

### 1. Pipeline Canvas (Visual Builder)

```typescript
interface PipelineCanvasProps {
  pipeline: Pipeline;
  onChange: (pipeline: Pipeline) => void;
  readonly?: boolean;
}

// Node Types
type NodeType = 'receiver' | 'processor' | 'exporter' | 'connector';

interface PipelineNode {
  id: string;
  type: NodeType;
  position: { x: number; y: number };
  data: {
    label: string;
    config: Record<string, any>;
    validation: ValidationResult;
  };
}

// Features
- Drag-and-drop node creation
- Real-time validation
- Auto-layout algorithm
- Undo/redo support
- Import/export YAML
```

### 2. Experiment Management

```typescript
interface ExperimentBuilderProps {
  onSubmit: (experiment: ExperimentSpec) => Promise<void>;
  initialData?: Partial<ExperimentSpec>;
}

// Workflow States
enum ExperimentWorkflow {
  DRAFT = 'draft',
  CONFIGURATION = 'configuration',
  VALIDATION = 'validation',
  REVIEW = 'review',
  ACTIVE = 'active'
}

// Features
- Multi-step wizard
- Pipeline A/B configuration
- Target host selection
- Schedule configuration
- Success criteria definition
```

### 3. Metrics Dashboard

```typescript
interface MetricsDashboardProps {
  experimentId: string;
  timeRange: TimeRange;
  refreshInterval?: number;
}

// Metric Panels
interface MetricPanel {
  id: string;
  type: 'chart' | 'stat' | 'table' | 'heatmap';
  metric: MetricQuery;
  layout: GridLayout;
}

// Features
- Real-time metric updates
- Customizable dashboard layouts
- Metric comparison views
- Export capabilities
- Alerting integration
```

## API Integration

### API Client Service

```typescript
class PhoenixAPIClient {
  private httpClient: AxiosInstance;
  private grpcClient: PhoenixServiceClient;

  constructor(config: APIConfig) {
    this.httpClient = axios.create({
      baseURL: config.apiUrl,
      timeout: config.timeout || 30000,
    });
    
    this.grpcClient = new PhoenixServiceClient(
      config.grpcUrl,
      config.grpcCredentials
    );
  }

  // REST endpoints for CRUD operations
  async listExperiments(params: ListParams): Promise<ExperimentList> {
    return this.httpClient.get('/v1/experiments', { params });
  }

  // gRPC for real-time operations
  streamMetrics(request: MetricStreamRequest): Observable<MetricData> {
    return this.grpcClient.streamMetrics(request);
  }
}
```

### WebSocket Integration

```typescript
interface WebSocketManager {
  connect(): Promise<void>;
  subscribe(channel: string, handler: MessageHandler): Subscription;
  unsubscribe(subscription: Subscription): void;
  disconnect(): void;
}

// Real-time updates
const channels = {
  experiments: 'experiments:updates',
  metrics: 'metrics:live',
  alerts: 'alerts:new',
  system: 'system:status'
};
```

## Routing

```typescript
const routes: RouteConfig[] = [
  {
    path: '/',
    component: Dashboard,
    protected: true,
  },
  {
    path: '/experiments',
    component: ExperimentList,
    protected: true,
  },
  {
    path: '/experiments/new',
    component: ExperimentBuilder,
    protected: true,
    permissions: ['experiment:create'],
  },
  {
    path: '/experiments/:id',
    component: ExperimentDetail,
    protected: true,
  },
  {
    path: '/pipelines',
    component: PipelineLibrary,
    protected: true,
  },
  {
    path: '/metrics',
    component: MetricsDashboard,
    protected: true,
  },
  {
    path: '/login',
    component: Login,
    protected: false,
  },
];
```

## Authentication & Authorization

### JWT Token Management

```typescript
class AuthService {
  private tokenKey = 'phoenix_token';
  private refreshTokenKey = 'phoenix_refresh_token';

  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await api.post('/auth/login', credentials);
    this.setTokens(response.data);
    return response.data;
  }

  async refreshToken(): Promise<string> {
    const refreshToken = this.getRefreshToken();
    const response = await api.post('/auth/refresh', { 
      refreshToken 
    });
    this.setTokens(response.data);
    return response.data.accessToken;
  }

  hasPermission(permission: string): boolean {
    const token = this.getDecodedToken();
    return token?.permissions?.includes(permission) ?? false;
  }
}
```

### Protected Route Component

```typescript
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ 
  children, 
  permissions = [] 
}) => {
  const { isAuthenticated, hasPermissions } = useAuth();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }
  
  if (permissions.length > 0 && !hasPermissions(permissions)) {
    return <AccessDenied />;
  }
  
  return <>{children}</>;
};
```

## Performance Optimization

### Code Splitting

```typescript
// Lazy load heavy components
const ExperimentBuilder = lazy(() => 
  import('./components/ExperimentBuilder')
);

const MetricsDashboard = lazy(() => 
  import('./components/MetricsDashboard')
);

// Route-based splitting
const routes = [
  {
    path: '/experiments/new',
    component: () => (
      <Suspense fallback={<Loading />}>
        <ExperimentBuilder />
      </Suspense>
    ),
  },
];
```

### Memoization Strategy

```typescript
// Expensive computations
const useProcessedMetrics = (rawMetrics: Metric[]) => {
  return useMemo(() => {
    return processMetrics(rawMetrics);
  }, [rawMetrics]);
};

// Heavy components
const MemoizedPipelineCanvas = memo(PipelineCanvas, (prev, next) => {
  return (
    prev.pipeline.id === next.pipeline.id &&
    prev.pipeline.version === next.pipeline.version &&
    prev.readonly === next.readonly
  );
});
```

### Virtual Scrolling

```typescript
// For large lists
const VirtualExperimentList: React.FC<Props> = ({ experiments }) => {
  const rowRenderer = ({ index, style }) => (
    <div style={style}>
      <ExperimentRow experiment={experiments[index]} />
    </div>
  );

  return (
    <AutoSizer>
      {({ height, width }) => (
        <List
          height={height}
          width={width}
          rowCount={experiments.length}
          rowHeight={72}
          rowRenderer={rowRenderer}
        />
      )}
    </AutoSizer>
  );
};
```

## Testing Strategy

### Unit Tests

```typescript
// Component testing
describe('ExperimentBuilder', () => {
  it('should validate pipeline configuration', async () => {
    const { getByRole, getByText } = render(
      <ExperimentBuilder onSubmit={jest.fn()} />
    );
    
    // Add invalid configuration
    fireEvent.click(getByRole('button', { name: 'Add Processor' }));
    fireEvent.change(getByRole('textbox', { name: 'Name' }), {
      target: { value: '' }
    });
    
    // Verify validation error
    expect(getByText('Name is required')).toBeInTheDocument();
  });
});
```

### Integration Tests

```typescript
// API integration testing
describe('PhoenixAPIClient', () => {
  it('should handle token refresh on 401', async () => {
    const client = new PhoenixAPIClient(config);
    
    // Mock expired token response
    mockServer.use(
      rest.get('/v1/experiments', (req, res, ctx) => {
        return res.once(ctx.status(401));
      })
    );
    
    // Should automatically refresh and retry
    const experiments = await client.listExperiments({});
    expect(experiments).toBeDefined();
  });
});
```

### E2E Tests

```typescript
// Cypress E2E tests
describe('Experiment Creation Flow', () => {
  it('should create new experiment', () => {
    cy.login();
    cy.visit('/experiments/new');
    
    // Fill experiment details
    cy.get('[data-cy=experiment-name]').type('Test Experiment');
    cy.get('[data-cy=baseline-pipeline]').select('baseline-v1');
    
    // Configure variant
    cy.get('[data-cy=add-variant]').click();
    cy.get('[data-cy=variant-pipeline]').select('optimized-v1');
    
    // Submit and verify
    cy.get('[data-cy=submit]').click();
    cy.location('pathname').should('match', /\/experiments\/[\w-]+$/);
  });
});
```

## Build & Deployment

### Build Configuration

```javascript
// webpack.config.js
module.exports = {
  entry: './src/index.tsx',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: '[name].[contenthash].js',
    publicPath: '/',
  },
  optimization: {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          priority: 10,
        },
        common: {
          minChunks: 2,
          priority: 5,
          reuseExistingChunk: true,
        },
      },
    },
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './public/index.html',
      favicon: './public/favicon.ico',
    }),
    new CompressionPlugin({
      algorithm: 'gzip',
      test: /\.(js|css|html|svg)$/,
      threshold: 8192,
      minRatio: 0.8,
    }),
  ],
};
```

### Docker Configuration

```dockerfile
# Multi-stage build
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy build artifacts
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Security headers
RUN echo 'add_header X-Frame-Options "SAMEORIGIN" always;' \
      >> /etc/nginx/conf.d/security-headers.conf && \
    echo 'add_header X-Content-Type-Options "nosniff" always;' \
      >> /etc/nginx/conf.d/security-headers.conf && \
    echo 'add_header Content-Security-Policy "default-src '"'"'self'"'"';" always;' \
      >> /etc/nginx/conf.d/security-headers.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Environment Configuration

```typescript
// config/environment.ts
interface Environment {
  production: boolean;
  apiUrl: string;
  grpcUrl: string;
  wsUrl: string;
  features: FeatureFlags;
}

const environments: Record<string, Environment> = {
  development: {
    production: false,
    apiUrl: 'http://localhost:8080',
    grpcUrl: 'localhost:9090',
    wsUrl: 'ws://localhost:8080/ws',
    features: {
      experimentBuilder: true,
      metricsStreaming: true,
      advancedAnalytics: false,
    },
  },
  production: {
    production: true,
    apiUrl: 'https://api.phoenix.example.com',
    grpcUrl: 'api.phoenix.example.com:443',
    wsUrl: 'wss://api.phoenix.example.com/ws',
    features: {
      experimentBuilder: true,
      metricsStreaming: true,
      advancedAnalytics: true,
    },
  },
};
```

## Monitoring & Observability

### Frontend Monitoring

```typescript
// Error boundary with reporting
class ErrorBoundary extends Component<Props, State> {
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log to monitoring service
    monitoringService.logError({
      error: error.toString(),
      componentStack: errorInfo.componentStack,
      props: this.props,
      url: window.location.href,
      userAgent: navigator.userAgent,
    });
  }
}

// Performance monitoring
const performanceObserver = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    monitoringService.logMetric({
      name: entry.name,
      duration: entry.duration,
      type: entry.entryType,
    });
  }
});

performanceObserver.observe({ 
  entryTypes: ['navigation', 'resource', 'paint'] 
});
```

### User Analytics

```typescript
// Track user interactions
const useAnalytics = () => {
  const trackEvent = useCallback((event: AnalyticsEvent) => {
    if (config.analytics.enabled) {
      analytics.track(event.name, {
        ...event.properties,
        timestamp: Date.now(),
        sessionId: getSessionId(),
      });
    }
  }, []);

  return { trackEvent };
};

// Usage
const ExperimentBuilder = () => {
  const { trackEvent } = useAnalytics();
  
  const handleSubmit = async (data: ExperimentSpec) => {
    trackEvent({
      name: 'experiment_created',
      properties: {
        pipelineType: data.pipelineType,
        targetCount: data.targets.length,
      },
    });
    
    await createExperiment(data);
  };
};
```

## Security Considerations

### Content Security Policy

```typescript
// CSP configuration
const cspDirectives = {
  'default-src': ["'self'"],
  'script-src': ["'self'", "'unsafe-inline'", 'https://cdn.example.com'],
  'style-src': ["'self'", "'unsafe-inline'"],
  'img-src': ["'self'", 'data:', 'https:'],
  'connect-src': ["'self'", 'https://api.phoenix.example.com'],
  'font-src': ["'self'"],
  'object-src': ["'none'"],
  'media-src': ["'self'"],
  'frame-src': ["'none'"],
};
```

### XSS Prevention

```typescript
// Sanitize user input
import DOMPurify from 'dompurify';

const SafeMarkdown: React.FC<{ content: string }> = ({ content }) => {
  const sanitized = DOMPurify.sanitize(content, {
    ALLOWED_TAGS: ['p', 'b', 'i', 'em', 'strong', 'a', 'ul', 'li'],
    ALLOWED_ATTR: ['href', 'target'],
  });
  
  return <div dangerouslySetInnerHTML={{ __html: sanitized }} />;
};
```

## Performance Requirements

- **Initial Load Time**: < 3 seconds on 3G connection
- **Time to Interactive**: < 5 seconds
- **API Response Time**: < 200ms for 95th percentile
- **Frame Rate**: 60 FPS for animations
- **Bundle Size**: < 500KB gzipped

## Browser Support

- Chrome/Edge: Last 2 versions
- Firefox: Last 2 versions
- Safari: Last 2 versions
- Mobile: iOS Safari 14+, Chrome Android 90+

## Accessibility Requirements

- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader compatibility
- High contrast mode support
- Focus indicators
- ARIA labels and roles

## Development Workflow

### Local Development

```bash
# Install dependencies
npm install

# Start development server
npm start

# Run tests
npm test

# Build for production
npm run build

# Analyze bundle
npm run analyze
```

### Code Quality Tools

```json
{
  "scripts": {
    "lint": "eslint src --ext .ts,.tsx",
    "format": "prettier --write 'src/**/*.{ts,tsx,css}'",
    "type-check": "tsc --noEmit",
    "test": "jest --coverage",
    "test:e2e": "cypress run",
    "storybook": "start-storybook -p 6006"
  }
}
```

## Compliance with Static Analysis Rules

### Folder Structure Validation

```yaml
dashboard_structure:
  pattern: "dashboard/*"
  requires:
    - src/components/
    - src/services/
    - src/types/
    - src/hooks/
    - src/store/
    - public/
    - tests/
  enforced_by:
    - ESLint import rules
    - TypeScript path mappings
    - CI/CD validation scripts
```

### Import Rules

```javascript
// .eslintrc.js
module.exports = {
  rules: {
    'import/no-restricted-paths': ['error', {
      zones: [
        {
          target: './src/components',
          from: './src/pages',
          message: 'Components should not import from pages'
        },
        {
          target: './src/services',
          from: './src/components',
          message: 'Services should not depend on components'
        }
      ]
    }]
  }
};
```