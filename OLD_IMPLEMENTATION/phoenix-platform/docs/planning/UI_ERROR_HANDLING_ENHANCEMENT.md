# UI Error Handling Enhancement Plan

## Overview
This document outlines improvements to error handling and user feedback in the Phoenix dashboard, addressing the functional gap where deployment failures and configuration errors may not be clearly communicated to users.

## Current State Analysis

### Existing Error Handling
- Basic error state in stores (`error: string | null`)
- Simple try-catch blocks that set error messages
- Console logging for debugging
- No structured error types or codes

### Gaps Identified
1. **Generic Error Messages**: Users see `error.message` which may be technical
2. **No Error Context**: Missing deployment phase, specific failure reasons
3. **Limited Recovery Guidance**: No actionable steps for users
4. **Silent Failures**: Some errors only logged to console
5. **No Real-time Updates**: Deployment failures may not surface immediately

## Enhancement Design

### 1. Structured Error System

```typescript
// src/types/errors.ts
export enum ErrorCode {
  // Deployment Errors
  DEPLOYMENT_CONFIG_INVALID = 'DEPLOYMENT_CONFIG_INVALID',
  DEPLOYMENT_NODES_NOT_FOUND = 'DEPLOYMENT_NODES_NOT_FOUND',
  DEPLOYMENT_TIMEOUT = 'DEPLOYMENT_TIMEOUT',
  DEPLOYMENT_RESOURCE_LIMIT = 'DEPLOYMENT_RESOURCE_LIMIT',
  
  // Pipeline Errors
  PIPELINE_INVALID_CONFIG = 'PIPELINE_INVALID_CONFIG',
  PIPELINE_TEMPLATE_NOT_FOUND = 'PIPELINE_TEMPLATE_NOT_FOUND',
  PIPELINE_PARAMETER_MISSING = 'PIPELINE_PARAMETER_MISSING',
  
  // Experiment Errors
  EXPERIMENT_OVERLAP_DETECTED = 'EXPERIMENT_OVERLAP_DETECTED',
  EXPERIMENT_QUOTA_EXCEEDED = 'EXPERIMENT_QUOTA_EXCEEDED',
  EXPERIMENT_VALIDATION_FAILED = 'EXPERIMENT_VALIDATION_FAILED',
  
  // System Errors
  API_UNAVAILABLE = 'API_UNAVAILABLE',
  WEBSOCKET_CONNECTION_LOST = 'WEBSOCKET_CONNECTION_LOST',
  AUTHENTICATION_EXPIRED = 'AUTHENTICATION_EXPIRED',
}

export interface PhoenixError {
  code: ErrorCode
  message: string
  details?: string
  userAction?: string
  technicalDetails?: any
  timestamp: Date
  experimentId?: string
  deploymentPhase?: string
}

export class ErrorHandler {
  static createError(code: ErrorCode, context?: any): PhoenixError {
    const errorMappings = {
      [ErrorCode.DEPLOYMENT_CONFIG_INVALID]: {
        message: 'Invalid deployment configuration',
        userAction: 'Check your pipeline configuration for syntax errors',
      },
      [ErrorCode.DEPLOYMENT_NODES_NOT_FOUND]: {
        message: 'No nodes match your target selector',
        userAction: 'Verify your label selectors match existing nodes',
      },
      [ErrorCode.DEPLOYMENT_TIMEOUT]: {
        message: 'Deployment timed out',
        userAction: 'Check cluster resources and try again',
      },
      // ... more mappings
    }
    
    const mapping = errorMappings[code] || {
      message: 'An unexpected error occurred',
      userAction: 'Please try again or contact support',
    }
    
    return {
      code,
      message: mapping.message,
      userAction: mapping.userAction,
      details: context?.details,
      technicalDetails: context?.technical,
      timestamp: new Date(),
      experimentId: context?.experimentId,
      deploymentPhase: context?.phase,
    }
  }
}
```

### 2. Enhanced Store Error Handling

```typescript
// Updated useExperimentStore.ts
interface ExperimentState {
  // ... existing state
  errors: PhoenixError[] // Multiple errors can occur
  warnings: string[]     // Non-critical issues
  
  // New actions
  addError: (error: PhoenixError) => void
  addWarning: (warning: string) => void
  clearErrors: () => void
  clearWarnings: () => void
}

// Enhanced error handling in actions
startExperiment: async (id: string) => {
  set({ loading: true, errors: [] })
  try {
    const response = await apiService.startExperiment(id)
    
    // Check for deployment warnings
    if (response.warnings) {
      response.warnings.forEach(w => get().addWarning(w))
    }
    
    // Update experiment with deployment status
    set((state) => ({
      experiments: state.experiments.map((exp) =>
        exp.id === id
          ? { 
              ...exp, 
              status: 'initializing',
              deploymentStatus: response.deploymentStatus,
              startedAt: Date.now() 
            }
          : exp
      ),
      loading: false,
    }))
    
  } catch (error) {
    const phoenixError = ErrorHandler.createError(
      error.code || ErrorCode.EXPERIMENT_VALIDATION_FAILED,
      {
        experimentId: id,
        details: error.response?.data?.message,
        technical: error,
      }
    )
    
    set((state) => ({
      errors: [...state.errors, phoenixError],
      loading: false,
    }))
  }
},
```

### 3. Real-time Status Updates

```typescript
// src/hooks/useDeploymentStatus.ts
export function useDeploymentStatus(experimentId: string) {
  const [status, setStatus] = useState<DeploymentStatus | null>(null)
  const [errors, setErrors] = useState<PhoenixError[]>([])
  const { addError } = useExperimentStore()
  
  useEffect(() => {
    const ws = new WebSocket(`${WS_URL}/experiments/${experimentId}/status`)
    
    ws.onmessage = (event) => {
      const update = JSON.parse(event.data)
      
      switch (update.type) {
        case 'deployment_progress':
          setStatus(update.status)
          break
          
        case 'deployment_error':
          const error = ErrorHandler.createError(
            update.errorCode,
            {
              experimentId,
              phase: update.phase,
              details: update.message,
            }
          )
          setErrors(prev => [...prev, error])
          addError(error)
          break
          
        case 'deployment_warning':
          // Handle warnings
          break
      }
    }
    
    return () => ws.close()
  }, [experimentId])
  
  return { status, errors }
}
```

### 4. UI Components for Error Display

```typescript
// src/components/ErrorAlert.tsx
export function ErrorAlert({ error }: { error: PhoenixError }) {
  return (
    <Alert severity="error" className="mb-4">
      <AlertTitle>{error.message}</AlertTitle>
      {error.details && (
        <Typography variant="body2">{error.details}</Typography>
      )}
      {error.userAction && (
        <Box mt={1}>
          <Typography variant="body2" fontWeight="bold">
            What to do:
          </Typography>
          <Typography variant="body2">{error.userAction}</Typography>
        </Box>
      )}
      {error.deploymentPhase && (
        <Typography variant="caption" color="text.secondary">
          Failed during: {error.deploymentPhase}
        </Typography>
      )}
    </Alert>
  )
}

// src/components/DeploymentStatusCard.tsx
export function DeploymentStatusCard({ experimentId }: { experimentId: string }) {
  const { status, errors } = useDeploymentStatus(experimentId)
  
  return (
    <Card>
      <CardContent>
        <Typography variant="h6">Deployment Status</Typography>
        
        {/* Progress indicator */}
        <DeploymentProgress 
          phase={status?.phase}
          progress={status?.progress}
        />
        
        {/* Error display */}
        {errors.map((error, idx) => (
          <ErrorAlert key={idx} error={error} />
        ))}
        
        {/* Detailed status */}
        {status?.details && (
          <Box mt={2}>
            <Typography variant="body2" color="text.secondary">
              {status.details.message}
            </Typography>
            {status.details.nodesReady && (
              <Typography variant="caption">
                Nodes ready: {status.details.nodesReady}/{status.details.nodesTotal}
              </Typography>
            )}
          </Box>
        )}
      </CardContent>
    </Card>
  )
}
```

### 5. Error Recovery Flows

```typescript
// src/components/ExperimentActions.tsx
export function ExperimentActions({ experiment }: { experiment: Experiment }) {
  const { errors } = useExperimentStore()
  const relevantErrors = errors.filter(e => e.experimentId === experiment.id)
  
  const handleRetry = async () => {
    // Clear previous errors
    clearErrors()
    
    // Retry based on error type
    const lastError = relevantErrors[relevantErrors.length - 1]
    if (lastError?.code === ErrorCode.DEPLOYMENT_TIMEOUT) {
      // Retry with extended timeout
      await retryExperiment(experiment.id, { timeout: '10m' })
    } else {
      // Standard retry
      await retryExperiment(experiment.id)
    }
  }
  
  return (
    <Box>
      {experiment.status === 'failed' && (
        <Button
          onClick={handleRetry}
          startIcon={<RefreshIcon />}
          variant="outlined"
        >
          Retry Deployment
        </Button>
      )}
      
      {relevantErrors.map((error) => (
        <ErrorRecoveryActions key={error.timestamp} error={error} />
      ))}
    </Box>
  )
}
```

### 6. Global Error Boundary

```typescript
// src/components/ErrorBoundary.tsx
export class PhoenixErrorBoundary extends React.Component {
  state = { hasError: false, error: null }
  
  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }
  
  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log to monitoring service
    console.error('Phoenix Error Boundary:', error, errorInfo)
    
    // Send to error tracking
    if (window.Sentry) {
      window.Sentry.captureException(error, {
        contexts: { react: errorInfo }
      })
    }
  }
  
  render() {
    if (this.state.hasError) {
      return (
        <Container>
          <Alert severity="error">
            <AlertTitle>Something went wrong</AlertTitle>
            <Typography>
              The application encountered an unexpected error. 
              Please refresh the page to continue.
            </Typography>
            <Button onClick={() => window.location.reload()}>
              Refresh Page
            </Button>
          </Alert>
        </Container>
      )
    }
    
    return this.props.children
  }
}
```

## Implementation Plan

### Phase 1: Core Error System (3 days)
1. Implement error types and ErrorHandler
2. Update stores with structured error handling
3. Add error persistence to prevent loss on navigation

### Phase 2: UI Components (2 days)
1. Create ErrorAlert component
2. Add DeploymentStatusCard
3. Implement error recovery actions
4. Add global error boundary

### Phase 3: Real-time Updates (3 days)
1. Enhance WebSocket integration
2. Add deployment status streaming
3. Implement progress indicators

### Phase 4: Testing & Polish (2 days)
1. Unit tests for error handling
2. E2E tests for error scenarios
3. Documentation updates

## Success Metrics

1. **Error Visibility**: 100% of deployment errors shown in UI
2. **Recovery Rate**: 80% of retryable errors successfully recovered
3. **User Understanding**: 90% reduction in support tickets about errors
4. **Response Time**: Errors displayed within 2 seconds of occurrence

## Future Enhancements

1. **Error Analytics**: Track common errors and patterns
2. **Predictive Warnings**: Alert users before errors occur
3. **Guided Troubleshooting**: Step-by-step resolution wizards
4. **Integration with Monitoring**: Link to logs/metrics for debugging

---

This enhancement plan addresses the functional gap in error reporting, ensuring users have clear visibility into deployment issues and actionable guidance for resolution.