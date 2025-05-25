# ADR-003: Visual Pipeline Builder as Primary Interface

## Status
Accepted

## Context
OpenTelemetry pipeline configurations are complex YAML files that require deep knowledge of OTel components. Users need an easier way to create and modify pipeline configurations.

## Decision
The PRIMARY interface for creating OpenTelemetry pipelines will be a visual drag-and-drop builder using React Flow. Direct YAML editing is secondary.

## Rationale
1. **Usability**: Visual interface is more intuitive than YAML
2. **Validation**: Real-time validation prevents invalid configurations
3. **Learning Curve**: Users don't need to learn OTel YAML syntax
4. **Error Prevention**: Drag-drop prevents structural errors
5. **Discoverability**: Visual components show available options

## Implementation
### Frontend (React Flow)
```typescript
// Component types
- ReceiverNode: Data ingestion points
- ProcessorNode: Data transformation
- ExporterNode: Data destinations
- Edges: Data flow connections

// Features
- Drag from component library
- Drop onto canvas
- Connect with edges
- Configure via property panels
- Export to YAML
```

### Backend Validation
```go
// Every visual configuration validated before saving
- Structural validation (proper connections)
- Semantic validation (compatible components)
- Performance validation (resource estimates)
```

## Consequences
### Positive
- Dramatically lower barrier to entry
- Fewer configuration errors
- Better user experience
- Visual representation aids understanding

### Negative
- Complex UI implementation
- Must maintain YAML compatibility
- Advanced users may prefer text editing

## UI/UX Principles
1. **Progressive Disclosure**: Simple by default, advanced when needed
2. **Immediate Feedback**: Real-time validation
3. **Template Starting Points**: Pre-built configurations
4. **Import/Export**: YAML compatibility maintained

## Alternatives Considered
1. **YAML-Only**: Too complex for average users
2. **Form-Based**: Less intuitive than visual
3. **Wizard-Based**: Too restrictive for complex pipelines
4. **Code Generation**: Still requires understanding syntax

## References
- UI mockups in TECHNICAL_SPEC_DASHBOARD.md
- User journey in user-guide.md
- React Flow implementation in PipelineCanvas.tsx