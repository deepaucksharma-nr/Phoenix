# Phoenix Platform Development Tips

## 🚀 Quick Development Tips

### Building the Phoenix CLI

```bash
# Navigate to Phoenix CLI directory
cd projects/phoenix-cli

# Build the CLI
go build -o bin/phoenix .

# Run directly
./bin/phoenix --help

# Or install to PATH
go install .
```

### Common Development Tasks

#### 1. Adding New Commands
When adding new commands to Phoenix CLI:
```go
// 1. Create new command file in cmd/
// cmd/mycommand.go

package cmd

import (
    "github.com/phoenix/platform/projects/phoenix-cli/internal/client"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Short description",
    RunE:  runMyCommand,
}

func init() {
    // Add to parent command
    rootCmd.AddCommand(myCmd)
    
    // Add flags
    myCmd.Flags().String("flag", "default", "description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

#### 2. Working with API Client
```go
// Get authenticated API client
apiClient, err := getAPIClient()
if err != nil {
    return err
}

// Make API calls
experiments, err := apiClient.ListExperiments(client.ListExperimentsRequest{
    Status: "running",
    PageSize: 10,
})
```

#### 3. Output Formatting
The CLI supports multiple output formats:
```go
// Use the output package for consistent formatting
import "github.com/phoenix/platform/projects/phoenix-cli/internal/output"

// Print based on user preference
outputFormat := viper.GetString("output")
switch outputFormat {
case "json":
    return output.PrintJSON(data)
case "yaml":
    return output.PrintYAML(data)
default:
    // Custom formatting
    fmt.Printf("Result: %v\n", data)
}
```

### Testing

#### Unit Tests
```bash
# Run all tests
cd projects/phoenix-cli
go test ./...

# Run specific test
go test ./cmd -run TestExperimentCreate

# With coverage
go test -cover ./...
```

#### Integration Tests
```bash
# Build CLI first
go build -o bin/phoenix .

# Run integration tests
cd ../../tests/integration
go test -v ./...
```

### Debugging

#### Enable Debug Output
```bash
# Set debug flag
./bin/phoenix --debug experiment list

# Or via environment variable
PHOENIX_DEBUG=true ./bin/phoenix experiment list
```

#### Using Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the CLI
dlv debug . -- experiment list
```

### Code Style Guidelines

1. **Error Handling**
   ```go
   // Always wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to create experiment: %w", err)
   }
   ```

2. **Command Structure**
   - Keep command logic in `cmd/` directory
   - Business logic in `internal/` packages
   - Reusable components in `internal/output`, `internal/client`, etc.

3. **User Experience**
   - Provide clear error messages
   - Show progress for long operations
   - Confirm destructive operations

### Common Patterns

#### Progress Indicators
```go
fmt.Println("Creating experiment...")
// Long operation
fmt.Println("✓ Experiment created successfully!")
```

#### Confirmation Prompts
```go
if !force {
    confirmed, err := output.Confirm("Are you sure you want to delete?")
    if err != nil || !confirmed {
        return fmt.Errorf("operation cancelled")
    }
}
```

#### Table Output
```go
w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
fmt.Fprintln(w, "NAME\tSTATUS\tCREATED")
for _, exp := range experiments {
    fmt.Fprintf(w, "%s\t%s\t%s\n", exp.Name, exp.Status, exp.CreatedAt)
}
w.Flush()
```

### Environment Variables

The CLI respects these environment variables:
- `PHOENIX_API_ENDPOINT` - API server URL
- `PHOENIX_AUTH_TOKEN` - Authentication token
- `PHOENIX_OUTPUT_FORMAT` - Default output format (json/yaml/table)
- `PHOENIX_DEBUG` - Enable debug output

### Configuration

The CLI uses Viper for configuration:
```go
// Read config value
endpoint := viper.GetString("api.endpoint")

// Set config value
viper.Set("api.endpoint", "http://localhost:8080")

// Save config
viper.WriteConfig()
```

### Performance Tips

1. **Concurrent Operations**
   ```go
   var wg sync.WaitGroup
   for _, item := range items {
       wg.Add(1)
       go func(i Item) {
           defer wg.Done()
           // Process item
       }(item)
   }
   wg.Wait()
   ```

2. **Reuse HTTP Clients**
   - The API client should be created once and reused
   - Don't create new clients for each request

3. **Batch Operations**
   - When possible, use batch APIs instead of individual calls
   - Implement pagination for large result sets

### Troubleshooting

#### Module Issues
```bash
# If you see module errors
go mod tidy
go work sync
```

#### Build Issues
```bash
# Clean build cache
go clean -cache
go build -o bin/phoenix .
```

#### Import Errors
Ensure imports use the correct path:
```go
// Correct
import "github.com/phoenix/platform/projects/phoenix-cli/internal/client"

// Incorrect (old)
import "github.com/phoenix-vnext/platform/cmd/phoenix-cli/internal/client"
```

### Release Process

1. Update version in `cmd/version.go`
2. Run tests: `go test ./...`
3. Build release binaries:
   ```bash
   # Linux
   GOOS=linux GOARCH=amd64 go build -o phoenix-linux-amd64
   
   # macOS
   GOOS=darwin GOARCH=amd64 go build -o phoenix-darwin-amd64
   GOOS=darwin GOARCH=arm64 go build -o phoenix-darwin-arm64
   
   # Windows
   GOOS=windows GOARCH=amd64 go build -o phoenix-windows-amd64.exe
   ```

### Contributing

1. Create feature branch
2. Make changes following patterns above
3. Add tests for new functionality
4. Update documentation if needed
5. Submit PR with clear description

---

Happy coding! 🚀 The Phoenix CLI is now ready for continued development.