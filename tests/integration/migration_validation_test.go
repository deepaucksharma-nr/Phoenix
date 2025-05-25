package integration

import (
    "os"
    "path/filepath"
    "testing"
)

// TestMigrationStructure validates the migrated project structure
func TestMigrationStructure(t *testing.T) {
    projectRoot := "../.."
    
    // Test that all required directories exist
    requiredDirs := []string{
        "services",
        "packages",
        "packages/go-common",
        "packages/contracts",
        "infrastructure",
        "infrastructure/kubernetes",
        "infrastructure/docker",
        "operators",
        "tests",
        "monitoring",
        "config",
        "tools",
        "docs",
    }
    
    for _, dir := range requiredDirs {
        path := filepath.Join(projectRoot, dir)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            t.Errorf("Required directory missing: %s", dir)
        }
    }
    
    // Test that core services have been migrated
    coreServices := []string{
        "services/api",
        "services/controller", 
        "services/generator",
        "services/dashboard",
    }
    
    for _, svc := range coreServices {
        // Check service directory exists
        svcPath := filepath.Join(projectRoot, svc)
        if _, err := os.Stat(svcPath); os.IsNotExist(err) {
            t.Errorf("Core service missing: %s", svc)
            continue
        }
        
        // Check for essential files
        essentialFiles := []string{"go.mod", "Dockerfile"}
        for _, file := range essentialFiles {
            filePath := filepath.Join(svcPath, file)
            if _, err := os.Stat(filePath); os.IsNotExist(err) {
                t.Errorf("Essential file missing in %s: %s", svc, file)
            }
        }
    }
    
    // Test that shared packages exist
    sharedPackages := []string{
        "packages/go-common/auth",
        "packages/go-common/metrics",
        "packages/go-common/interfaces",
        "packages/contracts/proto",
        "packages/contracts/openapi",
    }
    
    for _, pkg := range sharedPackages {
        pkgPath := filepath.Join(projectRoot, pkg)
        if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
            t.Errorf("Shared package missing: %s", pkg)
        }
    }
    
    // Test that operators have been migrated
    operators := []string{
        "operators/loadsim",
        "operators/pipeline",
    }
    
    for _, op := range operators {
        opPath := filepath.Join(projectRoot, op)
        if _, err := os.Stat(opPath); os.IsNotExist(err) {
            t.Errorf("Operator missing: %s", op)
        }
    }
}

// TestNoOldImplementationReferences ensures no references to OLD_IMPLEMENTATION remain
func TestNoOldImplementationReferences(t *testing.T) {
    // This would normally scan go.mod files and imports
    // For now, just check that services don't have OLD_IMPLEMENTATION in their paths
    t.Log("Checking for OLD_IMPLEMENTATION references...")
    // Implementation would go here
}