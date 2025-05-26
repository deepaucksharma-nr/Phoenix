package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// Plugin represents a CLI plugin
type Plugin struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Commands    []string `json:"commands"`
	Executable  string   `json:"executable"`
	Path        string   `json:"-"`
}

// PluginManager manages CLI plugins
type PluginManager struct {
	pluginDir string
	plugins   map[string]*Plugin
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir string) *PluginManager {
	return &PluginManager{
		pluginDir: pluginDir,
		plugins:   make(map[string]*Plugin),
	}
}

// LoadPlugins discovers and loads all plugins
func (pm *PluginManager) LoadPlugins() error {
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		// Plugin directory doesn't exist, create it
		if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory: %w", err)
		}
		return nil
	}

	entries, err := os.ReadDir(pm.pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(pm.pluginDir, entry.Name())
		manifestPath := filepath.Join(pluginPath, "plugin.json")

		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		plugin, err := pm.loadPlugin(manifestPath)
		if err != nil {
			fmt.Printf("Warning: failed to load plugin %s: %v\n", entry.Name(), err)
			continue
		}

		plugin.Path = pluginPath
		pm.plugins[plugin.Name] = plugin
	}

	return nil
}

// loadPlugin loads a single plugin from its manifest
func (pm *PluginManager) loadPlugin(manifestPath string) (*Plugin, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	var plugin Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to parse plugin manifest: %w", err)
	}

	// Validate required fields
	if plugin.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	if plugin.Executable == "" {
		return nil, fmt.Errorf("plugin executable is required")
	}

	// Check if executable exists
	execPath := filepath.Join(filepath.Dir(manifestPath), plugin.Executable)
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin executable not found: %s", execPath)
	}

	return &plugin, nil
}

// GetPlugins returns all loaded plugins
func (pm *PluginManager) GetPlugins() map[string]*Plugin {
	return pm.plugins
}

// GetPlugin returns a specific plugin by name
func (pm *PluginManager) GetPlugin(name string) (*Plugin, bool) {
	plugin, ok := pm.plugins[name]
	return plugin, ok
}

// ExecutePlugin executes a plugin command
func (pm *PluginManager) ExecutePlugin(pluginName string, args []string) error {
	plugin, ok := pm.plugins[pluginName]
	if !ok {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	execPath := filepath.Join(plugin.Path, plugin.Executable)
	cmd := exec.Command(execPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PHOENIX_PLUGIN_NAME=%s", plugin.Name),
		fmt.Sprintf("PHOENIX_PLUGIN_VERSION=%s", plugin.Version),
	)

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}

// CreatePluginCommands creates cobra commands for all loaded plugins
func (pm *PluginManager) CreatePluginCommands() []*cobra.Command {
	var commands []*cobra.Command

	for _, plugin := range pm.plugins {
		cmd := &cobra.Command{
			Use:   plugin.Name,
			Short: plugin.Description,
			Long:  fmt.Sprintf("%s (v%s by %s)", plugin.Description, plugin.Version, plugin.Author),
			RunE: func(cmd *cobra.Command, args []string) error {
				return pm.ExecutePlugin(plugin.Name, args)
			},
			DisableFlagParsing: true,
		}

		commands = append(commands, cmd)
	}

	return commands
}

// InstallPlugin installs a plugin from a given path or URL
func (pm *PluginManager) InstallPlugin(source string) error {
	// For now, support installing from local directory
	// Future: support downloading from URLs, archives, etc.
	
	if !filepath.IsAbs(source) {
		abs, err := filepath.Abs(source)
		if err != nil {
			return fmt.Errorf("failed to resolve source path: %w", err)
		}
		source = abs
	}

	manifestPath := filepath.Join(source, "plugin.json")
	plugin, err := pm.loadPlugin(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to validate plugin: %w", err)
	}

	// Check if plugin already exists
	if _, exists := pm.plugins[plugin.Name]; exists {
		return fmt.Errorf("plugin %s is already installed", plugin.Name)
	}

	// Copy plugin to plugin directory
	destPath := filepath.Join(pm.pluginDir, plugin.Name)
	if err := copyDir(source, destPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	// Load the plugin
	plugin.Path = destPath
	pm.plugins[plugin.Name] = plugin

	fmt.Printf("Plugin %s v%s installed successfully\n", plugin.Name, plugin.Version)
	return nil
}

// UninstallPlugin removes a plugin
func (pm *PluginManager) UninstallPlugin(name string) error {
	plugin, ok := pm.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s is not installed", name)
	}

	// Remove plugin directory
	if err := os.RemoveAll(plugin.Path); err != nil {
		return fmt.Errorf("failed to remove plugin directory: %w", err)
	}

	// Remove from loaded plugins
	delete(pm.plugins, name)

	fmt.Printf("Plugin %s uninstalled successfully\n", name)
	return nil
}

// ListPlugins returns a list of installed plugins
func (pm *PluginManager) ListPlugins() []*Plugin {
	var plugins []*Plugin
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(in)
	if err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, si.Mode())
}

// ValidatePluginName checks if a plugin name is valid
func ValidatePluginName(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	// Plugin names should not conflict with built-in commands
	builtinCommands := []string{
		"auth", "config", "experiment", "pipeline", "version", "completion", "help",
	}

	for _, builtin := range builtinCommands {
		if strings.EqualFold(name, builtin) {
			return fmt.Errorf("plugin name %s conflicts with built-in command", name)
		}
	}

	// Plugin names should be valid command names (alphanumeric + dashes)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return fmt.Errorf("plugin name contains invalid character: %c", char)
		}
	}

	return nil
}