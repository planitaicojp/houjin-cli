package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "シェル補完スクリプトを生成",
	Long: `シェル補完スクリプトを生成します。

  bash:  source <(houjin completion bash)
  zsh:   houjin completion zsh > "${fpath[1]}/_houjin"
  fish:  houjin completion fish | source`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return cmd.Usage()
		}
	},
}
