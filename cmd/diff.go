package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	diffFrom string
	diffTo   string
	diffPage int
	diffAll  bool
)

func init() {
	diffCmd.Flags().StringVar(&diffFrom, "from", "", "開始日 (YYYY-MM-DD) (必須)")
	diffCmd.Flags().StringVar(&diffTo, "to", "", "終了日 (YYYY-MM-DD) (必須)")
	diffCmd.MarkFlagRequired("from")
	diffCmd.MarkFlagRequired("to")
	diffCmd.Flags().IntVar(&diffPage, "page", 0, "ページ番号を指定 (分割番号)")
	diffCmd.Flags().BoolVar(&diffAll, "all", false, "全ページを自動取得")
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
		opts := api.DiffOptions{Divide: diffPage}

		var resp *model.Response
		if diffAll {
			resp, err = client.DiffAllPages(diffFrom, diffTo, opts)
		} else {
			resp, err = client.GetDiff(diffFrom, diffTo, opts)
		}
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
