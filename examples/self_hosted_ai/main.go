package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

func main() {
	provider := flag.String("provider", "ollama", "Self-hosted provider: ollama, openai-compatible, vllm, lmstudio, llamacpp, or textgen-webui")
	baseURL := flag.String("base-url", "http://localhost:11434", "Provider base URL (ex: http://localhost:11434 for Ollama)")
	model := flag.String("model", "llama3", "Model name hosted by your provider")
	timeout := flag.Duration("timeout", 30*time.Second, "HTTP timeout for the request")
	prompt := flag.String("prompt", "", "Prompt to send to your self-hosted model")
	promptFile := flag.String("prompt-file", "", "Path to a prompt file to send to your self-hosted model")
	flag.Parse()

	if strings.TrimSpace(*prompt) == "" && strings.TrimSpace(*promptFile) == "" {
		fmt.Fprintln(os.Stderr, "prompt is required (use --prompt or --prompt-file)")
		os.Exit(1)
	}
	if strings.TrimSpace(*prompt) != "" && strings.TrimSpace(*promptFile) != "" {
		fmt.Fprintln(os.Stderr, "provide only one of --prompt or --prompt-file")
		os.Exit(1)
	}

	finalPrompt := strings.TrimSpace(*prompt)
	if finalPrompt == "" {
		contents, err := os.ReadFile(*promptFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading prompt file:", err)
			os.Exit(1)
		}
		finalPrompt = strings.TrimSpace(string(contents))
		if finalPrompt == "" {
			fmt.Fprintln(os.Stderr, "prompt file is empty")
			os.Exit(1)
		}
	}

	client := &http.Client{Timeout: *timeout}

	var responseText string
	var err error

	switch normalizeProvider(*provider) {
	case "ollama":
		responseText, err = callOllama(client, *baseURL, *model, finalPrompt)
	case "openai-compatible":
		responseText, err = callOpenAICompatible(client, *baseURL, *model, finalPrompt)
	default:
		err = fmt.Errorf("unsupported provider %q (supported: %s)", *provider, supportedProviders())
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Println(responseText)
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "ollama":
		return "ollama"
	case "openai-compatible", "openai", "openai-compatible-chat":
		return "openai-compatible"
	case "vllm":
		return "openai-compatible"
	case "lmstudio", "lm-studio":
		return "openai-compatible"
	case "llamacpp", "llama.cpp", "llama-cpp":
		return "openai-compatible"
	case "textgen-webui", "text-generation-webui", "oobabooga":
		return "openai-compatible"
	default:
		return ""
	}
}

func supportedProviders() string {
	return "ollama, openai-compatible, vllm, lmstudio, llamacpp, textgen-webui"
}

func callOllama(client *http.Client, baseURL, model, prompt string) (string, error) {
	payload := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error (%d): %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	var decoded ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	return decoded.Response, nil
}

func callOpenAICompatible(client *http.Client, baseURL, model, prompt string) (string, error) {
	payload := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal openai-compatible request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("build openai-compatible request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send openai-compatible request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai-compatible error (%d): %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	var decoded openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode openai-compatible response: %w", err)
	}

	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("openai-compatible response missing choices")
	}

	return decoded.Choices[0].Message.Content, nil
}
