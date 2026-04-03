package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	searchMode  string
	searchPref  string
	searchCity  string
	searchClose bool
	searchPage  int
	searchAll   bool
	searchType  string
)

func init() {
	searchCmd.Flags().StringVar(&searchMode, "mode", "prefix", "検索モード: prefix(前方一致), partial(部分一致)")
	searchCmd.Flags().StringVar(&searchPref, "pref", "", "都道府県コード (01-47, 99=海外)")
	searchCmd.Flags().StringVar(&searchCity, "city", "", "市区町村コード")
	searchCmd.Flags().BoolVar(&searchClose, "close", false, "閉鎖法人を含める")
	searchCmd.Flags().IntVar(&searchPage, "page", 0, "ページ番号を指定 (分割番号)")
	searchCmd.Flags().BoolVar(&searchAll, "all", false, "全ページを自動取得")
	searchCmd.Flags().StringVar(&searchType, "type", "", "法人種別フィルタ (01:国の機関, 02:地方公共団体, 03:設立登記法人, 04:その他)")
	searchCmd.MarkFlagsMutuallyExclusive("page", "all")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <法人名>",
	Short: "法人名で検索",
	Long:  "法人名を指定して法人情報を検索します。前方一致（デフォルト）または部分一致で検索可能。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appID, err := getAppID()
		if err != nil {
			return err
		}

		client := api.NewClient(appID, api.WithVerbose(flagVerbose))
		opts := api.SearchOptions{
			Mode:   searchMode,
			Pref:   searchPref,
			City:   searchCity,
			Close:  searchClose,
			Divide: searchPage,
			Kind:   searchType,
		}

		var resp *model.Response
		if searchAll {
			resp, err = client.SearchAllPages(args[0], opts)
		} else {
			resp, err = client.SearchByName(args[0], opts)
		}
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
