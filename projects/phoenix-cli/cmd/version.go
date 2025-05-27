package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// These variables are set during build time
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information of Phoenix CLI including version number, git commit, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func printVersion() {
	switch outputFormat {
	case "json":
		fmt.Printf(`{"version":"%s","gitCommit":"%s","buildDate":"%s","goVersion":"%s","platform":"%s/%s"}%s`,
			Version, GitCommit, BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH, "\n")
	case "yaml":
		fmt.Printf("version: %s\ngitCommit: %s\nbuildDate: %s\ngoVersion: %s\nplatform: %s/%s\n",
			Version, GitCommit, BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	default:
		fmt.Printf("Phoenix CLI\n")
		fmt.Printf("  Version:    %s\n", Version)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	}
}
