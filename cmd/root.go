package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/sestrella/autocommitmsg/internal/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "autocommitmsg COMMIT_MSG_FILE",
	Short: "Generates a commit message from a git diff using AI",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		preCommitDetected := os.Getenv("PRE_COMMIT") == "1"
		if preCommitDetected {
			log.Println("pre-commit detected")
		}

		commitMsgFile := args[0]
		var commitSource string
		if preCommitDetected {
			commitSource = os.Getenv("PRE_COMMIT_COMMIT_MSG_SOURCE")
		}
		if commitSource != "" {
			log.Printf("Commit source '%s' is not empty, skipping commit generation\n", commitSource)
			return
		}

		gitDiff, err := exec.Command("git", "diff", "--cached").Output()
		if err != nil {
			cobra.CheckErr(err)
		}
		if len(gitDiff) == 0 {
			cobra.CheckErr("git diff is empty")
		}

		gitDiffStr := string(gitDiff)
		gitDiffLoc := strings.Count(gitDiffStr, "\n")
		diffThreshold := viper.GetInt("diff-threshold")
		var model string
		if gitDiffLoc < diffThreshold {
			model = viper.GetString("short-model")
			if model == "" {
				cobra.CheckErr("short-model cannot be empty")
			}
			log.Printf("git diff LOC %d under %d threshold, using model for short diffs: %s\n", gitDiffLoc, diffThreshold, model)
		} else {
			model = viper.GetString("long-model")
			if model == "" {
				cobra.CheckErr("long-model cannot be empty")
			}
			log.Printf("git diff LOC %d over %d threshold, using model for long diffs: %s\n", gitDiffLoc, diffThreshold, model)
		}

		apiKeyEnvName := viper.GetString("api-key")
		if apiKeyEnvName == "" {
			cobra.CheckErr("api-key environment variable name cannot be empty")
		}

		apiKey := os.Getenv(apiKeyEnvName)
		if apiKey == "" {
			cobra.CheckErr(fmt.Sprintf("environment variable %s is required", apiKeyEnvName))
		}

		baseUrl := viper.GetString("base-url")
		if baseUrl == "" {
			cobra.CheckErr("base-url cannot be empty")
		}

		client := openai.NewClient(baseUrl, apiKey)
		res, err := client.CreateChatCompletion(model, []openai.Message{
			{
				Role:    "developer",
				Content: "You are an assistant that writes concise, conventional commit messages based on the provided git diff. Return the commit message without any quotes.",
			},
			{
				Role:    "user",
				Content: gitDiffStr,
			},
		})
		if err != nil {
			cobra.CheckErr(err)
		}
		if len(res.Choices) == 0 {
			cobra.CheckErr(fmt.Sprintf("expects response to include at least one choice: %+v", res))
		}

		commitMsg := res.Choices[0].Message.Content
		err = os.WriteFile(commitMsgFile, []byte(commitMsg), 0644)
		if err != nil {
			cobra.CheckErr(err)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .autocommitmsg.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".autocommitmsg" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigType("yml")
		viper.SetConfigName(".autocommitmsg")
	}

	viper.SetDefault("base-url", "https://generativelanguage.googleapis.com/v1beta/openai")
	viper.SetDefault("api-key", "GEMINI_API_KEY")
	viper.SetDefault("short-model", "gemini-2.5-flash-lite")
	viper.SetDefault("long-model", "gemini-2.5-flash")
	viper.SetDefault("diff-threshold", 500)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
