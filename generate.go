package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func generateMain() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run . generate <prompt.json> <testcases.json> [num_cases]")
		os.Exit(1)
	}

	promptFile := os.Args[2]
	testFile := os.Args[3]
	numCases := 10
	
	if len(os.Args) > 4 {
		if n, err := strconv.Atoi(os.Args[4]); err == nil {
			numCases = n
		}
	}

	// Load prompt config
	promptConfig, err := loadPromptConfig(promptFile)
	if err != nil {
		fmt.Printf("Error loading prompt file %s: %v\n", promptFile, err)
		os.Exit(1)
	}

	// Load existing test cases (if file exists)
	var existingTestCases []TestCase
	if testData, err := os.ReadFile(testFile); err == nil {
		if err := json.Unmarshal(testData, &existingTestCases); err != nil {
			fmt.Printf("Error parsing existing test cases: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d existing test cases\n", len(existingTestCases))
	}

	fmt.Printf("Generating %d new test cases...\n", numCases)

	newTestCases, err := generateTestCasesWithPrompt(promptConfig, numCases)
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

	if err := os.WriteFile(testFile, data, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d new test cases and saved %d total to %s\n", len(newTestCases), len(allTestCases), testFile)
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

	result, err := callOpenAI(
		"You are a helpful assistant that generates test cases. Output only valid JSON, no additional text.",
		generationPrompt,
	)
	if err != nil {
		return nil, err
	}

	var testCases []TestCase
	if err := json.Unmarshal([]byte(result), &testCases); err != nil {
		return nil, fmt.Errorf("failed to parse generated test cases: %v", err)
	}

	return testCases, nil
}

func generateTestCases(systemPrompt string, numCases int) ([]TestCase, error) {
	generationPrompt := fmt.Sprintf(`Given this system prompt: "%s"

Generate %d diverse test cases as JSON array in this exact format:
[
  {
    "input": "example input text",
    "expected": "expected output"
  }
]

Make the test cases varied and realistic. Include edge cases and different scenarios that would test the system prompt thoroughly.`, systemPrompt, numCases)

	result, err := callOpenAI(
		"You are a helpful assistant that generates test cases. Output only valid JSON, no additional text.",
		generationPrompt,
	)
	if err != nil {
		return nil, err
	}

	var testCases []TestCase
	if err := json.Unmarshal([]byte(result), &testCases); err != nil {
		return nil, fmt.Errorf("failed to parse generated test cases: %v", err)
	}

	return testCases, nil
}
