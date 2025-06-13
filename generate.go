package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func generateMain() {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	promptFile := generateCmd.String("prompt", "", "prompt config file (required)")
	testFile := generateCmd.String("testcases", "", "test cases file (required)")
	numCases := generateCmd.Int("num", 10, "number of test cases to generate")

	generateCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s generate -prompt <file> -testcases <file> [-num <cases>]\n", os.Args[0])
		generateCmd.PrintDefaults()
	}

	if err := generateCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *promptFile == "" || *testFile == "" {
		generateCmd.Usage()
		os.Exit(1)
	}

	// Load prompt config
	promptConfig, err := loadPromptConfig(*promptFile)
	if err != nil {
		fmt.Printf("Error loading prompt file %s: %v\n", *promptFile, err)
		os.Exit(1)
	}

	// Load existing test cases (if file exists)
	var existingTestCases []TestCase
	if testData, err := os.ReadFile(*testFile); err == nil {
		if err := json.Unmarshal(testData, &existingTestCases); err != nil {
			fmt.Printf("Error parsing existing test cases: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d existing test cases\n", len(existingTestCases))
	}

	fmt.Printf("Generating %d new test cases...\n", *numCases)

	newTestCases, err := generateTestCasesWithPrompt(promptConfig, *numCases)
	if err != nil {
		fmt.Printf("Error generating test cases: %v\n", err)
		os.Exit(1)
	}

	// Combine existing and new test cases
	allTestCases := append(existingTestCases, newTestCases...)

	data, err := json.MarshalIndent(allTestCases, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling test cases: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*testFile, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d new test cases and saved %d total to %s\n", len(newTestCases), len(allTestCases), *testFile)
}

func generateTestCasesWithPrompt(promptConfig PromptConfig, numCases int) ([]TestCase, error) {
	// Extract system prompt from the config for generation context
	var systemPrompt string
	for _, msg := range promptConfig.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			break
		}
	}

	generationPrompt := fmt.Sprintf(`Given this system prompt: "%s"

Generate %d diverse test cases as JSON array in this exact format:
[
  {
    "input": "example input text",
    "expected": "expected output"
  }
]

Make the test cases varied and realistic. Include edge cases and different scenarios that would test the system prompt thoroughly.`, systemPrompt, numCases)

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
					Content: generationPrompt,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API for test generation: %w", err)
	}

	result := strings.TrimSpace(resp.Choices[0].Message.Content)

	var testCases []TestCase
	if err := json.Unmarshal([]byte(result), &testCases); err != nil {
		return nil, fmt.Errorf("failed to parse generated test cases: %w", err)
	}

	return testCases, nil
}
