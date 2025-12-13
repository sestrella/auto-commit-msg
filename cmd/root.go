package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/sestrella/auto-commit-msg/internal/git"
	"github.com/sestrella/auto-commit-msg/internal/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Trace    bool           `mapstructure:"trace"`
	Provider ProviderConfig `mapstructure:"provider"`
	Diff     DiffConfig     `mapstructure:"diff"`
}

type ProviderConfig struct {
	BaseUrl string `mapstructure:"base_url"`
	ApiKey  string `mapstructure:"api_key"`
}

type DiffConfig struct {
	ShortModel string `mapstructure:"short_model"`
	LongModel  string `mapstructure:"long_model"`
	Threshold  int    `mapstructure:"threshold"`
}

type Trace struct {
	Language      string  `json:"language"`
	Model         string  `json:"model"`
	Version       string  `json:"version"`
	ResponseTime  float64 `json:"response_time"`
	ExecutionTime float64 `json:"execution_time"`
}

type TraceWrapper struct {
	Trace Trace `json:"auto-commit-msg"`
}

var configFile string
var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "auto-commit-msg COMMIT_MSG_FILE",
	Short: "Generates a commit message from a git diff using AI",
	RunE: func(cmd *cobra.Command, args []string) error {
		var executionTime time.Time
		if config.Trace {
			executionTime = time.Now()
		}

		preCommitDetected := os.Getenv("PRE_COMMIT") == "1"
		if preCommitDetected {
			log.Println("pre-commit detected")
		}

		var commitMsgFile string
		if len(args) > 0 {
			commitMsgFile = args[0]
		}

		var commitSource string
		if preCommitDetected {
			commitSource = os.Getenv("PRE_COMMIT_COMMIT_MSG_SOURCE")
		}
		if commitSource != "" {
			log.Printf("Commit source '%s' is not empty, skipping commit message generation\n", commitSource)
			return nil
		}

		cachedGitDiff, err := git.DiffCached()
		if err != nil {
			return err
		}

		if cachedGitDiff == "" {
			return errors.New("git diff is empty")
		}

		stats, err := git.DiffCachedStats()
		if err != nil {
			return err
		}

		totalChanges := stats.Insertions + stats.Deletions
		totalChangesThreshold := config.Diff.Threshold
		var model string
		if totalChanges < totalChangesThreshold {
			model = config.Diff.ShortModel
			if model == "" {
				return errors.New("short_model cannot be empty")
			}
			log.Printf("git diff total changes %d under %d threshold, using model for short diffs: %s\n", totalChanges, totalChangesThreshold, model)
		} else {
			model = config.Diff.LongModel
			if model == "" {
				return errors.New("long_model cannot be empty")
			}
			log.Printf("git diff total changes %d over %d threshold, using model for long diffs: %s\n", totalChanges, totalChangesThreshold, model)
		}
		if config.Provider.ApiKey == "" {
			return errors.New("api_key environment variable name cannot be empty")
		}

		apiKey := os.Getenv(config.Provider.ApiKey)
		if apiKey == "" {
			return fmt.Errorf("environment variable %s is required", config.Provider.ApiKey)
		}
		if config.Provider.BaseUrl == "" {
			return errors.New("base_url cannot be empty")
		}

		client := openai.NewClient(config.Provider.BaseUrl, apiKey)
		var responseTime time.Time
		if config.Trace {
			responseTime = time.Now()
		}
		res, err := client.CreateChatCompletion(model, []openai.RequestMessage{
			{
				Role:    "developer",
				Content: "You are an assistant that writes concise, conventional commit messages based on the provided git diff. Return the commit message without any quotes.",
			},
			{
				Role:    "user",
				Content: cachedGitDiff,
			},
		})
		var responseDuration time.Duration
		if config.Trace {
			responseDuration = time.Since(responseTime)
		}
		if err != nil {
			return err
		}
		if len(res.Choices) == 0 {
			return fmt.Errorf("expects response to include at least one choice: %+v", res)
		}

		commitMsg := res.Choices[0].Message.Content
		if config.Trace {
			executionDuration := time.Since(executionTime)
			trace, err := json.Marshal(TraceWrapper{
				Trace: Trace{
					Language:      "go",
					Model:         model,
					Version:       strings.TrimSpace(cmd.Version),
					ResponseTime:  math.Round(responseDuration.Seconds()*100) / 100,
					ExecutionTime: math.Round(executionDuration.Seconds()*100) / 100,
				},
			})
			if err != nil {
				return err
			}

			commitMsg = fmt.Sprintf("%s\n---\n%s", commitMsg, trace)
		}

		if commitMsgFile == "" {
			fmt.Println(commitMsg)
		} else {
			err = os.WriteFile(commitMsgFile, []byte(commitMsg), 0644)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is .auto-commit-msg.toml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".auto-commit-msg")
		viper.SetConfigType("toml")
	}

	viper.SetDefault("trace", false)
	viper.SetDefault("provider.base_url", "https://generativelanguage.googleapis.com/v1beta/openai")
	viper.SetDefault("provider.api_key", "GEMINI_API_KEY")
	viper.SetDefault("diff.short_model", "gemini-2.5-flash-lite")
	viper.SetDefault("diff.long_model", "gemini-2.5-flash")
	viper.SetDefault("diff.threshold", 200)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
	if err := viper.Unmarshal(&config); err != nil {
		cobra.CheckErr(err)
	}
}
