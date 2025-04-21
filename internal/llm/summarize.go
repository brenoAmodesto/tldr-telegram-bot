package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Summarize sends a request to the Ollama LLM server and waits until done is true.
func Summarize(text string, lang string) (string, error) {
	ollamaAPIURL := os.Getenv("OLLAMA_API_URL")
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		return "", fmt.Errorf("OLLAMA_MODEL environment variable is not set")
	}

	prompt := constructPrompt(text, lang)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(ollamaAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response from Ollama API: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}
	summary, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("response field not found in the API response")
	}
	return summary, nil
}

// constructPrompt creates a prompt for the LLM based on the specified language.
func constructPrompt(text string, lang string) string {
	switch lang {
	case "pt":
		return fmt.Sprintf("Resuma a seguinte conversa do Telegram em Português:\n%s", text)
	case "en":
		return fmt.Sprintf("Summarize the following Telegram chat in English:\n%s", text)
	case "es":
		return fmt.Sprintf("Resume el siguiente Telegram chat en español:\n%s", text)
	default:
		return text // fallback para texto puro
	}
}

// func a ser chamada pela func SummarizeGeminiLimited
func constructPromptLimited(text string, lang string) string {
	switch lang {
	case "pt":
		return fmt.Sprintf("Resuma objetivamente a conversa do Telegram abaixo com até 400 caracteres:\n\n%s", text)
	case "en":
		return fmt.Sprintf("Summarize the following Telegram conversation in no more than 400 characters:\n\n%s", text)
	case "es":
		return fmt.Sprintf("Resume el siguiente chat de Telegram en no más de 400 caracteres:\n\n%s", text)
	default:
		return text
	}
}
