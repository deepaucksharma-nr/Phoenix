#!/bin/bash

# Script to analyze dashboard redundancies and generate detailed report

set -e

DASHBOARD_DIR="projects/dashboard/src"
REPORT_FILE="DASHBOARD_REDUNDANCY_ANALYSIS.md"

echo "# Dashboard Redundancy Analysis Report" > "$REPORT_FILE"
echo "Generated on: $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Function to count lines in a file
count_lines() {
    local file=$1
    if [ -f "$file" ]; then
        wc -l < "$file" | tr -d ' '
    else
        echo "0"
    fi
}

# 1. Check for duplicate API services
echo "## 1. API Service Duplication" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

if [ -f "$DASHBOARD_DIR/services/api.service.ts" ] && [ -f "$DASHBOARD_DIR/services/api/api.service.ts" ]; then
    echo "### Duplicate API Service Files Found:" >> "$REPORT_FILE"
    echo "- services/api.service.ts: $(count_lines "$DASHBOARD_DIR/services/api.service.ts") lines" >> "$REPORT_FILE"
    echo "- services/api/api.service.ts: $(count_lines "$DASHBOARD_DIR/services/api/api.service.ts") lines" >> "$REPORT_FILE"
    
    # Check if files are identical
    if diff -q "$DASHBOARD_DIR/services/api.service.ts" "$DASHBOARD_DIR/services/api/api.service.ts" > /dev/null; then
        echo "- **Status**: Files are IDENTICAL" >> "$REPORT_FILE"
    else
        echo "- **Status**: Files are DIFFERENT" >> "$REPORT_FILE"
    fi
    echo "" >> "$REPORT_FILE"
fi

# 2. Analyze WebSocket implementations
echo "## 2. WebSocket Implementation Analysis" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

websocket_files=(
    "$DASHBOARD_DIR/hooks/useWebSocket.ts"
    "$DASHBOARD_DIR/services/websocket/WebSocketService.ts"
    "$DASHBOARD_DIR/components/WebSocket/WebSocketProvider.tsx"
)

total_ws_lines=0
for file in "${websocket_files[@]}"; do
    if [ -f "$file" ]; then
        lines=$(count_lines "$file")
        echo "- ${file#$DASHBOARD_DIR/}: $lines lines" >> "$REPORT_FILE"
        total_ws_lines=$((total_ws_lines + lines))
    fi
done
echo "- **Total WebSocket code**: $total_ws_lines lines" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# 3. Analyze Builder vs Viewer components
echo "## 3. Builder vs Viewer Components" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "### Pipeline Builder Components:" >> "$REPORT_FILE"
builder_files=(
    "$DASHBOARD_DIR/components/PipelineBuilder/PipelineBuilder.tsx"
    "$DASHBOARD_DIR/components/PipelineBuilder/EnhancedPipelineBuilder.tsx"
    "$DASHBOARD_DIR/components/PipelineBuilder/ConfigurationPanel.tsx"
    "$DASHBOARD_DIR/components/PipelineBuilder/ProcessorLibrary.tsx"
    "$DASHBOARD_DIR/components/ExperimentBuilder/PipelineCanvas.tsx"
)

total_builder_lines=0
for file in "${builder_files[@]}"; do
    if [ -f "$file" ]; then
        lines=$(count_lines "$file")
        echo "- ${file#$DASHBOARD_DIR/}: $lines lines" >> "$REPORT_FILE"
        total_builder_lines=$((total_builder_lines + lines))
    fi
done
echo "- **Total Builder code**: $total_builder_lines lines" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "### Viewer Components:" >> "$REPORT_FILE"
if [ -f "$DASHBOARD_DIR/components/PipelineBuilder/PipelineViewer.tsx" ]; then
    echo "- PipelineViewer.tsx: $(count_lines "$DASHBOARD_DIR/components/PipelineBuilder/PipelineViewer.tsx") lines" >> "$REPORT_FILE"
fi
echo "" >> "$REPORT_FILE"

# 4. Analyze state management
echo "## 4. State Management Analysis" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "### Redux Implementation:" >> "$REPORT_FILE"
redux_slices=(
    "$DASHBOARD_DIR/store/slices/authSlice.ts"
    "$DASHBOARD_DIR/store/slices/experimentSlice.ts"
    "$DASHBOARD_DIR/store/slices/pipelineSlice.ts"
    "$DASHBOARD_DIR/store/slices/notificationSlice.ts"
    "$DASHBOARD_DIR/store/slices/uiSlice.ts"
)

for slice in "${redux_slices[@]}"; do
    if [ -f "$slice" ]; then
        echo "- ${slice#$DASHBOARD_DIR/store/slices/}: $(count_lines "$slice") lines" >> "$REPORT_FILE"
    fi
done
echo "" >> "$REPORT_FILE"

# Check for custom store usage
echo "### Custom Store Usage (potential redundancy):" >> "$REPORT_FILE"
echo "Searching for useAuthStore, useExperimentStore patterns..." >> "$REPORT_FILE"
grep -r "useAuthStore\|useExperimentStore" "$DASHBOARD_DIR" --include="*.ts" --include="*.tsx" 2>/dev/null | head -10 >> "$REPORT_FILE" || echo "No custom store usage found" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# 5. Analyze notification systems
echo "## 5. Notification System Analysis" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

notification_files=(
    "$DASHBOARD_DIR/components/Notifications/NotificationProvider.tsx"
    "$DASHBOARD_DIR/components/Notifications/RealTimeNotifications.tsx"
    "$DASHBOARD_DIR/store/slices/notificationSlice.ts"
)

total_notif_lines=0
for file in "${notification_files[@]}"; do
    if [ -f "$file" ]; then
        lines=$(count_lines "$file")
        echo "- ${file#$DASHBOARD_DIR/}: $lines lines" >> "$REPORT_FILE"
        total_notif_lines=$((total_notif_lines + lines))
    fi
done
echo "- **Total Notification code**: $total_notif_lines lines" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# 6. Find unused or test components
echo "## 6. Simple/Test Components" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

simple_files=(
    "$DASHBOARD_DIR/App.simple.tsx"
    "$DASHBOARD_DIR/main.simple.tsx"
)

for file in "${simple_files[@]}"; do
    if [ -f "$file" ]; then
        echo "- ${file#$DASHBOARD_DIR/}: $(count_lines "$file") lines" >> "$REPORT_FILE"
    fi
done
echo "" >> "$REPORT_FILE"

# 7. Dependencies analysis
echo "## 7. Package Dependencies" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

if [ -f "projects/dashboard/package.json" ]; then
    echo "### UI/Builder Dependencies (potentially removable for MVP):" >> "$REPORT_FILE"
    grep -E "reactflow|react-dnd|react-beautiful-dnd|react-sortable" projects/dashboard/package.json >> "$REPORT_FILE" || echo "No drag-and-drop dependencies found" >> "$REPORT_FILE"
fi
echo "" >> "$REPORT_FILE"

# 8. Summary
echo "## Summary of Redundancies" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "### Immediate Actions:" >> "$REPORT_FILE"
echo "1. Remove duplicate API service file (223 lines)" >> "$REPORT_FILE"
echo "2. Remove simple app versions (~50 lines)" >> "$REPORT_FILE"
echo "3. Consolidate WebSocket implementations (~$total_ws_lines lines to ~200 lines)" >> "$REPORT_FILE"
echo "4. Remove builder components for MVP (~$total_builder_lines lines)" >> "$REPORT_FILE"
echo "5. Unify notification systems (~$total_notif_lines lines to ~100 lines)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

potential_reduction=$((223 + 50 + total_ws_lines - 200 + total_builder_lines + total_notif_lines - 100))
echo "**Potential code reduction: ~$potential_reduction lines**" >> "$REPORT_FILE"

echo "Dashboard redundancy analysis complete. Report saved to $REPORT_FILE"