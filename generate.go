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

func generateMain() error {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	promptFile := generateCmd.String("prompt", "", "prompt config file (required)")
	testFile := generateCmd.String("testcases", "", "test cases file (required)")
	numCases := generateCmd.Int("num", 10, "number of test cases to generate")

	generateCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s generate -prompt <file> -testcases <file> [-num <cases>]\n", os.Args[0])
		generateCmd.PrintDefaults()
	}

	if err := generateCmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *promptFile == "" || *testFile == "" {
		generateCmd.Usage()
		return fmt.Errorf("missing required flags: prompt and testcases are required")
	}

	// Load prompt config
	promptConfig, err := loadPromptConfig(*promptFile)
	if err != nil {
		return fmt.Errorf("loading prompt file %s: %w", *promptFile, err)
	}

	// Load existing test cases (if file exists)
	var existingTestCases []TestCase
	if _, err := os.Stat(*testFile); err == nil {
		existingTestCases, err = loadTestCases(*testFile)
		if err != nil {
			return fmt.Errorf("loading existing test cases: %w", err)
		}
		fmt.Printf("Found %d existing test cases\n", len(existingTestCases))
	}

	fmt.Printf("Generating %d new test cases...\n", *numCases)

	newTestCases, err := generateTestCasesWithPrompt(promptConfig, *numCases)
	if err != nil {
		return fmt.Errorf("generating test cases: %w", err)
	}

	// Combine existing and new test cases
	allTestCases := append(existingTestCases, newTestCases...)

	data, err := json.MarshalIndent(allTestCases, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling test cases: %w", err)
	}

	if err := os.WriteFile(*testFile, data, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", *testFile, err)
	}

	fmt.Printf("Generated %d new test cases and saved %d total to %s\n", len(newTestCases), len(allTestCases), *testFile)
	return nil
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
