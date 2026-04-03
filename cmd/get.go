package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	getHistory bool
	getClose   bool
)

func init() {
	getCmd.Flags().BoolVar(&getHistory, "history", false, "履歴情報を含める")
	getCmd.Flags().BoolVar(&getClose, "close", false, "閉鎖法人を含める")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <法人番号> [法人番号...]",
	Short: "法人番号を指定して法人情報を取得",
	Long:  "13桁の法人番号を指定して法人情報を取得します。複数番号を空白区切りで指定可能。",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, num := range args {
			if err := model.ValidateCorporateNumber(num); err != nil {
				return &cerrors.ValidationError{
					Field:   "corporate_number",
					Message: fmt.Sprintf("%s: %v", num, err),
				}
			}
		}

		appID, err := getAppID()
		if err != nil {
			return err
		}

		client := api.NewClient(appID, api.WithVerbose(flagVerbose))
		resp, err := client.GetByNumber(args, api.GetOptions{
			History: getHistory,
			Close:   getClose,
		})
		if err != nil {
			return err
		}

		formatter := output.New(getFormat())
		return formatter.Format(os.Stdout, resp)
	},
}
