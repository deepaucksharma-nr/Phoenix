# Phoenix Platform PRD Status - Dashboard Streamlining

## Executive Summary

The Phoenix dashboard currently contains ~2,933 lines of redundant code that can be eliminated without impacting MVP functionality. This streamlining will reduce complexity by 40% and improve maintainability while focusing on core monitoring and viewing capabilities required by the PRD.

## Current State Analysis

### Redundancies Identified

1. **Duplicate API Services** (222 lines)
   - Identical files: `api.service.ts` and `api/api.service.ts`
   - Impact: Confusion, potential maintenance issues

2. **WebSocket Over-Engineering** (593 lines → 200 lines)
   - Three separate implementations
   - No unified state management approach

3. **Non-MVP Builder Components** (1,760 lines)
   - Complex drag-and-drop pipeline builders
   - Not required for Sprint 0 Process-Metrics MVP
   - MVP requires view-only functionality

4. **Redundant Notification Systems** (607 lines → 100 lines)
   - Multiple overlapping implementations
   - No single source of truth

5. **Test/Simple Components** (36 lines)
   - Unused placeholder files

## MVP Alignment

### PRD Requirements (Process-Metrics Sprint 0)

#### ✅ Required Features (Keep/Enhance)
- **FR-WEB-001**: Real-time metrics display
- **FR-WEB-005**: WebSocket connectivity 
- **FR-WEB-010**: Experiment monitoring
- **FR-WEB-020**: Metrics comparison
- **FR-WEB-030**: Pipeline catalog viewing

#### ❌ Not Required for MVP (Remove)
- Pipeline visual building/editing
- Drag-and-drop configuration
- Complex multi-step wizards
- Experiment creation UI (CLI-first approach)

## Streamlining Impact

### Code Reduction
- **Before**: ~10,000 lines in dashboard
- **After**: ~7,000 lines (30% reduction)
- **Removed**: 2,933 lines of redundant/non-MVP code

### Complexity Reduction
- Single state management system (Redux only)
- One WebSocket implementation
- Unified notification system
- Read-only pipeline viewing

### Performance Improvements
- Smaller bundle size (est. 25-30% reduction)
- Faster initial load time
- Reduced memory footprint
- Fewer dependencies

## Implementation Plan

### Phase 1: Quick Wins (Day 1)
- [x] Remove duplicate API service
- [x] Remove simple/test components
- [x] Clean up file structure

### Phase 2: State Unification (Day 2-3)
- [ ] Convert custom stores to Redux
- [ ] Integrate WebSocket with Redux
- [ ] Remove state management redundancies

### Phase 3: MVP Focus (Day 4-5)
- [ ] Remove all builder components
- [ ] Convert to viewer-only mode
- [ ] Simplify navigation

### Phase 4: System Consolidation (Day 6-7)
- [ ] Unify notification system
- [ ] Consolidate WebSocket implementation
- [ ] Update all component imports

## Risk Mitigation

1. **Backup Strategy**: All changes backed up before modification
2. **Incremental Approach**: Phase-by-phase implementation
3. **Testing**: Comprehensive testing after each phase
4. **Rollback Plan**: Easy restoration from backups

## Success Metrics

- [ ] All MVP features functional
- [ ] No regression in existing features
- [ ] Build size reduced by 25%+
- [ ] All tests passing
- [ ] Clean dependency tree
- [ ] No console errors/warnings

## Next Steps

1. **Execute streamlining script**: `./scripts/streamline-dashboard.sh`
2. **Run validation**: `npm test && npm run build`
3. **Update router**: Remove builder routes
4. **Clean dependencies**: Remove unused packages
5. **Documentation**: Update component documentation

## Conclusion

This streamlining effort directly supports the PRD's emphasis on simplicity and monitoring-first approach. By removing ~3,000 lines of non-essential code, we create a cleaner, more maintainable codebase that delivers exactly what the MVP requires without unnecessary complexity.