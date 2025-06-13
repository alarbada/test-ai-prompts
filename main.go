package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"

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

func callOpenAIWithPrompt(promptConfig PromptConfig, input string) (string, error) {
	req := promptConfig.ChatCompletionRequest
	req.Messages = append(req.Messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: input,
	})

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run . generate -prompt <file> -testcases <file> [-num <cases>] - generate test cases")
		fmt.Println("  go run . test -prompt <file> -testcases <file> - run tests with custom prompt")
		fmt.Println("  go run . run -prompt <file> -testcases <file> -index <n> - run single test case")
		os.Exit(1)
	}

	var err error
	command := os.Args[1]
	switch command {
	case "generate":
		err = generateMain()
	case "test":
		err = testMain()
	case "run":
		err = runMain()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: generate, test, run")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
