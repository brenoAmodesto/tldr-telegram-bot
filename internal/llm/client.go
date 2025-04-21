package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type OllamaClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ModelName  string
}

func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		BaseURL:    os.Getenv("OLLAMA_API_URL"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		ModelName:  os.Getenv("OLLAMA_MODEL"),
	}
}

func (c *OllamaClient) GenerateSummary(prompt string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"model":  c.ModelName,
		"prompt": prompt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response from Ollama: %s", resp.Status)
	}

	var response struct {
		Summary string `json:"summary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response from Ollama: %w", err)
	}

	return response.Summary, nil
}

type GeminiClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ModelName  string
	APIkey     string
}

func NewGeminiClient() *GeminiClient {
	return &GeminiClient{
		// In this case, BaseURL expects the base endpoint without model string and query parameters.
		// For example: "https://generativelanguage.googleapis.com/v1beta/models"
		BaseURL:    os.Getenv("GEMINI_API_URL"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		ModelName:  os.Getenv("GEMINI_MODEL"),
		APIkey:     os.Getenv("GEMINI_API_KEY"),
	}
}

func GenerateSummaryGemini(prompt string) (string, error) {
	client := NewGeminiClient()

	// Construct the full URL using the environment variables.
	// GEMINI_API_URL should be "https://generativelanguage.googleapis.com/v1beta/models"
	// GEMINI_MODEL should be, e.g., "gemini-2.0-flash"
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", client.BaseURL, client.ModelName, client.APIkey)

	// Build request body based on the curl example.
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Remove the following header if the API key is supposed to be in the URL.
	// req.Header.Set("Authorization", "Bearer "+client.APIkey)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response from Gemini: %s", resp.Status)
	}

	// Assuming the response returns a candidates array with an "output" field.
	var response struct {
		Candidates []struct {
			Output string `json:"output"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response from Gemini: %w", err)
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no summary candidates returned")
	}

	return response.Candidates[0].Output, nil
}

// SummarizeGemini summarizes text using the Gemini API
func SummarizeGemini(text string, lang string) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Initialize the client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	// Create the model
	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))

	// Configure parameters for better summarization
	model.SetTemperature(0.2)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(1024)

	// Set a timeout for the context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create prompt based on language
	prompt := constructPrompt(text, lang)

	// Generate content
	resp, err := model.GenerateContent(ctxWithTimeout, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	// Extract the text from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	// Get the text from the first candidate's first part
	summary, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return string(summary), nil
}

func SummarizeGeminiLimited(text string, lang string) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel(os.Getenv("GEMINI_MODEL"))
	model.SetTemperature(0.2)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(1024)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// chama func prompt
	prompt := constructPromptLimited(text, lang)

	resp, err := model.GenerateContent(ctxWithTimeout, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	summary, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return string(summary), nil
}