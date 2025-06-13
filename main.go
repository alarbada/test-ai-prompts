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
		fmt.Println("  go run . <eval-file.yaml> - run evaluation")
		os.Exit(1)
	}

	var err error
	command := os.Args[1]

	if command == "generate" {
		err = generateMain()
	} else {
		// Assume it's an eval file
		err = evalMain(command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
