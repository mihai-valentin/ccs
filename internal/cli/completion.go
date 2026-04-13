package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <bash|zsh|fish|powershell>",
		Short: "Generate shell completions",
		Long: `Generate shell completion scripts.

To load completions:

Bash:
  $ source <(ccs completion bash)
  # Or add to ~/.bashrc:
  $ ccs completion bash >> ~/.bashrc

Zsh:
  $ source <(ccs completion zsh)
  # Or add to fpath:
  $ ccs completion zsh > "${fpath[1]}/_ccs"

Fish:
  $ ccs completion fish | source
  # Or persist:
  $ ccs completion fish > ~/.config/fish/completions/ccs.fish

PowerShell:
  PS> ccs completion powershell | Out-String | Invoke-Expression
  # Or add to $PROFILE:
  PS> ccs completion powershell >> $PROFILE
`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
