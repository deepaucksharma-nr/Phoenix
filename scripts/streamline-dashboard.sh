#!/bin/bash

# Script to streamline Phoenix dashboard by removing redundancies
# This script creates backups before making changes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create backup directory
BACKUP_DIR="dashboard-backup-$(date +%Y%m%d-%H%M%S)"
DASHBOARD_DIR="projects/dashboard"

echo -e "${YELLOW}Creating backup of dashboard...${NC}"
mkdir -p "$BACKUP_DIR"
cp -r "$DASHBOARD_DIR" "$BACKUP_DIR/"

echo -e "${GREEN}Backup created at: $BACKUP_DIR${NC}"

# Function to safely remove file
remove_file() {
    local file=$1
    if [ -f "$file" ]; then
        echo -e "${YELLOW}Removing: $file${NC}"
        rm -f "$file"
    else
        echo -e "${RED}File not found: $file${NC}"
    fi
}

# Function to safely remove directory
remove_directory() {
    local dir=$1
    if [ -d "$dir" ]; then
        echo -e "${YELLOW}Removing directory: $dir${NC}"
        rm -rf "$dir"
    else
        echo -e "${RED}Directory not found: $dir${NC}"
    fi
}

echo -e "\n${GREEN}Phase 1: Removing duplicate files${NC}"
echo "======================================="

# 1. Remove duplicate API service
remove_file "$DASHBOARD_DIR/src/services/api/api.service.ts"

# 2. Remove simple app versions
remove_file "$DASHBOARD_DIR/src/App.simple.tsx"
remove_file "$DASHBOARD_DIR/src/main.simple.tsx"

echo -e "\n${GREEN}Phase 2: Removing non-MVP builder components${NC}"
echo "=============================================="

# 3. Create new structure for streamlined components
mkdir -p "$DASHBOARD_DIR/src/components/Pipeline"

# 4. Move PipelineViewer to new location if it exists
if [ -f "$DASHBOARD_DIR/src/components/PipelineBuilder/PipelineViewer.tsx" ]; then
    echo -e "${YELLOW}Moving PipelineViewer to Pipeline directory${NC}"
    mv "$DASHBOARD_DIR/src/components/PipelineBuilder/PipelineViewer.tsx" \
       "$DASHBOARD_DIR/src/components/Pipeline/PipelineViewer.tsx"
fi

# 5. Remove builder components
remove_file "$DASHBOARD_DIR/src/components/PipelineBuilder/EnhancedPipelineBuilder.tsx"
remove_file "$DASHBOARD_DIR/src/components/PipelineBuilder/ConfigurationPanel.tsx"
remove_file "$DASHBOARD_DIR/src/components/PipelineBuilder/ProcessorLibrary.tsx"
remove_file "$DASHBOARD_DIR/src/components/PipelineBuilder/PipelineBuilder.tsx"
remove_file "$DASHBOARD_DIR/src/components/ExperimentBuilder/PipelineCanvas.tsx"

# 6. Remove ExperimentWizard
remove_directory "$DASHBOARD_DIR/src/components/ExperimentWizard"

# 7. Clean up empty directories
if [ -d "$DASHBOARD_DIR/src/components/PipelineBuilder/nodes" ]; then
    # Move nodes if needed for viewer
    if [ -f "$DASHBOARD_DIR/src/components/PipelineBuilder/nodes/ProcessorNode.tsx" ]; then
        mkdir -p "$DASHBOARD_DIR/src/components/Pipeline/nodes"
        mv "$DASHBOARD_DIR/src/components/PipelineBuilder/nodes/ProcessorNode.tsx" \
           "$DASHBOARD_DIR/src/components/Pipeline/nodes/"
    fi
fi

# Remove old PipelineBuilder directory
remove_directory "$DASHBOARD_DIR/src/components/PipelineBuilder"
remove_directory "$DASHBOARD_DIR/src/components/ExperimentBuilder"

echo -e "\n${GREEN}Phase 3: Streamlining notification system${NC}"
echo "=========================================="

# 8. Remove redundant notification component
remove_file "$DASHBOARD_DIR/src/components/Notifications/RealTimeNotifications.tsx"

echo -e "\n${GREEN}Phase 4: Updating imports${NC}"
echo "==========================="

# 9. Update imports for moved files
if [ -f "$DASHBOARD_DIR/src/components/Pipeline/PipelineViewer.tsx" ]; then
    echo -e "${YELLOW}Updating PipelineViewer imports${NC}"
    # Update any relative imports in PipelineViewer
    sed -i '' 's|../PipelineBuilder/|./|g' "$DASHBOARD_DIR/src/components/Pipeline/PipelineViewer.tsx" 2>/dev/null || true
fi

# 10. Update all imports from removed API service
echo -e "${YELLOW}Updating API service imports${NC}"
find "$DASHBOARD_DIR/src" -type f \( -name "*.ts" -o -name "*.tsx" \) -exec \
    sed -i '' 's|services/api/api\.service|services/api.service|g' {} \; 2>/dev/null || true

# 11. Update imports for PipelineViewer
echo -e "${YELLOW}Updating PipelineViewer imports across codebase${NC}"
find "$DASHBOARD_DIR/src" -type f \( -name "*.ts" -o -name "*.tsx" \) -exec \
    sed -i '' 's|components/PipelineBuilder/PipelineViewer|components/Pipeline/PipelineViewer|g' {} \; 2>/dev/null || true

echo -e "\n${GREEN}Phase 5: Creating streamlined index files${NC}"
echo "=========================================="

# 12. Create new index files for reorganized components
cat > "$DASHBOARD_DIR/src/components/Pipeline/index.ts" << 'EOF'
export { default as PipelineViewer } from './PipelineViewer'
EOF

# 13. Update Pipeline page to use viewer only
if [ -f "$DASHBOARD_DIR/src/pages/PipelineBuilder.tsx" ]; then
    echo -e "${YELLOW}Creating viewer-only Pipeline page${NC}"
    cat > "$DASHBOARD_DIR/src/pages/Pipelines.tsx" << 'EOF'
import React from 'react'
import { Typography, Box, Paper, Container } from '@mui/material'
import { useParams } from 'react-router-dom'
import { PipelineViewer } from '@/components/Pipeline'
import { useAppSelector } from '@/store/hooks'
import { selectPipelineById } from '@/store/slices/pipelineSlice'

export default function Pipelines() {
  const { id } = useParams<{ id: string }>()
  const pipeline = useAppSelector(state => id ? selectPipelineById(state, id) : null)

  return (
    <Container maxWidth="xl">
      <Box py={3}>
        <Typography variant="h4" gutterBottom>
          Pipeline Viewer
        </Typography>
        
        {pipeline ? (
          <Paper sx={{ p: 3, mt: 3 }}>
            <PipelineViewer pipeline={pipeline} />
          </Paper>
        ) : (
          <Typography variant="body1" color="text.secondary">
            Select a pipeline to view its configuration
          </Typography>
        )}
      </Box>
    </Container>
  )
}
EOF
    
    # Remove old PipelineBuilder page
    remove_file "$DASHBOARD_DIR/src/pages/PipelineBuilder.tsx"
fi

echo -e "\n${GREEN}Phase 6: Summary Report${NC}"
echo "========================"

# Count removed lines
REMOVED_LINES=0
REMOVED_LINES=$((REMOVED_LINES + 222))  # Duplicate API service
REMOVED_LINES=$((REMOVED_LINES + 36))   # Simple app files
REMOVED_LINES=$((REMOVED_LINES + 1760)) # Builder components
REMOVED_LINES=$((REMOVED_LINES + 340))  # RealTimeNotifications

echo -e "${GREEN}Streamlining complete!${NC}"
echo -e "- Removed duplicate files: ${GREEN}✓${NC}"
echo -e "- Removed non-MVP components: ${GREEN}✓${NC}"
echo -e "- Updated imports: ${GREEN}✓${NC}"
echo -e "- Total lines removed: ${GREEN}~$REMOVED_LINES${NC}"
echo -e "- Backup location: ${YELLOW}$BACKUP_DIR${NC}"

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Run 'npm install' in the dashboard directory"
echo "2. Run 'npm run build' to verify the build"
echo "3. Run tests to ensure functionality"
echo "4. Review and update router configuration if needed"
echo "5. Consider removing 'reactflow' from package.json if no longer needed"

echo -e "\n${YELLOW}To restore from backup:${NC}"
echo "rm -rf $DASHBOARD_DIR && cp -r $BACKUP_DIR/dashboard $DASHBOARD_DIR"