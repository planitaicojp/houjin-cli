package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/planitaicojp/houjin-cli/internal/config"
	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
)

var (
	version = "dev"

	flagFormat  string
	flagVerbose bool
	flagConfig  string
)

var rootCmd = &cobra.Command{
	Use:           "houjin",
	Short:         "法人番号システム Web-API CLI",
	Long:          "国税庁法人番号公表サイトのWeb-APIを操作するCLIツール",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "", "出力形式: json, table, csv (デフォルト: json)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "詳細ログ出力")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "設定ファイルパス")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cerrors.GetExitCode(err))
	}
}

// getFormat returns the output format from flag > env > config > default.
func getFormat() string {
	if flagFormat != "" {
		return flagFormat
	}
	if f := config.EnvOr(config.EnvFormat, ""); f != "" {
		return f
	}
	cfg, err := loadConfig()
	if err != nil {
		return config.DefaultFormat
	}
	return cfg.Format
}

// loadConfig loads the config file, respecting --config flag.
func loadConfig() (*config.Config, error) {
	if flagConfig != "" {
		os.Setenv(config.EnvConfigDir, flagConfig)
	}
	return config.Load()
}

// getAppID returns the app ID or exits with config error.
func getAppID() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	appID := config.GetAppID(cfg)
	if appID == "" {
		return "", &cerrors.ConfigError{
			Message: "アプリケーションIDが設定されていません。\n" +
				"  環境変数: export HOUJIN_APP_ID=your-id\n" +
				"  設定ファイル: ~/.config/houjin/config.yaml に app_id を設定",
		}
	}
	return appID, nil
}
