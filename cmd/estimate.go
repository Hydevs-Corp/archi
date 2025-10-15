package cmd

import (
	"github.com/spf13/cobra"

	"archi/internal/app"
)

var estimateCmd = &cobra.Command{
	Use:     "estimate [directory]",
	Short:   "Estimate files, folders, and processing time",
	Long:    `Estimate files and folders in the target directory and how long a full analysis would take. Useful for large projects to understand scope before running a full analysis.`,
	Aliases: []string{"count"}, // backward compatible alias
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigFromGlobal()
		if err != nil {
			return err
		}

		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		application := app.New(cfg)
		return application.PerformCountAnalysis(targetDir)
	},
}

func init() {
	rootCmd.AddCommand(estimateCmd)
}
