package cmd

import (
	"github.com/spf13/cobra"

	"archi/internal/app"
)

var archiCmd = &cobra.Command{
	Use:   "architecture",
	Short: "Generate architectural recommendations",
	Long: `Analyze existing output.json and output.md files to generate 
comprehensive architectural recommendations. This command requires that 
you have already run a full analysis to generate the input files.`,
	Aliases: []string{"arch", "archi"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigFromGlobal()
		if err != nil {
			return err
		}

		application := app.New(cfg)
		return application.PerformArchitectureAnalysis()
	},
}

func init() {
	rootCmd.AddCommand(archiCmd)
}
