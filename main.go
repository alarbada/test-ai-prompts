package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/yaml.v3"

	openai "github.com/sashabaranov/go-openai"
)

func newOpenaiClient() *openai.Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}
	return openai.NewClient(apiKey)
}

var client = newOpenaiClient()

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type PromptConfig struct {
	Schema      string                         `json:"$schema,omitempty" yaml:"$schema,omitempty"`
	Model       string                         `json:"model" yaml:"model"`
	Messages    []openai.ChatCompletionMessage `json:"messages" yaml:"messages"`
	Temperature *float32                       `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	MaxTokens   *int                           `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`
	TopP        *float32                       `json:"top_p,omitempty" yaml:"top_p,omitempty"`
	Stop        []string                       `json:"stop,omitempty" yaml:"stop,omitempty"`
}

func loadPromptConfig(filename string) (PromptConfig, error) {
	var config PromptConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		err = json.Unmarshal(data, &config)
	default:
		return config, fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}

	return config, err
}

func callOpenAIWithPrompt(promptConfig PromptConfig, input string) (string, error) {
	// Create a copy of messages and add user input
	messages := make([]openai.ChatCompletionMessage, len(promptConfig.Messages))
	copy(messages, promptConfig.Messages)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	})

	req := openai.ChatCompletionRequest{
		Model:    promptConfig.Model,
		Messages: messages,
	}

	if promptConfig.Temperature != nil {
		req.Temperature = *promptConfig.Temperature
	}
	if promptConfig.MaxTokens != nil {
		req.MaxTokens = *promptConfig.MaxTokens
	}
	if promptConfig.TopP != nil {
		req.TopP = *promptConfig.TopP
	}
	if promptConfig.Stop != nil {
		req.Stop = promptConfig.Stop
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func callOpenAI(systemPrompt, input string) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4Dot1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleDeveloper,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func runTestWithPrompt(testCase TestCase, promptConfig PromptConfig) bool {
	result, err := callOpenAIWithPrompt(promptConfig, testCase.Input)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return false
	}

	passed := strings.EqualFold(strings.TrimSpace(result), strings.TrimSpace(testCase.Expected))

	fmt.Printf("Input: %s\n", testCase.Input)
	fmt.Printf("Expected: %s\n", testCase.Expected)
	fmt.Printf("Got: %s\n", result)
	fmt.Printf("PASSED: %v\n\n", passed)

	return passed
}

func testMain() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run . test <prompt.json> <testcases.json>")
		os.Exit(1)
	}

	promptFile := os.Args[2]
	testFile := os.Args[3]

	// Load prompt config
	promptConfig, err := loadPromptConfig(promptFile)
	if err != nil {
		fmt.Printf("Error loading prompt file %s: %v\n", promptFile, err)
		os.Exit(1)
	}

	// Load test cases
	testData, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Printf("Error reading test file %s: %v\n", testFile, err)
		os.Exit(1)
	}

	var testCases []TestCase
	if err := json.Unmarshal(testData, &testCases); err != nil {
		fmt.Printf("Error parsing test cases: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running %d test cases with prompt from %s:\n\n", len(testCases), promptFile)

	passed := 0
	for i, testCase := range testCases {
		fmt.Printf("Test %d:\n", i+1)
		if runTestWithPrompt(testCase, promptConfig) {
			passed++
		}
	}

	fmt.Printf("Results: %d/%d tests passed\n", passed, len(testCases))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run . generate <prompt.json> <testcases.json> [num_cases] - generate test cases")
		fmt.Println("  go run . test <prompt.json> <testcases.json> - run tests with custom prompt")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "generate":
		generateMain()
	case "test":
		testMain()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: generate, eval, test")
		os.Exit(1)
	}
}
