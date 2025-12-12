package main

import (
	"fmt"
	"os"

	"github.com/baldator/iac-recert-engine/internal/config"
	"github.com/baldator/iac-recert-engine/internal/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the recertification process",
	Long:  `Scans the repository and creates PRs for files needing recertification.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Config
		var cfg config.Config
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		if err := cfg.Validate(); err != nil {
			fmt.Println("Config validation error:", err)
			os.Exit(1)
		}

		// 2. Setup Logger
		logger, _ := zap.NewProduction()
		if cfg.Global.VerboseLogging {
			logger, _ = zap.NewDevelopment()
		}
		defer logger.Sync()

		// 3. Init Engine
		eng, err := engine.NewEngine(cfg, logger)
		if err != nil {
			logger.Fatal("failed to init engine", zap.Error(err))
		}

		// 4. Run
		if err := eng.Run(cmd.Context()); err != nil {
			logger.Fatal("run failed", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	runCmd.Flags().Bool("dry-run", false, "Run without creating PRs")
	viper.BindPFlag("global.dry_run", runCmd.Flags().Lookup("dry-run"))

	runCmd.Flags().String("repo-url", "", "Repository URL to scan (overrides config)")
	viper.BindPFlag("repository.url", runCmd.Flags().Lookup("repo-url"))
}
