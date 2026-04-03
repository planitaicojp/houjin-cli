package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	diffFrom string
	diffTo   string
)

func init() {
	diffCmd.Flags().StringVar(&diffFrom, "from", "", "開始日 (YYYY-MM-DD) (必須)")
	diffCmd.Flags().StringVar(&diffTo, "to", "", "終了日 (YYYY-MM-DD) (必須)")
	diffCmd.MarkFlagRequired("from")
	diffCmd.MarkFlagRequired("to")
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "指定期間内の更新法人一覧を取得",
	Long:  "指定した期間内に更新された法人の一覧を取得します。",
	RunE: func(cmd *cobra.Command, args []string) error {
		appID, err := getAppID()
		if err != nil {
			return err
		}

		client := api.NewClient(appID, api.WithVerbose(flagVerbose))
		resp, err := client.GetDiff(diffFrom, diffTo)
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
