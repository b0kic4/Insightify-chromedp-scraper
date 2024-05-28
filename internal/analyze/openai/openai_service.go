package openai

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
)

type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type GPTRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

func SendPromptToGPT(request GPTRequest) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := resty.New()

	resp, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(apiEndpoint)
	if err != nil {
		log.Printf("Error sending request to OpenAI: %v", err)
		return "", fmt.Errorf("error sending request to OpenAI: %v", err)
	}

	// Log the full response for debugging
	log.Printf("Full API response: %s", resp.String())

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return "", fmt.Errorf("error decoding JSON response: %v", err)
	}

	if errorInfo, found := data["error"]; found {
		log.Printf("API returned an error: %v", errorInfo)
		return "", fmt.Errorf("API error: %v", errorInfo)
	}

	choices, ok := data["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Printf("No choices returned or wrong data type")
		return "", fmt.Errorf("no choices returned or wrong data type")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		log.Printf("No message field or wrong type in choice")
		return "", fmt.Errorf("no message field or wrong type in choice")
	}

	content, ok := message["content"].(string)
	if !ok {
		log.Printf("Content field missing or not a string")
		return "", fmt.Errorf("content field missing or not a string")
	}

	log.Println("Received content:", content)
	return content, nil
}
