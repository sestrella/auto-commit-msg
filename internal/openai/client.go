package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	httpClient http.Client
	baseUrl    string
}

type openAITransport struct {
	base   http.RoundTripper
	apiKey string
}

type createResponse struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CreatedResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ChoiceMessage `json:"message"`
}

type ChoiceMessage struct {
	Content string `json:"content"`
}

func (transport openAITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", transport.apiKey))
	req.Header.Set("Content-Type", "application/json")
	return transport.base.RoundTrip(req)
}

func NewClient(baseUrl string, apiKey string) Client {
	transport := openAITransport{
		base:   http.DefaultTransport,
		apiKey: apiKey,
	}
	httpClient := http.Client{Transport: transport}
	return Client{httpClient, baseUrl}
}

func (client Client) CreateChatCompletion(model string, input []Message) (*CreatedResponse, error) {
	body, err := json.Marshal(createResponse{model, input})
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

	var data CreatedResponse
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}