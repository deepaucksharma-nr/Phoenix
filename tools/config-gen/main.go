package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment   string
	ValuesFile    string
	TemplateFile  string
	OutputFile    string
}

func main() {
	var cfg Config
	
	flag.StringVar(&cfg.Environment, "env", "development", "Environment name")
	flag.StringVar(&cfg.ValuesFile, "values", "", "Values file path")
	flag.StringVar(&cfg.TemplateFile, "template", "", "Template file path")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output file path")
	flag.Parse()
	
	if cfg.TemplateFile == "" {
		fmt.Fprintf(os.Stderr, "Error: template file is required\n")
		os.Exit(1)
	}
	
	// If values file not specified, try to find it based on environment
	if cfg.ValuesFile == "" {
		cfg.ValuesFile = fmt.Sprintf("configs/monitoring/prometheus/environments/%s.yaml", cfg.Environment)
	}
	
	// If output file not specified, generate based on template name
	if cfg.OutputFile == "" {
		base := filepath.Base(cfg.TemplateFile)
		cfg.OutputFile = base[:len(base)-5] // Remove .tmpl extension
	}
	
	if err := generateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Generated config: %s\n", cfg.OutputFile)
}

func generateConfig(cfg Config) error {
	// Read values file
	valuesData, err := ioutil.ReadFile(cfg.ValuesFile)
	if err != nil {
		return fmt.Errorf("reading values file: %w", err)
	}
	
	// Parse values
	var values map[string]interface{}
	if err := yaml.Unmarshal(valuesData, &values); err != nil {
		return fmt.Errorf("parsing values: %w", err)
	}
	
	// Add generated timestamp
	values["GeneratedAt"] = time.Now().Format(time.RFC3339)
	
	// Read template file
	templateData, err := ioutil.ReadFile(cfg.TemplateFile)
	if err != nil {
		return fmt.Errorf("reading template file: %w", err)
	}
	
	// Parse template
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}).Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}
	
	// Create output directory if needed
	outputDir := filepath.Dir(cfg.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	
	// Create output file
	output, err := os.Create(cfg.OutputFile)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer output.Close()
	
	// Execute template
	if err := tmpl.Execute(output, values); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	
	return nil
}