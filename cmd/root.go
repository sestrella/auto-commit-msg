package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

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

type TraceInfo struct {
	Version       string
	Model         string
	ResponseTime  time.Duration
	ExecutionTime time.Duration
}

var configFile string
var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "auto-commit-msg COMMIT_MSG_FILE",
	Short: "Generates a commit message from a git diff using AI",
	Run: func(cmd *cobra.Command, args []string) {
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
			return
		}

		gitDiff, err := exec.Command("git", "diff", "--cached").Output()
		if err != nil {
			cobra.CheckErr(err)
		}
		if len(gitDiff) == 0 {
			cobra.CheckErr("git diff is empty")
		}

		shortStat, err := exec.Command("git", "diff", "--cached", "--shortstat").Output()
		if err != nil {
			cobra.CheckErr(err)
		}
		shortStatStr := string(shortStat)

		insertionMatches := regexp.MustCompile(`(\d+)\s+insertions?\(\+\)`).FindStringSubmatch(shortStatStr)
		insertions := 0
		if len(insertionMatches) > 1 {
			insertions, err = strconv.Atoi(insertionMatches[1])
			if err != nil {
				cobra.CheckErr(err)
			}
		}

		deletionMatches := regexp.MustCompile(`(\d+)\s+deletions?\(\-\)`).FindStringSubmatch(shortStatStr)
		deletions := 0
		if len(deletionMatches) > 1 {
			deletions, err = strconv.Atoi(deletionMatches[1])
			if err != nil {
				cobra.CheckErr(err)
			}
		}

		totalChanges := insertions + deletions
		totalChangesThreshold := config.Diff.Threshold
		var model string
		if totalChanges < totalChangesThreshold {
			model = config.Diff.ShortModel
			if model == "" {
				cobra.CheckErr("short_model cannot be empty")
			}
			log.Printf("git diff total changes %d under %d threshold, using model for short diffs: %s\n", totalChanges, totalChangesThreshold, model)
		} else {
			model = config.Diff.LongModel
			if model == "" {
				cobra.CheckErr("long_model cannot be empty")
			}
			log.Printf("git diff total changes %d over %d threshold, using model for long diffs: %s\n", totalChanges, totalChangesThreshold, model)
		}
		if config.Provider.ApiKey == "" {
			cobra.CheckErr("api_key environment variable name cannot be empty")
		}

		apiKey := os.Getenv(config.Provider.ApiKey)
		if apiKey == "" {
			cobra.CheckErr(fmt.Sprintf("environment variable %s is required", config.Provider.ApiKey))
		}
		if config.Provider.BaseUrl == "" {
			cobra.CheckErr("base_url cannot be empty")
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
				Content: string(gitDiff),
			},
		})
		var responseDuration time.Duration
		if config.Trace {
			responseDuration = time.Since(responseTime)
		}
		if err != nil {
			cobra.CheckErr(err)
		}
		if len(res.Choices) == 0 {
			cobra.CheckErr(fmt.Sprintf("expects response to include at least one choice: %+v", res))
		}

		commitMsg := res.Choices[0].Message.Content
		if config.Trace {
			executionDuration := time.Since(executionTime)
			commitMsg = fmt.Sprintf("%s\n\nauto-commit-msg%+v", commitMsg, TraceInfo{
				Version:       strings.TrimSpace(cmd.Version),
				Model:         model,
				ResponseTime:  responseDuration,
				ExecutionTime: executionDuration,
			})
		}

		if commitMsgFile == "" {
			fmt.Println(commitMsg)
		} else {
			err = os.WriteFile(commitMsgFile, []byte(commitMsg), 0644)
			if err != nil {
				cobra.CheckErr(err)
			}
		}
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
