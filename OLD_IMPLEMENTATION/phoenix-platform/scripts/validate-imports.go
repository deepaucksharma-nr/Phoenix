package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Rule definitions for import validation
var importRules = []struct {
	name        string
	check       func(pkg, importPath string) bool
	message     string
}{
	{
		name: "No cmd imports",
		check: func(pkg, importPath string) bool {
			return !strings.Contains(pkg, "_test.go") && 
				   strings.Contains(importPath, "phoenix/platform/cmd/")
		},
		message: "packages should not import from cmd directories",
	},
	{
		name: "Internal package isolation",
		check: func(pkg, importPath string) bool {
			if !strings.Contains(pkg, "/internal/") {
				return false
			}
			
			// Extract service name from package path
			parts := strings.Split(pkg, "/")
			var serviceName string
			for i, part := range parts {
				if part == "cmd" && i+1 < len(parts) {
					serviceName = parts[i+1]
					break
				}
			}
			
			// Check if importing another service's internal package
			if serviceName != "" && strings.Contains(importPath, "/internal/") {
				return !strings.Contains(importPath, fmt.Sprintf("cmd/%s/internal", serviceName))
			}
			
			return false
		},
		message: "internal packages should not be imported across service boundaries",
	},
	{
		name: "Operator isolation",
		check: func(pkg, importPath string) bool {
			// Operators should only import from pkg, not from other operators
			if strings.Contains(pkg, "/operators/") && strings.Contains(importPath, "/operators/") {
				pkgOperator := extractOperatorName(pkg)
				importOperator := extractOperatorName(importPath)
				return pkgOperator != importOperator
			}
			return false
		},
		message: "operators should not import from other operators",
	},
}

func extractOperatorName(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "operators" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

type ImportViolation struct {
	File       string
	ImportPath string
	Rule       string
	Message    string
}

func main() {
	violations := []ImportViolation{}
	
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip vendor, .git, and other non-Go directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git" || 
			info.Name() == "node_modules" || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}
		
		// Process only Go files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// Skip generated files
		if strings.Contains(path, "generated") || strings.Contains(path, ".pb.go") {
			return nil
		}
		
		violations = append(violations, checkFile(path)...)
		
		return nil
	})
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}
	
	// Report results
	if len(violations) == 0 {
		fmt.Println("✅ No import violations found")
		os.Exit(0)
	}
	
	fmt.Printf("❌ Found %d import violations:\n\n", len(violations))
	
	// Group by rule
	ruleViolations := make(map[string][]ImportViolation)
	for _, v := range violations {
		ruleViolations[v.Rule] = append(ruleViolations[v.Rule], v)
	}
	
	for rule, vList := range ruleViolations {
		fmt.Printf("Rule: %s\n", rule)
		for _, v := range vList {
			fmt.Printf("  - %s imports %s\n    %s\n", v.File, v.ImportPath, v.Message)
		}
		fmt.Println()
	}
	
	os.Exit(1)
}

func checkFile(filename string) []ImportViolation {
	violations := []ImportViolation{}
	
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		return violations
	}
	
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		
		// Check against rules
		for _, rule := range importRules {
			if rule.check(filename, importPath) {
				violations = append(violations, ImportViolation{
					File:       filename,
					ImportPath: importPath,
					Rule:       rule.name,
					Message:    rule.message,
				})
			}
		}
	}
	
	return violations
}