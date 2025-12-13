package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIClient struct {
	httpClient http.Client
	baseUrl    string
}

type OpenAITransport struct {
	base   http.RoundTripper
	apiKey string
}

type RequestChatCompletions struct {
	Model    string           `json:"model"`
	Messages []RequestMessage `json:"messages"`
}

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseChatCompletion struct {
	Choices []ResponseChoice `json:"choices"`
}

type ResponseChoice struct {
	Message ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	Content string `json:"content"`
}

func (transport OpenAITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", transport.apiKey))
	req.Header.Set("Content-Type", "application/json")
	return transport.base.RoundTrip(req)
}

func NewClient(baseUrl string, apiKey string) OpenAIClient {
	transport := OpenAITransport{
		base:   http.DefaultTransport,
		apiKey: apiKey,
	}
	httpClient := http.Client{Transport: transport}
	return OpenAIClient{httpClient, baseUrl}
}

func (client OpenAIClient) CreateChatCompletion(model string, input []RequestMessage) (*ResponseChatCompletion, error) {
	body, err := json.Marshal(RequestChatCompletions{model, input})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", client.baseUrl), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("response status code is not 200: %+v", res)
	}

	var data ResponseChatCompletion
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
