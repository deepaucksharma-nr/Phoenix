package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorReset  = "\033[0m"
)

type ImportViolation struct {
	File       string
	ImportPath string
	Violation  string
}

func main() {
	projectRoot := getProjectRoot()
	violations := []ImportViolation{}
	warnings := []string{}

	fmt.Println("🔍 Validating Go imports...")

	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor and .git directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "/.git/") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			fmt.Printf("%sWarning: Could not parse %s: %v%s\n", colorYellow, path, err, colorReset)
			warnings = append(warnings, fmt.Sprintf("Parse error in %s", path))
			return nil
		}

		// Check imports
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			
			// Check for violations
			if violation := checkImportViolation(path, importPath, projectRoot); violation != "" {
				violations = append(violations, ImportViolation{
					File:       path,
					ImportPath: importPath,
					Violation:  violation,
				})
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("%sError walking directory: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Report results
	fmt.Println("\n=== Import Validation Results ===")
	
	if len(violations) == 0 {
		fmt.Printf("%s✅ No import violations found%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("%s❌ Found %d import violations:%s\n", colorRed, len(violations), colorReset)
		for _, v := range violations {
			fmt.Printf("\n%sFile: %s%s\n", colorYellow, v.File, colorReset)
			fmt.Printf("  Import: %s\n", v.ImportPath)
			fmt.Printf("  %s❌ %s%s\n", colorRed, v.Violation, colorReset)
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("\n%s⚠️  Warnings:%s\n", colorYellow, colorReset)
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if len(violations) > 0 {
		os.Exit(1)
	}
}

func checkImportViolation(filePath, importPath, projectRoot string) string {
	// Get relative path for better context
	relPath, _ := filepath.Rel(projectRoot, filePath)
	
	// Rule 1: No cross-service internal imports
	if strings.Contains(importPath, "/cmd/") && strings.Contains(importPath, "/internal") {
		// Check if it's importing from a different service
		fileService := getServiceFromPath(relPath)
		importService := getServiceFromImport(importPath)
		
		if fileService != "" && importService != "" && fileService != importService {
			return fmt.Sprintf("Cross-service internal import: service '%s' cannot import internal packages from service '%s'", fileService, importService)
		}
	}

	// Rule 2: No relative imports beyond module boundary
	if strings.HasPrefix(importPath, "../") && strings.Count(importPath, "..") > 2 {
		return "Relative import goes too far up the directory tree"
	}

	// Rule 3: Internal packages can only be imported from same module
	if strings.Contains(importPath, "/internal/") {
		fileModule := getModuleFromPath(relPath)
		importModule := getModuleFromImport(importPath)
		
		if fileModule != importModule {
			return fmt.Sprintf("Internal package import across module boundary: '%s' cannot import from '%s'", fileModule, importModule)
		}
	}

	// Rule 4: No imports from phoenix-platform root in service code
	if strings.Contains(relPath, "/cmd/") && strings.HasPrefix(importPath, "phoenix-platform") && !strings.Contains(importPath, "/pkg/") {
		if !strings.Contains(importPath, "/cmd/"+getServiceFromPath(relPath)) {
			return "Services should only import from pkg/ or their own packages"
		}
	}

	// Rule 5: Operators should not import from cmd/
	if strings.Contains(relPath, "/operators/") && strings.Contains(importPath, "/cmd/") {
		return "Operators cannot import from cmd/ services"
	}

	return ""
}

func getServiceFromPath(path string) string {
	if strings.Contains(path, "/cmd/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "cmd" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}

func getServiceFromImport(importPath string) string {
	if strings.Contains(importPath, "/cmd/") {
		parts := strings.Split(importPath, "/")
		for i, part := range parts {
			if part == "cmd" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}

func getModuleFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		if parts[0] == "cmd" || parts[0] == "operators" {
			if len(parts) > 1 {
				return parts[0] + "/" + parts[1]
			}
		}
		return parts[0]
	}
	return ""
}

func getModuleFromImport(importPath string) string {
	if strings.Contains(importPath, "phoenix-platform/") {
		afterPrefix := strings.Split(importPath, "phoenix-platform/")[1]
		return getModuleFromPath(afterPrefix)
	}
	return ""
}

func getProjectRoot() string {
	// Try to find project root by looking for go.mod
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("%sError getting current directory: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Default to current directory
	return cwd
}