package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate shell completion script for Phoenix CLI.

To load completions:

Bash:
  $ source <(phoenix completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ phoenix completion bash > /etc/bash_completion.d/phoenix
  
  # macOS:
  $ phoenix completion bash > /usr/local/etc/bash_completion.d/phoenix

Zsh:
  $ source <(phoenix completion zsh)

  # To load completions for each session, execute once:
  $ phoenix completion zsh > "${fpath[1]}/_phoenix"

  # You may need to start a new shell for this setup to take effect.

Fish:
  $ phoenix completion fish | source

  # To load completions for each session, execute once:
  $ phoenix completion fish > ~/.config/fish/completions/phoenix.fish

PowerShell:
  PS> phoenix completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> phoenix completion powershell > phoenix.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}