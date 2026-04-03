package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/api"
	"github.com/planitaicojp/houjin-cli/internal/batch"
	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
	"github.com/planitaicojp/houjin-cli/internal/model"
	"github.com/planitaicojp/houjin-cli/internal/output"
)

var (
	getHistory bool
	getClose   bool
	getFile    string
)

func init() {
	getCmd.Flags().BoolVar(&getHistory, "history", false, "履歴情報を含める")
	getCmd.Flags().BoolVar(&getClose, "close", false, "閉鎖法人を含める")
	getCmd.Flags().StringVar(&getFile, "file", "", "法人番号リストファイル (- で標準入力)")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <法人番号> [法人番号...]",
	Short: "法人番号を指定して法人情報を取得",
	Long:  "13桁の法人番号を指定して法人情報を取得します。複数番号を空白区切りで指定可能。",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var numbers []string

		if getFile != "" {
			var r io.Reader
			if getFile == "-" {
				r = os.Stdin
			} else {
				f, err := os.Open(getFile)
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer f.Close()
				r = f
			}
			var err error
			numbers, err = batch.ReadNumbers(r)
			if err != nil {
				return fmt.Errorf("reading numbers: %w", err)
			}
			if len(numbers) == 0 {
				return &cerrors.ValidationError{
					Field:   "file",
					Message: "ファイルに法人番号が含まれていません",
				}
			}
		} else {
			if len(args) == 0 {
				return &cerrors.ValidationError{
					Field:   "args",
					Message: "法人番号を指定するか、--file でファイルを指定してください",
				}
			}
			numbers = args
		}

		for _, num := range numbers {
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
		resp, err := client.GetByNumber(numbers, api.GetOptions{
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
