package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン情報を表示",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("houjin version %s\n", version)
	},
}
