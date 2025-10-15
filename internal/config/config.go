package config

import (
	"fmt"
	"os"
	"strings"
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
	// Mode controls the analysis behavior: "full", "description-only", or "folder-only"
	Mode                      string        `mapstructure:"mode"`
	// Single-model (backward compatible). If set, must be a Mistral model when sent as string to the API
	FileAnalysisModel         string        `mapstructure:"fileAnalysisModel"`
	FolderAnalysisModel       string        `mapstructure:"folderAnalysisModel"`
	ArchitectureAnalysisModel string        `mapstructure:"architectureAnalysisModel"`
	ImageAnalysisModel        string        `mapstructure:"imageAnalysisModel"`
	// Multi-model configuration (new). If provided and non-empty, these take precedence and are sent as an array of provider/model objects
	FileAnalysisModels         []ProviderModel `mapstructure:"fileAnalysisModels"`
	FolderAnalysisModels       []ProviderModel `mapstructure:"folderAnalysisModels"`
	ArchitectureAnalysisModels []ProviderModel `mapstructure:"architectureAnalysisModels"`
	ImageAnalysisModels        []ProviderModel `mapstructure:"imageAnalysisModels"`
	MaxFileSize               int64         `mapstructure:"maxFileSize"`
	RequestDelayStr           string        `mapstructure:"requestDelay"`
	RequestDelay              time.Duration `mapstructure:"-"`
	BatchSize                 int           `mapstructure:"batchSize"`
	Concurrency               ConcurrencyConfig `mapstructure:"concurrency"`
}

type ConcurrencyConfig struct {
	ArchiAnalysis  int `mapstructure:"archiAnalysis" json:"archiAnalysis"`
	ReportChunking int `mapstructure:"reportChunking" json:"reportChunking"`
}

// ProviderModel represents an entry of the new array-based model selection API
type ProviderModel struct {
	Provider string `mapstructure:"provider" json:"provider"`
	Model    string `mapstructure:"model" json:"model"`
}

func GetDefaultConfig() *Config {
	return &Config{
		APIBaseURL:                "http://localhost:3005",
		DefaultOutputDir:          ".",
		JSONOutputFile:            "output.json",
		MarkdownOutputFile:        "output.md",
		ReportOutputFile:          "report.md",
		EstimationFile:            "estimation.md",
		Mode:                      "full",
		FileAnalysisModel:         "mistral-small-2501",
		FolderAnalysisModel:       "mistral-small-2501",
		ArchitectureAnalysisModel: "mistral-small-2501",
		ImageAnalysisModel:        "magistral-small-2509",
		FileAnalysisModels:         nil,
		FolderAnalysisModels:       nil,
		ArchitectureAnalysisModels: nil,
		ImageAnalysisModels:        nil,
		MaxFileSize:               1024 * 1024, // 1MB
		RequestDelayStr:           "200ms",
		RequestDelay:              200 * time.Millisecond,
		BatchSize:                 5,
		Concurrency:               ConcurrencyConfig{ArchiAnalysis: 4, ReportChunking: 4},
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
	v.SetDefault("mode", config.Mode)
	v.SetDefault("fileAnalysisModel", config.FileAnalysisModel)
	v.SetDefault("folderAnalysisModel", config.FolderAnalysisModel)
	v.SetDefault("architectureAnalysisModel", config.ArchitectureAnalysisModel)
	v.SetDefault("imageAnalysisModel", config.ImageAnalysisModel)
	// Multi-model defaults (empty arrays)
	v.SetDefault("fileAnalysisModels", config.FileAnalysisModels)
	v.SetDefault("folderAnalysisModels", config.FolderAnalysisModels)
	v.SetDefault("architectureAnalysisModels", config.ArchitectureAnalysisModels)
	v.SetDefault("imageAnalysisModels", config.ImageAnalysisModels)
	v.SetDefault("maxFileSize", config.MaxFileSize)
	v.SetDefault("requestDelay", config.RequestDelayStr)
	v.SetDefault("batchSize", config.BatchSize)
	v.SetDefault("concurrency.archiAnalysis", config.Concurrency.ArchiAnalysis)
	v.SetDefault("concurrency.reportChunking", config.Concurrency.ReportChunking)

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

	// Clamp concurrency to a safe maximum to avoid overwhelming the API or local resources
	const maxConcurrency = 32
	if config.Concurrency.ArchiAnalysis > maxConcurrency {
		fmt.Printf("⚠️  concurrency.archiAnalysis value %d is higher than allowed max %d, clamping to %d\n", config.Concurrency.ArchiAnalysis, maxConcurrency, maxConcurrency)
		config.Concurrency.ArchiAnalysis = maxConcurrency
	}
	if config.Concurrency.ReportChunking > maxConcurrency {
		fmt.Printf("⚠️  concurrency.reportChunking value %d is higher than allowed max %d, clamping to %d\n", config.Concurrency.ReportChunking, maxConcurrency, maxConcurrency)
		config.Concurrency.ReportChunking = maxConcurrency
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

	// Warn when both single and array models are provided; arrays take precedence at runtime
	warnBoth := func(single string, multi []ProviderModel, name string) {
		if strings.TrimSpace(single) != "" && len(multi) > 0 {
			fmt.Printf("⚠️  %s: both single and array provided; using the array (takes precedence)\n", name)
		}
	}
	warnBoth(config.FileAnalysisModel, config.FileAnalysisModels, "fileAnalysisModel(s)")
	warnBoth(config.FolderAnalysisModel, config.FolderAnalysisModels, "folderAnalysisModel(s)")
	warnBoth(config.ArchitectureAnalysisModel, config.ArchitectureAnalysisModels, "architectureAnalysisModel(s)")
	warnBoth(config.ImageAnalysisModel, config.ImageAnalysisModels, "imageAnalysisModel(s)")

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
	if config.BatchSize <= 0 {
		return fmt.Errorf("batchSize must be >= 1")
	}
	if config.Concurrency.ArchiAnalysis <= 0 {
		return fmt.Errorf("concurrency.archiAnalysis must be >= 1")
	}
	if config.Concurrency.ReportChunking <= 0 {
		return fmt.Errorf("concurrency.reportChunking must be >= 1")
	}
	switch strings.ToLower(strings.TrimSpace(config.Mode)) {
	case "", "full", "description-only", "folder-only":
		if strings.TrimSpace(config.Mode) == "" {
			config.Mode = "full"
		}
	default:
		return fmt.Errorf("mode must be one of: full, description-only, folder-only")
	}
	// Validate model configuration: each operation must have either a single string (Mistral-only) or a non-empty array of provider/models
	type pair struct {
		single string
		multi  []ProviderModel
		name   string
	}
	checks := []pair{
		{single: config.FileAnalysisModel, multi: config.FileAnalysisModels, name: "fileAnalysisModel(s)"},
		{single: config.FolderAnalysisModel, multi: config.FolderAnalysisModels, name: "folderAnalysisModel(s)"},
		{single: config.ArchitectureAnalysisModel, multi: config.ArchitectureAnalysisModels, name: "architectureAnalysisModel(s)"},
		{single: config.ImageAnalysisModel, multi: config.ImageAnalysisModels, name: "imageAnalysisModel(s)"},
	}
	for _, c := range checks {
		if strings.TrimSpace(c.single) == "" && len(c.multi) == 0 {
			return fmt.Errorf("%s: provide either the single string model (Mistral-only) or a non-empty array of {provider, model}", c.name)
		}
		for i, pm := range c.multi {
			if strings.TrimSpace(pm.Provider) == "" || strings.TrimSpace(pm.Model) == "" {
				return fmt.Errorf("%s[%d]: both provider and model must be non-empty", c.name, i)
			}
		}
	}

	return nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
