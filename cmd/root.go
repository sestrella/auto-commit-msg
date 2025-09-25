package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type OpenAIClient struct {
	httpClient http.Client
	baseUrl    string
}

type OpenAITransport struct {
	base   http.RoundTripper
	apiKey string
}

func (transport OpenAITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", transport.apiKey))
	req.Header.Set("Content-Type", "application/json")
	return transport.base.RoundTrip(req)
}

func newOpenAIClient(apiKey string) OpenAIClient {
	transport := OpenAITransport{
		base:   http.DefaultTransport,
		apiKey: apiKey,
	}
	httpClient := http.Client{Transport: transport}
	return OpenAIClient{
		httpClient: httpClient,
		baseUrl:    "https://api.openai.com/v1",
	}
}

func (client OpenAIClient) createResponse(model string, input []map[string]any) (map[string]any, error) {
	body, err := json.Marshal(map[string]any{
		"model": model,
		"input": input,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", client.baseUrl, "/responses"), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Printf("%v\n", res)
		// TODO: Extract error message from response
		return nil, errors.New("TODO")
	}

	var data map[string]any
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "autocommitmsg",
	Short: "A brief description of your application",
	Run: func(cmd *cobra.Command, args []string) {
		var commitMsgFile = args[0]

		gitDiff, err := exec.Command("git", "diff", "--cached").Output()
		if err != nil {
			cobra.CheckErr(err)
		}

		if len(gitDiff) == 0 {
			cobra.CheckErr("git diff is empty")
		}

		client := newOpenAIClient(os.Getenv("OPENAI_API_KEY"))
		res, err := client.createResponse("gpt-4-turbo", []map[string]any{
			{
				"role":    "developer",
				"content": "You are an assistant that writes concise, conventional commit messages based on the provided git diff. Return the commit message without any quotes.",
			},
			{
				"role":    "user",
				"content": gitDiff,
			},
		})
		if err != nil {
			cobra.CheckErr(err)
		}

		// TODO: Create a response struct
		commitMsg := res["output"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
		os.WriteFile(commitMsgFile, []byte(commitMsg), 0644)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.autocommitmsg.yaml)")

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
		viper.SetConfigType("yaml")
		viper.SetConfigName(".autocommitmsg")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
