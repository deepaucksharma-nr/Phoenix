package cmd

import "fmt"

func getReadmeTemplate(name, description, language string) string {
	executableName := getExecutableName(language)
	
	template := `# %s

%s

## Installation

` + "```bash" + `
phoenix plugin install .
` + "```" + `

## Usage

` + "```bash" + `
phoenix %s [arguments]
` + "```" + `

## Development

This plugin is written in %s. To modify:

1. Edit the main executable file
2. Update the plugin.json manifest if needed
3. Reinstall the plugin: ` + "`" + `phoenix plugin install . --force` + "`" + `

## Plugin Structure

- ` + "`" + `plugin.json` + "`" + ` - Plugin manifest with metadata
- ` + "`" + `%s` + "`" + ` - Main executable
- ` + "`" + `README.md` + "`" + ` - This file

## Environment Variables

When executed, the plugin has access to:
- ` + "`" + `PHOENIX_PLUGIN_NAME` + "`" + ` - The plugin name
- ` + "`" + `PHOENIX_PLUGIN_VERSION` + "`" + ` - The plugin version

## API Integration

To interact with the Phoenix API from your plugin, you can:

1. Use the phoenix CLI commands as subprocesses
2. Make direct HTTP requests to the API
3. Use the Phoenix API token from the user's configuration

Example API call:
` + "```bash" + `
# Get auth token from phoenix config
API_TOKEN=$(phoenix config get auth_token)
API_URL=$(phoenix config get api_url)

# Make API request
curl -H "Authorization: Bearer $API_TOKEN" \
     "$API_URL/api/v1/experiments"
` + "```" + `
`
	
	return fmt.Sprintf(template, name, description, name, language, executableName)
}