package cmd

import (
	"github.com/spf13/cobra"

	"archi/internal/app"
)

var countCmd = &cobra.Command{
	Use:   "count [directory]",
	Short: "Count files and folders and estimate processing time",
	Long: `Count files and folders in the target directory and estimate 
how long a full analysis would take. This is useful for large projects 
where you want to understand the scope before running a full analysis.`,
	Args: cobra.MaximumNArgs(1),
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
	rootCmd.AddCommand(countCmd)
}
