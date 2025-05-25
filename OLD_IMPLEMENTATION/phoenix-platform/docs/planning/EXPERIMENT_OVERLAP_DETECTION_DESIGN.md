# Experiment Overlap Detection Design

## Overview
This document outlines the implementation of experiment overlap detection to prevent conflicting experiments from running on the same nodes simultaneously, addressing a functional gap identified in the platform review.

## Problem Statement

Currently, Phoenix allows multiple experiments to target the same nodes without warning. This can cause:
- Resource contention between multiple collector instances
- Misleading metrics due to interference
- Unpredictable experiment results
- Potential system instability

## Design Goals

1. **Prevent Conflicts**: Block experiments that would interfere with each other
2. **Provide Clear Feedback**: Explain why an experiment cannot proceed
3. **Allow Overrides**: Support forced execution with explicit acknowledgment
4. **Maintain Performance**: Overlap checking should be fast (<100ms)

## Implementation Design

### 1. Overlap Detection Logic

```go
// pkg/api/overlap_detector.go
package api

import (
    "context"
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
)

type OverlapDetector interface {
    CheckOverlap(ctx context.Context, experiment *Experiment) (*OverlapResult, error)
}

type OverlapResult struct {
    HasOverlap       bool
    ConflictingExpIDs []string
    AffectedNodes    []string
    OverlapType      OverlapType
    Severity         OverlapSeverity
    Message          string
    Suggestions      []string
}

type OverlapType string

const (
    OverlapTypeNone       OverlapType = "none"
    OverlapTypePartial    OverlapType = "partial"    // Some nodes overlap
    OverlapTypeComplete   OverlapType = "complete"   // All nodes overlap
    OverlapTypeNamespace  OverlapType = "namespace"  // Same namespace
)

type OverlapSeverity string

const (
    SeverityNone     OverlapSeverity = "none"
    SeverityWarning  OverlapSeverity = "warning"  // Can proceed with caution
    SeverityError    OverlapSeverity = "error"    // Should not proceed
    SeverityBlocking OverlapSeverity = "blocking" // Cannot proceed
)

type overlapDetectorImpl struct {
    store       ExperimentStore
    k8sClient   kubernetes.Interface
}

func (o *overlapDetectorImpl) CheckOverlap(ctx context.Context, exp *Experiment) (*OverlapResult, error) {
    // Get all active experiments
    activeExps, err := o.store.ListActiveExperiments(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list active experiments: %w", err)
    }
    
    result := &OverlapResult{
        HasOverlap: false,
        OverlapType: OverlapTypeNone,
        Severity: SeverityNone,
    }
    
    // Parse target selector for new experiment
    newSelector, err := metav1.LabelSelectorAsSelector(&exp.TargetNodes)
    if err != nil {
        return nil, fmt.Errorf("invalid target selector: %w", err)
    }
    
    // Get nodes matching new experiment
    newNodes, err := o.getMatchingNodes(ctx, newSelector)
    if err != nil {
        return nil, err
    }
    
    // Check against each active experiment
    for _, activeExp := range activeExps {
        if activeExp.ID == exp.ID {
            continue // Skip self
        }
        
        activeSelector, err := metav1.LabelSelectorAsSelector(&activeExp.TargetNodes)
        if err != nil {
            continue // Skip invalid selectors
        }
        
        activeNodes, err := o.getMatchingNodes(ctx, activeSelector)
        if err != nil {
            continue
        }
        
        // Find overlapping nodes
        overlap := findNodeOverlap(newNodes, activeNodes)
        if len(overlap) > 0 {
            result.HasOverlap = true
            result.ConflictingExpIDs = append(result.ConflictingExpIDs, activeExp.ID)
            result.AffectedNodes = append(result.AffectedNodes, overlap...)
            
            // Determine overlap type and severity
            o.classifyOverlap(result, len(overlap), len(newNodes), len(activeNodes))
        }
    }
    
    // Generate user-friendly message and suggestions
    if result.HasOverlap {
        o.generateUserGuidance(result, exp)
    }
    
    return result, nil
}

func (o *overlapDetectorImpl) classifyOverlap(result *OverlapResult, overlapCount, newCount, activeCount int) {
    overlapPercentNew := float64(overlapCount) / float64(newCount) * 100
    overlapPercentActive := float64(overlapCount) / float64(activeCount) * 100
    
    // Determine type
    if overlapPercentNew >= 100 {
        result.OverlapType = OverlapTypeComplete
    } else if overlapPercentNew > 0 {
        result.OverlapType = OverlapTypePartial
    }
    
    // Determine severity
    if overlapPercentNew >= 100 || overlapPercentActive >= 100 {
        result.Severity = SeverityBlocking
    } else if overlapPercentNew >= 50 || overlapPercentActive >= 50 {
        result.Severity = SeverityError
    } else {
        result.Severity = SeverityWarning
    }
}

func (o *overlapDetectorImpl) generateUserGuidance(result *OverlapResult, exp *Experiment) {
    conflictCount := len(result.ConflictingExpIDs)
    nodeCount := len(result.AffectedNodes)
    
    // Generate message
    if result.Severity == SeverityBlocking {
        result.Message = fmt.Sprintf(
            "Cannot start experiment: All target nodes are already running %d other experiment(s)",
            conflictCount,
        )
    } else if result.Severity == SeverityError {
        result.Message = fmt.Sprintf(
            "Significant overlap detected: %d nodes are already running experiments",
            nodeCount,
        )
    } else {
        result.Message = fmt.Sprintf(
            "Minor overlap detected: %d nodes are running other experiments",
            nodeCount,
        )
    }
    
    // Generate suggestions
    result.Suggestions = []string{}
    
    if result.Severity == SeverityBlocking {
        result.Suggestions = append(result.Suggestions,
            "Wait for the existing experiments to complete",
            "Choose different target nodes using label selectors",
            "Stop the conflicting experiments if they are no longer needed",
        )
    } else {
        result.Suggestions = append(result.Suggestions,
            "Consider using different nodes to avoid interference",
            "Proceed with caution - results may be affected by other experiments",
            fmt.Sprintf("Conflicting experiments: %v", result.ConflictingExpIDs),
        )
    }
}
```

### 2. API Integration

```go
// Update experiment creation endpoint
func (s *ExperimentService) CreateExperiment(ctx context.Context, req *CreateExperimentRequest) (*Experiment, error) {
    // Validate request
    if err := s.validateExperimentRequest(req); err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
    }
    
    // Create experiment object
    exp := &Experiment{
        Name: req.Name,
        TargetNodes: req.TargetNodes,
        // ... other fields
    }
    
    // Check for overlaps
    overlapResult, err := s.overlapDetector.CheckOverlap(ctx, exp)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "overlap check failed: %v", err)
    }
    
    // Handle overlap based on severity
    if overlapResult.HasOverlap {
        switch overlapResult.Severity {
        case SeverityBlocking:
            // Cannot proceed
            return nil, status.Errorf(codes.FailedPrecondition,
                "Experiment blocked due to overlap: %s. Suggestions: %v",
                overlapResult.Message,
                overlapResult.Suggestions,
            )
            
        case SeverityError:
            // Check if force flag is set
            if !req.ForceOverlap {
                return nil, status.Errorf(codes.FailedPrecondition,
                    "Significant overlap detected: %s. Use force flag to override. Suggestions: %v",
                    overlapResult.Message,
                    overlapResult.Suggestions,
                )
            }
            // Log warning but proceed
            s.logger.Warn("Experiment created with overlap override",
                zap.String("experiment", exp.Name),
                zap.Any("overlap", overlapResult),
            )
            
        case SeverityWarning:
            // Log warning and proceed
            s.logger.Info("Experiment created with minor overlap",
                zap.String("experiment", exp.Name),
                zap.Int("affected_nodes", len(overlapResult.AffectedNodes)),
            )
        }
    }
    
    // Store overlap information with experiment
    exp.OverlapInfo = overlapResult
    
    // Proceed with creation
    return s.store.CreateExperiment(ctx, exp)
}
```

### 3. UI Integration

```typescript
// src/components/ExperimentForm/OverlapWarning.tsx
import { Alert, AlertTitle, Box, Button, Chip, List, ListItem } from '@mui/material'
import { OverlapResult } from '../../types/experiment'

interface OverlapWarningProps {
  overlap: OverlapResult
  onProceed: () => void
  onCancel: () => void
  onModify: () => void
}

export function OverlapWarning({ overlap, onProceed, onCancel, onModify }: OverlapWarningProps) {
  const getSeverityColor = () => {
    switch (overlap.severity) {
      case 'blocking': return 'error'
      case 'error': return 'error'
      case 'warning': return 'warning'
      default: return 'info'
    }
  }
  
  return (
    <Alert severity={getSeverityColor()} sx={{ mt: 2 }}>
      <AlertTitle>Experiment Overlap Detected</AlertTitle>
      
      <Box sx={{ mt: 1 }}>
        <Typography>{overlap.message}</Typography>
        
        {overlap.conflictingExpIds.length > 0 && (
          <Box sx={{ mt: 2 }}>
            <Typography variant="subtitle2">Conflicting Experiments:</Typography>
            {overlap.conflictingExpIds.map(id => (
              <Chip 
                key={id} 
                label={id} 
                size="small" 
                sx={{ mr: 1, mt: 0.5 }}
                clickable
                onClick={() => window.open(`/experiments/${id}`, '_blank')}
              />
            ))}
          </Box>
        )}
        
        {overlap.affectedNodes.length > 0 && (
          <Box sx={{ mt: 2 }}>
            <Typography variant="subtitle2">
              Affected Nodes ({overlap.affectedNodes.length}):
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {overlap.affectedNodes.slice(0, 5).join(', ')}
              {overlap.affectedNodes.length > 5 && ` and ${overlap.affectedNodes.length - 5} more...`}
            </Typography>
          </Box>
        )}
        
        {overlap.suggestions.length > 0 && (
          <Box sx={{ mt: 2 }}>
            <Typography variant="subtitle2">Suggestions:</Typography>
            <List dense>
              {overlap.suggestions.map((suggestion, idx) => (
                <ListItem key={idx}>
                  <Typography variant="body2">• {suggestion}</Typography>
                </ListItem>
              ))}
            </List>
          </Box>
        )}
      </Box>
      
      <Box sx={{ mt: 3, display: 'flex', gap: 1 }}>
        {overlap.severity !== 'blocking' && (
          <Button 
            variant="contained" 
            color="warning"
            onClick={onProceed}
          >
            Proceed Anyway
          </Button>
        )}
        <Button 
          variant="outlined"
          onClick={onModify}
        >
          Modify Target Nodes
        </Button>
        <Button 
          variant="text"
          onClick={onCancel}
        >
          Cancel
        </Button>
      </Box>
    </Alert>
  )
}
```

### 4. CLI Integration

```bash
# Check for overlaps before creating
phoenix experiment create --name "test-opt" \
  --baseline process-baseline-v1 \
  --candidate process-topk-v1 \
  --target-selector "app=frontend" \
  --check-overlap

# Output:
Warning: Experiment overlap detected
Severity: error
Message: Significant overlap detected: 5 nodes are already running experiments
Conflicting experiments:
  - exp-abc123 (running)
  - exp-def456 (running)
Affected nodes: node-1, node-2, node-3, node-4, node-5

Suggestions:
  - Consider using different nodes to avoid interference
  - Proceed with caution - results may be affected by other experiments

Do you want to proceed anyway? [y/N]: n
Experiment creation cancelled.

# Force creation despite overlap
phoenix experiment create --name "test-opt" \
  --baseline process-baseline-v1 \
  --candidate process-topk-v1 \
  --target-selector "app=frontend" \
  --force
```

### 5. Database Schema

```sql
-- migrations/006_add_overlap_tracking.sql
ALTER TABLE experiments ADD COLUMN overlap_info JSONB;
ALTER TABLE experiments ADD COLUMN force_overlap BOOLEAN DEFAULT FALSE;

-- Add index for efficient overlap queries
CREATE INDEX idx_experiments_active_status 
ON experiments(status) 
WHERE status IN ('running', 'initializing', 'pending');

-- Track overlap history
CREATE TABLE experiment_overlaps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID REFERENCES experiments(id),
    conflicting_exp_id UUID REFERENCES experiments(id),
    overlap_type VARCHAR(50),
    severity VARCHAR(50),
    affected_nodes JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Testing Strategy

### Unit Tests
```go
func TestOverlapDetection(t *testing.T) {
    tests := []struct {
        name           string
        newExp         *Experiment
        activeExps     []*Experiment
        expectedResult *OverlapResult
    }{
        {
            name: "no overlap",
            newExp: &Experiment{
                TargetNodes: metav1.LabelSelector{
                    MatchLabels: map[string]string{"app": "frontend"},
                },
            },
            activeExps: []*Experiment{
                {
                    TargetNodes: metav1.LabelSelector{
                        MatchLabels: map[string]string{"app": "backend"},
                    },
                },
            },
            expectedResult: &OverlapResult{
                HasOverlap: false,
                Severity: SeverityNone,
            },
        },
        {
            name: "complete overlap",
            newExp: &Experiment{
                TargetNodes: metav1.LabelSelector{
                    MatchLabels: map[string]string{"app": "web"},
                },
            },
            activeExps: []*Experiment{
                {
                    TargetNodes: metav1.LabelSelector{
                        MatchLabels: map[string]string{"app": "web"},
                    },
                },
            },
            expectedResult: &OverlapResult{
                HasOverlap: true,
                OverlapType: OverlapTypeComplete,
                Severity: SeverityBlocking,
            },
        },
    }
    
    // Run tests...
}
```

### Integration Tests
- Test with real Kubernetes cluster
- Verify node matching logic
- Test concurrent experiment creation
- Verify UI displays warnings correctly

## Rollout Plan

### Phase 1: Backend Implementation (3 days)
1. Implement OverlapDetector service
2. Add database schema changes
3. Integrate with experiment creation API
4. Add unit tests

### Phase 2: UI Integration (2 days)
1. Create OverlapWarning component
2. Integrate into experiment creation flow
3. Add visual indicators for affected nodes
4. Test user flows

### Phase 3: CLI Support (2 days)
1. Add --check-overlap flag
2. Implement interactive prompts
3. Add --force flag for overrides
4. Update documentation

## Configuration Options

```yaml
# values.yaml
experimentController:
  overlap:
    enabled: true
    defaultAction: "warn"  # Options: warn, block, allow
    thresholds:
      warningPercent: 25   # Warn if >25% nodes overlap
      errorPercent: 50     # Error if >50% nodes overlap
      blockingPercent: 100 # Block if 100% nodes overlap
    allowForceOverride: true
    cacheTimeout: 60s      # Cache node lists for performance
```

## Performance Considerations

1. **Node List Caching**: Cache Kubernetes node lists for 60s
2. **Indexed Queries**: Use database indexes for active experiments
3. **Parallel Checks**: Check multiple experiments concurrently
4. **Early Exit**: Stop checking once blocking overlap found

## Future Enhancements

1. **Overlap Scheduling**: Queue experiments to run after conflicts end
2. **Resource-Based Detection**: Consider CPU/memory limits, not just nodes
3. **Namespace Isolation**: Allow overlaps in different namespaces
4. **Overlap Analytics**: Track overlap patterns and suggest optimizations

---

This design provides comprehensive overlap detection that protects users from accidental conflicts while maintaining flexibility for advanced use cases.