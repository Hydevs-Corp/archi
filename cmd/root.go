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
	cfgFile string
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

	// Default behavior: run full analysis when no subcommand provided
	return application.PerformFullAnalysis(targetDir)
}

func loadConfigFromGlobal() (*config.Config, error) {
	return config.LoadConfig(cfgFile)
}
