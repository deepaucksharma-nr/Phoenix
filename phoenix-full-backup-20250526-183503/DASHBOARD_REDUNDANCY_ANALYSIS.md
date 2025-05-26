# Dashboard Redundancy Analysis Report
Generated on: Mon May 26 18:19:09 IST 2025

## 1. API Service Duplication

### Duplicate API Service Files Found:
- services/api.service.ts: 222 lines
- services/api/api.service.ts: 222 lines
- **Status**: Files are IDENTICAL

## 2. WebSocket Implementation Analysis

- hooks/useWebSocket.ts: 201 lines
- services/websocket/WebSocketService.ts: 202 lines
- components/WebSocket/WebSocketProvider.tsx: 190 lines
- **Total WebSocket code**: 593 lines

## 3. Builder vs Viewer Components

### Pipeline Builder Components:
- components/PipelineBuilder/PipelineBuilder.tsx: 48 lines
- components/PipelineBuilder/EnhancedPipelineBuilder.tsx: 708 lines
- components/PipelineBuilder/ConfigurationPanel.tsx: 449 lines
- components/PipelineBuilder/ProcessorLibrary.tsx: 337 lines
- components/ExperimentBuilder/PipelineCanvas.tsx: 218 lines
- **Total Builder code**: 1760 lines

### Viewer Components:
- PipelineViewer.tsx: 171 lines

## 4. State Management Analysis

### Redux Implementation:
- authSlice.ts: 169 lines
- experimentSlice.ts: 257 lines
- pipelineSlice.ts: 177 lines
- notificationSlice.ts: 118 lines
- uiSlice.ts: 112 lines

### Custom Store Usage (potential redundancy):
Searching for useAuthStore, useExperimentStore patterns...
projects/dashboard/src/components/Auth/RoleGuard.tsx:  const { user } = useAuthStore()
projects/dashboard/src/components/Auth/__tests__/PrivateRoute.test.tsx:import { useAuthStore } from '@/store/useAuthStore'
projects/dashboard/src/components/Auth/__tests__/PrivateRoute.test.tsx:vi.mock('@/store/useAuthStore')
projects/dashboard/src/components/Auth/__tests__/PrivateRoute.test.tsx:    vi.mocked(useAuthStore).mockReturnValue({
projects/dashboard/src/components/Auth/__tests__/PrivateRoute.test.tsx:    vi.mocked(useAuthStore).mockReturnValue({
projects/dashboard/src/components/Auth/__tests__/PrivateRoute.test.tsx:    vi.mocked(useAuthStore).mockReturnValue({
projects/dashboard/src/components/Auth/__tests__/PrivateRoute.test.tsx:    vi.mocked(useAuthStore).mockReturnValue({
projects/dashboard/src/components/Auth/__tests__/AuthComponents.test.tsx:import { useAuthStore } from '../../../store/useAuthStore'
projects/dashboard/src/components/Auth/__tests__/AuthComponents.test.tsx:vi.mock('../../../store/useAuthStore')
projects/dashboard/src/components/Auth/__tests__/AuthComponents.test.tsx:const mockUseAuthStore = useAuthStore as any

## 5. Notification System Analysis

- components/Notifications/NotificationProvider.tsx: 149 lines
- components/Notifications/RealTimeNotifications.tsx: 340 lines
- store/slices/notificationSlice.ts: 118 lines
- **Total Notification code**: 607 lines

## 6. Simple/Test Components

- App.simple.tsx: 28 lines
- main.simple.tsx: 8 lines

## 7. Package Dependencies

### UI/Builder Dependencies (potentially removable for MVP):
    "reactflow": "^11.11.4",

## Summary of Redundancies

### Immediate Actions:
1. Remove duplicate API service file (223 lines)
2. Remove simple app versions (~50 lines)
3. Consolidate WebSocket implementations (~593 lines to ~200 lines)
4. Remove builder components for MVP (~1760 lines)
5. Unify notification systems (~607 lines to ~100 lines)

**Potential code reduction: ~2933 lines**
