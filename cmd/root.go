package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"archi/internal/app"
	"archi/internal/config"
)

var (
	cfgFile     string
	onlyFolders bool
	noContent   bool
	countOnly   bool
	betterArchi bool
)

var rootCmd = &cobra.Command{
	Use:   "archi [directory]",
	Short: "Analyze directory structure and generate AI-powered insights",
	Long: `Archi is a directory analysis tool that uses AI to understand and describe
the structure and content of your projects. It can:

- Analyze file contents and generate descriptions
- Create visual directory trees
- Generate architectural recommendations
- Estimate processing time for large projects
- Support various file formats (PDF, DOCX, XLSX, images)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAnalysis,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	rootCmd.Flags().BoolVar(&onlyFolders, "only-folders", false, "Only show folders in the output")
	rootCmd.Flags().BoolVar(&noContent, "no-content", false, "Exclude file content from the JSON output")
	rootCmd.Flags().BoolVar(&countOnly, "count-only", false, "Only count files and folders, estimate execution time")
	rootCmd.Flags().BoolVar(&betterArchi, "better-archi", false, "Analyze existing outputs to generate architectural recommendations")
}

func initConfig() {
	if cfgFile != "" {

		viper.SetConfigFile(cfgFile)
	} else {

		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.archi")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("📋 Using config file: %s\n", viper.ConfigFileUsed())
	}
}

func runAnalysis(cmd *cobra.Command, args []string) error {

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	if cfgFile != "" || config.FileExists("config.yaml") {
		fmt.Printf("📡 API endpoint: %s\n", cfg.APIBaseURL)
	}

	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	application := app.New(cfg)

	switch {
	case betterArchi:
		return application.PerformArchitectureAnalysis()
	case countOnly:
		return application.PerformCountAnalysis(targetDir)
	default:
		return application.PerformFullAnalysis(targetDir, onlyFolders, noContent)
	}
}

func loadConfigFromGlobal() (*config.Config, error) {
	return config.LoadConfig(cfgFile)
}
