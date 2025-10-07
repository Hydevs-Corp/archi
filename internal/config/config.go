package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	APIBaseURL                string        `mapstructure:"apiBaseURL"`
	DefaultOutputDir          string        `mapstructure:"defaultOutputDir"`
	JSONOutputFile            string        `mapstructure:"jsonOutputFile"`
	MarkdownOutputFile        string        `mapstructure:"markdownOutputFile"`
	ReportOutputFile          string        `mapstructure:"reportOutputFile"`
	EstimationFile            string        `mapstructure:"estimationFile"`
	FileAnalysisModel         string        `mapstructure:"fileAnalysisModel"`
	FolderAnalysisModel       string        `mapstructure:"folderAnalysisModel"`
	ArchitectureAnalysisModel string        `mapstructure:"architectureAnalysisModel"`
	ImageAnalysisModel        string        `mapstructure:"imageAnalysisModel"`
	MaxFileSize               int64         `mapstructure:"maxFileSize"`
	RequestDelayStr           string        `mapstructure:"requestDelay"`
	RequestDelay              time.Duration `mapstructure:"-"`
}

func GetDefaultConfig() *Config {
	return &Config{
		APIBaseURL:                "http://localhost:3005",
		DefaultOutputDir:          ".",
		JSONOutputFile:            "output.json",
		MarkdownOutputFile:        "output.md",
		ReportOutputFile:          "report.md",
		EstimationFile:            "estimation.md",
		FileAnalysisModel:         "mistral-small-2501",
		FolderAnalysisModel:       "mistral-small-2501",
		ArchitectureAnalysisModel: "mistral-small-2501",
		ImageAnalysisModel:        "magistral-small-2509",
		MaxFileSize:               1024 * 1024, // 1MB
		RequestDelayStr:           "200ms",
		RequestDelay:              200 * time.Millisecond,
	}
}

func LoadConfig(configPath string) (*Config, error) {
	config := GetDefaultConfig()

	v := viper.New()

	v.SetDefault("apiBaseURL", config.APIBaseURL)
	v.SetDefault("defaultOutputDir", config.DefaultOutputDir)
	v.SetDefault("jsonOutputFile", config.JSONOutputFile)
	v.SetDefault("markdownOutputFile", config.MarkdownOutputFile)
	v.SetDefault("reportOutputFile", config.ReportOutputFile)
	v.SetDefault("estimationFile", config.EstimationFile)
	v.SetDefault("fileAnalysisModel", config.FileAnalysisModel)
	v.SetDefault("folderAnalysisModel", config.FolderAnalysisModel)
	v.SetDefault("architectureAnalysisModel", config.ArchitectureAnalysisModel)
	v.SetDefault("imageAnalysisModel", config.ImageAnalysisModel)
	v.SetDefault("maxFileSize", config.MaxFileSize)
	v.SetDefault("requestDelay", config.RequestDelayStr)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.archi")
		v.AddConfigPath("/etc/archi")
	}

	v.SetEnvPrefix("ARCHI")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if configPath != "" {
				fmt.Printf("⚠️  Config file not found: %s, using defaults\n", configPath)
			}
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		fmt.Printf("📋 Loading configuration from: %s\n", v.ConfigFileUsed())
	}

	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if config.RequestDelayStr != "" {
		var err error
		config.RequestDelay, err = time.ParseDuration(config.RequestDelayStr)
		if err != nil {
			return nil, fmt.Errorf("invalid requestDelay format '%s': %w", config.RequestDelayStr, err)
		}
	} else {
		if v.IsSet("requestDelay") {
			delayStr := v.GetString("requestDelay")
			var err error
			config.RequestDelay, err = time.ParseDuration(delayStr)
			if err != nil {
				return nil, fmt.Errorf("invalid requestDelay format '%s': %w", delayStr, err)
			}
		}
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	if config.DefaultOutputDir != "." && config.DefaultOutputDir != "" {
		if err := os.MkdirAll(config.DefaultOutputDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating output directory '%s': %w", config.DefaultOutputDir, err)
		}
	}

	return config, nil
}

func validateConfig(config *Config) error {
	if config.APIBaseURL == "" {
		return fmt.Errorf("apiBaseURL cannot be empty")
	}
	if config.JSONOutputFile == "" {
		return fmt.Errorf("jsonOutputFile cannot be empty")
	}
	if config.MarkdownOutputFile == "" {
		return fmt.Errorf("markdownOutputFile cannot be empty")
	}
	if config.MaxFileSize <= 0 {
		return fmt.Errorf("maxFileSize must be positive")
	}
	if config.RequestDelay < 0 {
		return fmt.Errorf("requestDelay cannot be negative")
	}

	return nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
